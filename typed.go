package typed

import (
	"github.com/d1vbyz3r0/typed/generator"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/labstack/echo/v4"
	"reflect"
	"runtime"
	"strings"
)

func IsJwtMiddleware(mw echo.MiddlewareFunc) bool {
	funcName := runtime.FuncForPC(reflect.ValueOf(mw).Pointer()).Name()
	return strings.Contains(funcName, "github.com/labstack/echo-jwt")
}

func AddPathParams(op *openapi3.Operation, r *generator.RouteInfo) {
	for _, p := range r.PathParams {
		if p == nil {
			continue
		}

		param := openapi3.NewPathParameter(p.Name)
		param.Required = p.Required
		op.AddParameter(param)
	}
}

func AddQueryParams(op *openapi3.Operation, r *generator.RouteInfo) {
	for _, p := range r.Handler.QueryParams {
		if p == nil {
			continue
		}

		var ptype string
		switch p.Type {
		case "string":
			ptype = openapi3.TypeString

		case "array":
			ptype = openapi3.TypeArray

		case "int", "int32", "int64", "uint", "uint32", "uint16", "uint64":
			ptype = openapi3.TypeInteger

		case "float32", "float64":
			ptype = openapi3.TypeNumber

		case "bool":
			ptype = openapi3.TypeBoolean

		default:
			ptype = openapi3.TypeString
		}

		param := openapi3.NewQueryParameter(p.Name)
		param.Required = p.Required
		param.Schema = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:     &openapi3.Types{ptype},
				Nullable: !p.Required,
			},
		}
		op.AddParameter(param)
	}
}

func AddRequestBody(
	op *openapi3.Operation,
	r *generator.RouteInfo,
	openapiGen *openapi3gen.Generator,
	registry map[string]any,
) {
	dto := r.Handler.RequestDTO
	if dto == nil {
		return
	}

	body := openapi3.NewRequestBody()

	if dto.IsForm {
		body.Content = map[string]*openapi3.MediaType{
			dto.ContentType: {
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:       &openapi3.Types{openapi3.TypeObject},
						Properties: make(map[string]*openapi3.SchemaRef),
					},
				},
			},
		}

		for _, f := range dto.FormFields {
			if f.IsFile {
				var (
					items *openapi3.SchemaRef
					ftype = openapi3.TypeString
				)

				if f.IsArray {
					ftype = openapi3.TypeArray
					items = &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:   &openapi3.Types{openapi3.TypeString},
							Format: "binary",
						},
					}
				}

				body.Content[dto.ContentType].Schema.Value.Properties[f.Name] = &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:  &openapi3.Types{ftype},
						Items: items,
					},
				}
			} else {
				body.Content[dto.ContentType].Schema.Value.Properties[f.Name] = &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type:   &openapi3.Types{openapi3.TypeString},
						Format: f.Type,
					},
				}
			}
		}
	} else {
		bodyType := reflect.TypeOf(registry[dto.TypeName])
		schemaRef := openapiGen.Types[bodyType]
		body.Content = map[string]*openapi3.MediaType{
			dto.ContentType: {
				Schema: schemaRef,
			},
		}
	}

	op.RequestBody = &openapi3.RequestBodyRef{
		Value: body,
	}
}

func AddResponses(
	op *openapi3.Operation,
	r *generator.RouteInfo,
	openapiGen *openapi3gen.Generator,
	registry map[string]any,
) {
	for status, resp := range r.Handler.Responses {
		if resp == nil {
			continue
		}

		respType := reflect.TypeOf(registry[resp.TypeName])
		schemaRef := openapiGen.Types[respType]

		var response *openapi3.Response
		if resp.IsArray {
			response = &openapi3.Response{
				Content: map[string]*openapi3.MediaType{
					resp.ContentType: {
						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type:  &openapi3.Types{openapi3.TypeArray},
								Items: schemaRef,
							},
						},
					},
				},
				Description: &resp.TypeName,
			}

		} else {
			response = &openapi3.Response{
				Content: map[string]*openapi3.MediaType{
					resp.ContentType: {
						Schema: schemaRef,
					},
				},
				Description: &resp.TypeName,
			}
		}

		op.AddResponse(status, response)
	}
}
