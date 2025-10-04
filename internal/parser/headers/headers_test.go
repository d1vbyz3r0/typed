package headers

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"log/slog"
	"reflect"
	"testing"
)

func Test_NewInlineHeaders(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	want := []Header{
		{
			Name:     "h1",
			Type:     reflect.TypeOf(""),
			Required: false,
		},
		{
			Name:     "h2",
			Type:     reflect.TypeOf(int64(0)),
			Required: false,
		},
		{
			Name:     "h3",
			Type:     reflect.TypeOf(uuid.UUID{}),
			Required: false,
		},
		{
			Name:     "h4",
			Type:     reflect.TypeOf(false),
			Required: false,
		},
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedName,
	}, "../../../testdata/request/headers")
	require.NoError(t, err)

	file := pkgs[0].Syntax[0]
	typesInfo := pkgs[0].TypesInfo

	ast.Inspect(file, func(n ast.Node) bool {
		decl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		got := NewInlineHeaders(decl, typesInfo)
		require.ElementsMatch(t, want, got)
		return true
	})
}

func Test_NewStructQueryParams(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	type Struct struct {
		Name string `header:"name"`
		Age  int    `header:"age"`
		Opt  *bool  `header:"opt"`
	}

	got, err := NewStructHeaders(reflect.TypeOf(Struct{}))
	require.NoError(t, err)

	want := []Header{
		{
			Name:     "name",
			Type:     reflect.TypeOf(""),
			Required: true,
		},
		{
			Name:     "age",
			Type:     reflect.TypeOf(int(0)),
			Required: true,
		},
		{
			Name:     "opt",
			Type:     reflect.TypeOf(new(bool)),
			Required: false,
		},
	}

	require.ElementsMatch(t, want, got)
}
