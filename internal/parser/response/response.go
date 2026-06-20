package response

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"net/http"
	"reflect"
	"strconv"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser/headers"
	"github.com/d1vbyz3r0/typed/internal/parser/response/codes"
	"github.com/d1vbyz3r0/typed/internal/parser/response/mime"
	"github.com/d1vbyz3r0/typed/logging"
)

var stringType = reflect.TypeFor[string]()

type StatusCodeMapping map[int][]Response

type Response struct {
	// ContentType is a content type retrieved from func usage context. It's empty for Redirect and NoContent
	ContentType string
	ModelType   *typing.Type
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
			logging.Debug("skipping function call since it's not echo context response", "call", types.ExprString(call))
			return true
		}

		statusCode, err := resp.StatusCode()
		if err != nil {
			logging.Error("failed to get status code", "error", err)
			return true
		}

		contentType, err := resp.ContentType()
		if err != nil {
			logging.Error("failed to get content type", "error", err)
			return true
		}

		model, err := resp.ModelType()
		if err != nil {
			logging.Error("failed to get model type info", "error", err)
			return true
		}

		respHeaders := findHeaders(funcDecl, call.Pos(), typesInfo)
		logging.Debug("extracted response headers", "headers", respHeaders)

		m[statusCode] = append(m[statusCode], Response{
			ContentType: contentType,
			ModelType:   model,
			Headers:     respHeaders,
		})

		return true
	})

	if hasWebSocketUsages(funcDecl, typesInfo) {
		logging.Debug("found websocket usage, extending headers", "handler", funcDecl.Name.String())
		m[http.StatusSwitchingProtocols] = append(m[http.StatusSwitchingProtocols], Response{
			Headers: []headers.Header{
				{
					Name:     "Connection",
					Type:     stringType,
					Required: true,
					Value:    "Upgrade",
				},
				{
					Name:     "Upgrade",
					Type:     stringType,
					Required: true,
					Value:    "websocket",
				},
			},
		})
	}
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
				logging.Error("unquote basic lit vale", "error", err)
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
