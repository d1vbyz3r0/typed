package typing

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFillType_Basic(t *testing.T) {
	var got Type

	err := fillType(types.Typ[types.Int], &got)
	require.NoError(t, err)

	assert.Equal(t, TypeKindBasic, got.kind)
	assert.Equal(t, "int", got.name)
	assert.Empty(t, got.pkg)
	assert.Nil(t, got.elem)
	assert.Zero(t, got.size)
	assert.Empty(t, got.params)
}

func TestFillType_Named(t *testing.T) {
	pkg := types.NewPackage("github.com/example/foo", "foo")

	obj := types.NewTypeName(token.NoPos, pkg, "MyType", nil)
	named := types.NewNamed(obj, types.Typ[types.String], nil)

	var got Type

	err := fillType(named, &got)
	require.NoError(t, err)

	assert.Equal(t, TypeKindNamed, got.kind)
	assert.Equal(t, "MyType", got.name)
	assert.Equal(t, "github.com/example/foo", got.pkg)
	assert.Nil(t, got.elem)
}

func TestFillType_Pointer(t *testing.T) {
	pkg := types.NewPackage("github.com/example/foo", "foo")

	obj := types.NewTypeName(token.NoPos, pkg, "User", nil)
	named := types.NewNamed(obj, types.Typ[types.String], nil)

	ptr := types.NewPointer(named)

	var got Type

	err := fillType(ptr, &got)
	require.NoError(t, err)

	assert.Equal(t, TypeKindPointer, got.kind)
	require.NotNil(t, got.elem)

	assert.Equal(t, TypeKindNamed, got.elem.kind)
	assert.Equal(t, "User", got.elem.name)
	assert.Equal(t, "github.com/example/foo", got.elem.pkg)
}

func TestFillType_Slice(t *testing.T) {
	slice := types.NewSlice(types.Typ[types.Int])

	var got Type

	err := fillType(slice, &got)
	require.NoError(t, err)

	assert.Equal(t, TypeKindSlice, got.kind)
	require.NotNil(t, got.elem)

	assert.Equal(t, TypeKindBasic, got.elem.kind)
	assert.Equal(t, "int", got.elem.name)
}

func TestFillType_Array(t *testing.T) {
	array := types.NewArray(types.Typ[types.String], 5)

	var got Type

	err := fillType(array, &got)
	require.NoError(t, err)

	assert.Equal(t, TypeKindArray, got.kind)
	assert.EqualValues(t, 5, got.size)

	require.NotNil(t, got.elem)
	assert.Equal(t, TypeKindBasic, got.elem.kind)
	assert.Equal(t, "string", got.elem.name)
}

func TestFillType_Map(t *testing.T) {
	mapType := types.NewMap(
		types.Typ[types.String],
		types.Typ[types.Int],
	)

	var got Type

	err := fillType(mapType, &got)
	require.NoError(t, err)

	assert.Equal(t, TypeKindMap, got.kind)

	require.Len(t, got.params, 2)

	assert.Equal(t, TypeKindBasic, got.params[0].kind)
	assert.Equal(t, "string", got.params[0].name)

	assert.Equal(t, TypeKindBasic, got.params[1].kind)
	assert.Equal(t, "int", got.params[1].name)
}

func TestFillType_NestedRecursive(t *testing.T) {
	// []*[]int

	innerSlice := types.NewSlice(types.Typ[types.Int])
	ptr := types.NewPointer(innerSlice)
	outerSlice := types.NewSlice(ptr)

	var got Type

	err := fillType(outerSlice, &got)
	require.NoError(t, err)

	assert.Equal(t, TypeKindSlice, got.kind)

	require.NotNil(t, got.elem)
	assert.Equal(t, TypeKindPointer, got.elem.kind)

	require.NotNil(t, got.elem.elem)
	assert.Equal(t, TypeKindSlice, got.elem.elem.kind)

	require.NotNil(t, got.elem.elem.elem)
	assert.Equal(t, TypeKindBasic, got.elem.elem.elem.kind)
	assert.Equal(t, "int", got.elem.elem.elem.name)
}

