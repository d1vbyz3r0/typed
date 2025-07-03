package form

import (
	"fmt"
	"github.com/d1vbyz3r0/typed/internal/common/meta"
	"github.com/d1vbyz3r0/typed/internal/common/typing"
	"go/ast"
	"log/slog"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"
)

var multipartType = reflect.TypeOf(new(multipart.FileHeader))

// NewInlineForm builds reflect.Struct from inline form usages and reports if any form fields found
func NewInlineForm(funcDecl *ast.FuncDecl) (reflect.Type, bool) {
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

		switch callName {
		case "FormFile":
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok {
				slog.Debug("skipping non BasicLit value")
				return true
			}

			paramName, _ := strconv.Unquote(lit.Value)
			fields = append(fields, reflect.StructField{
				Name: strings.Title(paramName),
				Type: multipartType,
				Tag:  reflect.StructTag(fmt.Sprintf(`form:"%s"`, paramName)),
			})

		case "FormValue":
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

				if typing.IsParamUsage(call, "FormValue", paramName) {
					funcName, ok := meta.GetCalledFuncName(call)
					if !ok {
						slog.Debug("failed to get func name", "param", paramName)
						return true
					}

					pkgName, ok := meta.GetCalledFuncPkg(call)
					if !ok {
						slog.Debug("failed to get func pkg", "param", paramName)
						return true
					}

					paramType, ok = typing.GetTypeFromUsageContext(pkgName, funcName)
					return false
				}

				return true
			})

			fields = append(fields, reflect.StructField{
				Name: strings.Title(paramName),
				Type: reflect.PointerTo(paramType), // We make fields pointers, since we can't determine if it's really required, at least now...
				Tag:  reflect.StructTag(fmt.Sprintf(`form:"%s"`, paramName)),
			})
		}

		return true
	})

	return reflect.StructOf(fields), len(fields) > 0
}
