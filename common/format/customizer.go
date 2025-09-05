package format

import (
	"reflect"
	"strings"

	"github.com/d1vbyz3r0/typed/common/meta"
	"github.com/getkin/kin-openapi/openapi3"
)

type Meta struct {
	Required bool
	Pattern  string
	Format   string
}

func Customizer(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			v, ok := field.Tag.Lookup("validate")
			if !ok {
				return nil
			}

			rules := strings.Split(v, ",")
			for _, rule := range rules {
				if strings.Contains(rule, "=") {
					// TODO: skip complex rules for now...
					continue
				}

				fmt, ok := Formats[rule]
				if !ok {
					continue
				}

				if fmt.Required {
					schema.Required = append(schema.Required, meta.GetFieldNameByTag(field))
				}
			}
		}

		return nil
	}

	v, ok := tag.Lookup("validate")
	if !ok {
		return nil
	}

	rules := strings.Split(v, ",")
	for _, rule := range rules {
		if strings.Contains(rule, "=") {
			continue
		}

		fmt, ok := Formats[rule]
		if !ok {
			continue
		}

		// patterns applied only for string types
		if fmt.Pattern != "" && t.Kind() == reflect.String {
			schema.Pattern = fmt.Pattern
		}

		if fmt.Format != "" {
			schema.Format = string(fmt.Format)
		}

		if fmt.Required {
			if t.Kind() == reflect.Slice || t.Kind() == reflect.Map || t.Kind() == reflect.Ptr {
				schema.Nullable = false
			}

			if schema.Type.Is(openapi3.TypeInteger) || schema.Type.Is(openapi3.TypeNumber) {
				schema.Not = &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Enum: []any{0},
					},
				}
			}

			if schema.Type.Is(openapi3.TypeString) && schema.Format != Binary {
				schema.Pattern = NonEmptyRegexString
			}
		}
	}

	return nil
}
