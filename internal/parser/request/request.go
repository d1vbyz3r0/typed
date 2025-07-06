package request

import (
	"github.com/d1vbyz3r0/typed/internal/common/meta"
	"github.com/d1vbyz3r0/typed/internal/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser/request/binding"
	"github.com/d1vbyz3r0/typed/internal/parser/request/form"
	"github.com/d1vbyz3r0/typed/internal/parser/request/path"
	"github.com/d1vbyz3r0/typed/internal/parser/request/query"
	"github.com/labstack/echo/v4"
	"go/ast"
	"go/types"
	"log/slog"
	"reflect"
)

var echoBodyBindTags = []string{
	"json",
	"xml",
	"form",
}

var echoParamsBindTags = []string{
	"param",
	"query",
}

type ContentTypeMapping map[string]Body

type Body struct {
	// Form contains reflect.Struct, build from inline form usages with form.NewInlineForm
	Form reflect.Type
}

type Request struct {
	// BindModel as it's used in code: pkg.TypeName
	BindModel string
	// Full path to BindModel package
	BindModelPkg string
	// ContentTypeMapping contains mapping of content-type to request body
	ContentTypeMapping ContentTypeMapping
	PathParams         []path.Param
	QueryParams        []query.Param
}

func New(funcDecl *ast.FuncDecl, info *types.Info, opts ...ParseOpt) *Request {
	parseOpts := new(requestParseOpts)
	for _, opt := range opts {
		opt(parseOpts)
	}

	r := &Request{
		ContentTypeMapping: make(ContentTypeMapping),
		PathParams:         nil,
		QueryParams:        nil,
	}

	if parseOpts.parseInlinePathParams {
		r.PathParams = path.NewInlinePathParams(funcDecl)
	}

	if parseOpts.parseInlineQueryParams {
		r.QueryParams = query.NewInlineQueryParams(funcDecl)
	}

	if parseOpts.parseInlineForms {
		f, hasFiles, found := form.NewInlineForm(funcDecl)
		if found {
			if !hasFiles {
				// if form doesn't contain files, content-type can be both application/x-www-form-urlencoded and multipart/form-data
				r.ContentTypeMapping[echo.MIMEApplicationForm] = Body{
					Form: f,
				}
			}

			r.ContentTypeMapping[echo.MIMEMultipartForm] = Body{
				Form: f,
			}
		}
	}

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if !binding.IsBindCall(call) || len(call.Args) != 1 {
			return true
		}

		argType := info.TypeOf(call.Args[0])
		named, ok := typing.GetUnderlyingNamedType(argType)
		if !ok {
			slog.Error("failed to get underlying named type", "arg_type", argType)
			return true
		}

		s, ok := typing.GetUnderlyingStruct(argType)
		if !ok {
			slog.Error("expected struct as bind arg", "got", argType)
			return true
		}

		if s.NumFields() == 0 {
			slog.Debug("ignoring empty struct", "struct", named)
			return true
		}

		typeName, err := meta.GetTypeName(named)
		if err != nil {
			slog.Error("failed to get type", "type", named, "err", err)
			return true
		}

		pkgPath, err := meta.GetPkgPath(named)
		if err != nil {
			slog.Error("failed to get package path", "type", named, "err", err)
			return true
		}

		r.BindModel = typeName
		r.BindModelPkg = pkgPath

		if !binding.HasTags(s, echoBodyBindTags) && !binding.HasTags(s, echoParamsBindTags) {
			// If no tags provided, echo will try to bind xml + json: https://echo.labstack.com/docs/binding
			r.ContentTypeMapping[echo.MIMEApplicationJSON] = Body{}
			r.ContentTypeMapping[echo.MIMEApplicationXML] = Body{}
			return true
		}

		if binding.HasTag(s, "form") {
			if !binding.HasFiles(s) {
				r.ContentTypeMapping[echo.MIMEApplicationForm] = Body{}
			}
			r.ContentTypeMapping[echo.MIMEMultipartForm] = Body{}
		}

		if binding.HasTag(s, "json") {
			r.ContentTypeMapping[echo.MIMEApplicationJSON] = Body{}
		}

		if binding.HasTag(s, "xml") {
			r.ContentTypeMapping[echo.MIMEApplicationXML] = Body{}
		}

		return true
	})

	return r
}
