package typing

// Name will create named type from provided descriptor.
// Type itself can be generic with arbitrary args, described as *Type
func Named(pkg string, name string, args ...*Type) *Type {
	return &Type{
		kind:   TypeKindNamed,
		pkg:    pkg,
		name:   name,
		params: args,
	}
}

// Basic should create basic type descriptor,
// it does no validation on provided name, so it's up to caller to ensure that name is basic type name
func Basic(name string) *Type {
	return &Type{
		kind: TypeKindBasic,
		name: name,
	}
}

func Slice(elem *Type) *Type {
	return &Type{
		kind: TypeKindSlice,
		elem: elem,
	}
}

func Array(elem *Type, size int64) *Type {
	return &Type{
		kind: TypeKindArray,
		elem: elem,
		size: size,
	}
}

func Map(key *Type, value *Type) *Type {
	return &Type{
		kind: TypeKindMap,
		params: []*Type{
			key,
			value,
		},
	}
}

func Pointer(elem *Type) *Type {
	return &Type{
		kind: TypeKindPointer,
		elem: elem,
	}
}
