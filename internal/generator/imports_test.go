package generator

import (
	"net/http"
	"testing"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser"
	"github.com/d1vbyz3r0/typed/internal/parser/request"
	"github.com/d1vbyz3r0/typed/internal/parser/response"
	"github.com/stretchr/testify/require"
)

func TestCreateImportMappings(t *testing.T) {
	tests := []struct {
		name    string
		results []parser.Result
		want    []*importMapping
	}{
		{
			name: "unique package names do not need aliases",
			results: []parser.Result{
				{
					Handlers: []parser.Handler{
						{
							Name: "CreateUser",
							Pkg:  "github.com/acme/app/api",
							Request: &request.Request{
								ModelType: typing.Named(
									"github.com/acme/api/user",
									"CreateUserRequest",
								),
							},
							Responses: response.StatusCodeMapping{
								http.StatusOK: {
									{
										ModelType: typing.Named(
											"github.com/acme/api/order",
											"OrderResponse",
										),
									},
								},
							},
						},
					},
				},
			},
			want: []*importMapping{
				{
					Alias: "api",
					Pkg:   "github.com/acme/app/api",
				},
				{
					Alias: "user",
					Pkg:   "github.com/acme/api/user",
				},
				{
					Alias: "order",
					Pkg:   "github.com/acme/api/order",
				},
			},
		},
		{
			name: "same package name gets numbered aliases",
			results: []parser.Result{
				{
					Handlers: []parser.Handler{
						{
							Name: "CreateUser",
							Pkg:  "github.com/acme/app/api",
							Request: &request.Request{
								ModelType: typing.Named(
									"github.com/acme/public/model",
									"CreateUserRequest",
								),
							},
							Responses: response.StatusCodeMapping{
								http.StatusOK: {
									{
										ModelType: typing.Named(
											"github.com/acme/internal/model",
											"UserResponse",
										),
									},
								},
							},
						},
					},
				},
			},
			want: []*importMapping{
				{
					Alias: "api",
					Pkg:   "github.com/acme/app/api",
				},
				{
					Alias: "model",
					Pkg:   "github.com/acme/public/model",
				},
				{
					Alias: "model1",
					Pkg:   "github.com/acme/internal/model",
				},
			},
		},
		{
			name: "deduplicates same package used multiple times",
			results: []parser.Result{
				{
					Handlers: []parser.Handler{
						{
							Name: "CreateUser",
							Pkg:  "github.com/acme/app/api",
							Request: &request.Request{
								ModelType: typing.Named(
									"github.com/acme/api/model",
									"CreateUserRequest",
								),
							},
							Responses: response.StatusCodeMapping{
								http.StatusOK: {
									{
										ModelType: typing.Named(
											"github.com/acme/api/model",
											"UserResponse",
										),
									},
								},
								http.StatusBadRequest: {
									{
										ModelType: typing.Named(
											"github.com/acme/api/model",
											"ErrorResponse",
										),
									},
								},
							},
						},
					},
					AdditionalModels: []*typing.Type{
						typing.Named(
							"github.com/acme/api/model",
							"ExtraModel",
						),
					},
				},
			},
			want: []*importMapping{
				{
					Alias: "api",
					Pkg:   "github.com/acme/app/api",
				},
				{
					Alias: "model",
					Pkg:   "github.com/acme/api/model",
				},
			},
		},
		{
			name: "includes additional models",
			results: []parser.Result{
				{
					Handlers: []parser.Handler{
						{
							Name: "CreateUser",
							Pkg:  "github.com/acme/app/api",
						},
					},
					AdditionalModels: []*typing.Type{
						typing.Named(
							"github.com/acme/api/model",
							"ExtraModel",
						),
					},
				},
			},
			want: []*importMapping{
				{
					Alias: "api",
					Pkg:   "github.com/acme/app/api",
				},
				{
					Alias: "model",
					Pkg:   "github.com/acme/api/model",
				},
			},
		},
		{
			name: "ignores zero and basic types",
			results: []parser.Result{
				{
					Handlers: []parser.Handler{
						{
							Name: "CreateUser",
							Pkg:  "github.com/acme/app/api",
							Request: &request.Request{
								ModelType: typing.Basic("string"),
							},
							Responses: response.StatusCodeMapping{
								http.StatusOK: {
									{ModelType: nil},
									{ModelType: typing.Named("", "")},
								},
							},
						},
					},
					AdditionalModels: []*typing.Type{
						nil,
						typing.Named("", ""),
					},
				},
			},
			want: []*importMapping{
				{
					Alias: "api",
					Pkg:   "github.com/acme/app/api",
				},
			},
		},
		{
			name: "unwraps nested model types",
			results: []parser.Result{
				{
					Handlers: []parser.Handler{
						{
							Name: "CreateUser",
							Pkg:  "github.com/acme/app/api",
							Request: &request.Request{
								ModelType: typing.Pointer(
									typing.Named(
										"github.com/acme/api/user",
										"CreateUserRequest",
									),
								),
							},
							Responses: response.StatusCodeMapping{
								http.StatusOK: {
									{
										ModelType: typing.Slice(
											typing.Named(
												"github.com/acme/api/order",
												"OrderResponse",
											),
										),
									},
									{
										ModelType: typing.Map(
											typing.Basic("string"),
											typing.Named(
												"github.com/acme/api/meta",
												"Meta",
											),
										),
									},
								},
							},
						},
					},
				},
			},
			want: []*importMapping{
				{
					Alias: "api",
					Pkg:   "github.com/acme/app/api",
				},
				{
					Alias: "user",
					Pkg:   "github.com/acme/api/user",
				},
				{
					Alias: "order",
					Pkg:   "github.com/acme/api/order",
				},
				{
					Alias: "meta",
					Pkg:   "github.com/acme/api/meta",
				},
			},
		},
		{
			name: "same package basename across handler and model packages",
			results: []parser.Result{
				{
					Handlers: []parser.Handler{
						{
							Name: "CreateUser",
							Pkg:  "github.com/acme/transport/http/dto",
							Request: &request.Request{
								ModelType: typing.Named(
									"github.com/acme/domain/dto",
									"CreateUserRequest",
								),
							},
						},
					},
					AdditionalModels: []*typing.Type{
						typing.Named(
							"github.com/acme/generated/dto",
							"GeneratedUser",
						),
					},
				},
			},
			want: []*importMapping{
				{
					Alias: "dto",
					Pkg:   "github.com/acme/transport/http/dto",
				},
				{
					Alias: "dto1",
					Pkg:   "github.com/acme/domain/dto",
				},
				{
					Alias: "dto2",
					Pkg:   "github.com/acme/generated/dto",
				},
			},
		},
		{
			name: "aliases handler packages with same last path segment",
			results: []parser.Result{
				{
					Handlers: []parser.Handler{
						{
							Name: "PublicCreateUser",
							Pkg:  "github.com/acme/public/handler",
						},
						{
							Name: "AdminCreateUser",
							Pkg:  "github.com/acme/admin/handler",
						},
					},
				},
			},
			want: []*importMapping{
				{
					Alias: "handler",
					Pkg:   "github.com/acme/public/handler",
				},
				{
					Alias: "handler1",
					Pkg:   "github.com/acme/admin/handler",
				},
			},
		},
		{
			name: "deduplicates same full handler package path",
			results: []parser.Result{
				{
					Handlers: []parser.Handler{
						{
							Name: "CreateUser",
							Pkg:  "github.com/acme/service/handler",
						},
						{
							Name: "DeleteUser",
							Pkg:  "github.com/acme/service/handler",
						},
					},
				},
			},
			want: []*importMapping{
				{
					Alias: "handler",
					Pkg:   "github.com/acme/service/handler",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createImportMappings(tt.results, nil)
			require.NoError(t, err)
			require.ElementsMatch(t, stripIdx(tt.want), stripIdx(got))
		})
	}
}

func stripIdx(items []*importMapping) []any {
	type item struct {
		Alias string
		Pkg   string
	}

	res := make([]any, 0, len(items))
	for _, v := range items {
		res = append(res, item{
			Alias: v.Alias,
			Pkg:   v.Pkg,
		})
	}

	return res
}
