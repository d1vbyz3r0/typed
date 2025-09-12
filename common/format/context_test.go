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
			tag:  reflect.StructTag(`validate:"required,eq=1|eq=2"`),
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
			ctx := NewFieldContext(reflect.TypeOf(""), tc.tag)
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

func assertFieldContextEqual(t *testing.T, want, got *FieldContext) {
	if want == nil {
		require.Nil(t, got)
		return
	}
	require.NotNil(t, got)
	require.ElementsMatch(t, want.ValidationRules, got.ValidationRules)

	if want.child != nil || got.child != nil {
		assertFieldContextEqual(t, want.child, got.child)
	}
}

func TestFieldContext_Children(t *testing.T) {
	type testStruct struct {
		A string           `validate:"required"`
		B []string         `validate:"dive,required"`
		C [][]int          `validate:"dive,dive,gt=0"`
		D map[string][]int `validate:"dive,keys,required,endkeys,dive,lt=10"`
		E string           `validate:"-"`
		F string           `validate:"required,email"`
	}

	st := reflect.TypeOf(testStruct{})

	cases := []struct {
		field string
		typ   reflect.Type
		tag   reflect.StructTag
		want  *FieldContext
	}{
		{
			field: "A",
			typ:   st.Field(0).Type,
			tag:   st.Field(0).Tag,
			want: &FieldContext{
				ValidationRules: []string{"required"},
			},
		},
		{
			field: "B",
			typ:   st.Field(1).Type,
			tag:   st.Field(1).Tag,
			want: &FieldContext{
				ValidationRules: nil,
				child: &FieldContext{
					ValidationRules: []string{"required"},
				},
			},
		},
		{
			field: "C",
			typ:   st.Field(2).Type,
			tag:   st.Field(2).Tag,
			want: &FieldContext{
				ValidationRules: nil,
				child: &FieldContext{
					ValidationRules: nil,
					child: &FieldContext{
						ValidationRules: []string{"gt=0"},
					},
				},
			},
		},
		{
			field: "D",
			typ:   st.Field(3).Type,
			tag:   st.Field(3).Tag,
			want: &FieldContext{
				ValidationRules: nil,
				child: &FieldContext{
					ValidationRules: []string{"required", "keys", "endkeys"},
					child: &FieldContext{
						ValidationRules: []string{"lt=10"},
					},
				},
			},
		},
		{
			field: "E",
			typ:   st.Field(4).Type,
			tag:   st.Field(4).Tag,
			want:  nil,
		},
		{
			field: "F",
			typ:   st.Field(5).Type,
			tag:   st.Field(5).Tag,
			want: &FieldContext{
				ValidationRules: []string{"required", "email"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.field, func(t *testing.T) {
			ctx := NewFieldContext(c.typ, c.tag)
			assertFieldContextEqual(t, c.want, ctx)
		})
	}
}

func TestFieldContext_NestedDive(t *testing.T) {
	type testStruct struct {
		A [][]int    `validate:"dive,dive,gt=0"`
		B [][]int    `validate:"dive,gt=0,dive,gt=1"`
		C [][][]int  `validate:"dive,dive,dive,lt=5"`
		D [][]string `validate:"dive,required,dive,gt=3"`
	}

	st := reflect.TypeOf(testStruct{})

	cases := []struct {
		field string
		want  *FieldContext
	}{
		{
			field: "A",
			want: &FieldContext{
				ValidationRules: nil,
				child: &FieldContext{
					ValidationRules: nil,
					child: &FieldContext{
						ValidationRules: []string{"gt=0"},
					},
				},
			},
		},
		{
			field: "B",
			want: &FieldContext{
				ValidationRules: nil,
				child: &FieldContext{
					ValidationRules: []string{"gt=0"},
					child: &FieldContext{
						ValidationRules: []string{"gt=1"},
					},
				},
			},
		},
		{
			field: "C",
			want: &FieldContext{
				ValidationRules: nil,
				child: &FieldContext{
					ValidationRules: nil,
					child: &FieldContext{
						ValidationRules: nil,
						child: &FieldContext{
							ValidationRules: []string{"lt=5"},
						},
					},
				},
			},
		},
		{
			field: "D",
			want: &FieldContext{
				ValidationRules: nil,
				child: &FieldContext{
					ValidationRules: []string{"required"},
					child: &FieldContext{
						ValidationRules: []string{"gt=3"},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.field, func(t *testing.T) {
			f, ok := st.FieldByName(c.field)
			require.True(t, ok)

			ctx := NewFieldContext(f.Type, f.Tag)
			require.NotNil(t, ctx)

			assertFieldContextEqual(t, c.want, ctx)
		})
	}
}

func TestFieldContext_ComparisonTags(t *testing.T) {
	type testStruct struct {
		A float64 `validate:"lt=10"`
		B float64 `validate:"lte=20"`
		C float64 `validate:"gt=5"`
		D float64 `validate:"gte=15"`
		E float64 `validate:"lt=5,gt=1"`
	}

	st := reflect.TypeOf(testStruct{})

	cases := []struct {
		field string
		key   string
		want  float64
	}{
		{"A", "lt", 10},
		{"B", "lte", 20},
		{"C", "gt", 5},
		{"D", "gte", 15},
		{"E", "lt", 5},
		{"E", "gt", 1},
	}

	for _, c := range cases {
		t.Run(c.field+"_"+c.key, func(t *testing.T) {
			f, _ := st.FieldByName(c.field)
			ctx := NewFieldContext(reflect.TypeOf(0.0), f.Tag)
			val, err := ctx.LookupFloat(c.key)
			require.NoError(t, err)
			require.Equal(t, c.want, val)
		})
	}
}
