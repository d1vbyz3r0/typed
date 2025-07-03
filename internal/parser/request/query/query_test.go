package query

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func Test_NewInlineQuery(t *testing.T) {
	want := reflect.StructOf([]reflect.StructField{
		{
			Name:      "Q1",
			PkgPath:   "",
			Type:      reflect.TypeOf(""),
			Tag:       `query:"q1"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
		{
			Name:      "Q2",
			PkgPath:   "",
			Type:      reflect.TypeOf(int64(0)),
			Tag:       `query:"q2"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
		{
			Name:      "Q3",
			PkgPath:   "",
			Type:      reflect.TypeOf(uuid.UUID{}),
			Tag:       `query:"q3"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
		{
			Name:      "Q4",
			PkgPath:   "",
			Type:      reflect.TypeOf(false),
			Tag:       `query:"q4"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
	})

	src := `
package test

func Handler(c echo.Context) error {
	q1 := c.QueryParam("q1")
	q2, err := strconv.ParseInt(c.QueryParam("q2"), 10, 64)
	q3 := uuid.MustParse(c.QueryParam("q3"))
	q4, err := strconv.ParseBool(c.QueryParam("q4"))
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

		got, ok := NewInlineQuery(decl)
		require.True(t, ok)
		for i := 0; i < want.NumField(); i++ {
			wantField := want.Field(i)
			gotField := got.Field(i)
			require.Equal(t, wantField, gotField)
		}

		return true
	})
}