func TestFillType_UnaliasesAliasChain(t *testing.T) {
	pkg := types.NewPackage("github.com/example/foo", "foo")

	// type Base int
	baseObj := types.NewTypeName(token.NoPos, pkg, "Base", nil)
	baseNamed := types.NewNamed(baseObj, types.Typ[types.Int], nil)

	// type Alias1 = Base
	alias1Obj := types.NewTypeName(token.NoPos, pkg, "Alias1", nil)
	alias1 := types.NewAlias(alias1Obj, baseNamed)

	// type Alias2 = Alias1
	alias2Obj := types.NewTypeName(token.NoPos, pkg, "Alias2", nil)
	alias2 := types.NewAlias(alias2Obj, alias1)

	var got Type

	err := fillType(alias2, &got)
	require.NoError(t, err)

	// Should resolve to first non-alias type.
	assert.Equal(t, TypeKindNamed, got.kind)
	assert.Equal(t, "Base", got.name)
	assert.Equal(t, "github.com/example/foo", got.pkg)
}

func TestFillType_MapWithComplexTypes(t *testing.T) {
	pkg := types.NewPackage("github.com/example/foo", "foo")

	obj := types.NewTypeName(token.NoPos, pkg, "User", nil)
	userType := types.NewNamed(obj, types.Typ[types.String], nil)

	mapType := types.NewMap(
		types.NewPointer(userType),
		types.NewSlice(types.Typ[types.Int]),
	)

	var got Type

	err := fillType(mapType, &got)
	require.NoError(t, err)

	assert.Equal(t, TypeKindMap, got.kind)
	require.Len(t, got.params, 2)

	key := got.params[0]
	require.NotNil(t, key)
	assert.Equal(t, TypeKindPointer, key.kind)
	require.NotNil(t, key.elem)
	assert.Equal(t, "User", key.elem.name)

	value := got.params[1]
	require.NotNil(t, value)
	assert.Equal(t, TypeKindSlice, value.kind)
	require.NotNil(t, value.elem)
	assert.Equal(t, "int", value.elem.name)
}

func TestFillType_GenericNamedType(t *testing.T) {
	pkg := types.NewPackage("github.com/example/foo", "foo")

	// type Box[T any] struct{}
	tParamObj := types.NewTypeName(token.NoPos, pkg, "T", nil)
	tParam := types.NewTypeParam(
		tParamObj,
		types.Universe.Lookup("any").Type(),
	)

	obj := types.NewTypeName(token.NoPos, pkg, "Box", nil)
	named := types.NewNamed(
		obj,
		types.NewStruct(nil, nil),
		nil,
	)
	named.SetTypeParams([]*types.TypeParam{tParam})

	// Box[string]
	inst, err := types.Instantiate(
		nil,
		named,
		[]types.Type{types.Typ[types.String]},
		false,
	)
	require.NoError(t, err)

	var got Type

	err = fillType(inst, &got)
	require.NoError(t, err)

	assert.Equal(t, TypeKindNamed, got.kind)
	assert.Equal(t, "Box", got.name)
	assert.Equal(t, "github.com/example/foo", got.pkg)

	require.Len(t, got.params, 1)

	assert.Equal(t, TypeKindBasic, got.params[0].kind)
	assert.Equal(t, "string", got.params[0].name)
}

func TestFillType_GenericSliceWrapper(t *testing.T) {
	pkg := types.NewPackage("github.com/example/foo", "foo")

	// type List[T any] []T

	tParamObj := types.NewTypeName(token.NoPos, pkg, "T", nil)
	tParam := types.NewTypeParam(
		tParamObj,
		types.Universe.Lookup("any").Type(),
	)

	obj := types.NewTypeName(token.NoPos, pkg, "List", nil)

	named := types.NewNamed(
		obj,
		types.NewSlice(tParam),
		nil,
	)

	named.SetTypeParams([]*types.TypeParam{tParam})

	// List[int]
	inst, err := types.Instantiate(
		nil,
		named,
		[]types.Type{types.Typ[types.Int]},
		false,
	)
	require.NoError(t, err)

	var got Type

	err = fillType(inst, &got)
	require.NoError(t, err)

	assert.Equal(t, TypeKindNamed, got.kind)
	assert.Equal(t, "List", got.name)

	require.Len(t, got.params, 1)
	assert.Equal(t, "int", got.params[0].name)
}

