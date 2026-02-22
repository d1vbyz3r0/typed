package typed

import (
	"mime/multipart"
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/stretchr/testify/require"
)

func TestMakeFieldsRequiredSkipsFileHeader(t *testing.T) {
	schema := openapi3.NewStringSchema()

	err := makeFieldsRequired("", reflect.TypeOf(multipart.FileHeader{}), "", schema)
	require.NoError(t, err)
	require.Empty(t, schema.Required)
}

func TestMakeFieldsRequiredStillWorksForRegularStruct(t *testing.T) {
	type payload struct {
		Name  string            `json:"name"`
		Count int               `json:"count"`
		Meta  map[string]string `json:"meta"`
		Ref   *string           `json:"ref"`
	}

	schema := openapi3.NewObjectSchema()

	err := makeFieldsRequired("", reflect.TypeOf(payload{}), "", schema)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"name", "count"}, schema.Required)
}

func TestCustomizerFileHeaderHasNoRequired(t *testing.T) {
	schemas := make(openapi3.Schemas)
	g := openapi3gen.NewGenerator(openapi3gen.SchemaCustomizer(Customizer))

	ref, err := g.NewSchemaRefForValue(&multipart.FileHeader{}, schemas)
	require.NoError(t, err)
	require.NotNil(t, ref.Value)
	require.Equal(t, &openapi3.Types{openapi3.TypeString}, ref.Value.Type)
	require.Equal(t, "binary", ref.Value.Format)
	require.Empty(t, ref.Value.Required)
}
