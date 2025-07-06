# Typed - OpenAPI Specification Generator for Echo Framework

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**Typed** is a Go tool that automatically generates OpenAPI 3.0 specifications for projects using the Echo web
framework.
It uses a sophisticated two-stage approach combining AST (Abstract Syntax Tree) parsing and reflection to analyze your
code and produce accurate, comprehensive API documentation.

## 🚀 Features

- **Two-Stage Generation Process**: Combines AST analysis with runtime reflection for maximum accuracy
- **Code-first OpenAPI Generation approach**: Generates OpenAPI 3.0 specifications from your echo server, no magic
  comments required
- **Intelligent Parameter Detection**: Automatically detects path, query, and form parameters from inline usage
- **Type Inference**: Determines parameter types from usage context (e.g. `strconv.Atoi`, `uuid.Parse`)
- **Extensible Architecture**: Customizable type providers and schema customizers
- **Multiple Parameter Sources**: Supports path parameters, query parameters, and form data (including file uploads)
- **Type Registry Generation**: Creates Go code with type registry for reflection-based analysis
- **Multiple Output Formats**: Supports both YAML and JSON output formats
- **Echo Framework Integration**: Specifically designed for Echo-based projects
- **Library Mode**: Generated code can be used as a library

---

## 📦 Installation

```bash
go install github.com/d1vbyz3r0/typed/cmd/typed@latest
go get github.com/d1vbyz3r0/typed@latest
```

---

## 🛠️ Usage

### Basic Usage

To generate OpenAPI specification for your Echo project, write yaml config and add go generate directives to your code (or run them manually).
Example of config file can be found [here](./examples/typed.yaml)

```bash
//go:generate typed -config ../typed.yaml
//go:generate go run ../gen/spec_gen.go
```

### Command Line Options

```bash
typed [flags]

Flags:
  -config string Path to config file
```

### Supported Output Formats

The tool automatically detects the output format based on file extension:

- **YAML**: `.yaml` or `.yml` extensions
- **JSON**: `.json` extension

---

## 🧠 Intelligent Parameter Detection

One of Typed's most powerful features is its ability to automatically detect and analyze parameter usage from your Echo
handlers:

### Path Parameters

```go
func GetUser(c echo.Context) error {
    // Typed detects 'id' parameter and infers int type from strconv.Atoi usage. Note that for now you should pass c.Param directly as argument
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        return err
    }
    // Handler logic...
    return c.JSON(http.StatusOK, user)
}

// Route: e.GET("/users/:id", GetUser)
// Result: OpenAPI path parameter 'id' with integer type
```

### Query Parameters

```go
func SearchUsers(c echo.Context) error {
    // Typed detects 'limit' as integer and 'active' as boolean
    limit, _ := strconv.Atoi(c.QueryParam("limit"))
    active, _ := strconv.ParseBool(c.QueryParam("active"))
    
    // Handler logic...
    return c.JSON(http.StatusOK, users)
}

// Result: OpenAPI query parameters with correct types
```

### Form Parameters & File Uploads

```go
func UpdateProfile(c echo.Context) error {
    // Typed detects form fields and file uploads
    name := c.FormValue("name")
    email := c.FormValue("email")
    
    // File upload detection
    avatar, err := c.FormFile("avatar")
    if err != nil {
        return err
    }
    
    // Handler logic...
    return c.JSON(http.StatusOK, response)
}

// Result: OpenAPI form schema with string fields and binary file field
```

