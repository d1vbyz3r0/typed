package binding

import (
	"github.com/d1vbyz3r0/typed/common/typing"
	"go/ast"
	"go/types"
	"reflect"
	"slices"
)

func HasTag(s *types.Struct, tag string) bool {
	for i := 0; i < s.NumFields(); i++ {
		t := reflect.StructTag(s.Tag(i))
		tagVal := t.Get(tag)
		if tagVal != "" && tagVal != "-" {
			return true
		}
	}
	return false
}

func HasTags(s *types.Struct, tags []string) bool {
	for i := 0; i < s.NumFields(); i++ {
		t := reflect.StructTag(s.Tag(i))
		contains := slices.ContainsFunc(tags, func(s string) bool {
			tagVal := t.Get(s)
			return tagVal != "" && tagVal != "-"
		})

		if !contains {
			return false
		}
	}

	return true
}

func HasFiles(s *types.Struct) bool {
	const (
		mimePkgPath = "mime/multipart"
		fileHeader  = "FileHeader"
	)

	for i := 0; i < s.NumFields(); i++ {
		field := s.Field(i)
		fieldType := field.Type()
		named, ok := typing.GetUnderlyingNamedType(fieldType)
		if ok {
			pkgPath := named.Obj().Pkg().Path()
			objName := named.Obj().Name()
			if pkgPath == mimePkgPath && objName == fileHeader {
				return true
			}
		} else {
			slice, ok := typing.GetUnderlyingSlice(fieldType)
			if !ok {
				continue
			}

			named, ok = typing.GetUnderlyingNamedType(slice.Elem())
			if ok {
				pkgPath := named.Obj().Pkg().Path()
				objName := named.Obj().Name()
				if pkgPath == mimePkgPath && objName == fileHeader {
					return true
				}
			}
		}
	}

	return false
}

func IsBindCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name == "Bind" {
		return true
	}

	return false
}
