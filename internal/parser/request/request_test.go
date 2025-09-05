package request

import (
	"go/ast"
	"log/slog"
	"mime/multipart"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestNewRequest_JSON(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../../testdata/request/jsontest")
	require.NoError(t, err)

	want := &Request{
		BindModel:    "jsontest.JsonDTO",
		BindModelPkg: "github.com/d1vbyz3r0/typed/testdata/request/jsontest",
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEApplicationJSON: Body{},
		},
		PathParams:  nil,
		QueryParams: nil,
	}

	pkg := pkgs[0]
	file := pkg.Syntax[0]
	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		req := New(decl, pkg.TypesInfo, ParseInlineForms(), ParseInlinePathParams(), ParseInlineQueryParams())
		require.Equal(t, want, req)
		return false
	})
}

func TestNewRequest_XML(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../../testdata/request/xmltest")
	require.NoError(t, err)

	want := &Request{
		BindModel:    "xmltest.XMLDto",
		BindModelPkg: "github.com/d1vbyz3r0/typed/testdata/request/xmltest",
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEApplicationXML: Body{},
		},
		PathParams:  nil,
		QueryParams: nil,
	}

	pkg := pkgs[0]
	file := pkg.Syntax[0]
	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		req := New(decl, pkg.TypesInfo, ParseInlineForms(), ParseInlinePathParams(), ParseInlineQueryParams())
		require.Equal(t, want, req)
		return false
	})
}

func TestNewRequest_EmptyTags(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../../testdata/request/emptytest")
	require.NoError(t, err)

	want := &Request{
		BindModel:    "emptytest.NoTags",
		BindModelPkg: "github.com/d1vbyz3r0/typed/testdata/request/emptytest",
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEApplicationJSON: Body{},
			echo.MIMEApplicationXML:  Body{},
		},
		PathParams:  nil,
		QueryParams: nil,
	}

	pkg := pkgs[0]
	file := pkg.Syntax[0]
	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		req := New(decl, pkg.TypesInfo, ParseInlineForms(), ParseInlinePathParams(), ParseInlineQueryParams())
		require.Equal(t, want, req)
		return false
	})
}

func TestNewRequest_FormTagsNoFiles(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../../testdata/request/formtest/nofile")
	require.NoError(t, err)

	want := &Request{
		BindModel:    "nofile.Form",
		BindModelPkg: "github.com/d1vbyz3r0/typed/testdata/request/formtest/nofile",
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEApplicationForm: Body{},
			echo.MIMEMultipartForm:   Body{},
		},
		PathParams:  nil,
		QueryParams: nil,
	}

	pkg := pkgs[0]
	file := pkg.Syntax[0]
	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		req := New(decl, pkg.TypesInfo, ParseInlineForms(), ParseInlinePathParams(), ParseInlineQueryParams())
		require.Equal(t, want, req)
		return false
	})
}

func TestNewRequest_FormWithFile(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../../testdata/request/formtest/file")
	require.NoError(t, err)

	want := &Request{
		BindModel:    "file.Form",
		BindModelPkg: "github.com/d1vbyz3r0/typed/testdata/request/formtest/file",
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEMultipartForm: Body{},
		},
		PathParams:  nil,
		QueryParams: nil,
	}

	pkg := pkgs[0]
	file := pkg.Syntax[0]
	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		req := New(decl, pkg.TypesInfo, ParseInlineForms(), ParseInlinePathParams(), ParseInlineQueryParams())
		require.Equal(t, want, req)
		return false
	})
}

func TestNewRequest_FormWithFiles(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../../testdata/request/formtest/files")
	require.NoError(t, err)

	want := &Request{
		BindModel:    "files.Form",
		BindModelPkg: "github.com/d1vbyz3r0/typed/testdata/request/formtest/files",
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEMultipartForm: Body{},
		},
		PathParams:  nil,
		QueryParams: nil,
	}

	pkg := pkgs[0]
	file := pkg.Syntax[0]
	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		req := New(decl, pkg.TypesInfo, ParseInlineForms(), ParseInlinePathParams(), ParseInlineQueryParams())
		require.Equal(t, want, req)
		return false
	})
}

