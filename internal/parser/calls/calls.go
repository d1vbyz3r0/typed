package calls

import "go/ast"

func IsEchoContextMethodCall(call *ast.CallExpr) bool {
	if call == nil {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	if ident.Obj == nil || ident.Obj.Decl == nil {
		return false
	}

	decl, ok := ident.Obj.Decl.(*ast.Field)
	if !ok {
		return false
	}

	declSel, ok := decl.Type.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	declXIdent, ok := declSel.X.(*ast.Ident)
	if !ok {
		return false
	}

	if declSel.Sel == nil {
		return false
	}

	return declXIdent.Name == "echo" && declSel.Sel.Name == "Context"
}
