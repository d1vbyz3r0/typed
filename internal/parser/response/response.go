package response

import (
	"github.com/d1vbyz3r0/typed/internal/parser/headers"
	"github.com/d1vbyz3r0/typed/internal/parser/response/codes"
	"github.com/d1vbyz3r0/typed/internal/parser/response/mime"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"log/slog"
	"reflect"
	"strconv"
)

var stringType = reflect.TypeOf("")

type StatusCodeMapping map[int][]Response

type Response struct {
	// ContentType is a content type retrieved from func usage context. It's empty for Redirect and NoContent
	ContentType string
	// TypeName is a type name like it's used in code, with package name as prefix (except for std types).
	// Field is empty for responses with empty body
	TypeName string
	// TypePkgPath is a full pkg path for type. Field is empty for responses with empty body
	TypePkgPath string
	Headers     []headers.Header
}

// NewStatusCodeMapping builds StatusCodeMapping from provided handler function declaration
func NewStatusCodeMapping(
	funcDecl *ast.FuncDecl,
	cr *codes.Resolver,
	mr *mime.Resolver,
	typesInfo *types.Info,
) StatusCodeMapping {
	m := make(StatusCodeMapping)
	m.extractResponses(funcDecl, cr, mr, typesInfo)
	return m
}

func (m StatusCodeMapping) extractResponses(
	funcDecl *ast.FuncDecl,
	cr *codes.Resolver,
	mr *mime.Resolver,
	typesInfo *types.Info,
) {
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		resp, supported := newContextResponseType(call, cr, mr, typesInfo)
		if !supported {
			slog.Debug("skipping function call", "call", call.Fun)
			return true
		}

		statusCode, err := resp.StatusCode()
		if err != nil {
			slog.Error("failed to get status code", "error", err)
			return true
		}

		contentType, err := resp.ContentType()
		if err != nil {
			slog.Error("failed to get content type", "error", err)
			return true
		}

		typeName, err := resp.TypeName()
		if err != nil {
			slog.Error("failed to get type name", "error", err)
			return true
		}

		pkgPath, err := resp.TypePkgPath()
		if err != nil {
			slog.Error("failed to get type package path", "error", err)
			return true
		}

		respHeaders := findHeaders(funcDecl, call.Pos(), typesInfo)
		slog.Debug("extracted response headers", "headers", respHeaders)

		m[statusCode] = append(m[statusCode], Response{
			ContentType: contentType,
			TypeName:    typeName,
			TypePkgPath: pkgPath,
			Headers:     respHeaders,
		})

		return true
	})
}

func findHeaders(
	funcDecl *ast.FuncDecl,
	returnPos token.Pos,
	typesInfo *types.Info,
) []headers.Header {
	var res []headers.Header
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if n == nil || n.Pos() >= returnPos {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if !headers.IsHttpHeaderMethod(call, typesInfo) {
			return true
		}

		funcName := sel.Sel.Name
		if funcName != "Set" && funcName != "Add" {
			return true
		}

		switch arg := call.Args[0].(type) {
		case *ast.BasicLit:
			v, err := strconv.Unquote(arg.Value)
			if err != nil {
				slog.Error("unquote basic lit vale", "error", err)
				return true
			}

			res = append(res, headers.Header{
				Name:     v,
				Type:     stringType,
				Required: false,
			})

		case *ast.Ident:
			obj := typesInfo.ObjectOf(arg)
			c, ok := obj.(*types.Const)
			if !ok {
				return true
			}

			res = append(res, headers.Header{
				Name:     constant.StringVal(c.Val()),
				Type:     stringType,
				Required: false,
			})
		}

		return true
	})

	return res
}
