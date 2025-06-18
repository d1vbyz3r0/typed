package generator

import (
	"go/ast"
	"strings"
)

type PathParam struct {
	Name     string
	Required bool   // Always true for path params
	Type     string // Usually string
}

func extractPathParams(path string) []*PathParam {
	params := make([]*PathParam, 0)
	parts := strings.Split(path, "/")

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, &PathParam{
				Name:     strings.TrimPrefix(part, ":"),
				Required: true,
				Type:     "string",
			})
		}
	}
	return params
}

func paramTypeFromContext(funcDecl *ast.FuncDecl, paramName string) string {
	paramType := "string"

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			// Check if this conversion is using our parameter
			if isParamUsage(call, paramName) {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if x, ok := sel.X.(*ast.Ident); ok && x.Name == "strconv" {
						switch sel.Sel.Name {
						case "ParseInt", "ParseUint", "Atoi":
							paramType = "int"

						case "ParseFloat":
							paramType = "float64"
						}
					}
				}
			}
		}
		return true
	})

	return paramType
}

func isParamUsage(call *ast.CallExpr, paramName string) bool {
	if len(call.Args) == 0 {
		return false
	}

	// Check if first argument is c.Param() or c.QueryParam() call
	arg := call.Args[0]
	if callExpr, ok := arg.(*ast.CallExpr); ok {
		if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
			if sel.Sel.Name == "Param" || sel.Sel.Name == "QueryParam" {
				// Check if param name matches
				if len(callExpr.Args) > 0 {
					if lit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
						return strings.Trim(lit.Value, "\"") == paramName
					}
				}
			}
		}
	}

	return false
}