func TestFillType_NestedGenericInstantiation(t *testing.T) {
	pkg := types.NewPackage("github.com/example/foo", "foo")

	// type Pair[K, V any] struct{}

	kObj := types.NewTypeName(token.NoPos, pkg, "K", nil)
	vObj := types.NewTypeName(token.NoPos, pkg, "V", nil)

	kParam := types.NewTypeParam(
		kObj,
		types.Universe.Lookup("any").Type(),
	)

	vParam := types.NewTypeParam(
		vObj,
		types.Universe.Lookup("any").Type(),
	)

	obj := types.NewTypeName(token.NoPos, pkg, "Pair", nil)

	named := types.NewNamed(
		obj,
		types.NewStruct(nil, nil),
		nil,
	)

	named.SetTypeParams(
		[]*types.TypeParam{
			kParam,
			vParam,
		},
	)

	// Pair[string, []int]
	inst, err := types.Instantiate(
		nil,
		named,
		[]types.Type{
			types.Typ[types.String],
			types.NewSlice(types.Typ[types.Int]),
		},
		false,
	)
	require.NoError(t, err)

	var got Type

	err = fillType(inst, &got)
	require.NoError(t, err)

	assert.Equal(t, TypeKindNamed, got.kind)
	assert.Equal(t, "Pair", got.name)

	require.Len(t, got.params, 2)

	assert.Equal(t, "string", got.params[0].name)

	assert.Equal(t, TypeKindSlice, got.params[1].kind)
	require.NotNil(t, got.params[1].elem)
	assert.Equal(t, "int", got.params[1].elem.name)
}

func TestTypeString(t *testing.T) {
	cases := []struct {
		name  string
		_type *Type
		want  string
	}{
		{
			name:  "basic type",
			_type: Basic("string"),
			want:  "string",
		},
		{
			name:  "array of basic types",
			_type: Array(Basic("int"), 10),
			want:  "[10]int",
		},
		{
			name:  "slice of basic types",
			_type: Slice(Basic("string")),
			want:  "[]string",
		},
		{
			name:  "map of basic types",
			_type: Map(Basic("string"), Basic("int")),
			want:  "map[string]int",
		},
		{
			name:  "pointer to basic type",
			_type: Pointer(Basic("int")),
			want:  "*int",
		},
		{
			name:  "named type",
			_type: Named("github.com/example/foo", "Named"),
			want:  "github.com/example/foo.Named",
		},
		{
			name:  "named generic type with basic generic args",
			_type: Named("github.com/example/foo", "Generic", Basic("int"), Basic("string")),
			want:  "github.com/example/foo.Generic[int,string]",
		},
		{
			name: "named generic type with named generic args",
			_type: Named(
				"github.com/example/foo", "Generic",
				Named("github.com/example/bar", "Arg1"),
				Named("github.com/example/baz", "Arg2"),
			),
			want: "github.com/example/foo.Generic[github.com/example/bar.Arg1,github.com/example/baz.Arg2]",
		},
		{
			name: "named generic type with map and pointer to slice of pointers",
			_type: Named(
				"github.com/example/foo", "Generic",
				Map(Basic("int"), Named("github.com/example/bar", "MapVal")),
				Pointer(Slice(Pointer(Basic("string")))),
			),
			want: "github.com/example/foo.Generic[map[int]github.com/example/bar.MapVal,*[]*string]",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc._type.String()
			require.Equal(t, tc.want, got, "got unexpected result")
		})
	}
}
