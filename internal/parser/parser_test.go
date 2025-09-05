package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/d1vbyz3r0/typed/internal/parser/request"
	"github.com/d1vbyz3r0/typed/internal/parser/request/path"
	"github.com/d1vbyz3r0/typed/internal/parser/request/query"
	"github.com/d1vbyz3r0/typed/internal/parser/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"golang.org/x/tools/go/packages"
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
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../testdata/parser/c1/")
	require.NoError(t, err)

	require.Len(t, pkgs, 1)

	pkg := pkgs[0]
	require.Len(t, pkg.Errors, 0)

	p, err := New()
	require.NoError(t, err)

	res, err := p.Parse(pkg, ParseInlineForms(), ParseInlineQueryParams(), ParseInlinePathParams(), ParseEnums())
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
				Pkg:  "github.com/d1vbyz3r0/typed/testdata/parser/c1",
				Request: &request.Request{
					BindModel:    "c1.Form",
					BindModelPkg: "github.com/d1vbyz3r0/typed/testdata/parser/c1",
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
							ContentType: echo.MIMETextPlain,
							TypeName:    "string",
							TypePkgPath: "",
						},
					},
					http.StatusBadRequest: []response.Response{
						{
							ContentType: echo.MIMEApplicationXML,
							TypeName:    "c1.Error",
							TypePkgPath: "github.com/d1vbyz3r0/typed/testdata/parser/c1",
						},
						{
							ContentType: echo.MIMEApplicationJSON,
							TypeName:    "c1.Error",
							TypePkgPath: "github.com/d1vbyz3r0/typed/testdata/parser/c1",
						},
					},
					http.StatusOK: []response.Response{
						{
							ContentType: echo.MIMEApplicationJSON,
							TypeName:    "c1.Result",
							TypePkgPath: "github.com/d1vbyz3r0/typed/testdata/parser/c1",
						},
					},
				},
			},
			{
				Doc:  "Other handler is other handler",
				Name: "OtherHandler",
				Pkg:  "github.com/d1vbyz3r0/typed/testdata/parser/c1",
				Request: &request.Request{
					BindModel:    "c1.User",
					BindModelPkg: "github.com/d1vbyz3r0/typed/testdata/parser/c1",
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
							TypeName:    "map[string]string",
							TypePkgPath: "",
						},
					},
					http.StatusOK: []response.Response{
						{
							ContentType: "",
							TypeName:    "",
							TypePkgPath: "",
						},
					},
				},
			},
		},
	}

	require.Equal(t, want.Enums, res.Enums)
	for i, h := range res.Handlers {
		require.Equal(t, want.Handlers[i].Name, h.Name)
		require.Equal(t, want.Handlers[i].Pkg, h.Pkg)
		require.Equal(t, want.Handlers[i].Request.BindModel, h.Request.BindModel)
		require.ElementsMatch(t, want.Handlers[i].Request.PathParams, h.Request.PathParams)
		require.ElementsMatch(t, want.Handlers[i].Request.QueryParams, h.Request.QueryParams)
		require.Equal(t, want.Handlers[i].Request.ContentTypeMapping, h.Request.ContentTypeMapping)
		require.Equal(t, want.Handlers[i].Responses, h.Responses)
	}
}

func TestParserAllModels(t *testing.T) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../testdata/parser/allmodels/")
	require.NoError(t, err)

	require.Len(t, pkgs, 1)

	pkg := pkgs[0]
	require.Len(t, pkg.Errors, 0)

	p, err := New()
	require.NoError(t, err)

	res, err := p.Parse(pkg, ParseAllModels())
	require.NoError(t, err)

	want := Result{
		Enums:    nil,
		Handlers: nil,
		AdditionalModels: []Model{
			{
				Name:    "allmodels.Form",
				PkgPath: "github.com/d1vbyz3r0/typed/testdata/parser/allmodels",
			},
			{
				Name:    "allmodels.Error",
				PkgPath: "github.com/d1vbyz3r0/typed/testdata/parser/allmodels",
			},
			{
				Name:    "allmodels.Result",
				PkgPath: "github.com/d1vbyz3r0/typed/testdata/parser/allmodels",
			},
			{
				Name:    "allmodels.User",
				PkgPath: "github.com/d1vbyz3r0/typed/testdata/parser/allmodels",
			},
			{
				Name:    "string",
				PkgPath: "",
			},
			{
				Name:    "map[string]string",
				PkgPath: "",
			},
		},
	}

	unique := make(map[Model]struct{})
	for _, m := range res.AdditionalModels {
		unique[m] = struct{}{}
	}

	require.ElementsMatch(t, want.AdditionalModels, maps.Keys(unique))
}
