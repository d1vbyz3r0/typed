package typing

import (
	"errors"
	"fmt"
	"go/types"
	"strings"
)

var ErrTypeUnsupported = errors.New("unsupported type")

type NamerFunc func(t *Type) (string, string)

// Namer describes how to name type in string representation.
// Default namer returns <full_pkg_path>.<TypeName>
var Namer = func(t *Type) (string, string) {
	return t.pkg, t.name
}

type TypeKind int

const (
	TypeKindUndefined TypeKind = iota
	TypeKindNamed
	TypeKindBasic
	TypeKindPointer
	TypeKindSlice
	TypeKindArray
	TypeKindMap
	TypeKindEnum
)

type Type struct {
	kind TypeKind
	// name is a type name as it was used with type XXX ... declaration.
	// It will be empty for TypeKindPointer, TypeKindSlice, TypeKindArray and TypeKindMap
	name string
	// pkg holds full package path for type.
	// It will be empty for TypeKindBasic, TypeKindPointer, TypeKindSlice and TypeKindArray
	pkg string
	// elem holds type info for pointers, slices and arrays recursively
	elem *Type
	// size holds array size for TypeKindArray, otherwise 0
	size int64
	// params holds type infos for generic type params.
	// For maps it stores key type first and value type as second elem
	params []*Type
	// values holds possible values for TypeKindEnum
	values []any
}

func fillType(t types.Type, _type *Type) error {
	switch tt := t.(type) {
	case *types.Pointer:
		_type.kind = TypeKindPointer
		_type.elem = new(Type)
		return fillType(tt.Elem(), _type.elem)

	case *types.Alias:
		return fillType(types.Unalias(tt), _type)

	case *types.Slice:
		_type.kind = TypeKindSlice
		_type.elem = new(Type)
		return fillType(tt.Elem(), _type.elem)

	case *types.Array:
		_type.kind = TypeKindArray
		_type.elem = new(Type)
		_type.size = tt.Len()
		return fillType(tt.Elem(), _type.elem)

	case *types.Map:
		_type.kind = TypeKindMap
		_type.params = make([]*Type, 2)
		_type.params[0] = new(Type)
		_type.params[1] = new(Type)
		if err := fillType(tt.Key(), _type.params[0]); err != nil {
			return err
		}
		if err := fillType(tt.Elem(), _type.params[1]); err != nil {
			return err
		}
		return nil

	case *types.Basic:
		_type.kind = TypeKindBasic
		_type.name = tt.Name()
		return nil

	case *types.Named:
		obj := tt.Obj()
		pkgPath := ""
		if pkg := obj.Pkg(); pkg != nil {
			pkgPath = pkg.Path()
		}

		_type.kind = TypeKindNamed
		_type.name = obj.Name()
		_type.pkg = pkgPath

		if params := tt.TypeArgs(); params != nil {
			_type.params = make([]*Type, params.Len())
			i := 0
			for param := range params.Types() {
				_type.params[i] = new(Type)
				err := fillType(param, _type.params[i])
				if err != nil {
					return err
				}
				i++
			}
		}
		return nil

	default:
		return fmt.Errorf("%w: %T", ErrTypeUnsupported, tt)
	}
}

func NewType(t types.Type) (*Type, error) {
	_type := new(Type)
	err := fillType(t, _type)
	return _type, err
}

// Name returns type name
func (t *Type) Name() string {
	return t.name
}

// Pkg returns full package path for type
func (t *Type) Pkg() string {
	return t.pkg
}

func (t *Type) Kind() TypeKind {
	return t.kind
}

func (t *Type) IsGeneric() bool {
	return len(t.params) > 0
}

// KeyType returns type descriptor for key if current type is map, otherwise nil
func (t *Type) KeyType() *Type {
	if t.kind != TypeKindMap {
		return nil
	}
	return t.params[0]
}

// ValueType returns type descriptor for value if current type is map, otherwise nil
func (t *Type) ValueType() *Type {
	if t.kind != TypeKindMap {
		return nil
	}
	return t.params[1]
}

// ElemType returns type descriptor for elem if current type is pointer, slice or array, otherwise nil
func (t *Type) ElemType() *Type {
	switch t.kind {
	case TypeKindPointer, TypeKindArray, TypeKindSlice:
		return t.elem
	default:
		return nil
	}
}

// Params returns slice of params, if type is generic, otherwise slice will be nil
func (t *Type) Params() []*Type {
	return t.params
}

func (t *Type) String() string {
	switch t.kind {
	case TypeKindPointer:
		return "*" + t.elem.String()

	case TypeKindArray:
		return fmt.Sprintf("[%d]%s", t.size, t.elem)

	case TypeKindSlice:
		return "[]" + t.elem.String()

	case TypeKindMap:
		return fmt.Sprintf("map[%s]%s", t.params[0], t.params[1])

	case TypeKindBasic:
		return t.name

	case TypeKindNamed:
		pkg, name := Namer(t)
		res := pkg + "." + name
		if t.IsGeneric() {
			res += "["
			params := forEach(t.params, func(t *Type) string {
				return t.String()
			})
			res += strings.Join(params, ",")
			res += "]"
		}
		return res

	case TypeKindEnum:
		return t.elem.String()
	}

	return "UnsupportedType"
}

func forEach[T any](s []T, fn func(t T) string) []string {
	res := make([]string, 0, len(s))
	for _, v := range s {
		res = append(res, fn(v))
	}
	return res
}
