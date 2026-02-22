package meta

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"sort"
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
	case *types.Interface:
		return t.String(), nil

	case *types.Named:
		obj := t.Obj()
		pkg := obj.Pkg()
		pkgName := ""
		if pkg != nil {
			pkgName = pkg.Name() + "."
		}

		name := pkgName + obj.Name()
		withArgs, err := appendTypeArgs(name, getTypeArgs(t))
		if err != nil {
			return "", fmt.Errorf("append type args for named type: %w", err)
		}

		return withArgs, nil

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

		name := pkgName + obj.Name()
		withArgs, err := appendTypeArgs(name, getTypeArgs(t))
		if err != nil {
			return "", fmt.Errorf("append type args for alias type: %w", err)
		}

		return withArgs, nil

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

func appendTypeArgs(base string, args *types.TypeList) (string, error) {
	if args == nil || args.Len() == 0 {
		return base, nil
	}

	values := make([]string, 0, args.Len())
	for i := 0; i < args.Len(); i++ {
		v, err := resolveTypeName(args.At(i))
		if err != nil {
			return "", fmt.Errorf("resolve type arg: %w", err)
		}

		values = append(values, v)
	}

	return fmt.Sprintf("%s[%s]", base, strings.Join(values, ", ")), nil
}

func getTypeArgs(t types.Type) *types.TypeList {
	type argsProvider interface {
		TypeArgs() *types.TypeList
	}

	v, ok := t.(argsProvider)
	if !ok {
		return nil
	}

	return v.TypeArgs()
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

// GetPkgPaths returns unique package paths used by a type expression.
// It includes package paths from named types and all instantiated type arguments.
func GetPkgPaths(t types.Type) []string {
	if t == nil {
		return nil
	}

	result := make(map[string]struct{})
	visitPkgPaths(t, result)
	if len(result) == 0 {
		return nil
	}

	paths := make([]string, 0, len(result))
	for p := range result {
		paths = append(paths, p)
	}

	sort.Strings(paths)
	return paths
}

func visitPkgPaths(t types.Type, result map[string]struct{}) {
	t = types.Unalias(t)
	switch v := t.(type) {
	case *types.Named:
		obj := v.Obj()
		if obj != nil && obj.Pkg() != nil {
			result[obj.Pkg().Path()] = struct{}{}
		}

		if args := v.TypeArgs(); args != nil {
			for i := 0; i < args.Len(); i++ {
				visitPkgPaths(args.At(i), result)
			}
		}

	case *types.Pointer:
		visitPkgPaths(v.Elem(), result)

	case *types.Slice:
		visitPkgPaths(v.Elem(), result)

	case *types.Array:
		visitPkgPaths(v.Elem(), result)

	case *types.Map:
		visitPkgPaths(v.Key(), result)
		visitPkgPaths(v.Elem(), result)

	case *types.Chan:
		visitPkgPaths(v.Elem(), result)
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
	// TODO: filtering on last segment can cause wrong package matching
	return GetPkgName(parent) == GetPkgName(child)
}

func GetFieldNameByTag(field reflect.StructField) string {
	if v := field.Tag.Get("json"); v != "" && v != "-" {
		return v
	}

	if v := field.Tag.Get("form"); v != "" && v != "-" {
		return v
	}

	if v := field.Tag.Get("xml"); v != "" && v != "-" {
		return v
	}

	return field.Name
}
