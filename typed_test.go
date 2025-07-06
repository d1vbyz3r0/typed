package typed

import (
	"fmt"
	"github.com/d1vbyz3r0/typed/handlers"
	"github.com/d1vbyz3r0/typed/internal/parser"
	"github.com/d1vbyz3r0/typed/internal/parser/request"
	"github.com/d1vbyz3r0/typed/internal/parser/request/path"
	"github.com/d1vbyz3r0/typed/internal/parser/request/query"
	"github.com/d1vbyz3r0/typed/testdata/models"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"reflect"
	"testing"
)

var (
	g        *openapi3gen.Generator
	registry = map[string]any{
		"models.User":        new(models.User),
		"models.PathParams":  new(models.PathParams),
		"models.QueryParams": new(models.QueryParams),
	}
	enums = map[string][]any{
		"models.UserRole": {"Admin", "User"},
	}
)

func TestMain(m *testing.M) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	g = openapi3gen.NewGenerator(
		openapi3gen.UseAllExportedFields(),
		openapi3gen.CreateComponentSchemas(openapi3gen.ExportComponentSchemasOptions{
			ExportComponentSchemas: true,
			ExportTopLevelSchema:   true,
			ExportGenerics:         false,
		}),
		openapi3gen.SchemaCustomizer(Customizer),
		openapi3gen.CreateTypeNameGenerator(func(t reflect.Type) string {
			return t.String()
		}),
	)

	RegisterCustomizer(NewEnumsCustomizer(enums))

	m.Run()
}

func TestTagOperation(t *testing.T) {
	tests := []struct {
		ApiPrefix string
		Route     string
	}{
		{
			ApiPrefix: "/api/v1",
			Route:     "/api/v1/tasks/test/",
		},
		{
			ApiPrefix: "/api/v1/",
			Route:     "/api/v1/tasks/test",
		},
		{
			ApiPrefix: "/api/v1",
			Route:     "/api/v1/tasks/",
		},
		{
			ApiPrefix: "/api/v1/",
			Route:     "/api/v1/tasks",
		},
	}

	expected := []string{"Tasks"}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			op := openapi3.NewOperation()
			err := TagOperation(op, tc.Route, tc.ApiPrefix)
			require.NoError(t, err)
			assert.Equal(t, expected, op.Tags)
		})
	}
}

func TestAddPathParams(t *testing.T) {
	cases := []struct {
		Name    string
		Handler handlers.Handler
		Want    *openapi3.Operation
	}{
		{
			Name: "Bind model only",
			Handler: handlers.NewHandler(echo.Route{}, parser.Handler{
				Doc:  "",
				Name: "Test",
				Pkg:  "test",
				Request: &request.Request{
					BindModel:    "models.PathParams",
					BindModelPkg: "",
					PathParams:   nil,
				},
				Responses: nil,
			}),
			Want: &openapi3.Operation{
				Parameters: openapi3.Parameters{
					{
						Value: &openapi3.Parameter{
							Name:     "uuid",
							In:       "path",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "UUID",
								Value: &openapi3.Schema{
									Type:   &openapi3.Types{openapi3.TypeString},
									Format: "uuid",
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "name",
							In:       "path",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "string",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeString},
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "id",
							In:       "path",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "int",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeInteger},
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "role",
							In:       "path",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "UserRole",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeString},
									Enum: enums["models.UserRole"],
								}},
						},
					},
				},
			},
		},
		{
			Name: "Inline params only",
			Handler: handlers.NewHandler(echo.Route{}, parser.Handler{
				Doc:  "",
				Name: "Test",
				Pkg:  "test",
				Request: &request.Request{
					BindModel:    "",
					BindModelPkg: "",
					PathParams: []path.Param{
						{
							Name: "uuid",
							Type: reflect.TypeOf(uuid.UUID{}),
						},
						{
							Name: "name",
							Type: reflect.TypeOf(""),
						},
						{
							Name: "id",
							Type: reflect.TypeOf(int(0)),
						},
						{
							Name: "role",
							Type: reflect.TypeOf(""),
						},
					},
				},
				Responses: nil,
			}),
			Want: &openapi3.Operation{
				Parameters: openapi3.Parameters{
					{
						Value: &openapi3.Parameter{
							Name:     "uuid",
							In:       "path",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "UUID",
								Value: &openapi3.Schema{
									Type:   &openapi3.Types{openapi3.TypeString},
									Format: "uuid",
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "name",
							In:       "path",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "string",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeString},
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "id",
							In:       "path",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "int",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeInteger},
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "role",
							In:       "path",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "string",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeString},
								}},
						},
					},
				},
			},
		},
		{
			Name: "Mixed",
			Handler: handlers.NewHandler(echo.Route{}, parser.Handler{
				Doc:  "",
				Name: "Test",
				Pkg:  "test",
				Request: &request.Request{
					BindModel:    "models.PathParams",
					BindModelPkg: "",
					PathParams: []path.Param{
						{
							Name: "uuid",
							Type: reflect.TypeOf(uuid.UUID{}),
						},
						{
							Name: "name",
							Type: reflect.TypeOf(""),
						},
						{
							Name: "id",
							Type: reflect.TypeOf(int(0)),
						},
						{
							Name: "role",
							Type: reflect.TypeOf(""),
						},
					},
				},
				Responses: nil,
			}),
			Want: &openapi3.Operation{
				Parameters: openapi3.Parameters{
					{
						Value: &openapi3.Parameter{
							Name:     "uuid",
							In:       "path",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "UUID",
								Value: &openapi3.Schema{
									Type:   &openapi3.Types{openapi3.TypeString},
									Format: "uuid",
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "name",
							In:       "path",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "string",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeString},
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "id",
							In:       "path",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "int",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeInteger},
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "role",
							In:       "path",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "UserRole",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeString},
								}},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			op := openapi3.NewOperation()
			AddPathParams(op, tc.Handler, g, registry)

			got, err := op.MarshalJSON()
			require.NoError(t, err)
			fmt.Println(string(got))

			want, err := tc.Want.MarshalJSON()
			require.NoError(t, err)
			fmt.Println(string(want))

			require.Equal(t, string(want), string(got))
		})
	}
}

