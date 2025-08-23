package meta

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

// GetTypeName returns type name in format pkg.TypeName.
// For map it will return map[KeyType]ValueType, KeyType and ValueType will contain package name.
// For slices it will return []ValueType, ValueType will contain package name.
func GetTypeName(t types.Type) (string, error) {
	return resolveTypeName(t)
}

func resolveTypeName(t types.Type) (string, error) {
	switch t := t.(type) {
	case *types.Named:
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

// GetPkgPath returns full package path for type. Underlying type should be types.Named
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

// GetPkgName returns last part of pkg path
func GetPkgName(pkgPath string) string {
	parts := strings.Split(pkgPath, "/")
	return parts[len(parts)-1]
}

func IsSubPkg(parent string, child string) bool {
	return GetPkgName(parent) == GetPkgName(child)
}
