package meta

import (
	"github.com/stretchr/testify/require"
	"go/types"
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
			name:     "Special case for echo.Map",
			input:    namedEchoMap,
			expected: "map[string]any",
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
