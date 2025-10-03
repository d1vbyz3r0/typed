package format

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"regexp"
	"testing"
	"time"
)

func assertFieldContextEqual(t *testing.T, want, got *FieldContext) {
	if want == nil {
		require.Nil(t, got)
		return
	}
	require.NotNil(t, got)
	require.ElementsMatch(t, want.validationRules, got.validationRules)

	if want.child != nil || got.child != nil {
		assertFieldContextEqual(t, want.child, got.child)
	}
}

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
			want:     "^([a-zA-Z]+)$",
		},
		{
			name: "real pattern",
			tag:  reflect.StructTag(`validate:"required,alpha|alphanum"`),
			patterns: []string{
				AlphaRegexString,
				AlphaNumericRegexString,
			},
			want: "^(([a-zA-Z]+)|([a-zA-Z0-9]+))$",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := NewFieldContext(reflect.TypeOf(""), tc.tag)

			require.Equal(t, tc.want, ctx.pattern)
			_, err := regexp.Compile(ctx.pattern)
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
				validationRules: []string{"required"},
			},
		},
		{
			field: "B",
			typ:   st.Field(1).Type,
			tag:   st.Field(1).Tag,
			want: &FieldContext{
				validationRules: nil,
				child: &FieldContext{
					validationRules: []string{"required"},
				},
			},
		},
		{
			field: "C",
			typ:   st.Field(2).Type,
			tag:   st.Field(2).Tag,
			want: &FieldContext{
				validationRules: nil,
				child: &FieldContext{
					validationRules: nil,
					child: &FieldContext{
						validationRules: []string{"gt=0"},
					},
				},
			},
		},
		{
			field: "D",
			typ:   st.Field(3).Type,
			tag:   st.Field(3).Tag,
			want: &FieldContext{
				validationRules: nil,
				child: &FieldContext{
					validationRules: nil,
					child: &FieldContext{
						validationRules: []string{"lt=10"},
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
				validationRules: []string{"required", "email"},
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
		A [][]int          `validate:"dive,dive,gt=0"`
		B [][]int          `validate:"dive,gt=0,dive,gt=1"`
		C [][][]int        `validate:"dive,dive,dive,lt=5"`
		D [][]string       `validate:"dive,required,dive,gt=3"`
		E map[string][]int `validate:"required,dive,keys,required,endkeys,dive,gt=1"`
	}

	st := reflect.TypeOf(testStruct{})

	cases := []struct {
		field string
		want  *FieldContext
	}{
		{
			field: "A",
			want: &FieldContext{
				validationRules: nil,
				child: &FieldContext{
					validationRules: nil,
					child: &FieldContext{
						validationRules: []string{"gt=0"},
					},
				},
			},
		},
		{
			field: "B",
			want: &FieldContext{
				validationRules: nil,
				child: &FieldContext{
					validationRules: []string{"gt=0"},
					child: &FieldContext{
						validationRules: []string{"gt=1"},
					},
				},
			},
		},
		{
			field: "C",
			want: &FieldContext{
				validationRules: nil,
				child: &FieldContext{
					validationRules: nil,
					child: &FieldContext{
						validationRules: nil,
						child: &FieldContext{
							validationRules: []string{"lt=5"},
						},
					},
				},
			},
		},
		{
			field: "D",
			want: &FieldContext{
				validationRules: nil,
				child: &FieldContext{
					validationRules: []string{"required"},
					child: &FieldContext{
						validationRules: []string{"gt=3"},
					},
				},
			},
		},
		{
			field: "E",
			want: &FieldContext{
				validationRules: []string{"required"},
				child: &FieldContext{
					validationRules: nil,
					child: &FieldContext{
						validationRules: []string{"gt=1"},
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

func TestNewFieldContext(t *testing.T) {
	tests := []struct {
		name      string
		inputType reflect.Type
		tag       string
		expected  *FieldContext
		wantErr   bool
	}{
		// Test case for no validate tag or "-"
		{
			name:      "No validate tag",
			inputType: reflect.TypeOf(""),
			tag:       "",
			expected:  nil,
			wantErr:   false,
		},
		{
			name:      "Skip field with -",
			inputType: reflect.TypeOf(""),
			tag:       "-",
			expected:  nil,
			wantErr:   false,
		},

		// String format validations
		{
			name:      "alpha",
			inputType: reflect.TypeOf(""),
			tag:       "alpha",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + AlphaRegexString + ")$",
				validationRules: []string{"alpha"},
			},
			wantErr: false,
		},
		{
			name:      "alphanum",
			inputType: reflect.TypeOf(""),
			tag:       "alphanum",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + AlphaNumericRegexString + ")$",
				validationRules: []string{"alphanum"},
			},
			wantErr: false,
		},
		{
			name:      "alphaunicode",
			inputType: reflect.TypeOf(""),
			tag:       "alphaunicode",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + AlphaUnicodeRegexString + ")$",
				validationRules: []string{"alphaunicode"},
			},
			wantErr: false,
		},
		{
			name:      "alphanumunicode",
			inputType: reflect.TypeOf(""),
			tag:       "alphanumunicode",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + AlphaUnicodeNumericRegexString + ")$",
				validationRules: []string{"alphanumunicode"},
			},
			wantErr: false,
		},
		{
			name:      "numeric",
			inputType: reflect.TypeOf(""),
			tag:       "numeric",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + NumericRegexString + ")$",
				validationRules: []string{"numeric"},
			},
			wantErr: false,
		},
		{
			name:      "number",
			inputType: reflect.TypeOf(""),
			tag:       "number",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + NumberRegexString + ")$",
				validationRules: []string{"number"},
			},
			wantErr: false,
		},
		{
			name:      "hexadecimal",
			inputType: reflect.TypeOf(""),
			tag:       "hexadecimal",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + HexadecimalRegexString + ")$",
				validationRules: []string{"hexadecimal"},
			},
			wantErr: false,
		},
		{
			name:      "hexcolor",
			inputType: reflect.TypeOf(""),
			tag:       "hexcolor",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + HexColorRegexString + ")$",
				validationRules: []string{"hexcolor"},
			},
			wantErr: false,
		},
		{
			name:      "rgb",
			inputType: reflect.TypeOf(""),
			tag:       "rgb",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + RgbRegexString + ")$",
				validationRules: []string{"rgb"},
			},
			wantErr: false,
		},
		{
			name:      "rgba",
			inputType: reflect.TypeOf(""),
			tag:       "rgba",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + RgbaRegexString + ")$",
				validationRules: []string{"rgba"},
			},
			wantErr: false,
		},
		{
			name:      "hsl",
			inputType: reflect.TypeOf(""),
			tag:       "hsl",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + HslRegexString + ")$",
				validationRules: []string{"hsl"},
			},
			wantErr: false,
		},
		{
			name:      "hsla",
			inputType: reflect.TypeOf(""),
			tag:       "hsla",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + HslaRegexString + ")$",
				validationRules: []string{"hsla"},
			},
			wantErr: false,
		},
		{
			name:      "email",
			inputType: reflect.TypeOf(""),
			tag:       "email",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				Format:          "email",
				validationRules: []string{"email"},
			},
			wantErr: false,
		},
		{
			name:      "e164",
			inputType: reflect.TypeOf(""),
			tag:       "e164",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + E164RegexString + ")$",
				validationRules: []string{"e164"},
			},
			wantErr: false,
		},
		{
			name:      "base32",
			inputType: reflect.TypeOf(""),
			tag:       "base32",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + Base32RegexString + ")$",
				validationRules: []string{"base32"},
			},
			wantErr: false,
		},
		{
			name:      "base64",
			inputType: reflect.TypeOf(""),
			tag:       "base64",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + Base64RegexString + ")$",
				validationRules: []string{"base64"},
			},
			wantErr: false,
		},
		{
			name:      "base64url",
			inputType: reflect.TypeOf(""),
			tag:       "base64url",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + Base64URLRegexString + ")$",
				validationRules: []string{"base64url"},
			},
			wantErr: false,
		},
		{
			name:      "base64rawurl",
			inputType: reflect.TypeOf(""),
			tag:       "base64rawurl",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + Base64RawURLRegexString + ")$",
				validationRules: []string{"base64rawurl"},
			},
			wantErr: false,
		},
		{
			name:      "isbn",
			inputType: reflect.TypeOf(""),
			tag:       "isbn",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + ISBNRegexString + ")$",
				validationRules: []string{"isbn"},
			},
			wantErr: false,
		},
		{
			name:      "isbn10",
			inputType: reflect.TypeOf(""),
			tag:       "isbn10",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + ISBN10RegexString + ")$",
				validationRules: []string{"isbn10"},
			},
			wantErr: false,
		},
		{
			name:      "isbn13",
			inputType: reflect.TypeOf(""),
			tag:       "isbn13",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + ISBN13RegexString + ")$",
				validationRules: []string{"isbn13"},
			},
			wantErr: false,
		},
		{
			name:      "issn",
			inputType: reflect.TypeOf(""),
			tag:       "issn",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + ISSNRegexString + ")$",
				validationRules: []string{"issn"},
			},
			wantErr: false,
		},
		{
			name:      "uuid",
			inputType: reflect.TypeOf(""),
			tag:       "uuid",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				Format:          "uuid",
				validationRules: []string{"uuid"},
			},
			wantErr: false,
		},
		{
			name:      "uuid3",
			inputType: reflect.TypeOf(""),
			tag:       "uuid3",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + UUID3RegexString + ")$",
				validationRules: []string{"uuid3"},
			},
			wantErr: false,
		},
		{
			name:      "uuid4",
			inputType: reflect.TypeOf(""),
			tag:       "uuid4",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				Format:          "uuid",
				validationRules: []string{"uuid4"},
			},
			wantErr: false,
		},
		{
			name:      "uuid5",
			inputType: reflect.TypeOf(""),
			tag:       "uuid5",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + UUID5RegexString + ")$",
				validationRules: []string{"uuid5"},
			},
			wantErr: false,
		},
		{
			name:      "uuid_rfc4122",
			inputType: reflect.TypeOf(""),
			tag:       "uuid_rfc4122",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + UUIDRFC4122RegexString + ")$",
				validationRules: []string{"uuid_rfc4122"},
			},
			wantErr: false,
		},
		{
			name:      "uuid3_rfc4122",
			inputType: reflect.TypeOf(""),
			tag:       "uuid3_rfc4122",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + UUID3RFC4122RegexString + ")$",
				validationRules: []string{"uuid3_rfc4122"},
			},
			wantErr: false,
		},
		{
			name:      "uuid4_rfc4122",
			inputType: reflect.TypeOf(""),
			tag:       "uuid4_rfc4122",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + UUID4RFC4122RegexString + ")$",
				validationRules: []string{"uuid4_rfc4122"},
			},
			wantErr: false,
		},
		{
			name:      "uuid5_rfc4122",
			inputType: reflect.TypeOf(""),
			tag:       "uuid5_rfc4122",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + UUID5RFC4122RegexString + ")$",
				validationRules: []string{"uuid5_rfc4122"},
			},
			wantErr: false,
		},
		{
			name:      "ulid",
			inputType: reflect.TypeOf(""),
			tag:       "ulid",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + ULIDRegexString + ")$",
				validationRules: []string{"ulid"},
			},
			wantErr: false,
		},
		{
			name:      "md4",
			inputType: reflect.TypeOf(""),
			tag:       "md4",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + MD4RegexString + ")$",
				validationRules: []string{"md4"},
			},
			wantErr: false,
		},
		{
			name:      "md5",
			inputType: reflect.TypeOf(""),
			tag:       "md5",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + MD5RegexString + ")$",
				validationRules: []string{"md5"},
			},
			wantErr: false,
		},
		{
			name:      "sha256",
			inputType: reflect.TypeOf(""),
			tag:       "sha256",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + SHA256RegexString + ")$",
				validationRules: []string{"sha256"},
			},
			wantErr: false,
		},
		{
			name:      "sha384",
			inputType: reflect.TypeOf(""),
			tag:       "sha384",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + SHA384RegexString + ")$",
				validationRules: []string{"sha384"},
			},
			wantErr: false,
		},
		{
			name:      "sha512",
			inputType: reflect.TypeOf(""),
			tag:       "sha512",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + SHA512RegexString + ")$",
				validationRules: []string{"sha512"},
			},
			wantErr: false,
		},
		{
			name:      "ripemd128",
			inputType: reflect.TypeOf(""),
			tag:       "ripemd128",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + Ripemd128RegexString + ")$",
				validationRules: []string{"ripemd128"},
			},
			wantErr: false,
		},
		{
			name:      "ripemd160",
			inputType: reflect.TypeOf(""),
			tag:       "ripemd160",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + Ripemd160RegexString + ")$",
				validationRules: []string{"ripemd160"},
			},
			wantErr: false,
		},
		{
			name:      "tiger128",
			inputType: reflect.TypeOf(""),
			tag:       "tiger128",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + Tiger128RegexString + ")$",
				validationRules: []string{"tiger128"},
			},
			wantErr: false,
		},
		{
			name:      "tiger160",
			inputType: reflect.TypeOf(""),
			tag:       "tiger160",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + Tiger160RegexString + ")$",
				validationRules: []string{"tiger160"},
			},
			wantErr: false,
		},
		{
			name:      "tiger192",
			inputType: reflect.TypeOf(""),
			tag:       "tiger192",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + Tiger192RegexString + ")$",
				validationRules: []string{"tiger192"},
			},
			wantErr: false,
		},
		{
			name:      "ascii",
			inputType: reflect.TypeOf(""),
			tag:       "ascii",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        false,
				Nullable:        false,
				pattern:         "^(" + ASCIIRegexString + ")$",
				validationRules: []string{"ascii"},
			},
			wantErr: false,
		},
		{
			name:      "printascii",
			inputType: reflect.TypeOf(""),
			tag:       "printascii",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        false,
				Nullable:        false,
				pattern:         "^(" + PrintableASCIIRegexString + ")$",
				validationRules: []string{"printascii"},
			},
			wantErr: false,
		},
		{
			name:      "multibyte",
			inputType: reflect.TypeOf(""),
			tag:       "multibyte",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        false,
				Nullable:        false,
				pattern:         "^(" + MultibyteRegexString + ")$",
				validationRules: []string{"multibyte"},
			},
			wantErr: false,
		},
		{
			name:      "datauri",
			inputType: reflect.TypeOf(""),
			tag:       "datauri",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + DataURIRegexString + ")$",
				validationRules: []string{"datauri"},
			},
			wantErr: false,
		},
		{
			name:      "latitude string",
			inputType: reflect.TypeOf(""),
			tag:       "latitude",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + LatitudeRegexString + ")$",
				validationRules: []string{"latitude"},
			},
			wantErr: false,
		},
		{
			name:      "latitude float",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "latitude",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				Min:             ptr(float64(-90)),
				Max:             ptr(float64(90)),
				validationRules: []string{"latitude"},
			},
			wantErr: false,
		},
		{
			name:      "longitude string",
			inputType: reflect.TypeOf(""),
			tag:       "longitude",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + LongitudeRegexString + ")$",
				validationRules: []string{"longitude"},
			},
			wantErr: false,
		},
		{
			name:      "longitude float",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "longitude",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				Min:             ptr(float64(-180)),
				Max:             ptr(float64(180)),
				validationRules: []string{"longitude"},
			},
			wantErr: false,
		},
		{
			name:      "ssn",
			inputType: reflect.TypeOf(""),
			tag:       "ssn",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + SSNRegexString + ")$",
				validationRules: []string{"ssn"},
			},
			wantErr: false,
		},
		{
			name:      "hostname",
			inputType: reflect.TypeOf(""),
			tag:       "hostname",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + HostnameRegexStringRFC952 + ")$",
				validationRules: []string{"hostname"},
			},
			wantErr: false,
		},
		{
			name:      "hostname_rfc1123",
			inputType: reflect.TypeOf(""),
			tag:       "hostname_rfc1123",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				Format:          "hostname",
				validationRules: []string{"hostname_rfc1123"},
			},
			wantErr: false,
		},
		{
			name:      "fqdn",
			inputType: reflect.TypeOf(""),
			tag:       "fqdn",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + FqdnRegexStringRFC1123 + ")$",
				validationRules: []string{"fqdn"},
			},
			wantErr: false,
		},
		{
			name:      "btc_addr",
			inputType: reflect.TypeOf(""),
			tag:       "btc_addr",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + BtcAddressRegexString + ")$",
				validationRules: []string{"btc_addr"},
			},
			wantErr: false,
		},
		{
			name:      "btc_addr_bech32",
			inputType: reflect.TypeOf(""),
			tag:       "btc_addr_bech32",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + BtcAddressBech32RegexString + ")$",
				validationRules: []string{"btc_addr_bech32"},
			},
			wantErr: false,
		},
		{
			name:      "eth_addr",
			inputType: reflect.TypeOf(""),
			tag:       "eth_addr",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + EthAddressRegexString + ")$",
				validationRules: []string{"eth_addr"},
			},
			wantErr: false,
		},
		{
			name:      "url_encoded",
			inputType: reflect.TypeOf(""),
			tag:       "url_encoded",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + URLEncodedRegexString + ")$",
				validationRules: []string{"url_encoded"},
			},
			wantErr: false,
		},
		{
			name:      "html",
			inputType: reflect.TypeOf(""),
			tag:       "html",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + HTMLRegexString + ")$",
				validationRules: []string{"html"},
			},
			wantErr: false,
		},
		{
			name:      "html_encoded",
			inputType: reflect.TypeOf(""),
			tag:       "html_encoded",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + HTMLEncodedRegexString + ")$",
				validationRules: []string{"html_encoded"},
			},
			wantErr: false,
		},
		{
			name:      "jwt",
			inputType: reflect.TypeOf(""),
			tag:       "jwt",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + JWTRegexString + ")$",
				validationRules: []string{"jwt"},
			},
			wantErr: false,
		},
		{
			name:      "bic",
			inputType: reflect.TypeOf(""),
			tag:       "bic",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + BicRegexString + ")$",
				validationRules: []string{"bic"},
			},
			wantErr: false,
		},
		{
			name:      "semver",
			inputType: reflect.TypeOf(""),
			tag:       "semver",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + SemverRegexString + ")$",
				validationRules: []string{"semver"},
			},
			wantErr: false,
		},
		{
			name:      "dns_rfc1035_label",
			inputType: reflect.TypeOf(""),
			tag:       "dns_rfc1035_label",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + DNSRegexStringRFC1035Label + ")$",
				validationRules: []string{"dns_rfc1035_label"},
			},
			wantErr: false,
		},
		{
			name:      "cve",
			inputType: reflect.TypeOf(""),
			tag:       "cve",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + CVERegexString + ")$",
				validationRules: []string{"cve"},
			},
			wantErr: false,
		},
		{
			name:      "mongodb",
			inputType: reflect.TypeOf(""),
			tag:       "mongodb",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + MongodbIdRegexString + ")$",
				validationRules: []string{"mongodb"},
			},
			wantErr: false,
		},
		{
			name:      "mongodb_connection_string",
			inputType: reflect.TypeOf(""),
			tag:       "mongodb_connection_string",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + MongodbConnStringRegexString + ")$",
				validationRules: []string{"mongodb_connection_string"},
			},
			wantErr: false,
		},
		{
			name:      "cron",
			inputType: reflect.TypeOf(""),
			tag:       "cron",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + CronRegexString + ")$",
				validationRules: []string{"cron"},
			},
			wantErr: false,
		},
		{
			name:      "spicedb",
			inputType: reflect.TypeOf(""),
			tag:       "spicedb",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + SpicedbIDRegexString + ")$",
				validationRules: []string{"spicedb"},
			},
			wantErr: false,
		},
		{
			name:      "ein",
			inputType: reflect.TypeOf(""),
			tag:       "ein",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + EinRegexString + ")$",
				validationRules: []string{"ein"},
			},
			wantErr: false,
		},
		{
			name:      "cidr",
			inputType: reflect.TypeOf(""),
			tag:       "cidr",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + CIDRRegex + ")$",
				validationRules: []string{"cidr"},
			},
			wantErr: false,
		},
		{
			name:      "cidrv4",
			inputType: reflect.TypeOf(""),
			tag:       "cidrv4",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + CIDRRegex + ")$",
				validationRules: []string{"cidrv4"},
			},
			wantErr: false,
		},
		{
			name:      "hostname_port",
			inputType: reflect.TypeOf(""),
			tag:       "hostname_port",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				Format:          "hostname",
				validationRules: []string{"hostname_port"},
			},
			wantErr: false,
		},
		{
			name:      "ip",
			inputType: reflect.TypeOf(""),
			tag:       "ip",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + IPAnyRegexString + ")$",
				validationRules: []string{"ip"},
			},
			wantErr: false,
		},
		{
			name:      "ip_addr",
			inputType: reflect.TypeOf(""),
			tag:       "ip_addr",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + IPAnyRegexString + ")$",
				validationRules: []string{"ip_addr"},
			},
			wantErr: false,
		},
		{
			name:      "ipv4",
			inputType: reflect.TypeOf(""),
			tag:       "ipv4",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				Format:          "ipv4",
				validationRules: []string{"ipv4"},
			},
			wantErr: false,
		},
		{
			name:      "ip4_addr",
			inputType: reflect.TypeOf(""),
			tag:       "ip4_addr",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				Format:          "ipv4",
				validationRules: []string{"ip4_addr"},
			},
			wantErr: false,
		},
		{
			name:      "ipv6",
			inputType: reflect.TypeOf(""),
			tag:       "ipv6",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				Format:          "ipv6",
				validationRules: []string{"ipv6"},
			},
			wantErr: false,
		},
		{
			name:      "ip6_addr",
			inputType: reflect.TypeOf(""),
			tag:       "ip6_addr",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				Format:          "ipv6",
				validationRules: []string{"ip6_addr"},
			},
			wantErr: false,
		},
		{
			name:      "mac",
			inputType: reflect.TypeOf(""),
			tag:       "mac",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + MACRegexString + ")$",
				validationRules: []string{"mac"},
			},
			wantErr: false,
		},
		{
			name:      "tcp4_addr",
			inputType: reflect.TypeOf(""),
			tag:       "tcp4_addr",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + TCP4AddrRegexString + ")$",
				validationRules: []string{"tcp4_addr"},
			},
			wantErr: false,
		},
		{
			name:      "tcp6_addr",
			inputType: reflect.TypeOf(""),
			tag:       "tcp6_addr",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + TCP6AddrRegexString + ")$",
				validationRules: []string{"tcp6_addr"},
			},
			wantErr: false,
		},
		{
			name:      "tcp_addr",
			inputType: reflect.TypeOf(""),
			tag:       "tcp_addr",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + TCPAddrRegexString + ")$",
				validationRules: []string{"tcp_addr"},
			},
			wantErr: false,
		},
		{
			name:      "udp4_addr",
			inputType: reflect.TypeOf(""),
			tag:       "udp4_addr",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + UDP4AddrRegexString + ")$",
				validationRules: []string{"udp4_addr"},
			},
			wantErr: false,
		},
		{
			name:      "udp6_addr",
			inputType: reflect.TypeOf(""),
			tag:       "udp6_addr",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + UDP6AddrRegexString + ")$",
				validationRules: []string{"udp6_addr"},
			},
			wantErr: false,
		},
		{
			name:      "udp_addr",
			inputType: reflect.TypeOf(""),
			tag:       "udp_addr",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + UDPAddrRegexString + ")$",
				validationRules: []string{"udp_addr"},
			},
			wantErr: false,
		},
		{
			name:      "unix_addr",
			inputType: reflect.TypeOf(""),
			tag:       "unix_addr",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + UnixAddrRegexString + ")$",
				validationRules: []string{"unix_addr"},
			},
			wantErr: false,
		},
		{
			name:      "uri",
			inputType: reflect.TypeOf(""),
			tag:       "uri",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				Format:          "uri",
				validationRules: []string{"uri"},
			},
			wantErr: false,
		},
		{
			name:      "url",
			inputType: reflect.TypeOf(""),
			tag:       "url",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				Format:          "uri",
				validationRules: []string{"url"},
			},
			wantErr: false,
		},
		{
			name:      "urn_rfc2141",
			inputType: reflect.TypeOf(""),
			tag:       "urn_rfc2141",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(" + URNRFC2141RegexString + ")$",
				validationRules: []string{"urn_rfc2141"},
			},
			wantErr: false,
		},

		// Comparison validations
		{
			name:      "required",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "required",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				Not:             []any{0},
				validationRules: []string{"required"},
			},
			wantErr: false,
		},
		{
			name:      "required slice",
			inputType: reflect.TypeOf([]string{}),
			tag:       "required",
			expected: &FieldContext{
				Type:            reflect.TypeOf([]string{}),
				Required:        true,
				Nullable:        false,
				validationRules: []string{"required"},
			},
			wantErr: false,
		},
		{
			name:      "min float",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "min=5.5",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				Min:             ptr(float64(5.5)),
				validationRules: []string{"min=5.5"},
			},
			wantErr: false,
		},
		{
			name:      "min string",
			inputType: reflect.TypeOf(""),
			tag:       "min=5",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				MinLength:       5,
				validationRules: []string{"min=5"},
			},
			wantErr: false,
		},
		{
			name:      "min slice",
			inputType: reflect.TypeOf([]string{}),
			tag:       "min=3",
			expected: &FieldContext{
				Type:            reflect.TypeOf([]string{}),
				Required:        true,
				Nullable:        true,
				MinItems:        3,
				validationRules: []string{"min=3"},
			},
			wantErr: false,
		},
		{
			name:      "min duration",
			inputType: reflect.TypeOf(time.Duration(0)),
			tag:       "min=1h",
			expected: &FieldContext{
				Type:            reflect.TypeOf(time.Duration(0)),
				Required:        true,
				Nullable:        false,
				Min:             ptr(float64(time.Hour)),
				validationRules: []string{"min=1h"},
			},
			wantErr: false,
		},
		{
			name:      "max float",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "max=10.5",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				Max:             ptr(float64(10.5)),
				validationRules: []string{"max=10.5"},
			},
			wantErr: false,
		},
		{
			name:      "max string",
			inputType: reflect.TypeOf(""),
			tag:       "max=10",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				MaxLength:       10,
				validationRules: []string{"max=10"},
			},
			wantErr: false,
		},
		{
			name:      "max slice",
			inputType: reflect.TypeOf([]string{}),
			tag:       "max=5",
			expected: &FieldContext{
				Type:            reflect.TypeOf([]string{}),
				Required:        true,
				Nullable:        true,
				MaxItems:        5,
				validationRules: []string{"max=5"},
			},
			wantErr: false,
		},
		{
			name:      "max duration",
			inputType: reflect.TypeOf(time.Duration(0)),
			tag:       "max=2h",
			expected: &FieldContext{
				Type:            reflect.TypeOf(time.Duration(0)),
				Required:        true,
				Nullable:        false,
				Max:             ptr(float64(2 * time.Hour)),
				validationRules: []string{"max=2h"},
			},
			wantErr: false,
		},
		{
			name:      "len string",
			inputType: reflect.TypeOf(""),
			tag:       "len=5",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        false,
				Nullable:        false,
				MinLength:       5,
				MaxLength:       5,
				validationRules: []string{"len=5"},
			},
			wantErr: false,
		},
		{
			name:      "len slice",
			inputType: reflect.TypeOf([]string{}),
			tag:       "len=3",
			expected: &FieldContext{
				Type:            reflect.TypeOf([]string{}),
				Required:        false,
				Nullable:        true,
				MinItems:        3,
				MaxItems:        3,
				validationRules: []string{"len=3"},
			},
			wantErr: false,
		},
		{
			name:      "lt float",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "lt=10",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				Max:             ptr(float64(10)),
				ExclusiveMax:    true,
				validationRules: []string{"lt=10"},
			},
			wantErr: false,
		},
		{
			name:      "lt string",
			inputType: reflect.TypeOf(""),
			tag:       "lt=10",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				MaxLength:       9,
				validationRules: []string{"lt=10"},
			},
			wantErr: false,
		},
		{
			name:      "lt slice",
			inputType: reflect.TypeOf([]string{}),
			tag:       "lt=5",
			expected: &FieldContext{
				Type:            reflect.TypeOf([]string{}),
				Required:        true,
				Nullable:        true,
				MaxItems:        4,
				validationRules: []string{"lt=5"},
			},
			wantErr: false,
		},
		{
			name:      "lt duration",
			inputType: reflect.TypeOf(time.Duration(0)),
			tag:       "lt=1h",
			expected: &FieldContext{
				Type:            reflect.TypeOf(time.Duration(0)),
				Required:        true,
				Nullable:        false,
				Max:             ptr(float64(time.Hour)),
				ExclusiveMax:    true,
				validationRules: []string{"lt=1h"},
			},
			wantErr: false,
		},
		{
			name:      "lte float",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "lte=10",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				Max:             ptr(float64(10)),
				validationRules: []string{"lte=10"},
			},
			wantErr: false,
		},
		{
			name:      "lte string",
			inputType: reflect.TypeOf(""),
			tag:       "lte=10",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				MaxLength:       10,
				validationRules: []string{"lte=10"},
			},
			wantErr: false,
		},
		{
			name:      "lte slice",
			inputType: reflect.TypeOf([]string{}),
			tag:       "lte=5",
			expected: &FieldContext{
				Type:            reflect.TypeOf([]string{}),
				Required:        true,
				Nullable:        true,
				MaxItems:        5,
				validationRules: []string{"lte=5"},
			},
			wantErr: false,
		},
		{
			name:      "lte duration",
			inputType: reflect.TypeOf(time.Duration(0)),
			tag:       "lte=1h",
			expected: &FieldContext{
				Type:            reflect.TypeOf(time.Duration(0)),
				Required:        true,
				Nullable:        false,
				Max:             ptr(float64(time.Hour)),
				validationRules: []string{"lte=1h"},
			},
			wantErr: false,
		},
		{
			name:      "gt float",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "gt=5",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				Min:             ptr(float64(5)),
				validationRules: []string{"gt=5"},
			},
			wantErr: false,
		},
		{
			name:      "gt string",
			inputType: reflect.TypeOf(""),
			tag:       "gt=5",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				MaxLength:       5,
				validationRules: []string{"gt=5"},
			},
			wantErr: false,
		},
		{
			name:      "gt slice",
			inputType: reflect.TypeOf([]string{}),
			tag:       "gt=3",
			expected: &FieldContext{
				Type:            reflect.TypeOf([]string{}),
				Required:        true,
				Nullable:        true,
				MinItems:        3,
				validationRules: []string{"gt=3"},
			},
			wantErr: false,
		},
		{
			name:      "gt duration",
			inputType: reflect.TypeOf(time.Duration(0)),
			tag:       "gt=1h",
			expected: &FieldContext{
				Type:            reflect.TypeOf(time.Duration(0)),
				Required:        true,
				Nullable:        false,
				Min:             ptr(float64(time.Hour)),
				validationRules: []string{"gt=1h"},
			},
			wantErr: false,
		},
		{
			name:      "gte float",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "gte=5",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				Min:             ptr(float64(5)),
				validationRules: []string{"gte=5"},
			},
			wantErr: false,
		},
		{
			name:      "eq float",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "eq=5",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				Min:             ptr(float64(5)),
				Max:             ptr(float64(5)),
				validationRules: []string{"eq=5"},
			},
			wantErr: false,
		},
		{
			name:      "eq string",
			inputType: reflect.TypeOf(""),
			tag:       "eq=abc",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(abc)$",
				validationRules: []string{"eq=abc"},
			},
			wantErr: false,
		},
		{
			name:      "eq slice",
			inputType: reflect.TypeOf([]string{}),
			tag:       "eq=3",
			expected: &FieldContext{
				Type:            reflect.TypeOf([]string{}),
				Required:        true,
				Nullable:        true,
				MinItems:        3,
				MaxItems:        3,
				validationRules: []string{"eq=3"},
			},
			wantErr: false,
		},
		{
			name:      "eq duration",
			inputType: reflect.TypeOf(time.Duration(0)),
			tag:       "eq=1h",
			expected: &FieldContext{
				Type:            reflect.TypeOf(time.Duration(0)),
				Required:        true,
				Nullable:        false,
				Min:             ptr(float64(time.Hour)),
				Max:             ptr(float64(time.Hour)),
				ExclusiveMin:    true,
				ExclusiveMax:    true,
				validationRules: []string{"eq=1h"},
			},
			wantErr: false,
		},
		{
			name:      "ne float",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "ne=5",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				Not:             []any{float64(5)},
				validationRules: []string{"ne=5"},
			},
			wantErr: false,
		},
		{
			name:      "ne string",
			inputType: reflect.TypeOf(""),
			tag:       "ne=abc",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^(((?!abc).)*)$",
				validationRules: []string{"ne=abc"},
			},
			wantErr: false,
		},
		{
			name:      "ne duration",
			inputType: reflect.TypeOf(time.Duration(0)),
			tag:       "ne=1h",
			expected: &FieldContext{
				Type:            reflect.TypeOf(time.Duration(0)),
				Required:        true,
				Nullable:        false,
				Not:             []any{time.Hour},
				validationRules: []string{"ne=1h"},
			},
			wantErr: false,
		},
		{
			name:      "oneof string",
			inputType: reflect.TypeOf(""),
			tag:       "oneof=red green blue",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				OneOf:           []any{"red", "green", "blue"},
				validationRules: []string{"oneof=red green blue"},
			},
			wantErr: false,
		},
		{
			name:      "oneof float",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "oneof=1.5 2.5 3.5",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				OneOf:           []any{1.5, 2.5, 3.5},
				validationRules: []string{"oneof=1.5 2.5 3.5"},
			},
			wantErr: false,
		},

		// Dive tag with nested validation
		{
			name:      "dive with string",
			inputType: reflect.TypeOf([]string{}),
			tag:       "dive,alphanum",
			expected: &FieldContext{
				Type:            reflect.TypeOf([]string{}),
				Required:        false,
				Nullable:        true,
				validationRules: []string{},
				child: &FieldContext{
					Type:            reflect.TypeOf(""),
					Required:        true,
					Nullable:        false,
					pattern:         "^(" + AlphaNumericRegexString + ")$",
					validationRules: []string{"alphanum"},
				},
			},
			wantErr: false,
		},

		// OR condition
		{
			name:      "or condition",
			inputType: reflect.TypeOf(""),
			tag:       "hexcolor|rgb",
			expected: &FieldContext{
				Type:            reflect.TypeOf(""),
				Required:        true,
				Nullable:        false,
				pattern:         "^((" + HexColorRegexString + ")|(" + RgbRegexString + "))$",
				validationRules: []string{"hexcolor|rgb"},
				hasOrConditions: true,
			},
			wantErr: false,
		},

		// Invalid cases
		{
			name:      "invalid min duration",
			inputType: reflect.TypeOf(time.Duration(0)),
			tag:       "min=invalid",
			expected: &FieldContext{
				Type:            reflect.TypeOf(time.Duration(0)),
				Required:        true,
				Nullable:        false,
				validationRules: []string{"min=invalid"},
			},
			wantErr: true, // Expect error due to invalid duration
		},
		{
			name:      "invalid min float",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "min=invalid",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				validationRules: []string{"min=invalid"},
			},
			wantErr: true, // Expect error due to invalid float
		},
		{
			name:      "invalid oneof float",
			inputType: reflect.TypeOf(float64(0)),
			tag:       "oneof=invalid",
			expected: &FieldContext{
				Type:            reflect.TypeOf(float64(0)),
				Required:        true,
				Nullable:        false,
				validationRules: []string{"oneof=invalid"},
			},
			wantErr: true, // Expect error due to invalid float slice
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := reflect.StructTag(`validate:"` + tt.tag + `"`)
			ctx := NewFieldContext(tt.inputType, tag)

			if tt.expected == nil {
				if ctx != nil {
					t.Errorf("expected nil context, got %v", ctx)
				}
				return
			}

			if ctx == nil {
				t.Errorf("expected non-nil context, got nil")
				return
			}

			if ctx.Type != tt.expected.Type {
				t.Errorf("Type: got %v, want %v", ctx.Type, tt.expected.Type)
			}

			if ctx.Required != tt.expected.Required {
				t.Errorf("Required: got %v, want %v", ctx.Required, tt.expected.Required)
			}

			if ctx.Nullable != tt.expected.Nullable {
				t.Errorf("Nullable: got %v, want %v", ctx.Nullable, tt.expected.Nullable)
			}

			if ctx.Format != tt.expected.Format {
				t.Errorf("Format: got %v, want %v", ctx.Format, tt.expected.Format)
			}

			if ctx.pattern != tt.expected.pattern {
				t.Errorf("Pattern: got %v, want %v", ctx.pattern, tt.expected.pattern)
			}

			if tt.expected.Min != nil {
				if ctx.Min == nil || *ctx.Min != *tt.expected.Min {
					t.Errorf("Min: got %v, want %v", ctx.Min, tt.expected.Min)
				}
			} else if ctx.Min != nil {
				t.Errorf("Min: got %v, want nil", ctx.Min)
			}

			if tt.expected.Max != nil {
				if ctx.Max == nil || *ctx.Max != *tt.expected.Max {
					t.Errorf("Max: got %v, want %v", ctx.Max, tt.expected.Max)
				}
			} else if ctx.Max != nil {
				t.Errorf("Max: got %v, want nil", ctx.Max)
			}

			if ctx.ExclusiveMin != tt.expected.ExclusiveMin {
				t.Errorf("ExclusiveMin: got %v, want %v", ctx.ExclusiveMin, tt.expected.ExclusiveMin)
			}

			if ctx.ExclusiveMax != tt.expected.ExclusiveMax {
				t.Errorf("ExclusiveMax: got %v, want %v", ctx.ExclusiveMax, tt.expected.ExclusiveMax)
			}

			if ctx.MinItems != tt.expected.MinItems {
				t.Errorf("MinItems: got %v, want %v", ctx.MinItems, tt.expected.MinItems)
			}

			if ctx.MaxItems != tt.expected.MaxItems {
				t.Errorf("MaxItems: got %v, want %v", ctx.MaxItems, tt.expected.MaxItems)
			}

			if ctx.MinLength != tt.expected.MinLength {
				t.Errorf("MinLength: got %v, want %v", ctx.MinLength, tt.expected.MinLength)
			}

			if ctx.MaxLength != tt.expected.MaxLength {
				t.Errorf("MaxLength: got %v, want %v", ctx.MaxLength, tt.expected.MaxLength)
			}

			if !reflect.DeepEqual(ctx.Not, tt.expected.Not) {
				t.Errorf("Not: got %v, want %v", ctx.Not, tt.expected.Not)
			}

			if !reflect.DeepEqual(ctx.OneOf, tt.expected.OneOf) {
				t.Errorf("OneOf: got %v, want %v", ctx.OneOf, tt.expected.OneOf)
			}

			if !reflect.DeepEqual(ctx.validationRules, tt.expected.validationRules) {
				t.Errorf("validationRules: got %v, want %v", ctx.validationRules, tt.expected.validationRules)
			}

			if ctx.hasOrConditions != tt.expected.hasOrConditions {
				t.Errorf("hasOrConditions: got %v, want %v", ctx.hasOrConditions, tt.expected.hasOrConditions)
			}

			if tt.expected.child != nil {
				if ctx.child == nil {
					t.Errorf("expected child context, got nil")
				} else {
					// Recursively compare child context
					if ctx.child.Type != tt.expected.child.Type {
						t.Errorf("child.Type: got %v, want %v", ctx.child.Type, tt.expected.child.Type)
					}
					if ctx.child.Required != tt.expected.child.Required {
						t.Errorf("child.Required: got %v, want %v", ctx.child.Required, tt.expected.child.Required)
					}
					if ctx.child.pattern != tt.expected.child.pattern {
						t.Errorf("child.pattern: got %v, want %v", ctx.child.pattern, tt.expected.child.pattern)
					}
					if !reflect.DeepEqual(ctx.child.validationRules, tt.expected.child.validationRules) {
						t.Errorf("child.validationRules: got %v, want %v", ctx.child.validationRules, tt.expected.child.validationRules)
					}
				}
			} else if ctx.child != nil {
				t.Errorf("child: got %v, want nil", ctx.child)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
