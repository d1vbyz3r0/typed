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
	ContentType string
	TypeName    *string
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

		m[statusCode] = append(m[statusCode], Response{
			ContentType: contentType,
			TypeName:    typeName,
		})

		return true
	})
}
