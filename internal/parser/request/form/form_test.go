package form

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"mime/multipart"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_NewInlineFormNoFiles(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	want := reflect.StructOf([]reflect.StructField{
		{
			Name:      "V1",
			PkgPath:   "",
			Type:      reflect.TypeOf(""),
			Tag:       `form:"v1"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
		{
			Name:      "V2",
			PkgPath:   "",
			Type:      reflect.TypeOf(0),
			Tag:       `form:"v2"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
	})

	src := `
package test

func Handler(c echo.Context) error {
	v1 := c.FormValue("v1")
	v2 := strconv.Atoi(c.FormValue("v2"))
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

		got, hasFiles, ok := NewInlineForm(decl)
		require.Equal(t, false, hasFiles)

		require.True(t, ok)
		for i := 0; i < want.NumField(); i++ {
			wantField := want.Field(i)
			gotField := got.Field(i)
			require.Equal(t, wantField, gotField)
		}

		return true
	})
}

func Test_NewInlineFormWithFiles(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	want := reflect.StructOf([]reflect.StructField{
		{
			Name:      "V1",
			PkgPath:   "",
			Type:      reflect.TypeOf(""),
			Tag:       `form:"v1"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
		{
			Name:      "V2",
			PkgPath:   "",
			Type:      reflect.TypeOf(0),
			Tag:       `form:"v2"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
		{
			Name:      "F1",
			PkgPath:   "",
			Type:      reflect.TypeOf(new(multipart.FileHeader)),
			Tag:       `form:"f1"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
		{
			Name:      "F2",
			PkgPath:   "",
			Type:      reflect.TypeOf(new(multipart.FileHeader)),
			Tag:       `form:"f2"`,
			Offset:    0,
			Index:     nil,
			Anonymous: false,
		},
	})

	src := `
package test

func Handler(c echo.Context) error {
	v1 := c.FormValue("v1")
	v2 := strconv.Atoi(c.FormValue("v2"))
	f1 := c.FormFile("f1")
	f2 := c.FormFile("f2")
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

		got, hasFiles, ok := NewInlineForm(decl)
		require.Equal(t, true, hasFiles)

		require.True(t, ok)
		for i := 0; i < want.NumField(); i++ {
			wantField := want.Field(i)
			gotField := got.Field(i)
			require.Equal(t, wantField, gotField)
		}

		return true
	})
}
