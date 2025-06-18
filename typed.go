package typed

import (
	"encoding/json"
	"fmt"
	"github.com/d1vbyz3r0/typed/generator"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

type SpecFormat string

const (
	UndefinedFormat = SpecFormat("")
	YamlFormat      = SpecFormat("yaml")
	JsonFormat      = SpecFormat("json")
)

var NoContent = "No content"

func getSpecFormat(path string) SpecFormat {
	ext := filepath.Ext(path)
	switch ext {
	case ".yaml", ".yml":
		return YamlFormat

	case ".json":
		return JsonFormat

	default:
		return UndefinedFormat
	}
}

func SaveSpec(spec *openapi3.T, outPath string) error {
	f, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	format := getSpecFormat(outPath)
	switch format {
	case YamlFormat:
		enc := yaml.NewEncoder(f)
		enc.SetIndent(2)
		err := enc.Encode(spec)
		if err != nil {
			return fmt.Errorf("encode spec: %w", err)
		}

	case JsonFormat:
		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		err := enc.Encode(spec)
		if err != nil {
			return fmt.Errorf("encode spec: %w", err)
		}

	case UndefinedFormat:
		return fmt.Errorf("can't define spec format basing on path, check extension: %s", outPath)
	}

	return nil
}

func NormalizePathParams(path string) string {
	segments := strings.Split(path, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			trimmed := strings.TrimPrefix(segment, ":")
			segments[i] = "{" + trimmed + "}"
		}
	}
	return strings.Join(segments, "/")
}

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
		param.Schema = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:     &openapi3.Types{toOpenApiType(p.Type)},
				Nullable: false,
			},
		}
		op.AddParameter(param)
	}
}

func AddQueryParams(op *openapi3.Operation, r *generator.RouteInfo) {
	for _, p := range r.Handler.QueryParams {
		if p == nil {
			continue
		}

		param := openapi3.NewQueryParameter(p.Name)
		param.Required = p.Required
		param.Schema = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:     &openapi3.Types{toOpenApiType(p.Type)},
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

		} else if resp.IsMap {
			var hasType *bool
			if resp.ValueType == "any" {
				hasType = new(bool)
			} else {
				hasType = new(bool)
				*hasType = true
			}

			response = &openapi3.Response{
				Content: map[string]*openapi3.MediaType{
					resp.ContentType: {
						Schema: &openapi3.SchemaRef{
							Value: &openapi3.Schema{
								Type: &openapi3.Types{openapi3.TypeObject},
								AdditionalProperties: openapi3.AdditionalProperties{
									Has: hasType,
									Schema: &openapi3.SchemaRef{
										Value: &openapi3.Schema{
											Type:  &openapi3.Types{openapi3.TypeObject},
											Items: schemaRef,
										},
									},
								},
							},
						},
					},
				},
			}

		} else {
			if resp.NoContent {
				response = &openapi3.Response{
					Content:     map[string]*openapi3.MediaType{},
					Description: &NoContent,
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
		}

		op.AddResponse(status, response)
	}
}

func TagOperation(op *openapi3.Operation, path string, apiPrefix string) error {
	tag, err := extractOpTag(path, apiPrefix)
	if err != nil {
		return fmt.Errorf("extract operation tag: %w", err)
	}

	op.Tags = append(op.Tags, tag)
	return nil
}

func AddOperationId(op *openapi3.Operation, r *generator.RouteInfo) {
	if op != nil {
		op.OperationID = r.Handler.Name
	}
}

func extractOpTag(path string, prefix string) (string, error) {
	// /api/v1/tasks -> tasks
	path, found := strings.CutPrefix(path, prefix)
	if !found {
		return "", fmt.Errorf("bad api prefix '%s' for path: '%s'", prefix, path)
	}

	path, _ = strings.CutPrefix(path, "/")
	path = strings.Title(path)

	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("bad path or prefix provided, nothing to extract after prefix cutoff: %s", parts)
	}

	return parts[0], nil
}

func toOpenApiType(t string) string {
	var ptype string
	switch t {
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

	return ptype
}
