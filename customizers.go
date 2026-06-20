package typed

import (
	"reflect"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/logging"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

var nonBodyTags = []string{
	"query",
	"header",
	"param",
}

var customizers = []openapi3gen.SchemaCustomizerFn{
	uuidCustomizer,
	makeFieldsRequired,
	processFormFiles,
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
	// this one always should be called last, since other customizer may want to process these fields
	return excludeNonBodyFieldsFromGeneration(name, t, tag, schema)
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
	t = typing.DerefReflectPtr(t)
	if t.Kind() != reflect.Struct {
		return false
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

func NewEnumsCustomizer(registry *Registry) openapi3gen.SchemaCustomizerFn {
	return func(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		pkgPath := t.PkgPath()
		if pkgPath == "" {
			return nil
		}

		typeName := t.Name()
		vals, ok := registry.LookupEnumValues(pkgPath, typeName)
		if !ok {
			logging.Debug(
				"enum values not found in registry for enum customizer, skipping (not a enum?)",
				"pkg", pkgPath,
				"typename", typeName,
			)
			return nil
		}

		logging.Debug("enum values found in registry", "pkg", pkgPath, "typename", typeName)

		schema.Enum = make([]any, len(vals))
		copy(schema.Enum, vals)
		return nil
	}
}

func excludeNonBodyFieldsFromGeneration(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	if tag == "" {
		return nil
	}

	for _, tagName := range nonBodyTags {
		if v, ok := tag.Lookup(tagName); ok && v != "-" {
			logging.Debug("removed field from final schema since it's not in body", "pkg", t.PkgPath(), "name", t.Name(), "tag", tag)
			return new(openapi3gen.ExcludeSchemaSentinel)
		}
	}

	return nil
}

func makeFieldsRequired(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	if t.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		if isNonBodyField(f) {
			continue
		}

		fieldType := f.Type
		if fieldType.Kind() == reflect.Pointer ||
			fieldType.Kind() == reflect.Slice ||
			fieldType.Kind() == reflect.Map {
			continue
		}
		schema.Required = append(schema.Required, getFieldNameByTag(t.Field(i)))
	}

	return nil
}

func getFieldNameByTag(field reflect.StructField) string {
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

func isNonBodyField(f reflect.StructField) bool {
	for _, tag := range nonBodyTags {
		if f.Tag.Get(tag) != "" {
			return true
		}
	}
	return false
}
