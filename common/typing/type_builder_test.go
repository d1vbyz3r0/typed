package typing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicType(t *testing.T) {
	typ := Basic("string")
	require.NotNil(t, typ)
	assert.Equal(t, TypeKindBasic, typ.kind)
	assert.Equal(t, "string", typ.name)
	assert.Nil(t, typ.elem)
	assert.Empty(t, typ.pkg)
	assert.Empty(t, typ.params)
}

func TestNamedType(t *testing.T) {
	typ := Named("github.com/acme/user", "User")
	require.NotNil(t, typ)
	assert.Equal(t, TypeKindNamed, typ.kind)
	assert.Equal(t, "User", typ.name)
	assert.Equal(t, "github.com/acme/user", typ.pkg)
	assert.Nil(t, typ.elem)
}

func TestSliceType(t *testing.T) {
	typ := Slice(Basic("int"))
	require.NotNil(t, typ)
	assert.Equal(t, TypeKindSlice, typ.kind)
	require.NotNil(t, typ.elem)
	assert.Equal(t, TypeKindBasic, typ.elem.kind)
	assert.Equal(t, "int", typ.elem.name)
}

func TestPointerType(t *testing.T) {
	typ := Pointer(Named("github.com/acme/user", "User"))
	require.NotNil(t, typ)
	assert.Equal(t, TypeKindPointer, typ.kind)
	require.NotNil(t, typ.elem)
	assert.Equal(t, TypeKindNamed, typ.elem.kind)
	assert.Equal(t, "User", typ.elem.name)
}

func TestArrayType(t *testing.T) {
	typ := Array(Basic("string"), 5)
	require.NotNil(t, typ)
	assert.Equal(t, TypeKindArray, typ.kind)
	assert.EqualValues(t, 5, typ.size)
	require.NotNil(t, typ.elem)
	assert.Equal(t, TypeKindBasic, typ.elem.kind)
	assert.Equal(t, "string", typ.elem.name)
}

func TestMapType(t *testing.T) {
	typ := Map(
		Basic("string"),
		Basic("int"),
	)
	require.NotNil(t, typ)
	assert.Equal(t, TypeKindMap, typ.kind)
	require.Len(t, typ.params, 2)
	assert.Equal(t, "string", typ.params[0].name)
	assert.Equal(t, "int", typ.params[1].name)
}

func TestNestedType(t *testing.T) {
	typ := Slice(
		Pointer(
			Named("github.com/acme/user", "User"),
		),
	)
	require.NotNil(t, typ)
	assert.Equal(t, TypeKindSlice, typ.kind)
	require.NotNil(t, typ.elem)
	assert.Equal(t, TypeKindPointer, typ.elem.kind)
	require.NotNil(t, typ.elem.elem)
	assert.Equal(t, TypeKindNamed, typ.elem.elem.kind)
	assert.Equal(t, "User", typ.elem.elem.name)
}

func TestComplexType(t *testing.T) {
	typ := Map(
		Basic("string"),
		Slice(
			Named("github.com/acme/user", "User"),
		),
	)

	require.NotNil(t, typ)
	assert.Equal(t, TypeKindMap, typ.kind)

	require.NotNil(t, typ.params)
	require.Len(t, typ.params, 2)

	key := typ.params[0]
	val := typ.params[1]

	assert.Equal(t, "string", key.name)

	require.Equal(t, TypeKindSlice, val.kind)
	require.Equal(t, "User", val.elem.name)
}

func TestGenericStyleType(t *testing.T) {
	typ := Named(
		"github.com/acme/box",
		"Box",
		Named("github.com/acme/user", "User"),
	)

	require.NotNil(t, typ)

	assert.Equal(t, TypeKindNamed, typ.kind)
	assert.Equal(t, "Box", typ.name)

	require.Len(t, typ.params, 1)
	assert.Equal(t, "User", typ.params[0].name)
}

