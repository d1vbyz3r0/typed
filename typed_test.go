package typed

import (
	"testing"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestGenerateOptionsSetDefaults(t *testing.T) {
	registry := MustNewRegistry(T{
		Val:  new(string),
		Type: typing.Basic("string"),
	})
	opts := GenerateOptions{
		Spec:     &openapi3.T{},
		Registry: registry,
	}

	require.NoError(t, opts.setDefaults())
	require.NotNil(t, opts.Generator)
	require.NotNil(t, opts.Spec.Components)
	require.NotNil(t, opts.Spec.Components.Schemas)
	require.NotNil(t, opts.Spec.Components.SecuritySchemes)
	require.Positive(t, opts.Concurrency)
}

func TestGenerateOptionsSetDefaultsRequiresInputs(t *testing.T) {
	tests := []struct {
		name    string
		opts    GenerateOptions
		wantErr string
	}{
		{
			name:    "registry",
			opts:    GenerateOptions{Spec: &openapi3.T{}},
			wantErr: "registry is required",
		},
		{
			name: "spec",
			opts: GenerateOptions{
				Registry: MustNewRegistry(T{
					Val:  new(string),
					Type: typing.Basic("string"),
				}),
			},
			wantErr: "spec is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.EqualError(t, tt.opts.setDefaults(), tt.wantErr)
		})
	}
}

func TestGenerateWithDefaults(t *testing.T) {
	registry := MustNewRegistry(T{
		Val:  new(string),
		Type: typing.Basic("string"),
	})
	spec := &openapi3.T{}

	require.NoError(t, Generate(GenerateOptions{
		Spec:     spec,
		Registry: registry,
	}))
	require.NotNil(t, spec.Components.Schemas)
}

func TestCollectRoutes(t *testing.T) {
	handler := func(echo.Context) error { return nil }
	middleware := func(next echo.HandlerFunc) echo.HandlerFunc { return next }
	provider := &testRoutesProvider{
		route: echo.Route{
			Method: "GET",
			Path:   "/users",
			Name:   "users",
		},
		handler:    handler,
		middleware: []echo.MiddlewareFunc{middleware},
	}

	routes := CollectRoutes(provider)

	require.Len(t, routes, 1)
	require.Equal(t, provider.route, routes[0].Route)
	require.NotNil(t, routes[0].HandlerFunc)
	require.Equal(t, provider.middleware, routes[0].Middlewares)
}

type testRoutesProvider struct {
	callback   func(string, echo.Route, echo.HandlerFunc, []echo.MiddlewareFunc)
	route      echo.Route
	handler    echo.HandlerFunc
	middleware []echo.MiddlewareFunc
}

func (p *testRoutesProvider) OnRouteAdded(
	callback func(string, echo.Route, echo.HandlerFunc, []echo.MiddlewareFunc),
) {
	p.callback = callback
}

func (p *testRoutesProvider) ProvideRoutes() {
	p.callback("", p.route, p.handler, p.middleware)
}
