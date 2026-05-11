package typed

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/d1vbyz3r0/typed/handlers"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

type HandlerProcessingHookFn func(spec *openapi3.T, operation *openapi3.Operation, handler handlers.Handler)

var handlerProcessingHooks []HandlerProcessingHookFn

func RegisterHandlerProcessingHook(hook HandlerProcessingHookFn) {
	handlerProcessingHooks = append(handlerProcessingHooks, hook)
}

func RunHandlerHooks(spec *openapi3.T, operation *openapi3.Operation, handler handlers.Handler) {
	for _, hook := range handlerProcessingHooks {
		hook(spec, operation, handler)
	}
}

func GetMiddlewareFuncName(mw echo.MiddlewareFunc) string {
	return runtime.FuncForPC(reflect.ValueOf(mw).Pointer()).Name()
}

func EchoJWTMiddlewareHook(spec *openapi3.T, operation *openapi3.Operation, handler handlers.Handler) {
	const bearerAuthScheme = "bearerAuthScheme"
	if _, ok := spec.Components.Schemas[bearerAuthScheme]; !ok {
		spec.Components.SecuritySchemes[bearerAuthScheme] = &openapi3.SecuritySchemeRef{
			Value: &openapi3.SecurityScheme{
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			},
		}
	}

	const echoJwtMiddlewarePkg = "github.com/labstack/echo-jwt"
	for _, mw := range handler.Middlewares() {
		if strings.Contains(GetMiddlewareFuncName(mw), echoJwtMiddlewarePkg) {
			if operation.Security == nil {
				operation.Security = openapi3.NewSecurityRequirements()
			}

			operation.Security.With(openapi3.SecurityRequirement{
				bearerAuthScheme: []string{},
			})
		}
	}
}
