package main

import (
	"examples/lib-example/openapi"
	"examples/server"
	"fmt"
	"reflect"

	"github.com/d1vbyz3r0/typed"
	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/handlers"
	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	// in library mode you can customize typed any way you want
	typed.RegisterCustomizer(func(name string, t reflect.Type, tag reflect.StructTag, schema *openapi3.Schema) error {
		fmt.Println("using custom customizer for schema:", name)
		return nil
	})

	typed.RegisterHandlerProcessingHook(func(spec *openapi3.T, operation *openapi3.Operation, handler handlers.Handler) {
		fmt.Println("using custom handler hook for handler: ", handler.HandlerName())
	})

	typing.RegisterTypeProvider(func(pkg, funcName string) (reflect.Type, bool) {
		fmt.Printf("using custom type provider for pkg = %s, funcName = %s\n", pkg, funcName)
		return nil, false
	})

	registry := openapi.GetRegistry()
	spec := openapi.GetSpec()

	routesProvider := server.NewBuilder()
	err := typed.Generate(typed.GenerateOptions{
		Spec:        spec,
		Registry:    registry,
		Concurrency: 2,
		APIPrefix:   typed.MakePointer("/api/v1"),
		Routes:      typed.CollectRoutes(routesProvider), // You can skip implementation of typed.RoutesProvider, but in that case you should collect routes and middlewares manually somehow
		SearchPatterns: []handlers.SearchPattern{
			{
				Path:      ".",
				Recursive: true,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	// NOTE: path is still relative to directory when generator was called
	err = typed.SaveSpec(spec, "../lib-example/spec.yaml")
	if err != nil {
		panic(err)
	}
}
