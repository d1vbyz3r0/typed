package format

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"regexp"
	"testing"
)

func TestFieldContext_AddPattern(t *testing.T) {
	cases := []struct {
		name     string
		tag      reflect.StructTag
		patterns []string
		want     string
	}{
		{
			name:     "one value",
			tag:      reflect.StructTag(`validate:"required,alpha"`),
			patterns: []string{AlphaRegexString},
			want:     "^[a-zA-Z]+$",
		},
		{
			name: "simple pattern",
			tag:  reflect.StructTag(`validate:"required,eq=1|2"`),
			patterns: []string{
				"1",
				"2",
			},
			want: "^(1)|(2)$",
		},
		{
			name: "real pattern",
			tag:  reflect.StructTag(`validate:"required,alpha|alphanum"`),
			patterns: []string{
				AlphaRegexString,
				AlphaNumericRegexString,
			},
			want: "^([a-zA-Z]+)|([a-zA-Z0-9]+)$",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewFieldContext(nil, tc.tag)
			for _, pattern := range tc.patterns {
				ctx.AddPattern(pattern)
			}
			ctx.finalize()

			require.Equal(t, tc.want, ctx.Pattern)
			_, err := regexp.Compile(ctx.Pattern)
			require.NoError(t, err)
		})
	}
}

func TestFieldContext_OneOf(t *testing.T) {
	tagFn := Formats["oneof"]

	cases := []struct {
		name string
		typ  reflect.Type
		tag  reflect.StructTag
		want []any
	}{
		{
			name: "string oneof",
			typ:  reflect.TypeOf(""),
			tag:  reflect.StructTag(`validate:"required,oneof=red green"`),
			want: []any{"red", "green"},
		},
		{
			name: "string oneof with spaces",
			typ:  reflect.TypeOf(""),
			tag:  reflect.StructTag(`validate:"required,oneof='red green' 'blue yellow'"`),
			want: []any{"red green", "blue yellow"},
		},
		{
			name: "float oneof",
			typ:  reflect.TypeOf(float64(0)),
			tag:  reflect.StructTag(`validate:"required,oneof=1 2 3"`),
			want: []any{float64(1), float64(2), float64(3)},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewFieldContext(tc.typ, tc.tag)
			tagFn(ctx)
			require.ElementsMatch(t, tc.want, ctx.OneOf)
		})
	}
}
