package path

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

func Test_NewInlinePathParams(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	want := []Param{
		{
			Name: "p1",
			Type: reflect.TypeOf(""),
		},
		{
			Name: "p2",
			Type: reflect.TypeOf(int64(0)),
		},
		{
			Name: "p3",
			Type: reflect.TypeOf(uuid.UUID{}),
		},
	}

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

		got := NewInlinePathParams(decl)
		require.ElementsMatch(t, want, got)

		return true
	})
}

func Test_NewStructPathParams(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	type Struct struct {
		Name string `param:"name"`
		Age  int    `param:"age"`
	}

	got := NewStructPathParams(reflect.TypeOf(Struct{}))
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
