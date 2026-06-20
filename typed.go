package typed

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/d1vbyz3r0/typed/handlers"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/labstack/echo/v4"
)

// GenerateOptions configures OpenAPI generation from parsed handlers and
// routes registered at runtime.
type GenerateOptions struct {
	Generator      *openapi3gen.Generator
	Spec           *openapi3.T
	Registry       *Registry
	Routes         []handlers.EchoRoute
	SearchPatterns []handlers.SearchPattern
	APIPrefix      *string
	Concurrency    int
}

func (o *GenerateOptions) setDefaults() error {
	if o.Registry == nil {
		return errors.New("registry is required")
	}

	if o.Spec == nil {
		return errors.New("spec is required")
	}

	if o.Generator == nil {
		o.Generator = NewGenerator(o.Registry)
	}

	if o.Spec.Components == nil {
		o.Spec.Components = &openapi3.Components{}
	}

	if o.Spec.Components.Schemas == nil {
		o.Spec.Components.Schemas = make(openapi3.Schemas)
	}

	if o.Spec.Components.SecuritySchemes == nil {
		o.Spec.Components.SecuritySchemes = make(openapi3.SecuritySchemes)
	}

	if o.Concurrency <= 0 {
		o.Concurrency = runtime.GOMAXPROCS(0)
	}

	return nil
}

// Generate builds schemas and operations into opts.Spec.
func Generate(opts GenerateOptions) error {
	if err := opts.setDefaults(); err != nil {
		return err
	}

	if err := GenerateRefs(opts.Generator, opts.Spec.Components.Schemas, opts.Registry); err != nil {
		return fmt.Errorf("generate refs: %w", err)
	}

	finder, err := handlers.NewFinder()
	if err != nil {
		return fmt.Errorf("create handlers finder: %w", err)
	}

	err = finder.Find(opts.SearchPatterns, handlers.WithConcurrency(opts.Concurrency))
	if err != nil {
		return fmt.Errorf("run finder: %w", err)
	}

	matchedHandlers := finder.Match(opts.Routes)
	for _, handler := range matchedHandlers {
		b := NewOperationBuilder(
			opts.Generator,
			handler,
			opts.Registry,
		).
			AddPathParams().
			AddQueryParams().
			AddRequestBody(opts.Spec.Components.Schemas).
			AddResponses(opts.Spec.Components.Schemas).
			AddHeaders().
			AddOperationId().
			AddOperationDescription()

		if opts.APIPrefix != nil {
			b = b.AddOperationTag(*opts.APIPrefix)
		}

		op, err := b.Build()
		if err != nil {
			return fmt.Errorf("build operation %s: %w", handler.HandlerName(), err)
		}

		// TODO: move up from global state
		RunHandlerHooks(opts.Spec, op, handler)
		opts.Spec.AddOperation(handler.Path(), handler.Method(), op)
	}

	return nil
}

// CollectRoutes captures routes registered by provider.
func CollectRoutes(provider RoutesProvider) []handlers.EchoRoute {
	var routes []handlers.EchoRoute
	provider.OnRouteAdded(func(
		_ string,
		route echo.Route,
		handler echo.HandlerFunc,
		middleware []echo.MiddlewareFunc,
	) {
		routes = append(routes, handlers.EchoRoute{
			Route:       route,
			HandlerFunc: handler,
			Middlewares: middleware,
		})
	})
	provider.ProvideRoutes()
	return routes
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

// GenerateRefs generates schema references for all types in registry. It is
// useful for types that are not directly used, such as a model decoded from
// json.RawMessage.
func GenerateRefs(g *openapi3gen.Generator, schemas openapi3.Schemas, registry *Registry) error {
	for instance := range registry.Values() {
		_, err := g.NewSchemaRefForValue(instance, schemas)
		if err != nil {
			return fmt.Errorf("new schema ref for value: %w", err)
		}
	}
	return nil
}
