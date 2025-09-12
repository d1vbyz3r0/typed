package path

import (
	"fmt"
	"github.com/d1vbyz3r0/typed/common/meta"
	"github.com/d1vbyz3r0/typed/common/typing"
	"go/ast"
	"log/slog"
	"reflect"
	"strconv"
)

type Param struct {
	Name string
	Type reflect.Type
}

func NewInlinePathParams(funcDecl *ast.FuncDecl) []Param {
	var params []Param
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		callName, ok := meta.GetCalledFuncName(call)
		if !ok {
			return true
		}

		if callName != "Param" {
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

			if typing.IsParamUsage(call, "Param", paramName) {
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

		params = append(params, Param{
			Name: paramName,
			Type: paramType,
		})

		slog.Debug(
			"found inline path param usage",
			"handler", funcDecl.Name.Name,
			"param", paramName,
			"type", paramType,
		)

		return true
	})

	return params
}

func NewStructPathParams(s reflect.Type) ([]Param, error) {
	if s.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", s.Kind())
	}

	params := make([]Param, 0, s.NumField())
	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		tag, ok := field.Tag.Lookup("param")
		if !ok || tag == "-" || tag == "" {
			continue
		}

		params = append(params, Param{
			Name: tag,
			Type: field.Type,
		})

		slog.Debug(
			"found struct path param",
			"param", tag,
			"type", field.Type,
		)
	}

	return params, nil
}
