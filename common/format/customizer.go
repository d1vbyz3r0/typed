package format

import (
	"github.com/d1vbyz3r0/typed/common/meta"
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

		v, ok := field.Tag.Lookup("validate")
		if !ok || v == "-" {
			return nil
		}

		ctx := NewFieldContext(t, tag)
		if ctx == nil {
			continue
		}

		fieldName := meta.GetFieldNameByTag(field)
		markedRequired := slices.Contains(schema.Required, fieldName)

		for _, tagName := range ctx.TagNames() {
			tagFn, ok := Formats[tagName]
			if !ok {
				continue
			}

			tagFn(ctx)
		}

		ctx.finalize()

		if ctx.ShouldOmit && markedRequired {
			excludeFromRequired[fieldName] = struct{}{}
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
	// if t.Kind() == reflect.Struct {
	// 	return processStruct(name, t, tag, schema)
	// }
	//
	// v := tag.Get("validate")
	// if v == "" || v == "-" {
	// 	return nil
	// }
	//
	// rules := strings.Split(v, ",")
	// isRequired := false
	// for _, rule := range rules {
	// 	if strings.Contains(rule, "=") {
	// 		continue
	// 	}
	//
	// 	fmt, ok := Formats[rule]
	// 	if !ok {
	// 		continue
	// 	}
	//
	// 	// patterns applied only for string types
	// 	if fmt.Pattern != "" && t.Kind() == reflect.String {
	// 		schema.Pattern = fmt.Pattern
	// 	}
	//
	// 	if fmt.Format != "" {
	// 		schema.Format = fmt.Format
	// 	}
	//
	// 	if fmt.LtFunc != nil {
	// 		lt, ok := fmt.LtFunc(tag)
	// 		if ok {
	// 			schema.Max = &lt
	// 			schema.ExclusiveMax = true
	// 		}
	// 	}
	//
	// 	if fmt.LteFunc != nil {
	// 		lte, ok := fmt.LteFunc(tag)
	// 		if ok {
	// 			schema.Max = &lte
	// 			schema.ExclusiveMax = false
	// 		}
	// 	}
	//
	// 	if fmt.GtFunc != nil {
	// 		gt, ok := fmt.GtFunc(tag)
	// 		if ok {
	// 			schema.Min = &gt
	// 			schema.ExclusiveMax = true
	// 		}
	// 	}
	//
	// 	if fmt.GteFunc != nil {
	// 		gte, ok := fmt.GteFunc(tag)
	// 		if ok {
	// 			schema.Min = &gte
	// 			schema.ExclusiveMax = false
	// 		}
	// 	}
	//
	// 	if fmt.Required {
	// 		// we can't apply "required" processing rules, until other tags correctly processed cause of possible conflicts
	// 		isRequired = true
	// 	}
	// }
	//
	// if isRequired {
	// 	if t.Kind() == reflect.Slice || t.Kind() == reflect.Map || t.Kind() == reflect.Ptr {
	// 		schema.Nullable = false
	// 	}
	//
	// 	if schema.Type.Is(openapi3.TypeInteger) || schema.Type.Is(openapi3.TypeNumber) {
	// 		schema.Not = &openapi3.SchemaRef{
	// 			Value: &openapi3.Schema{
	// 				Enum: []any{0},
	// 			},
	// 		}
	// 	}
	//
	// 	if schema.Type.Is(openapi3.TypeString) && !IsKnown(schema.Format) {
	// 		schema.Pattern = NonEmptyRegexString
	// 	}
	// }

	return nil
}
