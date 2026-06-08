package typed

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/handlers"
	"github.com/d1vbyz3r0/typed/internal/parser/headers"
	"github.com/d1vbyz3r0/typed/internal/parser/request/path"
	"github.com/d1vbyz3r0/typed/internal/parser/request/query"
	"github.com/d1vbyz3r0/typed/logging"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

type OperationBuilder struct {
	op        *openapi3.Operation
	handler   handlers.Handler
	generator *openapi3gen.Generator
	registry  *Registry
	err       error
}

func NewOperationBuilder(
	g *openapi3gen.Generator,
	h handlers.Handler,
	reg *Registry,
) *OperationBuilder {
	return &OperationBuilder{
		op:        openapi3.NewOperation(),
		handler:   h,
		generator: g,
		registry:  reg,
	}
}

func (b *OperationBuilder) AddPathParams() *OperationBuilder {
	b.step("add path params", func() error {
		params := make(map[string]path.Param, len(b.handler.PathParams()))
		model := b.handler.BindModel()
		if model != nil {
			obj, ok := b.registry.LookupValue(model)
			if !ok {
				return fmt.Errorf("bind model not found in registry: %s", model)
			}

			typedParams, err := path.NewStructPathParams(typing.DerefReflectPtr(reflect.TypeOf(obj)))
			if err != nil {
				return fmt.Errorf("failed to extract path params from bind model: %w", err)
			}

			for _, p := range typedParams {
				// TODO: process required more precise
				param := openapi3.NewPathParameter(p.Name).WithRequired(true)
				schema, err := b.generator.GenerateSchemaRef(p.Type)
				if err != nil {
					return fmt.Errorf("failed to generate schema ref for param %s: %w", p.Name, err)
				}

				param.Schema = &openapi3.SchemaRef{
					Value: schema.Value,
				}
				params[p.Name] = p
				b.op.AddParameter(param)
			}
		}

		for _, p := range b.handler.PathParams() {
			_, ok := params[p.Name]
			if ok {
				// parameter was declared in struct, used with echo.Bind
				continue
			}

			param := openapi3.NewPathParameter(p.Name).WithRequired(true)
			schema, err := b.generator.GenerateSchemaRef(p.Type)
			if err != nil {
				return fmt.Errorf("failed to generate schema ref for param %s: %w", p.Name, err)
			}

			param.Schema = &openapi3.SchemaRef{
				Value: schema.Value,
			}
			b.op.AddParameter(param)
		}

		return nil
	})

	return b
}

func (b *OperationBuilder) AddQueryParams() *OperationBuilder {
	b.step("add query params", func() error {
		params := make(map[string]query.Param, len(b.handler.QueryParams()))
		model := b.handler.BindModel()
		if model != nil {
			obj, ok := b.registry.LookupValue(model)
			if !ok {
				return fmt.Errorf("bind model not found in registry: %s", model)
			}

			typedParams, err := query.NewStructQueryParams(typing.DerefReflectPtr(reflect.TypeOf(obj)))
			if err != nil {
				return fmt.Errorf("failed to extract query params from bind model: %w", err)
			}

			for _, p := range typedParams {
				isRequired := p.Type.Kind() != reflect.Pointer
				param := openapi3.NewQueryParameter(p.Name).WithRequired(isRequired)
				schema, err := b.generator.GenerateSchemaRef(p.Type)
				if err != nil {
					return fmt.Errorf("failed to generate schema ref for param %s: %w", p.Name, err)
				}

				param.Schema = &openapi3.SchemaRef{
					Value: schema.Value,
				}
				params[p.Name] = p
				b.op.AddParameter(param)
			}
		}

		for _, p := range b.handler.QueryParams() {
			_, ok := params[p.Name]
			if ok {
				// parameter was declared in struct, used with echo.Bind
				continue
			}

			param := openapi3.NewQueryParameter(p.Name).WithRequired(true)
			schema, err := b.generator.GenerateSchemaRef(p.Type)
			if err != nil {
				return fmt.Errorf("failed to generate schema ref for param %s: %w", p.Name, err)
			}

			param.Schema = &openapi3.SchemaRef{
				Value: schema.Value,
			}
			b.op.AddParameter(param)
		}

		return nil
	})

	return b
}

func (b *OperationBuilder) AddRequestBody(schemas openapi3.Schemas) *OperationBuilder {
	b.step("add request body", func() error {
		content := make(openapi3.Content)
		request := b.handler.Request()
		for contentType, reqBody := range request.ContentTypeMapping {
			if reqBody.Form != nil {
				ref, err := b.generator.GenerateSchemaRef(reqBody.Form)
				if err != nil {
					return fmt.Errorf("failed to generate schema ref for form: %w", err)
				}
				content[contentType] = openapi3.NewMediaType().WithSchemaRef(ref)
			}

			if request.ModelType == nil {
				logging.Debug("request contains empty bind model", "handler", b.handler.HandlerName())
				continue
			}

			obj, ok := b.registry.LookupValue(request.ModelType)
			if !ok {
				return fmt.Errorf("bind model not found in registry: %s", request.ModelType)
			}

			ref, err := b.generator.NewSchemaRefForValue(obj, schemas)
			if err != nil {
				return fmt.Errorf("failed to generate schema ref for bind model %s: %w", request.ModelType, err)
			}
			content[contentType] = openapi3.NewMediaType().WithSchemaRef(ref)
		}

		if len(content) > 0 {
			b.op.RequestBody = &openapi3.RequestBodyRef{
				Value: openapi3.NewRequestBody().WithContent(content),
			}
		}
		return nil
	})

	return b
}

