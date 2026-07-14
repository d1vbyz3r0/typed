# typed

[![Go](https://github.com/d1vbyz3r0/typed/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/d1vbyz3r0/typed/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/d1vbyz3r0/typed.svg)](https://pkg.go.dev/github.com/d1vbyz3r0/typed)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

`typed` generates an OpenAPI 3.0 document from an Echo application's Go
source and registered routes.

The generator parses handler source with `go/ast`, generates a Go program
containing a runtime type registry, and runs that program to build the
OpenAPI document with reflection. It does not require annotation comments.
The generated specification reflects the supported source-code patterns,
but behavior implemented through unsupported or dynamic constructs may require some tweaks on user side.

See the generated [example specification](./examples/standalone-example/spec.yaml) or
open it in [Swagger UI](https://petstore.swagger.io/?url=https://raw.githubusercontent.com/d1vbyz3r0/typed/refs/heads/master/examples/standalone-example/spec.yaml).

## Requirements

- Go 1.25 or newer
- Echo v4
- an application component that can register routes without starting the
  server and implements `typed.RoutesProvider` (or you have a script, that will collect routes and middlewares, see [lib example](./examples/lib-example/gen.go) for more details)

## Installation

Add the CLI as a Go tool dependency and the library as a module dependency:

```bash
go get -tool github.com/d1vbyz3r0/typed/cmd/typed@latest
go get github.com/d1vbyz3r0/typed@latest
```

The library dependency is required because the generated Go source imports
`github.com/d1vbyz3r0/typed`.

## Usage

### 1. Expose route registration

Implement `typed.RoutesProvider`:

```go
type RoutesProvider interface {
    OnRouteAdded(func(
        host string,
        route echo.Route,
        handler echo.HandlerFunc,
        middleware []echo.MiddlewareFunc,
    ))
    ProvideRoutes()
}
```

`OnRouteAdded` must install the callback used to collect routes.
`ProvideRoutes` must register the application's routes, but must not start the
server. The configured constructor must return the provider.

The implementation used by this repository is in
[`examples/server/builder.go`](./examples/server/builder.go).

### 2. Create a configuration file

Paths in the configuration are resolved relative to the directory from which
the generator or generated program is run.

```yaml
input:
  title: Example API
  version: 0.0.1
  servers:
    - url: http://localhost:8080

  # Optional. When set, the first path segment after this prefix becomes
  # the operation tag.
  api-prefix: /api/v1

  routes-provider-ctor: NewBuilder
  routes-provider-pkg: github.com/acme/service/internal/server

  handlers:
    - path: .
      recursive: true

  # Packages whose exported types and enums may be added to components.
  models:
    - path: ../dto
      recursive: true
      include:
        - pkg: "^publicdto$"
      exclude:
        - name: "^Internal"

output:
  # Generated Go source.
  path: ../gen/spec.go
  # Optional. Defaults to main.
  package: main
  # Generated OpenAPI document. Supported extensions: .yaml, .yml, .json.
  spec-path: ../gen/openapi.yaml

# Names of built-in typed hooks called for each matched handler.
processing-hooks:
  - EchoJWTMiddlewareHook

# Maximum concurrent package parsing operations. Values <= 0 use the number
# of loaded packages.
concurrency: 0
debug: false
```

Handler and model entries require a `path`. Setting `recursive: true` loads
subpackages. Model filters are regular expressions and support `path`,
`import-path`, `pkg`, and `name`. If include filters are present, a type must
match at least one of them. A type that subsequently matches an exclude filter
is rejected.

By default, the generated source is an executable in `package main`.
`input.routes-provider-ctor`, `input.routes-provider-pkg`, and
`output.spec-path` are required in this mode.

Setting `output.package` to another package name generates only the private
`spec` and `registry` variables. Routes provider settings and `spec-path` are
not required. A file in the same package can use these variables to call
`typed.Generate`, expose getters or wrappers, and save the resulting
specification.

The complete example configuration is
[`examples/configs/standalone.yaml`](./examples/configs/standalone.yaml)
or
[`examples/configs/lib.yaml`](./examples/configs/lib.yaml).

### 3. Generate the specification

```bash
go tool typed -config typed.yaml
go run ./path/to/generated/spec.go
```

The first command analyzes the configured packages and writes the generated
Go source. The second command registers routes and writes the OpenAPI document.

The commands can also be used with `go generate`:

```go
//go:generate go tool typed -config ../typed.yaml
//go:generate go run ../gen/spec.go
```

CLI flags:

```text
-config string
    path to config file
-version
    print version and exit
```

## Generated Data

For handlers that can be matched to registered Echo routes, `typed` currently
generates:

- paths and HTTP methods from registered Echo routes;
- operation IDs from handler function names;
- operation descriptions from handler documentation comments;
- optional tags derived from `input.api-prefix`;
- path, query, header, and form parameters found in inline Echo context calls;
- path, query, header, form, JSON, and XML inputs declared through a struct
  passed to `echo.Context.Bind`;
- response status codes, content types, models, and headers for supported Echo
  response methods;
- component schemas for discovered models and typed constants;
- UUID and time schemas inferred from supported conversion calls;
- YAML or JSON output, selected by `output.spec-path`;
- Response description for websocket usages. Supported libs are: 
  [github.com/coder/websocket](https://github.com/coder/websocket), 
  [github.com/gorilla/websocket](https://github.com/gorilla/websocket)
  and [golang.org/x/net/websocket](https://pkg.go.dev/golang.org/x/net/websocket).

Supported Echo response methods are:

```text
JSON, JSONPretty, JSONBlob
XML, XMLPretty, XMLBlob
String, Blob, Stream
Redirect, NoContent
```

Inline parameter types default to `string`. A different type is inferred when
the context call is passed directly to one of these functions:

| Package | Functions | Inferred Go type |
| --- | --- | --- |
| `strconv` | `Atoi` | `int` |
| `strconv` | `ParseInt` | `int64` |
| `strconv` | `ParseUint` | `uint` |
| `strconv` | `ParseFloat` | `float64` |
| `strconv` | `ParseBool` | `bool` |
| `github.com/google/uuid` | `Parse`, `MustParse` | `uuid.UUID` |
| `time` | `Parse` | `time.Time` |

For example:

```go
func GetUser(c echo.Context) error {
    id, err := uuid.Parse(c.Param("id"))
    if err != nil {
        return err
    }
    return c.JSON(http.StatusOK, loadUser(id))
}
```

## Extension Points

Custom inline type inference can be registered through
`common/typing.RegisterTypeProvider`:

```go
typing.RegisterTypeProvider(func(pkg, function string) (reflect.Type, bool) {
    if pkg == "decimal" && function == "Parse" {
        return reflect.TypeOf(decimal.Decimal{}), true
    }
    return nil, false
})
```

Schema customizers can be registered with `typed.RegisterCustomizer`.
Handler hooks can be registered with `typed.RegisterHandlerProcessingHook`.
These registrations must run in the process that performs handler parsing or
schema generation. Configuration-based `processing-hooks` only supports
functions exported by the `typed` package.

The generated executable can enable the built-in `EchoJWTMiddlewareHook`
through `processing-hooks`. It marks operations whose captured middleware
function name contains `github.com/labstack/echo-jwt` with a bearer JWT
security requirement.

## Programmatic Generation

The generated executable uses the same public API that is available to library
users:

```go
err := typed.Generate(typed.GenerateOptions{
    Spec:           spec,
    Registry:       registry,
    Routes:         typed.CollectRoutes(routesProvider),
    SearchPatterns: []handlers.SearchPattern{{Path: ".", Recursive: true}},
    APIPrefix:      typed.MakePointer("/api/v1"),
})
```

When `Generator` is omitted, `typed.Generate` creates the default schema
generator. It also initializes missing component maps on the provided
specification. A registry and specification must be provided explicitly.

## Current Limitations

`typed` is based on static pattern matching plus runtime route inspection. It
does not execute handler logic and does not infer arbitrary control flow.
Important current limitations are:

- only handlers found in configured handler packages and matched to a
  registered route are included; unmatched routes are skipped with a warning;
- handler discovery recognizes standard `func(echo.Context) error` handlers
  and wrapper functions returning `echo.HandlerFunc`;
- inline parameter inference expects recognizable direct calls such as
  `strconv.Atoi(c.QueryParam("limit"))`;
- response extraction only recognizes the Echo methods listed above;
- response status codes must be integer literals or `net/http` status
  constants;
- `Blob` and `Stream` content types must be string literals or Echo MIME
  constants;
- request-body inference is based on `c.Bind` and binding tags; multiple bind
  calls and complex binding flows are not represented reliably;
- required and nullable semantics are inferred from Go types and tags and may
  not match application validation rules;
- XML and form field naming has incomplete edge-case support (form tag name have priority over xml name, when used on one struct for same field, [see](https://github.com/d1vbyz3r0/typed/blob/master/field_name_generator.go#L13C6-L13C24));
- exported models in configured model packages are considered for generation,
  so model filters may be needed.

## TODO

- define and implement consistent more reliable required and nullable semantics;
- support `omitempty`, `omitzero`, and validation tags;
- improve handling of forms and multiple `echo.Context.Bind` calls;
- avoid exporting unrelated models by default;
- reduce direct dependencies on Echo internals;
- improve parser and generator performance.

## License

[MIT](./LICENSE)
