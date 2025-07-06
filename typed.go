package typed

import (
	"fmt"
	"github.com/d1vbyz3r0/typed/handlers"
	"github.com/d1vbyz3r0/typed/internal/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser/request/path"
	"github.com/d1vbyz3r0/typed/internal/parser/request/query"
	"github.com/d1vbyz3r0/typed/internal/parser/response"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"log/slog"
	"reflect"
	"strings"
)

func AddPathParams(
	op *openapi3.Operation,
	h handlers.Handler,
	openapiGen *openapi3gen.Generator,
	registry map[string]any,
) {
	params := make(map[string]path.Param, len(h.PathParams()))

	model := h.BindModel()
	if model != "" {
		obj, ok := registry[model]
		if ok {
			typedParams, err := path.NewStructPathParams(typing.DerefReflectPtr(reflect.TypeOf(obj)))
			if err != nil {
				slog.Error("get path params from bind model", "error", err)
			}

			for _, p := range typedParams {
				param := openapi3.NewPathParameter(p.Name)
				param.Required = true

				schema, err := openapiGen.GenerateSchemaRef(p.Type)
				if err != nil {
					slog.Error("generate schema ref for path param", "param", p.Name, "error", err)
					continue
				}

				param.Schema = &openapi3.SchemaRef{
					Value: schema.Value,
				}
				params[p.Name] = p
				op.AddParameter(param)
			}
		} else {
			slog.Warn("bind model not found in provided registry", "model", model)
		}
	}

	for _, p := range h.PathParams() {
		_, ok := params[p.Name]
		if ok {
			continue
		}

		param := openapi3.NewPathParameter(p.Name)
		param.Required = true
		schema, err := openapiGen.GenerateSchemaRef(p.Type)
		if err != nil {
			slog.Error("generate schema ref for path param", "param", p.Name, "error", err)
			continue
		}

		param.Schema = &openapi3.SchemaRef{
			Value: schema.Value,
		}
		op.AddParameter(param)
	}
}

func AddQueryParams(
	op *openapi3.Operation,
	h handlers.Handler,
	openapiGen *openapi3gen.Generator,
	registry map[string]any,
) {
	params := make(map[string]query.Param, len(h.QueryParams()))

	model := h.BindModel()
	if model != "" {
		obj, ok := registry[model]
		if ok {
			typedParams, err := query.NewStructQueryParams(typing.DerefReflectPtr(reflect.TypeOf(obj)))
			if err != nil {
				slog.Error("get query params from bind model", "error", err)
			}

			for _, p := range typedParams {
				param := openapi3.NewQueryParameter(p.Name)
				param.Required = p.Type.Kind() != reflect.Pointer

				schema, err := openapiGen.GenerateSchemaRef(p.Type)
				if err != nil {
					slog.Error("generate schema ref for query param", "param", p.Name, "error", err)
					continue
				}

				param.Schema = &openapi3.SchemaRef{
					Value: schema.Value,
				}
				params[p.Name] = p
				op.AddParameter(param)
			}
		} else {
			slog.Warn("bind model not found in provided registry", "model", model)
		}
	}

	for _, p := range h.QueryParams() {
		_, ok := params[p.Name]
		if ok {
			continue
		}

		param := openapi3.NewQueryParameter(p.Name)
		param.Required = true

		schema, err := openapiGen.GenerateSchemaRef(p.Type)
		if err != nil {
			slog.Error("generate schema ref for query param", "param", p.Name, "error", err)
			continue
		}

		param.Schema = &openapi3.SchemaRef{
			Value: schema.Value,
		}
		op.AddParameter(param)
	}
}

func AddRequestBody(
	op *openapi3.Operation,
	h handlers.Handler,
	openapiGen *openapi3gen.Generator,
	schemas openapi3.Schemas,
	registry map[string]any,
) {
	content := make(openapi3.Content)
	request := h.Request()

	for contentType, reqBody := range request.ContentTypeMapping {
		if reqBody.Form != nil {
			ref, err := openapiGen.GenerateSchemaRef(reqBody.Form)
			if err != nil {
				slog.Error("generate schema ref for request form", "form", reqBody.Form, "error", err)
				continue
			}

			content[contentType] = openapi3.NewMediaType().WithSchemaRef(ref)
		}

		if request.BindModel != "" {
			obj, ok := registry[request.BindModel]
			if !ok {
				slog.Warn("bind model not found in provided registry", "model", request.BindModel)
				continue
			}

			ref, err := openapiGen.NewSchemaRefForValue(obj, schemas)
			if err != nil {
				slog.Error("generate schema ref for bind model", "model", request.BindModel, "error", err)
				continue
			}

			content[contentType] = openapi3.NewMediaType().WithSchemaRef(ref)
		} else {
			slog.Debug("request contains empty bind model", "handler", h.HandlerName())
		}
	}

	op.RequestBody = &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody().WithContent(content),
	}
}

func AddResponses(
	op *openapi3.Operation,
	statusCodeMapping response.StatusCodeMapping,
	openapiGen *openapi3gen.Generator,
	schemas openapi3.Schemas,
	registry map[string]any,
) {
	for status, responses := range statusCodeMapping {
		r := openapi3.NewResponse()
		content := make(openapi3.Content, len(responses))
		for _, resp := range responses {
			mediaType := openapi3.NewMediaType()
			if resp.TypeName != "" {
				val, ok := registry[resp.TypeName]
				if !ok {
					slog.Warn("type not found in registry", "type", resp.TypeName, "pkg", resp.TypePkgPath)
					continue
				}

				ref, err := openapiGen.NewSchemaRefForValue(val, schemas)
				if err != nil {
					slog.Error("generate ref for value", "type", resp.TypeName, "error", err)
					continue
				}

				mediaType.WithSchemaRef(ref)
			}

			if resp.ContentType != "" {
				content[resp.ContentType] = mediaType
			}
		}

		r.WithContent(content)
		op.AddResponse(status, r)
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

func AddOperationId(op *openapi3.Operation, h handlers.Handler) {
	op.OperationID = h.HandlerName()
}

func extractOpTag(path string, prefix string) (string, error) {
	// /api/v1/tasks -> tasks
	p, found := strings.CutPrefix(path, prefix)
	if !found {
		return "", fmt.Errorf("bad api prefix '%s' for path: '%s'", prefix, path)
	}

	p, _ = strings.CutPrefix(p, "/")
	p = strings.Title(p)

	parts := strings.Split(p, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("bad path or prefix provided, nothing to extract after prefix cutoff: %s", parts)
	}

	return parts[0], nil
}
