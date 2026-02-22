package meta

import (
	"github.com/stretchr/testify/require"
	"go/types"
	"reflect"
	"testing"
)

func TestGetTypeName(t *testing.T) {
	type testCase struct {
		name        string
		input       types.Type
		expected    string
		expectError bool
	}

	echoPkg := types.NewPackage("github.com/labstack/echo/v4", "echo")
	customPkg := types.NewPackage("mypkg", "mypkg")

	namedEchoMap := types.NewNamed(types.NewTypeName(0, echoPkg, "Map", nil), nil, nil)
	namedMyStruct := types.NewNamed(types.NewTypeName(0, customPkg, "MyStruct", nil), nil, nil)

	tests := []testCase{
		{
			name:     "Pointer to named type",
			input:    types.NewPointer(namedMyStruct),
			expected: "*mypkg.MyStruct",
		},
		{
			name:     "Double pointer to named type",
			input:    types.NewPointer(types.NewPointer(namedMyStruct)),
			expected: "**mypkg.MyStruct",
		},
		{
			name:     "Slice of named type",
			input:    types.NewSlice(namedMyStruct),
			expected: "[]mypkg.MyStruct",
		},
		{
			name:        "Slice of pointers",
			input:       types.NewSlice(types.NewPointer(namedMyStruct)),
			expected:    "[]*mypkg.MyStruct",
			expectError: false,
		},
		{
			name:        "Slice of maps with pointer value",
			input:       types.NewSlice(types.NewMap(types.Typ[types.String], types.NewPointer(namedMyStruct))),
			expected:    "[]map[string]*mypkg.MyStruct",
			expectError: false,
		},
		{
			name:     "Map of named types",
			input:    types.NewMap(types.Typ[types.Int], namedMyStruct),
			expected: "map[int]mypkg.MyStruct",
		},
		{
			name:     "echo.Map",
			input:    namedEchoMap,
			expected: "echo.Map",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := GetTypeName(tc.input)
			if tc.expectError {
				require.Error(t, err)
				t.Logf("Expected error: %v", err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestGetPkgPath(t *testing.T) {
	echoPkg := types.NewPackage("github.com/labstack/echo/v4", "echo")
	customPkg := types.NewPackage("mypkg/types", "types")

	namedEchoMap := types.NewNamed(types.NewTypeName(0, echoPkg, "Map", nil), nil, nil)
	namedMyStruct := types.NewNamed(types.NewTypeName(0, customPkg, "MyStruct", nil), nil, nil)

	cases := []struct {
		Name      string
		Type      types.Type
		Pkg       string
		WantError bool
	}{
		{
			Name:      "Third party package",
			Type:      namedEchoMap,
			Pkg:       "github.com/labstack/echo/v4",
			WantError: false,
		},
		{
			Name:      "custom package",
			Type:      namedMyStruct,
			Pkg:       "mypkg/types",
			WantError: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			p, err := GetPkgPath(tc.Type)
			if err != nil && !tc.WantError {
				t.Fatalf("Unexpected error: %v", err)
			}
			require.Equal(t, tc.Pkg, p)
		})
	}
}

func TestGetTypeName_InstantiatedGeneric(t *testing.T) {
	constraint := types.NewInterfaceType(nil, nil)
	constraint.Complete()

	pkgDTO := types.NewPackage("example.com/dto", "dto")
	pkgShared := types.NewPackage("example.com/shared", "shared")

	sharedUser := types.NewNamed(types.NewTypeName(0, pkgShared, "User", nil), nil, nil)

	page := types.NewNamed(
		types.NewTypeName(0, pkgDTO, "Page", nil),
		types.NewStruct(nil, nil),
		nil,
	)
	page.SetTypeParams([]*types.TypeParam{
		types.NewTypeParam(types.NewTypeName(0, nil, "T", nil), constraint),
	})

	inst, err := types.Instantiate(nil, page, []types.Type{sharedUser}, false)
	require.NoError(t, err)

	got, err := GetTypeName(inst)
	require.NoError(t, err)
	require.Equal(t, "dto.Page[shared.User]", got)
}

func TestGetPkgPaths(t *testing.T) {
	constraint := types.NewInterfaceType(nil, nil)
	constraint.Complete()

	pkgDTO := types.NewPackage("example.com/dto", "dto")
	pkgShared := types.NewPackage("example.com/shared", "shared")

	sharedUser := types.NewNamed(types.NewTypeName(0, pkgShared, "User", nil), nil, nil)

	page := types.NewNamed(
		types.NewTypeName(0, pkgDTO, "Page", nil),
		types.NewStruct(nil, nil),
		nil,
	)
	page.SetTypeParams([]*types.TypeParam{
		types.NewTypeParam(types.NewTypeName(0, nil, "T", nil), constraint),
	})

	inst, err := types.Instantiate(nil, page, []types.Type{sharedUser}, false)
	require.NoError(t, err)

	got := GetPkgPaths(types.NewMap(types.Typ[types.String], inst))
	require.True(t, reflect.DeepEqual([]string{"example.com/dto", "example.com/shared"}, got), "unexpected pkg paths: %#v", got)
}