func TestNewRequest_NoBody(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../../testdata/request/nobody")
	require.NoError(t, err)

	want := &Request{
		BindModel:          "",
		ContentTypeMapping: ContentTypeMapping{},
		PathParams:         nil,
		QueryParams:        nil,
	}

	pkg := pkgs[0]
	file := pkg.Syntax[0]
	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		req := New(decl, pkg.TypesInfo, ParseInlineForms(), ParseInlinePathParams(), ParseInlineQueryParams())
		require.Equal(t, want, req)
		return false
	})
}

func TestNewRequest_NoBinds(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../../testdata/request/nobind")
	require.NoError(t, err)

	want := &Request{
		BindModel:          "",
		ContentTypeMapping: ContentTypeMapping{},
		PathParams:         nil,
		QueryParams:        nil,
	}

	pkg := pkgs[0]
	file := pkg.Syntax[0]
	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		req := New(decl, pkg.TypesInfo, ParseInlineForms(), ParseInlinePathParams(), ParseInlineQueryParams())
		require.Equal(t, want, req)
		return false
	})
}

func TestNewRequest_MultipleTags(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../../testdata/request/multiple")
	require.NoError(t, err)

	want := &Request{
		BindModel:    "multiple.Data",
		BindModelPkg: "github.com/d1vbyz3r0/typed/testdata/request/multiple",
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEApplicationJSON: Body{},
			echo.MIMEMultipartForm:   Body{},
			echo.MIMEApplicationForm: Body{},
			echo.MIMEApplicationXML:  Body{},
		},
		PathParams:  nil,
		QueryParams: nil,
	}

	pkg := pkgs[0]
	file := pkg.Syntax[0]
	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		req := New(decl, pkg.TypesInfo, ParseInlineForms(), ParseInlinePathParams(), ParseInlineQueryParams())
		require.Equal(t, want, req)
		return false
	})
}

func TestNewRequest_InlineFormNoFiles(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../../testdata/request/formtest/inline")
	require.NoError(t, err)

	f := reflect.StructOf([]reflect.StructField{
		{
			Name:      "Name",
			PkgPath:   "",
			Type:      reflect.TypeOf(""),
			Tag:       `form:"name"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
		{
			Name:      "Age",
			PkgPath:   "",
			Type:      reflect.TypeOf(0),
			Tag:       `form:"age"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
	})

	want := &Request{
		BindModel: "",
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEApplicationForm: Body{
				Form: f,
			},
			echo.MIMEMultipartForm: Body{
				Form: f,
			},
		},
		PathParams:  nil,
		QueryParams: nil,
	}

	pkg := pkgs[0]
	file := pkg.Syntax[0]
	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		req := New(decl, pkg.TypesInfo, ParseInlineForms(), ParseInlinePathParams(), ParseInlineQueryParams())
		require.Equal(t, want, req)
		return false
	})
}

func TestNewRequest_InlineFormWithFiles(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../../testdata/request/formtest/inlinefiles")
	require.NoError(t, err)

	f := reflect.StructOf([]reflect.StructField{
		{
			Name:      "Name",
			PkgPath:   "",
			Type:      reflect.TypeOf(""),
			Tag:       `form:"name"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
		{
			Name:      "File",
			PkgPath:   "",
			Type:      reflect.TypeOf(new(multipart.FileHeader)),
			Tag:       `form:"file"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
	})

	want := &Request{
		BindModel: "",
		ContentTypeMapping: ContentTypeMapping{
			echo.MIMEMultipartForm: Body{
				Form: f,
			},
		},
		PathParams:  nil,
		QueryParams: nil,
	}

	pkg := pkgs[0]
	file := pkg.Syntax[0]
	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		req := New(decl, pkg.TypesInfo, ParseInlineForms(), ParseInlinePathParams(), ParseInlineQueryParams())
		require.Equal(t, want.BindModel, req.BindModel)
		got := req.ContentTypeMapping[echo.MIMEMultipartForm].Form
		require.Equal(t, f, got)
		return false
	})
}
