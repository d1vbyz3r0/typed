package path

import (
	"fmt"
	"github.com/d1vbyz3r0/typed/internal/common/meta"
	"github.com/d1vbyz3r0/typed/internal/common/typing"
	"go/ast"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
)

func NewInlinePathParams(funcDecl *ast.FuncDecl) (reflect.Type, bool) {
	fields := make([]reflect.StructField, 0)
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

		fields = append(fields, reflect.StructField{
			Name: strings.Title(paramName),
			Type: paramType, // We make fields required, since we can't determine if it's optional, at least now...
			Tag:  reflect.StructTag(fmt.Sprintf(`path:"%s"`, paramName)),
		})

		return true
	})

	return reflect.StructOf(fields), len(fields) > 0
}
