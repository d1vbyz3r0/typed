package typing

import (
	"go/ast"
	"strconv"
)

// IsParamUsage checks if provided call use contextFuncName as first argument with paramName. Ex: strconv.Atoi(c.FormValue("param_name"))
func IsParamUsage(call *ast.CallExpr, contextFuncName string, paramName string) bool {
	if len(call.Args) == 0 {
		return false
	}

	arg := call.Args[0]
	call, ok := arg.(*ast.CallExpr)
	if !ok {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != contextFuncName {
		return false
	}

	if len(call.Args) == 0 {
		return false
	}

	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok {
		return false
	}

	name, err := strconv.Unquote(lit.Value)
	if err != nil {
		return false
	}

	return name == paramName
}
