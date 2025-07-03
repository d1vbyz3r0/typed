package path

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func Test_NewInlinePathParams(t *testing.T) {
	want := reflect.StructOf([]reflect.StructField{
		{
			Name:      "P1",
			PkgPath:   "",
			Type:      reflect.TypeOf(""),
			Tag:       `path:"p1"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
		{
			Name:      "P2",
			PkgPath:   "",
			Type:      reflect.TypeOf(int64(0)),
			Tag:       `path:"p2"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
		{
			Name:      "P3",
			PkgPath:   "",
			Type:      reflect.TypeOf(uuid.UUID{}),
			Tag:       `path:"p3"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
	})

	src := `
package test

func Handler(c echo.Context) error {
	p1 := c.Param("p1")
	p2, err := strconv.ParseInt(c.Param("p2"), 10, 64)
	p3 := uuid.MustParse(c.Param("p3"))
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	require.NoError(t, err)

	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		got, ok := NewInlinePathParams(decl)
		require.True(t, ok)
		for i := 0; i < want.NumField(); i++ {
			wantField := want.Field(i)
			gotField := got.Field(i)
			require.Equal(t, wantField, gotField)
		}

		return true
	})
}
