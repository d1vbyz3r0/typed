package response

import (
	"github.com/d1vbyz3r0/typed/internal/parser/response/codes"
	"github.com/d1vbyz3r0/typed/internal/parser/response/mime"
	"go/ast"
	"go/types"
	"log/slog"
)

type StatusCodeMapping map[int][]Response

type Response struct {
	// ContentType is a content type retrieved from func usage context. It's empty for Redirect and NoContent
	ContentType string
	// TypeName is a type name like it's used in code, with package name as prefix (except for std types).
	//Field is empty for responses with empty body
	TypeName string
	// TypePkgPath is a full pkg path for type. Field is empty for responses with empty body
	TypePkgPath string
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

		m[statusCode] = append(m[statusCode], Response{
			ContentType: contentType,
			TypeName:    typeName,
			TypePkgPath: pkgPath,
		})

		return true
	})
}
