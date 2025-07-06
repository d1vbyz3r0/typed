package typing

import (
	"go/types"
	"reflect"
)

// GetUnderlyingStruct returns underlying type struct, "dereferencing" pointers recursively
func GetUnderlyingStruct(t types.Type) (*types.Struct, bool) {
	t = t.Underlying()
	switch t := t.(type) {
	case *types.Struct:
		return t, true

	case *types.Pointer:
		return GetUnderlyingStruct(t.Elem())

	default:
		return nil, false
	}
}

func GetUnderlyingNamedType(t types.Type) (*types.Named, bool) {
	switch t := t.(type) {
	case *types.Named:
		return t, true

	case *types.Pointer:
		return GetUnderlyingNamedType(t.Elem())

	default:
		return nil, false
	}
}

func GetUnderlyingSlice(t types.Type) (*types.Slice, bool) {
	t = t.Underlying()
	switch t := t.(type) {
	case *types.Slice:
		return t, true

	case *types.Pointer:
		return GetUnderlyingSlice(t.Elem())

	default:
		return nil, false
	}
}

func IsPointer(t types.Type) bool {
	t = t.Underlying()
	_, ok := t.(*types.Pointer)
	return ok
}

func IsSlice(t types.Type) bool {
	t = t.Underlying()
	switch t := t.(type) {
	case *types.Slice:
		return true

	case *types.Pointer:
		return IsSlice(t.Elem())

	default:
		return false
	}
}

func IsMap(t types.Type) bool {
	t = t.Underlying()
	switch t := t.(type) {
	case *types.Map:
		return true

	case *types.Pointer:
		return IsMap(t.Elem())

	default:
		return false
	}
}

// GetUnderlyingElemType returns elem for slice or map. Elem can be empty interface or aliased "any" type
func GetUnderlyingElemType(t types.Type) (types.Type, bool) {
	switch t := t.(type) {
	case *types.Named:
		return t, true

	case *types.Basic:
		return t, true

	case *types.Alias:
		return GetUnderlyingElemType(t.Underlying())

	case *types.Interface:
		if IsAnyType(t) {
			return t, true
		}
		return nil, false

	case *types.Pointer:
		return GetUnderlyingElemType(t.Elem())

	case *types.Slice:
		return GetUnderlyingElemType(t.Elem())

	case *types.Map:
		return GetUnderlyingElemType(t.Elem())

	default:
		return nil, false
	}
}

func IsAnyType(t types.Type) bool {
	t = types.Unalias(t)
	switch t := t.(type) {
	case *types.Interface:
		return t.NumMethods() == 0
	default:
		return false
	}
}

func IsBasicType(t types.Type) bool {
	_, ok := t.(*types.Basic)
	return ok
}

func IsFunc(t types.Type) bool {
	t = t.Underlying()
	_, ok := t.(*types.Signature)
	return ok
}

// DerefReflectPtr returns t, if it's kind is not reflect.Ptr, otherwise recursively gets t.Elem(), while it's pointer
func DerefReflectPtr(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}