func (b *OperationBuilder) AddResponses(schemas openapi3.Schemas) *OperationBuilder {
	b.step("add responses", func() error {
		statusCodeMapping := b.handler.Responses()
		if len(statusCodeMapping) == 0 {
			// TODO: better solution ?
			b.op.AddResponse(0, &openapi3.Response{Description: new(string)})
			return nil
		}

		for status, responses := range statusCodeMapping {
			content := make(openapi3.Content, len(responses))
			mergedHeaders := make([]headers.Header, 0, len(responses))

			for _, resp := range responses {
				mediaType := openapi3.NewMediaType()
				if resp.ModelType != nil {
					val, ok := b.registry.LookupValue(resp.ModelType)
					if !ok {
						return fmt.Errorf("response model type not found in registry: %s", resp.ModelType)
					}

					ref, err := b.generator.NewSchemaRefForValue(val, schemas)
					if err != nil {
						return fmt.Errorf("failed to generate schema ref for response model %s: %w", resp.ModelType, err)
					}

					mediaType = mediaType.WithSchemaRef(ref)
					mergedHeaders = append(mergedHeaders, resp.Headers...)
				}

				if resp.ContentType != "" {
					content[resp.ContentType] = mediaType
				}
			}

			resp := openapi3.
				NewResponse().
				WithContent(content).
				WithDescription(http.StatusText(status))

			resp.Headers = make(openapi3.Headers, len(mergedHeaders))
			for _, header := range mergedHeaders {
				schema, err := b.generator.GenerateSchemaRef(header.Type)
				if err != nil {
					return fmt.Errorf("failed to generate schema ref for response header %s: %w", header.Name, err)
				}

				resp.Headers[header.Name] = &openapi3.HeaderRef{
					Value: &openapi3.Header{
						Parameter: openapi3.Parameter{
							Required: header.Required,
							Schema: &openapi3.SchemaRef{
								Value: schema.Value,
							},
						},
					},
				}
			}

			b.op.AddResponse(status, resp)
		}
		return nil
	})

	return b
}

func (b *OperationBuilder) AddHeaders() *OperationBuilder {
	b.step("add headers", func() error {
		model := b.handler.BindModel()
		request := b.handler.Request()
		params := make(map[string]struct{}, len(request.Headers))
		if model != nil {
			obj, ok := b.registry.LookupValue(model)
			if !ok {
				return fmt.Errorf("bind model type not found in registry: %s", model)
			}

			typedParams, err := headers.NewStructRequestHeaders(typing.DerefReflectPtr(reflect.TypeOf(obj)))
			if err != nil {
				return fmt.Errorf("failed to extract header params from bind model: %w", err)
			}

			for _, p := range typedParams {
				param := openapi3.NewHeaderParameter(p.Name).WithRequired(p.Required)
				schema, err := b.generator.GenerateSchemaRef(p.Type)
				if err != nil {
					return fmt.Errorf("failed to generate schema ref for header param %s: %w", p.Name, err)
				}

				param.Schema = &openapi3.SchemaRef{
					Value: schema.Value,
				}
				params[p.Name] = struct{}{}
				b.op.AddParameter(param)
			}
		}

		for _, header := range request.Headers {
			_, ok := params[header.Name]
			if ok {
				// parameter was declared in struct, used with echo.Bind
				continue
			}

			param := openapi3.NewHeaderParameter(header.Name).WithRequired(header.Required)
			schema, err := b.generator.GenerateSchemaRef(header.Type)
			if err != nil {
				return fmt.Errorf("failed to generate schema ref for header param %s: %w", header.Name, err)
			}

			param.Schema = &openapi3.SchemaRef{
				Value: schema.Value,
			}
			b.op.AddParameter(param)
		}

		return nil
	})

	return b
}

func (b *OperationBuilder) AddOperationTag(apiPrefix string) *OperationBuilder {
	b.step("add operation tag", func() error {
		tag, err := extractOpTag(b.handler.Path(), apiPrefix)
		if err != nil {
			return fmt.Errorf("extract operation tag: %w", err)
		}
		b.op.Tags = append(b.op.Tags, tag)
		return nil
	})
	return b
}

func (b *OperationBuilder) AddOperationId() *OperationBuilder {
	b.step("add operation id", func() error {
		b.op.OperationID = b.handler.HandlerName()
		return nil
	})
	return b
}

func (b *OperationBuilder) AddOperationDescription() *OperationBuilder {
	b.step("add operation description", func() error {
		b.op.Description = b.handler.Description()
		return nil
	})
	return b
}

func (b *OperationBuilder) Build() (*openapi3.Operation, error) {
	return b.op, b.err
}

func (b *OperationBuilder) step(name string, fn func() error) {
	if b.err != nil {
		return
	}

	if err := fn(); err != nil {
		b.err = fmt.Errorf("%s: %w", name, err)
		return
	}
}
