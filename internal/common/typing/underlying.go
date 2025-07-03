package typing

import "go/types"

// GetUnderlyingStruct returns underlying type struct, "dereferencing" pointers recursively. For not *types.Struct underlying type, second param will be always false
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
