package response

import (
	"go/ast"
	"go/types"
)

func hasWebSocketUsages(funcDecl *ast.FuncDecl, info *types.Info) bool {
	return hasXNetWebSocketUsages(funcDecl, info) ||
		hasGorillaWebSocketUsages(funcDecl, info) ||
		hasCoderWebsocketUsages(funcDecl, info)
}

func hasXNetWebSocketUsages(funcDecl *ast.FuncDecl, info *types.Info) (found bool) {
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if found {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "ServeHTTP" {
			return true
		}

		if isXNetWebSocketHandler(info.TypeOf(sel.X)) {
			found = true
			return false
		}

		return true
	})

	return found
}

func isXNetWebSocketHandler(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	return obj != nil &&
		obj.Name() == "Handler" &&
		obj.Pkg() != nil &&
		obj.Pkg().Path() == "golang.org/x/net/websocket"
}

func hasGorillaWebSocketUsages(funcDecl *ast.FuncDecl, info *types.Info) (found bool) {
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if found {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Upgrade" {
			return true
		}

		if isGorillaUpgrader(info.TypeOf(sel.X)) {
			found = true
			return false
		}

		return true
	})

	return found
}

func isGorillaUpgrader(t types.Type) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}

	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	return obj != nil &&
		obj.Name() == "Upgrader" &&
		obj.Pkg() != nil &&
		obj.Pkg().Path() == "github.com/gorilla/websocket"
}

func hasCoderWebsocketUsages(funcDecl *ast.FuncDecl, info *types.Info) (found bool) {
	ast.Inspect(funcDecl, func(n ast.Node) bool {
		if found {
			return false
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "Accept" {
			return true
		}

		if obj, ok := info.Uses[sel.Sel]; ok {
			if pkg := obj.Pkg(); pkg != nil {
				if pkg.Path() == "github.com/coder/websocket" {
					found = true
					return false
				}
			}
		}
		return true
	})

	return found
}
