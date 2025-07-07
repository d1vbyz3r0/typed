package typed

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/labstack/echo/v4"
	"reflect"
)

const (
	FieldNameOverrideFormKey = "x-typed-override-form-key"
	FieldNameOverrideXMLKey  = "x-typed-override-xml-key"
)

var customizers = []openapi3gen.SchemaCustomizerFn{
	//overrideNames,
	processFormFiles,
	uuidCustomizer,
}

func RegisterCustomizer(fn openapi3gen.SchemaCustomizerFn) {
	customizers = append(customizers, fn)
}

// Customizer is a top-level openapi3gen.SchemaCustomizerFn. It will call some default customizers and all registered with RegisterCustomizer
func Customizer(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	for _, customizeFn := range customizers {
		err := customizeFn(name, t, tag, schema)
		if err != nil {
			return err
		}
	}
	return nil
}

// OverrideFieldNames will replace all prop names to values extracted from tags and remove key FieldNameOverrideKey from Extensions.
func OverrideFieldNames(ref *openapi3.SchemaRef, schemas openapi3.Schemas, typeName string, contentType string) error {
	key := ""
	if contentType == echo.MIMEApplicationForm || contentType == echo.MIMEMultipartForm {
		key = FieldNameOverrideFormKey
	} else if contentType == echo.MIMEApplicationXML {
		key = FieldNameOverrideXMLKey
	} else {
		return nil
	}

	var (
		props     openapi3.Schemas
		schemaRef *openapi3.SchemaRef
	)
	if schemas != nil {
		schema, ok := schemas[typeName]
		if !ok {
			return fmt.Errorf("schema not found for type %s", typeName)
		}

		if schema.Value == nil {
			return fmt.Errorf("schema value for %s is nil", typeName)
		}

		props = schema.Value.Properties
		schemaRef = schema
	} else {
		if ref.Value == nil {
			return fmt.Errorf("ref value for %s is nil", ref.RefString())
		}
		props = ref.Value.Properties
		schemaRef = ref
	}

	renames := make(map[string]string, len(props))
	for fieldName, fieldSchema := range props {
		if override, ok := fieldSchema.Value.Extensions[key]; ok {
			overrideStr := override.(string)
			renames[fieldName] = overrideStr
		}
	}

	for oldName, newName := range renames {
		schema := schemaRef.Value.Properties[oldName]
		delete(schema.Value.Extensions, key)
		schemaRef.Value.Properties[newName] = schema
		delete(schemaRef.Value.Properties, oldName)
	}

	return nil
}

func overrideNames(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	if schema.Extensions == nil {
		schema.Extensions = make(map[string]any)
	}

	v, ok := tag.Lookup("form")
	if ok && v != "-" {
		schema.Extensions[FieldNameOverrideFormKey] = v
	}

	v, ok = tag.Lookup("xml")
	if ok && v != "-" {
		schema.Extensions[FieldNameOverrideXMLKey] = v
	}

	return nil
}

func processFormFiles(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	if isFileHeader(t) {
		schema.Type = &openapi3.Types{openapi3.TypeString}
		schema.Format = "binary"
		schema.Properties = nil
	}

	if t.Kind() == reflect.Slice && isFileHeader(t.Elem()) {
		schema.Type = &openapi3.Types{openapi3.TypeArray}
		schema.Items = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:   &openapi3.Types{openapi3.TypeString},
				Format: "binary",
			},
		}
	}

	return nil
}

func isFileHeader(t reflect.Type) bool {
	if t.Kind() != reflect.Ptr && t.Kind() != reflect.Struct {
		return false
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.PkgPath() == "mime/multipart" && t.Name() == "FileHeader"
}

func uuidCustomizer(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	if t.PkgPath() == "github.com/google/uuid" && t.Name() == "UUID" {
		schema.Type = &openapi3.Types{openapi3.TypeString}
		schema.Format = "uuid"
	}
	return nil
}

func NewEnumsCustomizer(enums map[string][]any) openapi3gen.SchemaCustomizerFn {
	return func(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		if values, ok := enums[t.String()]; ok {
			schema.Enum = make([]interface{}, len(values))
			for i, v := range values {
				schema.Enum[i] = v
			}
		}
		return nil
	}
}
