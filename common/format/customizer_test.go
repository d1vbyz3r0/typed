package format

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCustomizer_Required(t *testing.T) {
	type CustomString string

	type TestStruct struct {
		String       string         `json:"string" validate:"required"`
		Slice        []int          `json:"slice" validate:"required"`
		Map          map[string]any `json:"map" validate:"required"`
		Ptr          *int           `json:"ptr" validate:"required"`
		Uint         uint           `json:"uint" validate:"required"`
		Int          int            `json:"int" validate:"required"`
		Float        float64        `json:"float" validate:"required"`
		CustomString CustomString   `json:"customString" validate:"required"`
		KnownFormat  string         `json:"knownFormat" validate:"email"`
	}

	schemas := make(openapi3.Schemas)
	g := openapi3gen.NewGenerator(openapi3gen.SchemaCustomizer(Customizer))
	ref, err := g.NewSchemaRefForValue(new(TestStruct), schemas)
	require.NoError(t, err)

	require.Equal(t, NonEmptyRegexString, ref.Value.Properties["string"].Value.Pattern)
	require.Equal(t, false, ref.Value.Properties["slice"].Value.Nullable)
	require.Equal(t, false, ref.Value.Properties["map"].Value.Nullable)
	// require.Equal(t, false, ref.Value.Properties["ptr"].Value.Nullable)
	require.ElementsMatch(t, []any{0}, ref.Value.Properties["uint"].Value.Not.Value.Enum)
	require.ElementsMatch(t, []any{0}, ref.Value.Properties["int"].Value.Not.Value.Enum)
	require.ElementsMatch(t, []any{0}, ref.Value.Properties["float"].Value.Not.Value.Enum)
	require.Equal(t, NonEmptyRegexString, ref.Value.Properties["customString"].Value.Pattern)
	require.Equal(t, "", ref.Value.Properties["knownFormat"].Value.Pattern)
	require.Equal(t, Email, ref.Value.Properties["knownFormat"].Value.Format)

	require.ElementsMatch(t, []string{"string", "slice", "map", "ptr", "uint", "int", "float", "customString", "knownFormat"}, ref.Value.Required)
}
