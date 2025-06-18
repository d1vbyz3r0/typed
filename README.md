## Typed

Typed is an OpenAPI 3 specification generator for go echo web framework,
designed to combine static AST analysis and runtime reflection to produce accurate and flexible API specs.


It supports:

- OpenAPI 3.0 generation using [kin-openapi](https://github.com/getkin/kin-openapi)
- Extraction of status codes from echo handlers
- Extraction of used types for each status code
- Enums extraction from go constants (for now on const blocks only with no iota support)
- AST-level handler/method mapping
- For now it supports only json and forms (including file upload)

---

## Installation

```bash
go install github.com/d1vbyz3r0/typed/cmd/typed@latest
go get github.com/d1vbyz3r0/typed@latest
```


## Usage
The generator is driven by a YAML configuration file, you can find small example [here](./examples/typed.yaml)

Generation of spec requires 2 steps:
1. Generation of script, with registry of types and enums
2. Generation of spec, using previously generated script

So you want something like that in your go:generate directives:
```go
//go:generate typed -config ../typed.yaml
//go:generate go run ../gen/spec_gen.go
```