func TestAddQueryParams(t *testing.T) {
	cases := []struct {
		Name    string
		Handler handlers.Handler
		Want    *openapi3.Operation
	}{
		{
			Name: "Bind model only",
			Handler: handlers.NewHandler(echo.Route{}, parser.Handler{
				Doc:  "",
				Name: "Test",
				Pkg:  "test",
				Request: &request.Request{
					BindModel:    "models.QueryParams",
					BindModelPkg: "",
					PathParams:   nil,
				},
				Responses: nil,
			}),
			Want: &openapi3.Operation{
				Parameters: openapi3.Parameters{
					{
						Value: &openapi3.Parameter{
							Name:     "uuid",
							In:       "query",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "UUID",
								Value: &openapi3.Schema{
									Type:   &openapi3.Types{openapi3.TypeString},
									Format: "uuid",
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "name",
							In:       "query",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "string",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeString},
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "id",
							In:       "query",
							Required: false,
							Schema: &openapi3.SchemaRef{
								Ref: "int",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeInteger},
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "role",
							In:       "query",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "UserRole",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeString},
									Enum: enums["models.UserRole"],
								}},
						},
					},
				},
			},
		},
		{
			Name: "Inline params only",
			Handler: handlers.NewHandler(echo.Route{}, parser.Handler{
				Doc:  "",
				Name: "Test",
				Pkg:  "test",
				Request: &request.Request{
					BindModel:    "",
					BindModelPkg: "",
					QueryParams: []query.Param{
						{
							Name: "uuid",
							Type: reflect.TypeOf(uuid.UUID{}),
						},
						{
							Name: "name",
							Type: reflect.TypeOf(""),
						},
						{
							Name: "id",
							Type: reflect.TypeOf(int(0)),
						},
						{
							Name: "role",
							Type: reflect.TypeOf(""),
						},
					},
					PathParams: nil,
				},
				Responses: nil,
			}),
			Want: &openapi3.Operation{
				Parameters: openapi3.Parameters{
					{
						Value: &openapi3.Parameter{
							Name:     "uuid",
							In:       "query",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "UUID",
								Value: &openapi3.Schema{
									Type:   &openapi3.Types{openapi3.TypeString},
									Format: "uuid",
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "name",
							In:       "query",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "string",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeString},
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "id",
							In:       "query",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "int",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeInteger},
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "role",
							In:       "query",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "string",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeString},
								}},
						},
					},
				},
			},
		},
		{
			Name: "Mixed",
			Handler: handlers.NewHandler(echo.Route{}, parser.Handler{
				Doc:  "",
				Name: "Test",
				Pkg:  "test",
				Request: &request.Request{
					BindModel:    "models.QueryParams",
					BindModelPkg: "",
					PathParams: []path.Param{
						{
							Name: "uuid",
							Type: reflect.TypeOf(uuid.UUID{}),
						},
						{
							Name: "name",
							Type: reflect.TypeOf(""),
						},
						{
							Name: "id",
							Type: reflect.TypeOf(int(0)),
						},
						{
							Name: "role",
							Type: reflect.TypeOf(""),
						},
					},
				},
				Responses: nil,
			}),
			Want: &openapi3.Operation{
				Parameters: openapi3.Parameters{
					{
						Value: &openapi3.Parameter{
							Name:     "uuid",
							In:       "query",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "UUID",
								Value: &openapi3.Schema{
									Type:   &openapi3.Types{openapi3.TypeString},
									Format: "uuid",
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "name",
							In:       "query",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "string",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeString},
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name: "id",
							In:   "query",
							Schema: &openapi3.SchemaRef{
								Ref: "int",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeInteger},
								}},
						},
					},
					{
						Value: &openapi3.Parameter{
							Name:     "role",
							In:       "query",
							Required: true,
							Schema: &openapi3.SchemaRef{
								Ref: "UserRole",
								Value: &openapi3.Schema{
									Type: &openapi3.Types{openapi3.TypeString},
								}},
						},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			op := openapi3.NewOperation()
			AddQueryParams(op, tc.Handler, g, registry)

			got, err := op.MarshalJSON()
			require.NoError(t, err)
			fmt.Println(string(got))

			want, err := tc.Want.MarshalJSON()
			require.NoError(t, err)
			fmt.Println(string(want))

			require.Equal(t, string(want), string(got))
		})
	}
}
