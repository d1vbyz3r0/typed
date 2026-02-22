package request

import (
	"github.com/d1vbyz3r0/typed/common/meta"
	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser/headers"
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

var bodyBindingTags = []string{
	"json",
	"xml",
	"form",
}

var paramBindingTags = []string{
	"param",
	"query",
	"header",
}

type ContentTypeMapping map[string]Body

type Body struct {
	// Form contains reflect.Struct, built from inline form usages with form.NewInlineForm
	Form reflect.Type
}

type Request struct {
	// BindModel as it's used in code: pkg.TypeName
	BindModel string
	// Full path to BindModel package
	BindModelPkg string
	// BindModelTypeArgPkgs contains package paths used by bind model type arguments (excluding BindModelPkg).
	BindModelTypeArgPkgs []string
	// ContentTypeMapping contains mapping of content-type to request body
	ContentTypeMapping ContentTypeMapping
	PathParams         []path.Param
	QueryParams        []query.Param
	Headers            []headers.Header
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
		Headers:            nil,
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

	if parseOpts.parseInlineHeaders {
		r.Headers = headers.NewInlineRequestHeaders(funcDecl, info)
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
		allPkgPaths := meta.GetPkgPaths(argType)
		if len(allPkgPaths) > 0 {
			extra := make([]string, 0, len(allPkgPaths))
			for _, p := range allPkgPaths {
				if p == "" || p == pkgPath {
					continue
				}

				extra = append(extra, p)
			}

			if len(extra) > 0 {
				r.BindModelTypeArgPkgs = extra
			}
		}

		hasUntagged := binding.HasAtLeastOneFieldWithoutBindingTag(s, bodyBindingTags, paramBindingTags)

		if binding.HasTag(s, "form") {
			if !binding.HasFiles(s) {
				r.ContentTypeMapping[echo.MIMEApplicationForm] = Body{}
			}
			r.ContentTypeMapping[echo.MIMEMultipartForm] = Body{}
		}

		if binding.HasTag(s, "json") || hasUntagged {
			r.ContentTypeMapping[echo.MIMEApplicationJSON] = Body{}
		}

		if binding.HasTag(s, "xml") || hasUntagged {
			r.ContentTypeMapping[echo.MIMEApplicationXML] = Body{}
		}

		return true
	})

	return r
}
