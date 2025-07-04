package parser

import (
	"github.com/d1vbyz3r0/typed/internal/parser/request"
	"github.com/d1vbyz3r0/typed/internal/parser/request/path"
	"github.com/d1vbyz3r0/typed/internal/parser/request/query"
	"github.com/d1vbyz3r0/typed/internal/parser/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/packages"
	"log/slog"
	"net/http"
	"reflect"
	"testing"
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
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../testdata/parser/c1/")
	require.NoError(t, err)

	require.Len(t, pkgs, 1)

	pkg := pkgs[0]
	require.Len(t, pkg.Errors, 0)

	p, err := New()
	require.NoError(t, err)

	res, err := p.Parse(pkg)
	require.NoError(t, err)

	want := Result{
		Enums: map[string][]any{
			"c1.Role":   {"admin", "user", "guest"},
			"c1.Status": {1, 2},
		},
		Handlers: []Handler{
			{
				Doc:  "Handler 1",
				Name: "Handler",
				Pkg:  "c1",
				Request: &request.Request{
					BindModel: "c1.Form",
					ContentTypeMapping: request.ContentTypeMapping{
						echo.MIMEMultipartForm: request.Body{
							Form: reflect.StructOf([]reflect.StructField{}),
						},
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
							ContentType: echo.MIMETextPlain,
							TypeName:    "string",
						},
					},
					http.StatusBadRequest: []response.Response{
						{
							ContentType: echo.MIMEApplicationXML,
							TypeName:    "c1.Error",
						},
						{
							ContentType: echo.MIMEApplicationJSON,
							TypeName:    "c1.Error",
						},
					},
					http.StatusOK: []response.Response{
						{
							ContentType: echo.MIMEApplicationJSON,
							TypeName:    "c1.Result",
						},
					},
				},
			},
			{
				Doc:  "Other handler is other handler",
				Name: "OtherHandler",
				Pkg:  "c1",
				Request: &request.Request{
					BindModel: "c1.User",
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
							ContentType: echo.MIMETextPlain,
							TypeName:    "map[string]string",
						},
					},
					http.StatusNoContent: nil,
				},
			},
		},
	}

	require.Equal(t, want.Enums, res.Enums)
	for i, h := range res.Handlers {
		require.Equal(t, h.Name, res.Handlers[i].Name)
		require.Equal(t, h.Pkg, res.Handlers[i].Pkg)
		require.Equal(t, h.Request.BindModel, res.Handlers[i].Request.BindModel)
		require.ElementsMatch(t, h.Request.PathParams, res.Handlers[i].Request.PathParams)
		require.ElementsMatch(t, h.Request.QueryParams, res.Handlers[i].Request.QueryParams)
		require.Equal(t, h.Request.ContentTypeMapping, res.Handlers[i].Request.ContentTypeMapping)
		require.Equal(t, h.Responses, res.Handlers[i].Responses)
	}
}
