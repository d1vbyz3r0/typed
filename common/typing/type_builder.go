package typing

import (
	"fmt"
	"strings"
)

// Named will create named type from provided descriptor.
// Type itself can be generic with arbitrary args described with *Type
func Named(pkg string, name string, args ...*Type) *Type {
	return &Type{
		kind:   TypeKindNamed,
		pkg:    pkg,
		name:   name,
		params: args,
	}
}

// Basic will create basic type descriptor,
// it does no validation on provided name, so it's up to caller to ensure that name is basic type name
func Basic(name string) *Type {
	return &Type{
		kind: TypeKindBasic,
		name: name,
	}
}

// Slice will create slice with elem type of provided type descriptor
func Slice(elem *Type) *Type {
	return &Type{
		kind: TypeKindSlice,
		elem: elem,
	}
}

// Array will create array with elem type of provided type descriptor and specified size
func Array(elem *Type, size int64) *Type {
	return &Type{
		kind: TypeKindArray,
		elem: elem,
		size: size,
	}
}

// Map will create map with provided key and value types
func Map(key *Type, value *Type) *Type {
	return &Type{
		kind: TypeKindMap,
		params: []*Type{
			key,
			value,
		},
	}
}

// Pointer will create pointer for provided type
func Pointer(elem *Type) *Type {
	return &Type{
		kind: TypeKindPointer,
		elem: elem,
	}
}

// Enum creates enum for provided type descriptor
func Enum(elem *Type, values []any) *Type {
	return &Type{
		kind:       TypeKindEnum,
		name:       elem.name,
		pkg:        elem.pkg,
		elem:       elem,
		enumValues: values,
	}
}

// TypeTreeToString traverses provided type and builds a chain of constructors ready to use for templates in string format.
// pkg specifes package prefix for generated calls
func TypeTreeToString(pkg string, t *Type, namer NamerFunc) string {
	if namer == nil {
		namer = DefaultNamer
	}

	var walkFn func(t *Type) string
	walkFn = func(t *Type) string {
		switch t.kind {
		case TypeKindPointer:
			elem := walkFn(t.elem)
			next := fmt.Sprintf("%s.Pointer(%s)", pkg, elem)
			return next

		case TypeKindArray:
			elem := walkFn(t.elem)
			next := fmt.Sprintf("%s.Array(%s, %d)", pkg, elem, t.size)
			return next

		case TypeKindSlice:
			elem := walkFn(t.elem)
			next := fmt.Sprintf("%s.Slice(%s)", pkg, elem)
			return next

		case TypeKindMap:
			key := walkFn(t.params[0])
			value := walkFn(t.params[1])
			next := fmt.Sprintf("%s.Map(%s, %s)", pkg, key, value)
			return next

		case TypeKindBasic:
			return fmt.Sprintf(`%s.Basic("%s")`, pkg, t.name)

		case TypeKindNamed:
			tpkg, tname := namer(t)
			if t.IsGeneric() {
				params := forEach(t.params, func(t *Type) string {
					return walkFn(t)
				})
				return fmt.Sprintf(`%s.Named("%s", "%s", %s)`, pkg, tpkg, tname, strings.Join(params, ", "))
			}
			return fmt.Sprintf(`%s.Named("%s", "%s")`, pkg, tpkg, tname)

		case TypeKindEnum:
			elem := walkFn(t.elem)
			return fmt.Sprintf("%s.Enum(%s, %#v)", pkg, elem, t.enumValues)

		default:
			return "UnsupportedType"
		}
	}

	return walkFn(t)
}

// Traverse traverses type tree recursively, until it reaches TypeKindBasic or TypeKindNamed, then fn will be called for this leaf
// If t is nil Traverse exits immidiately
func Traverse(t *Type, fn func(n *Type)) error {
	if t == nil {
		return nil
	}

	switch t.kind {
	case TypeKindPointer:
		Traverse(t.elem, fn)
	case TypeKindArray:
		Traverse(t.elem, fn)
	case TypeKindSlice:
		Traverse(t.elem, fn)
	case TypeKindMap:
		Traverse(t.params[0], fn)
		Traverse(t.params[1], fn)
	case TypeKindBasic:
		fn(t)
		return nil
	case TypeKindNamed:
		fn(t)
		for _, param := range t.params {
			Traverse(param, fn)
		}
	case TypeKindEnum:
		Traverse(t, fn)
		return nil
	default:
		return fmt.Errorf("unsupported type kind: %s", t.kind)
	}

	return nil
}
