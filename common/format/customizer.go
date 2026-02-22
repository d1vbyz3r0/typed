package format

import (
	"fmt"
	"github.com/d1vbyz3r0/typed/common/meta"
	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
	"slices"
)

func processStruct(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	excludeFromRequired := make(map[string]struct{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		ctx := NewFieldContext(field.Type, field.Tag)
		if ctx == nil {
			continue
		}

		fieldName := meta.GetFieldNameByTag(field)
		markedRequired := slices.Contains(schema.Required, fieldName)

		if ctx.shouldOmit {
			excludeFromRequired[fieldName] = struct{}{}
			continue
		}

		if ctx.Required && !markedRequired {
			schema.Required = append(schema.Required, fieldName)
		}
	}

	if len(excludeFromRequired) > 0 {
		schema.Required = slices.DeleteFunc(schema.Required, func(s string) bool {
			_, ok := excludeFromRequired[s]
			return ok
		})
	}

	return nil
}

func Customizer(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	t = typing.DerefReflectPtr(t)
	if t.Kind() == reflect.Struct {
		err := processStruct(name, t, tag, schema)
		if err != nil {
			return fmt.Errorf("process struct: %w", err)
		}
	}

	ctx := NewFieldContext(t, tag)
	if ctx == nil {
		return nil
	}

	applyFieldContext(ctx, schema)
	return nil
}

func applyFieldContext(ctx *FieldContext, schema *openapi3.Schema) {
	if ctx.pattern != "" {
		schema.Pattern = ctx.pattern
	}

	if ctx.MinLength > 0 {
		schema.MinLength = ctx.MinLength
	}

	if ctx.MaxLength > 0 {
		schema.MaxLength = &ctx.MaxLength
	}

	if ctx.Format != "" {
		schema.Format = ctx.Format
	}

	if ctx.Min != nil {
		schema.Min = ctx.Min
		schema.ExclusiveMin = ctx.ExclusiveMin
	}

	if ctx.Max != nil {
		schema.Max = ctx.Max
		schema.ExclusiveMax = ctx.ExclusiveMax
	}

	if ctx.MinItems > 0 {
		schema.MinItems = ctx.MinItems
	}

	if ctx.MaxItems > 0 {
		schema.MaxItems = &ctx.MaxItems
	}

	if schema.Items != nil && ctx.child != nil {
		applyFieldContext(ctx.child, schema.Items.Value)
	}

	if ctx.child != nil && schema.AdditionalProperties.Schema != nil {
		applyFieldContext(ctx.child, schema.AdditionalProperties.Schema.Value)
	}

	schema.Nullable = ctx.Nullable

	if ctx.OneOf != nil {
		schema.Enum = append(schema.Enum, ctx.OneOf...)
	}

	if len(ctx.Not) > 0 {
		schema.Not = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Enum: ctx.Not,
			},
		}
	}
}
