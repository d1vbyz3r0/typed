package form

import (
	"fmt"
	"go/ast"
	"log/slog"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"

	"github.com/d1vbyz3r0/typed/common/meta"
	"github.com/d1vbyz3r0/typed/common/typing"
)

var multipartType = reflect.TypeOf(new(multipart.FileHeader))

// NewInlineForm builds reflect.Struct from inline form usages and reports if form contains files and any fields found
func NewInlineForm(funcDecl *ast.FuncDecl) (form reflect.Type, hasFiles bool, found bool) {
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
			hasFiles = true

			slog.Debug(
				"found inline form file usage",
				"handler", funcDecl.Name.Name,
				"param", paramName,
			)

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
						return false
					}

					pkgName, ok := meta.GetCalledFuncPkg(call)
					if !ok {
						slog.Debug("failed to get func pkg", "param", paramName)
						return false
					}

					t, ok := typing.GetTypeFromUsageContext(pkgName, funcName)
					if !ok {
						slog.Debug("failed to get func pkg", "param", paramName)
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
				Tag:  reflect.StructTag(fmt.Sprintf(`form:"%s"`, paramName)),
			})

			slog.Debug(
				"found inline form param usage",
				"handler", funcDecl.Name.Name,
				"param", paramName,
				"type", paramType,
			)
		}

		return true
	})

	return reflect.StructOf(fields), hasFiles, len(fields) > 0
}
