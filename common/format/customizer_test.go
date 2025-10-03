package format

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
	"time"
)

func TestProcessStruct(t *testing.T) {
	type TestStruct struct {
		RequiredField string `validate:"required" json:"req"`
		OmitField     string `validate:"omitempty" json:"omit"`
		NormalField   int    `json:"normal"`
		unexported    bool
		Email         string   `validate:"email" json:"email"`
		OneOf         string   `validate:"oneof=red green 'sky blue'" json:"oneof"`
		MinMax        int      `validate:"min=1,max=10" json:"minmax"`
		SliceMin      []string `validate:"min=2" json:"slicemin"`
	}

	schema := openapi3.NewObjectSchema()
	schema.Required = []string{"req", "omit"} // Изначально required both

	err := processStruct("TestStruct", reflect.TypeOf(TestStruct{}), reflect.StructTag(""), schema)
	assert.NoError(t, err)

	// Required: added RequiredField, removed OmitField
	assert.Contains(t, schema.Required, "req")
	assert.NotContains(t, schema.Required, "omit")
	assert.NotContains(t, schema.Required, "normal")
	assert.NotContains(t, schema.Required, "unexported")
	assert.Contains(t, schema.Required, "email")
	assert.Contains(t, schema.Required, "oneof")
	assert.Contains(t, schema.Required, "minmax")
	assert.Contains(t, schema.Required, "slicemin")
	fmt.Printf("%+v\n", schema)
}

func TestProcessStruct_Complex(t *testing.T) {
	type TestStruct struct {
		NestedSlice [][]string     `validate:"dive,dive,required,email|uri" json:"nested"`
		MapField    map[string]int `validate:"dive,keys,alpha,endkeys,gt=0" json:"map"`
	}

	schema := openapi3.NewObjectSchema()

	err := processStruct("TestStruct", reflect.TypeOf(TestStruct{}), reflect.StructTag(""), schema)
	assert.NoError(t, err)

	assert.Empty(t, schema.Required) // No top-level required
}

func TestProcessStruct_NoChanges(t *testing.T) {
	type TestStruct struct {
		Field int `json:"field"`
	}

	schema := openapi3.NewObjectSchema()

	err := processStruct("TestStruct", reflect.TypeOf(TestStruct{}), reflect.StructTag(""), schema)
	assert.NoError(t, err)

	assert.Empty(t, schema.Required)
}

func TestProcessStruct_NilContext(t *testing.T) {
	type TestStruct struct {
		Field int
	}

	schema := openapi3.NewObjectSchema()

	err := processStruct("TestStruct", reflect.TypeOf(TestStruct{}), reflect.StructTag(""), schema)
	assert.NoError(t, err)

	assert.Empty(t, schema.Required)
}

func TestCustomizer_NonStruct(t *testing.T) {
	schema := openapi3.NewStringSchema()

	err := Customizer("TestString", reflect.TypeOf(""), reflect.StructTag(`validate:"required,min=5,max=10,ne=foo"`), schema)
	assert.NoError(t, err)

	assert.Equal(t, uint64(5), schema.MinLength)
	assert.Equal(t, ptrUint(10), schema.MaxLength)
	assert.Equal(t, "^(((?!foo).)*)$", schema.Pattern)
}

func TestCustomizer_NilContext(t *testing.T) {
	schema := openapi3.NewStringSchema()

	err := Customizer("TestString", reflect.TypeOf(""), reflect.StructTag(""), schema)
	assert.NoError(t, err)

	assert.Empty(t, schema.Format)
	assert.Equal(t, uint64(0), schema.MinLength)
}

func TestApplyFieldContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      *FieldContext
		schema   *openapi3.Schema
		expected func(s *openapi3.Schema)
	}{
		{
			name: "Pattern and Format",
			ctx: &FieldContext{
				pattern: "^[a-z]+$",
				Format:  "email",
			},
			schema: openapi3.NewStringSchema(),
			expected: func(s *openapi3.Schema) {
				assert.Equal(t, "^[a-z]+$", s.Pattern)
				assert.Equal(t, "email", s.Format)
			},
		},
		{
			name: "MinMax Length and Values",
			ctx: &FieldContext{
				MinLength:    5,
				MaxLength:    10,
				Min:          ptrFloat(1.0),
				Max:          ptrFloat(100.0),
				ExclusiveMin: true,
				ExclusiveMax: true,
			},
			schema: openapi3.NewStringSchema(),
			expected: func(s *openapi3.Schema) {
				assert.Equal(t, uint64(5), s.MinLength)
				assert.Equal(t, ptrUint(10), s.MaxLength)
				assert.Equal(t, 1.0, *s.Min)
				assert.Equal(t, 100.0, *s.Max)
				assert.True(t, s.ExclusiveMin)
				assert.True(t, s.ExclusiveMax)
			},
		},
		{
			name: "Array Items",
			ctx: &FieldContext{
				MinItems: 2,
				MaxItems: 5,
				child: &FieldContext{
					Format: "uuid",
				},
			},
			schema: openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema()),
			expected: func(s *openapi3.Schema) {
				assert.Equal(t, uint64(2), s.MinItems)
				assert.Equal(t, ptrUint(5), s.MaxItems)
				assert.Equal(t, "uuid", s.Items.Value.Format)
			},
		},
		{
			name: "Object AdditionalProperties",
			ctx: &FieldContext{
				child: &FieldContext{
					Nullable: true,
				},
			},
			schema: openapi3.NewObjectSchema().WithAdditionalProperties(openapi3.NewStringSchema()),
			expected: func(s *openapi3.Schema) {
				assert.True(t, s.AdditionalProperties.Schema.Value.Nullable)
			},
		},
		{
			name: "Enum and Not",
			ctx: &FieldContext{
				OneOf: []interface{}{"a", "b"},
				Not:   []interface{}{0},
			},
			schema: openapi3.NewStringSchema(),
			expected: func(s *openapi3.Schema) {
				assert.ElementsMatch(t, []interface{}{"a", "b"}, s.Enum)
				assert.NotNil(t, s.Not)
				assert.ElementsMatch(t, []interface{}{0}, s.Not.Value.Enum)
			},
		},
		{
			name: "Nullable",
			ctx: &FieldContext{
				Nullable: true,
			},
			schema: openapi3.NewStringSchema(),
			expected: func(s *openapi3.Schema) {
				assert.True(t, s.Nullable)
			},
		},
		{
			name: "Nested Child Application",
			ctx: &FieldContext{
				child: &FieldContext{
					child: &FieldContext{
						pattern: "^email$",
						Format:  "email",
					},
				},
			},
			schema: openapi3.NewArraySchema().WithItems(openapi3.NewArraySchema().WithItems(openapi3.NewStringSchema())),
			expected: func(s *openapi3.Schema) {
				childSchema := s.Items.Value.Items.Value
				assert.Equal(t, "^email$", childSchema.Pattern)
				assert.Equal(t, "email", childSchema.Format)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyFieldContext(tt.ctx, tt.schema)
			tt.expected(tt.schema)
		})
	}
}

func ptrFloat(v float64) *float64 { return &v }
func ptrUint(v uint64) *uint64    { return &v }

func TestCustomizer_Integration(t *testing.T) {
	type TestStruct struct {
		Email string                   `validate:"email,required,min=5" json:"email"`
		Count int                      `validate:"required,gt=0,lte=100" json:"count"`
		Slice []string                 `validate:"dive,alphanum|alpha" json:"slice"`
		Map   map[string]time.Duration `validate:"dive,keys,required,endkeys,gt=1h" json:"map"`
	}

	schemas := make(openapi3.Schemas)
	g := openapi3gen.NewGenerator(openapi3gen.SchemaCustomizer(Customizer))
	ref, err := g.NewSchemaRefForValue(&TestStruct{}, schemas)
	require.NoError(t, err)

	schema := ref.Value
	require.NotNil(t, schema)

	emailProp := schema.Properties["email"].Value
	assert.Equal(t, "email", emailProp.Format)
	assert.Equal(t, uint64(5), emailProp.MinLength)
	assert.Contains(t, schema.Required, "email")

	countProp := schema.Properties["count"].Value
	assert.Equal(t, 0.0, *countProp.Min)
	assert.Equal(t, 100.0, *countProp.Max)
	assert.Contains(t, schema.Required, "count")

	sliceProp := schema.Properties["slice"].Value
	assert.Equal(t, fmt.Sprintf("^((%s)|(%s))$", AlphaNumericRegexString, AlphaRegexString), sliceProp.Items.Value.Pattern)
	assert.Equal(t, "", sliceProp.Items.Value.Format)

	mapProp := schema.Properties["map"].Value
	assert.Equal(t, float64(time.Hour), *mapProp.AdditionalProperties.Schema.Value.Min)
}