func TestDeepNesting(t *testing.T) {
	typ := Slice(
		Slice(
			Slice(
				Basic("int"),
			),
		),
	)

	cur := typ
	for i := 0; i < 3; i++ {
		require.Equal(t, TypeKindSlice, cur.kind)
		require.NotNil(t, cur.elem)
		cur = cur.elem
	}

	assert.Equal(t, TypeKindBasic, cur.kind)
	assert.Equal(t, "int", cur.name)
}

func TestEnumType(t *testing.T) {
	base := Named("github.com/acme/user", "Role")
	typ := Enum(base, []any{"admin", "user", "guest"})
	require.NotNil(t, typ)

	require.Equal(t, TypeKindEnum, typ.kind, "unexpected kind")
	require.Equal(t, base, typ.elem, "unexpected elem")
	require.ElementsMatch(t, []any{"admin", "user", "guest"}, typ.enumValues, "unexpected values")
}

func TestTypeTreeToString(t *testing.T) {
	cases := []struct {
		name  string
		_type *Type
		want  string
	}{
		{
			name:  "basic type",
			_type: Basic("string"),
			want:  `t.Basic("string")`,
		},
		{
			name:  "array of basic types",
			_type: Array(Basic("int"), 10),
			want:  `t.Array(t.Basic("int"), 10)`,
		},
		{
			name:  "slice of basic types",
			_type: Slice(Basic("string")),
			want:  `t.Slice(t.Basic("string"))`,
		},
		{
			name:  "map of basic types",
			_type: Map(Basic("string"), Basic("int")),
			want:  `t.Map(t.Basic("string"), t.Basic("int"))`,
		},
		{
			name:  "pointer to basic type",
			_type: Pointer(Basic("int")),
			want:  `t.Pointer(t.Basic("int"))`,
		},
		{
			name:  "named type",
			_type: Named("github.com/example/foo", "Named"),
			want:  `t.Named("github.com/example/foo", "Named")`,
		},
		{
			name:  "named generic type with basic generic args",
			_type: Named("github.com/example/foo", "Generic", Basic("int"), Basic("string")),
			want:  `t.Named("github.com/example/foo", "Generic", t.Basic("int"), t.Basic("string"))`,
		},
		{
			name: "named generic type with named generic args",
			_type: Named(
				"github.com/example/foo", "Generic",
				Named("github.com/example/bar", "Arg1"),
				Named("github.com/example/baz", "Arg2"),
			),
			want: `t.Named("github.com/example/foo", "Generic", t.Named("github.com/example/bar", "Arg1"), t.Named("github.com/example/baz", "Arg2"))`,
		},
		{
			name: "named generic type with map and pointer to slice of pointers",
			_type: Named(
				"github.com/example/foo", "Generic",
				Map(Basic("int"), Named("github.com/example/bar", "MapVal")),
				Pointer(Slice(Pointer(Basic("string")))),
			),
			want: `t.Named("github.com/example/foo", "Generic", t.Map(t.Basic("int"), t.Named("github.com/example/bar", "MapVal")), t.Pointer(t.Slice(t.Pointer(t.Basic("string")))))`,
		},
		{
			name:  "enum",
			_type: Enum(Named("github.com/acme/user", "Role"), []any{"admin", "user", "guest"}),
			want:  `t.Enum(t.Named("github.com/acme/user", "Role"), []any{"admin", "user", "guest"})`,
		},
		{
			name:  "array of named types",
			_type: Array(Named("github.com/example/foo", "Named"), 10),
			want:  `t.Array(t.Named("github.com/example/foo", "Named"), 10)`,
		},
		{
			name:  "slice of named types",
			_type: Slice(Named("github.com/example/foo", "Named")),
			want:  `t.Slice(t.Named("github.com/example/foo", "Named"))`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := TypeTreeToString("t", tc._type, Namer)
			require.Equal(t, tc.want, got, "got unexpected result")
		})
	}
}
