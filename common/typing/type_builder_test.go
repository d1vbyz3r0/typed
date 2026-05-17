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
