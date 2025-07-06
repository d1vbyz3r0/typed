package query

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"reflect"
	"testing"
)

func Test_NewInlineQuery(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	want := []Param{
		{
			Name: "q1",
			Type: reflect.TypeOf(""),
		},
		{
			Name: "q2",
			Type: reflect.TypeOf(int64(0)),
		},
		{
			Name: "q3",
			Type: reflect.TypeOf(uuid.UUID{}),
		},
		{
			Name: "q4",
			Type: reflect.TypeOf(false),
		},
	}

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

		got := NewInlineQueryParams(decl)
		require.ElementsMatch(t, want, got)
		return true
	})
}

func Test_NewStructQueryParams(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	type Struct struct {
		Name string `query:"name"`
		Age  int    `query:"age"`
	}

	got, err := NewStructQueryParams(reflect.TypeOf(Struct{}))
	require.NoError(t, err)

	want := []Param{
		{
			Name: "name",
			Type: reflect.TypeOf(""),
		},
		{
			Name: "age",
			Type: reflect.TypeOf(int(0)),
		},
	}

	require.ElementsMatch(t, want, got)
}
