package typed

import (
	"reflect"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

// NewGenerator creates the default OpenAPI schema generator.
func NewGenerator(registry *Registry) *openapi3gen.Generator {
	enumsCustomizer := NewEnumsCustomizer(registry)
	return openapi3gen.NewGenerator(
		openapi3gen.UseAllExportedFields(),
		openapi3gen.CreateComponentSchemas(openapi3gen.ExportComponentSchemasOptions{
			ExportComponentSchemas: true,
			ExportTopLevelSchema:   true,
			ExportGenerics:         false,
		}),
		openapi3gen.SchemaCustomizer(func(
			name string,
			t reflect.Type,
			tag reflect.StructTag,
			schema *openapi3.Schema,
		) error {
			// TODO: move up from global state
			if err := enumsCustomizer(name, t, tag, schema); err != nil {
				return err
			}
			return Customizer(name, t, tag, schema)
		}),
		openapi3gen.CreateTypeNameGenerator(NewTypeNameGenerator(registry)),
		openapi3gen.CreateFieldNameGenerator(FieldNameGenerator),
	)
}