Of course, you also can declare parameters as struct fields with necessary [tags](https://echo.labstack.com/docs/binding).
When both struct tag and inline usage are found, the struct field will have priority.

### Supported Type Inference

Typed automatically infers parameter types from common conversion functions:

| Package   | Function             | Inferred Type |
|-----------|----------------------|---------------|
| `strconv` | `Atoi`               | `int`         |
| `strconv` | `ParseInt`           | `int64`       |
| `strconv` | `ParseUint`          | `uint`        |
| `strconv` | `ParseFloat`         | `float64`     |
| `strconv` | `ParseBool`          | `bool`        |
| `uuid`    | `Parse`, `MustParse` | `uuid.UUID`   |
| `time`    | `Parse`              | `time.Time`   |

## 🔧 Extensibility

### Custom Type Providers

You can extend type inference by registering custom type providers:

```go
// Example from common/typing/type.go
func RegisterTypeProvider(p Provider) {
    providers = append(providers, p)
}

// Custom provider example
func customProvider(pkg string, funcName string) (reflect.Type, bool) {
    if pkg == "mypackage" && funcName == "ParseCustomType" {
        return reflect.TypeOf(MyCustomType{}), true
    }
    return nil, false
}
```

### Schema Customizers

Customize OpenAPI schema generation with custom functions:

```go
// Example from customizers.go
func RegisterCustomizer(fn openapi3gen.SchemaCustomizerFn) {
    customizers = append(customizers, fn)
}

// Built-in customizers include:
// - Field name overrides from struct tags
// - File upload handling
// - UUID format specification
// - Enum value support
```

## 🔌 Handler Processing Hooks

Typed provides a hook system that allows you to customize OpenAPI specification generation based on handler analysis. This is particularly useful for automatically detecting and documenting middleware-specific behavior.

### Built-in Hooks

#### JWT Authentication Hook

Typed includes a built-in hook that automatically detects Echo JWT middleware usage and adds appropriate security schemes to your OpenAPI specification:
To enable it, add following to your config file:
```yaml
processing-hooks:
  - "EchoJWTMiddlewareHook"
```

```go
func GetProtectedResource(c echo.Context) error {
    // Your protected handler logic
    return c.JSON(http.StatusOK, data)
}

// Route with JWT middleware
protected := e.Group("/api")
protected.Use(echojwt.WithConfig(echojwt.Config{
    SigningKey: []byte("secret"),
}))
protected.GET("/users", GetProtectedResource)
```

**Result**: Automatically adds Bearer token security scheme to the OpenAPI specification:

```yaml
components:
  securitySchemes:
    bearerAuthScheme:
      type: http
      scheme: bearer
      bearerFormat: JWT
paths:
  /api/users:
    get:
      security:
        - bearerAuthScheme: []
```

### Custom Hooks

You can register custom hooks to extend the specification generation process:

```go
// Example from middlewares.go
func RegisterHandlerProcessingHook(hook HandlerProcessingHookFn) {
    handlerProcessingHooks = append(handlerProcessingHooks, hook)
}

// Custom hook example
func CustomAuthHook(spec *openapi3.T, operation *openapi3.Operation, handler handlers.Handler) {
    // Analyze handler middlewares
    for _, mw := range handler.Middlewares() {
        middlewareName := typed.GetMiddlewareFuncName(mw)
        
        if strings.Contains(middlewareName, "myauth") {
            // Add custom security scheme
            if spec.Components.SecuritySchemes == nil {
                spec.Components.SecuritySchemes = make(map[string]*openapi3.SecuritySchemeRef)
            }
            
            spec.Components.SecuritySchemes["customAuth"] = &openapi3.SecuritySchemeRef{
                Value: &openapi3.SecurityScheme{
                    Type: "apiKey",
                    In:   "header",
                    Name: "X-API-Key",
                },
            }
            
            // Apply to operation
            if operation.Security == nil {
                operation.Security = openapi3.NewSecurityRequirements()
            }
            operation.Security.With(openapi3.SecurityRequirement{
                "customAuth": []string{},
            })
        }
    }
}

// Register your custom hook
func init() {
    typed.RegisterHandlerProcessingHook(CustomAuthHook)
}
```

### Hook Function Signature

```go
type HandlerProcessingHookFn func(spec *openapi3.T, operation *openapi3.Operation, handler handlers.Handler)
```

**Parameters:**
- `spec`: The OpenAPI specification being built
- `operation`: The current operation being processed
- `handler`: Handler information including middlewares, route, and metadata

### Use Cases for Custom Hooks

- **Authentication/Authorization**: Automatically detect auth middlewares and add security schemes
- **Rate Limiting**: Add rate limit headers and responses based on middleware detection
- **CORS**: Document CORS headers and preflight responses
- **Validation**: Add validation error responses based on validation middleware
- **Logging/Monitoring**: Add operation IDs or tags based on middleware configuration
- **Custom Headers**: Document custom headers added by middlewares

### Middleware Detection

Typed provides utilities to analyze middleware functions:

```go
// Get middleware function name for analysis
middlewareName := GetMiddlewareFuncName(middleware)

// Example middleware names:
// "github.com/labstack/echo-jwt.(*Config).ToMiddleware.func1"
// "github.com/labstack/echo/v4/middleware.CORS.func1"
// "myproject/middleware.CustomAuth"
```

This hook system makes Typed highly extensible and allows it to automatically document complex middleware behavior without manual specification.

---

## 🏗️ How It Works

Typed uses a sophisticated two-stage approach to overcome the limitations of pure AST analysis:

### Stage 1: Code Generation & Type Registry

1. **AST Parsing**: Analyzes your Go source code using Go's AST parser
2. **Type Discovery**: Identifies all types used in Echo handlers and routes
3. **Registry Generation**: Generates Go code with a type registry containing all discovered types
4. **Standard Functions**: Includes utility functions for reflection-based analysis
5. **Library Output**: The generated code can be used as a standalone library

### Stage 2: Specification Generation

1. **Registry Execution**: Runs the generated code to access `reflect.Type` information
2. **Echo Route Analysis**: Maps Echo routes to their corresponding handlers
3. **Schema Generation**: Uses [kin-openapi](https://github.com/getkin/kin-openapi) to create OpenAPI schemas from
   reflection data
4. **Specification Assembly**: Builds complete OpenAPI 3.0 specification with proper SchemaRefs
5. **Output Generation**: Saves specification in requested format (YAML/JSON)

### Why Two Stages?

The two-stage approach is necessary because:

- **AST Limitation**: During AST analysis, we cannot access `reflect.Type` information
- **Runtime Reflection**: We need actual type information to generate accurate schemas
- **Best of Both Worlds**: Combines compile-time analysis with runtime type information

### Key Dependencies

- **[kin-openapi](https://github.com/getkin/kin-openapi)**: Powers the OpenAPI 3.0 specification generation, schema
  creation, and SchemaRef handling
- **Go AST**: For source code analysis and type discovery
- **Go Reflection**: For runtime type information access

---

## 🤝 Contributing

Contributions are welcome! This project was created because similar tools weren't available for Echo framework projects.

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## Acknowledgments

- **[kin-openapi](https://github.com/getkin/kin-openapi)** - Essential for OpenAPI 3.0 specification generation and
  schema handling
- **[Echo Framework](https://echo.labstack.com/)** - High performance, minimalist Go web framework
- **Go AST & Reflection** - Powerful code analysis and runtime type inspection capabilities

---

## 🐛 Issues & Support

If you encounter any issues or have questions:

1. Check existing [Issues](https://github.com/d1vbyz3r0/typed/issues)
2. Create a new issue with detailed description
3. Include code examples and error messages

---

## 🔮 Roadmap

- [ ] Enhanced comment parsing for OpenAPI descriptions
- [ ] Headers support
- [ ] Add more std hooks 
