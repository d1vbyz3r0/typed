package typed

import (
	"github.com/d1vbyz3r0/typed/common/format"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"reflect"
)

var customizers = []openapi3gen.SchemaCustomizerFn{
	processFormFiles,
	uuidCustomizer,
	excludeNonBodyFieldsFromGeneration,
	makeFieldsRequired,
	format.Customizer,
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

func processFormFiles(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	if isFileHeader(t) {
		schema.Type = &openapi3.Types{openapi3.TypeString}
		schema.Format = "binary"
		schema.Properties = make(openapi3.Schemas)
		schema.Required = nil
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

func excludeNonBodyFieldsFromGeneration(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	shouldSkip := []string{
		"query",
		"header",
		"param",
	}

	for _, tagName := range shouldSkip {
		if v, ok := tag.Lookup(tagName); ok && v != "-" && v != "" {
			return new(openapi3gen.ExcludeSchemaSentinel)
		}
	}

	return nil
}

func makeFieldsRequired(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	if t.Kind() != reflect.Struct {
		return nil
	}

	// multipart.FileHeader is represented as string/binary in OpenAPI.
	// Marking its internal struct fields as required produces invalid noise.
	if isFileHeader(t) {
		return nil
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		fieldType := f.Type
		if fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Map {
			continue
		}
		schema.Required = append(schema.Required, getFieldNameByTag(t.Field(i)))
	}

	return nil
}

func getFieldNameByTag(field reflect.StructField) string {
	// TODO: strip omitempty and omitnil
	if v := field.Tag.Get("json"); v != "" && v != "-" {
		return v
	}

	if v := field.Tag.Get("form"); v != "" && v != "-" {
		return v
	}

	if v := field.Tag.Get("xml"); v != "" && v != "-" {
		return v
	}

	return field.Name
}
