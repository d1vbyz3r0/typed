package meta

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

const (
	echoMapName = "Map"
	echoPkgPath = "github.com/labstack/echo/v4"
)

// GetTypeName returns type name in format pkg.TypeName.
// For map it will return map[KeyType]ValueType, KeyType and ValueType will contain package name.
// For slices it will return []ValueType, ValueType will contain package name.
// For echo.Map it will return map[string]any
func GetTypeName(t types.Type) (string, error) {
	return resolveTypeName(t)
}

func resolveTypeName(t types.Type) (string, error) {
	switch t := t.(type) {
	case *types.Named:
		if t.Obj().Name() == echoMapName && t.Obj().Pkg().Path() == echoPkgPath {
			return "map[string]any", nil
		}

		obj := t.Obj()
		return obj.Pkg().Name() + "." + obj.Name(), nil

	case *types.Pointer:
		elem := t.Elem()
		res, err := resolveTypeName(elem)
		if err != nil {
			return "", fmt.Errorf("resolve underlying ptr type name: %w", err)
		}

		return "*" + res, nil

	case *types.Slice:
		elem, err := resolveTypeName(t.Elem())
		if err != nil {
			return "", fmt.Errorf("failed to resolve type name for clice item: %w", err)
		}

		name := fmt.Sprintf("[]%s", elem)
		return name, nil

	case *types.Alias:
		obj := t.Obj()
		pkg := obj.Pkg()
		pkgName := ""
		if pkg != nil {
			pkgName = pkg.Name() + "."
		}

		return pkgName + obj.Name(), nil

	case *types.Map:
		k, err := resolveTypeName(t.Key())
		if err != nil {
			return "", fmt.Errorf("get key type: %w", err)
		}

		elem, err := resolveTypeName(t.Elem())
		if err != nil {
			return "", fmt.Errorf("get elem type: %w", err)
		}

		name := fmt.Sprintf("map[%s]%s", k, elem)
		return name, nil

	case *types.Basic:
		return t.Name(), nil

	default:
		return "", fmt.Errorf("unsupported type: %T", t)
	}
}

// GetPkgPath returns full package path for type
func GetPkgPath(t types.Type) (string, error) {
	switch t := t.(type) {
	case *types.Named:
		obj := t.Obj()
		return obj.Pkg().Path(), nil

	case *types.Pointer:
		elem := t.Elem()
		return GetPkgPath(elem)

	default:
		return "", fmt.Errorf("unsupported type: %T", t)
	}
}

func GetFuncDocumentation(funcDecl *ast.FuncDecl) string {
	if funcDecl.Doc == nil {
		return ""
	}

	var doc strings.Builder
	for _, comment := range funcDecl.Doc.List {
		doc.WriteString(strings.TrimPrefix(comment.Text, "// "))
		doc.WriteString("\n")
	}

	return strings.TrimSpace(doc.String())
}

func GetCalledFuncPkg(call *ast.CallExpr) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", false
	}

	return ident.Name, true
}

func GetCalledFuncName(call *ast.CallExpr) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}

	return sel.Sel.Name, true
}
