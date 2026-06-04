package generator

import (
	"testing"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser"
	"github.com/d1vbyz3r0/typed/internal/parser/request"
	"github.com/d1vbyz3r0/typed/internal/parser/response"
	"github.com/stretchr/testify/require"
)

func Test_collectTypes(t *testing.T) {
	tests := []struct {
		name    string
		want    []*typing.Type
		results []parser.Result
	}{
		{
			name: "basic type in request and response",
			want: []*typing.Type{
				typing.Basic("string"),
				typing.Basic("int"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Request: &request.Request{
						ModelType: typing.Basic("string"),
					},
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Basic("int"),
						}},
					},
				}},
			}},
		},
		{
			name: "map of basic types in response",
			want: []*typing.Type{
				typing.Map(typing.Basic("string"), typing.Basic("string")),
				typing.Basic("string"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Map(typing.Basic("string"), typing.Basic("string")),
						}},
					},
				}},
			}},
		},
		{
			name: "slice of named types in response",
			want: []*typing.Type{
				typing.Slice(typing.Named("github.com/example/foo", "Model")),
				typing.Named("github.com/example/foo", "Model"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Slice(typing.Named("github.com/example/foo", "Model")),
						}},
					},
				}},
			}},
		},
		{
			name: "generic type in response",
			want: []*typing.Type{
				typing.Named("github.com/example/foo", "Model", typing.Basic("string"), typing.Basic("int")),
				typing.Basic("string"),
				typing.Basic("int"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Named("github.com/example/foo", "Model", typing.Basic("string"), typing.Basic("int")),
						}},
					},
				}},
			}},
		},
		{
			name: "array of generic types in response",
			want: []*typing.Type{
				typing.Array(typing.Named("github.com/example/foo", "Model", typing.Basic("string"), typing.Basic("int")), 10),
				typing.Named("github.com/example/foo", "Model", typing.Basic("string"), typing.Basic("int")),
				typing.Basic("string"), typing.Basic("int"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Array(typing.Named("github.com/example/foo", "Model", typing.Basic("string"), typing.Basic("int")), 10),
						}},
					},
				}},
			}},
		},
		{
			name: "enum type and same named type leaves enum only",
			want: []*typing.Type{
				typing.Enum(typing.Named("github.com/example/foo", "Enum"), []any{"1", "2", "3"}),
			},
			results: []parser.Result{{
				AdditionalModels: []*typing.Type{
					typing.Named("github.com/example/foo", "Enum"),
					typing.Enum(typing.Named("github.com/example/foo", "Enum"), []any{"1", "2", "3"}),
				},
			}},
		},
		{
			name: "map with named key and named value",
			want: []*typing.Type{
				typing.Map(
					typing.Named("github.com/example/foo", "Key"),
					typing.Named("github.com/example/bar", "Value"),
				),
				typing.Named("github.com/example/foo", "Key"),
				typing.Named("github.com/example/bar", "Value"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Map(
								typing.Named("github.com/example/foo", "Key"),
								typing.Named("github.com/example/bar", "Value"),
							),
						}},
					},
				}},
			}},
		},
		{
			name: "generic params contain imported named types",
			want: []*typing.Type{
				typing.Named(
					"github.com/example/foo",
					"Page",
					typing.Named("github.com/example/bar", "Item"),
					typing.Slice(typing.Named("github.com/example/baz", "Meta")),
				),
				typing.Named("github.com/example/bar", "Item"),
				typing.Slice(typing.Named("github.com/example/baz", "Meta")),
				typing.Named("github.com/example/baz", "Meta"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Named(
								"github.com/example/foo",
								"Page",
								typing.Named("github.com/example/bar", "Item"),
								typing.Slice(typing.Named("github.com/example/baz", "Meta")),
							),
						}},
					},
				}},
			}},
		},
		{
			name: "nested containers",
			want: []*typing.Type{
				typing.Slice(typing.Map(
					typing.Basic("string"),
					typing.Slice(typing.Named("github.com/example/foo", "Model")),
				)),
				typing.Map(
					typing.Basic("string"),
					typing.Slice(typing.Named("github.com/example/foo", "Model")),
				),
				typing.Slice(typing.Named("github.com/example/foo", "Model")),
				typing.Named("github.com/example/foo", "Model"),
				typing.Basic("string"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Slice(typing.Map(
								typing.Basic("string"),
								typing.Slice(typing.Named("github.com/example/foo", "Model")),
							)),
						}},
					},
				}},
			}},
		},
		{
			name: "request model is collected",
			want: []*typing.Type{
				typing.Named("github.com/example/foo", "CreateRequest"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Request: &request.Request{
						ModelType: typing.Named("github.com/example/foo", "CreateRequest"),
					},
				}},
			}},
		},
		{
			name: "additional models are collected",
			want: []*typing.Type{
				typing.Named("github.com/example/foo", "Additional"),
			},
			results: []parser.Result{{
				AdditionalModels: []*typing.Type{
					typing.Named("github.com/example/foo", "Additional"),
				},
			}},
		},
		{
			name: "duplicate named types are collected once",
			want: []*typing.Type{
				typing.Named("github.com/example/foo", "Model"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Request: &request.Request{
						ModelType: typing.Named("github.com/example/foo", "Model"),
					},
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Named("github.com/example/foo", "Model"),
						}},
					},
				}},
				AdditionalModels: []*typing.Type{
					typing.Named("github.com/example/foo", "Model"),
				},
			}},
		},
		{
			name: "multiple status codes and responses are collected",
			want: []*typing.Type{
				typing.Named("github.com/example/foo", "OK"),
				typing.Named("github.com/example/foo", "BadRequest"),
				typing.Named("github.com/example/foo", "InternalError"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Named("github.com/example/foo", "OK"),
						}},
						400: []response.Response{{
							ModelType: typing.Named("github.com/example/foo", "BadRequest"),
						}},
						500: []response.Response{{
							ModelType: typing.Named("github.com/example/foo", "InternalError"),
						}},
					},
				}},
			}},
		},
		{
			name: "pointer to named type in response",
			want: []*typing.Type{
				typing.Pointer(typing.Named("github.com/example/foo", "Model")),
				typing.Named("github.com/example/foo", "Model"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Pointer(typing.Named("github.com/example/foo", "Model")),
						}},
					},
				}},
			}},
		},
		{
			name: "slice of pointers to named types in response",
			want: []*typing.Type{
				typing.Slice(typing.Pointer(typing.Named("github.com/example/foo", "Model"))),
				typing.Pointer(typing.Named("github.com/example/foo", "Model")),
				typing.Named("github.com/example/foo", "Model"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Slice(typing.Pointer(typing.Named("github.com/example/foo", "Model"))),
						}},
					},
				}},
			}},
		},
		{
			name: "generic type with pointer argument",
			want: []*typing.Type{
				typing.Named(
					"github.com/example/foo",
					"Box",
					typing.Pointer(typing.Named("github.com/example/bar", "Item")),
				),
				typing.Pointer(typing.Named("github.com/example/bar", "Item")),
				typing.Named("github.com/example/bar", "Item"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Named(
								"github.com/example/foo",
								"Box",
								typing.Pointer(typing.Named("github.com/example/bar", "Item")),
							),
						}},
					},
				}},
			}},
		},
		{
			name: "map value is pointer to named type",
			want: []*typing.Type{
				typing.Map(
					typing.Basic("string"),
					typing.Pointer(typing.Named("github.com/example/foo", "Model")),
				),
				typing.Pointer(typing.Named("github.com/example/foo", "Model")),
				typing.Named("github.com/example/foo", "Model"),
				typing.Basic("string"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Map(
								typing.Basic("string"),
								typing.Pointer(typing.Named("github.com/example/foo", "Model")),
							),
						}},
					},
				}},
			}},
		},
		{
			name: "deep nested pointer containers",
			want: []*typing.Type{
				typing.Pointer(typing.Slice(typing.Map(
					typing.Basic("string"),
					typing.Pointer(typing.Named("github.com/example/foo", "Model")),
				))),
				typing.Slice(typing.Map(
					typing.Basic("string"),
					typing.Pointer(typing.Named("github.com/example/foo", "Model")),
				)),
				typing.Map(
					typing.Basic("string"),
					typing.Pointer(typing.Named("github.com/example/foo", "Model")),
				),
				typing.Pointer(typing.Named("github.com/example/foo", "Model")),
				typing.Named("github.com/example/foo", "Model"),
				typing.Basic("string"),
			},
			results: []parser.Result{{
				Handlers: []parser.Handler{{
					Responses: response.StatusCodeMapping{
						200: []response.Response{{
							ModelType: typing.Pointer(typing.Slice(typing.Map(
								typing.Basic("string"),
								typing.Pointer(typing.Named("github.com/example/foo", "Model")),
							))),
						}},
					},
				}},
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := collectTypes(tc.results)
			require.NoError(t, err)
			expectedResult(t, tc.want, got)
		})
	}
}

func expectedResult(t *testing.T, want []*typing.Type, got []*typing.Type) {
	t.Helper()
	wantMap := make(map[string]struct{})
	for _, item := range want {
		wantMap[item.String()] = struct{}{}
	}

	gotMap := make(map[string]struct{})
	for _, item := range got {
		gotMap[item.String()] = struct{}{}
	}

	require.Equal(t, wantMap, gotMap)
}
