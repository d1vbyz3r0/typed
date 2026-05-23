package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"testing"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser/request"
	"github.com/d1vbyz3r0/typed/internal/parser/request/path"
	"github.com/d1vbyz3r0/typed/internal/parser/request/query"
	"github.com/d1vbyz3r0/typed/internal/parser/response"
	"github.com/d1vbyz3r0/typed/internal/testsuite"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestIsEchoHandler_NormalHandler(t *testing.T) {
	src := `
package test

func Handler(c echo.Context) error {
	return nil
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	require.NoError(t, err)

	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		ok = isEchoHandler(decl)
		require.True(t, ok)
		return false
	})
}

func TestIsEchoHandler_StructMethod(t *testing.T) {
	src := `
package test
type Handler struct{}

func (h *Handler) Handler(c echo.Context) error {
	return nil
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	require.NoError(t, err)

	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		ok = isEchoHandler(decl)
		require.True(t, ok)
		return false
	})
}

func TestIsWrapperFunction(t *testing.T) {
	src := `
package test

func Handler() echo.HandlerFunc {
	return func(c echo.Context) error {
		return nil
	}
}`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	require.NoError(t, err)

	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		ok = isWrapperFunction(decl)
		require.True(t, ok)
		return false
	})
}

func TestInlineHandlerDecl(t *testing.T) {
	src := `
package main

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return nil
	})
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	require.NoError(t, err)

	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if decl.Name.Name == "main" {
			return true
		}

		ok = isEchoHandler(decl)
		require.True(t, ok)
		return false
	})
}

func TestParser(t *testing.T) {
	pkg := testsuite.LoadPackage(t, "../../testdata/parser/c1/")

	p, err := New()
	require.NoError(t, err)

	res, err := p.Parse(pkg, ParseInlineForms(), ParseInlineQueryParams(), ParseInlinePathParams(), ParseEnums())
	require.NoError(t, err)

	want := Result{
		AdditionalModels: []*typing.Type{
			typing.Enum(typing.Named("github.com/d1vbyz3r0/typed/testdata/parser/c1", "Role"), []any{"admin", "user", "guest"}),
			typing.Enum(typing.Named("github.com/d1vbyz3r0/typed/testdata/parser/c1", "Status"), []any{int64(1), int64(2)}),
		},
		Handlers: []Handler{
			{
				Doc:  "Handler 1",
				Name: "Handler",
				Pkg:  "github.com/d1vbyz3r0/typed/testdata/parser/c1",
				Request: &request.Request{
					ModelType: typing.Named("github.com/d1vbyz3r0/typed/testdata/parser/c1", "Form"),
					ContentTypeMapping: request.ContentTypeMapping{
						echo.MIMEMultipartForm: request.Body{},
					},
					PathParams: []path.Param{
						{
							Name: "id",
							Type: reflect.TypeOf(int64(0)),
						},
					},
					QueryParams: []query.Param{
						{
							Name: "x",
							Type: reflect.TypeOf(""),
						},
						{
							Name: "json",
							Type: reflect.TypeOf(true),
						},
					},
				},
				Responses: response.StatusCodeMapping{
					http.StatusInternalServerError: []response.Response{
						{
							ModelType:   typing.Basic("string"),
							ContentType: echo.MIMETextPlain,
						},
					},
					http.StatusBadRequest: []response.Response{
						{
							ContentType: echo.MIMEApplicationXML,
							ModelType:   typing.Named("github.com/d1vbyz3r0/typed/testdata/parser/c1", "Error"),
						},
						{
							ContentType: echo.MIMEApplicationJSON,
							ModelType:   typing.Named("github.com/d1vbyz3r0/typed/testdata/parser/c1", "Error"),
						},
					},
					http.StatusOK: []response.Response{
						{
							ContentType: echo.MIMEApplicationJSON,
							ModelType:   typing.Named("github.com/d1vbyz3r0/typed/testdata/parser/c1", "Result"),
						},
					},
				},
			},
			{
				Doc:  "Other handler is other handler",
				Name: "OtherHandler",
				Pkg:  "github.com/d1vbyz3r0/typed/testdata/parser/c1",
				Request: &request.Request{
					ModelType: typing.Named("github.com/d1vbyz3r0/typed/testdata/parser/c1", "User"),
					ContentTypeMapping: request.ContentTypeMapping{
						echo.MIMEMultipartForm:   request.Body{},
						echo.MIMEApplicationForm: request.Body{},
						echo.MIMEApplicationJSON: request.Body{},
					},
					PathParams:  nil,
					QueryParams: nil,
				},
				Responses: response.StatusCodeMapping{
					http.StatusInternalServerError: []response.Response{
						{
							ContentType: echo.MIMEApplicationJSON,
							ModelType:   typing.Map(typing.Basic("string"), typing.Basic("string")),
						},
					},
					http.StatusOK: []response.Response{
						{
							ContentType: "",
						},
					},
				},
			},
		},
	}

	require.ElementsMatch(t, want.AdditionalModels, res.AdditionalModels)
	for i, h := range res.Handlers {
		require.Equal(t, want.Handlers[i].Name, h.Name)
		require.Equal(t, want.Handlers[i].Pkg, h.Pkg)
		require.Equal(t, want.Handlers[i].Request.ModelType, h.Request.ModelType)
		require.ElementsMatch(t, want.Handlers[i].Request.PathParams, h.Request.PathParams)
		require.ElementsMatch(t, want.Handlers[i].Request.QueryParams, h.Request.QueryParams)
		require.Equal(t, want.Handlers[i].Request.ContentTypeMapping, h.Request.ContentTypeMapping)
		require.Equal(t, want.Handlers[i].Responses, h.Responses)
	}
}

func TestParserAllModels(t *testing.T) {
	pkg := testsuite.LoadPackage(t, "../../testdata/parser/allmodels/")

	p, err := New()
	require.NoError(t, err)

	res, err := p.Parse(pkg, ParseAllModels(), ParseEnums())
	require.NoError(t, err)

	models := Result{
		AdditionalModels: []*typing.Type{
			typing.Named("github.com/d1vbyz3r0/typed/testdata/parser/allmodels", "Form"),
			typing.Named("github.com/d1vbyz3r0/typed/testdata/parser/allmodels", "Error"),
			typing.Named("github.com/d1vbyz3r0/typed/testdata/parser/allmodels", "Result"),
			typing.Named("github.com/d1vbyz3r0/typed/testdata/parser/allmodels", "User"),
			typing.Enum(typing.Named("github.com/d1vbyz3r0/typed/testdata/parser/allmodels", "Role"), []any{"admin", "user"}),
		},
	}

	got := make(map[string]struct{})
	for _, m := range res.AdditionalModels {
		got[m.String()] = struct{}{}
	}
	fmt.Println(got)
	want := make([]string, 0, len(models.AdditionalModels))
	for _, m := range models.AdditionalModels {
		want = append(want, m.String())
	}

	require.ElementsMatch(t, want, slices.Collect(maps.Keys(got)))
}
