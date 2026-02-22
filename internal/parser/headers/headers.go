package headers

import (
	"fmt"
	"github.com/d1vbyz3r0/typed/common/meta"
	"github.com/d1vbyz3r0/typed/common/typing"
	"go/ast"
	"go/types"
	"log/slog"
	"reflect"
	"strconv"
)

type Header struct {
	Name     string
	Type     reflect.Type
	Required bool // TODO: determine if required or not
	Tag      reflect.StructTag
}

func IsHttpHeaderMethod(call *ast.CallExpr, typesInfo *types.Info) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	typeInfo := typesInfo.TypeOf(sel.X)
	if typeInfo == nil {
		return false
	}

	recvTypeName, err := meta.GetTypeName(typeInfo)
	if err != nil {
		return false
	}

	return recvTypeName == "http.Header"
}

func NewInlineRequestHeaders(funcDecl *ast.FuncDecl, typesInfo *types.Info) []Header {
	var headers []Header
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if !IsHttpHeaderMethod(call, typesInfo) {
			return true
		}

		callName, ok := meta.GetCalledFuncName(call)
		if !ok {
			return true
		}

		if callName != "Get" {
			return true
		}

		lit, ok := call.Args[0].(*ast.BasicLit)
		if !ok {
			slog.Debug("skipping non BasicLit value")
			return true
		}

		paramName, _ := strconv.Unquote(lit.Value)
		paramType := reflect.TypeOf("")

		ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if typing.IsParamUsage(call, "Get", paramName) {
				funcName, ok := meta.GetCalledFuncName(call)
				if !ok {
					slog.Debug("failed to get func name", "param", paramName)
					return false
				}

				pkgName, ok := meta.GetCalledFuncPkg(call)
				if !ok {
					slog.Debug("failed to get func pkg", "param", paramName)
					return false
				}

				t, ok := typing.GetTypeFromUsageContext(pkgName, funcName)
				if !ok {
					return false
				}

				paramType = t
				return false
			}

			return true
		})

		headers = append(headers, Header{
			Name:     paramName,
			Type:     paramType,
			Required: false,
		})

		slog.Debug(
			"found inline header usage",
			"handler", funcDecl.Name.Name,
			"param", paramName,
			"type", paramType,
		)

		return true
	})

	return headers
}

func NewStructRequestHeaders(s reflect.Type) ([]Header, error) {
	if s.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", s.Kind())
	}

	headers := make([]Header, 0, s.NumField())
	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		tag, ok := field.Tag.Lookup("header")
		if !ok || tag == "-" || tag == "" {
			continue
		}

		headers = append(headers, Header{
			Name:     tag,
			Type:     field.Type,
			Required: field.Type.Kind() != reflect.Ptr,
			Tag:      field.Tag,
		})

		slog.Debug(
			"found struct header param",
			"param", tag,
			"type", field.Type,
		)
	}

	return headers, nil
}
