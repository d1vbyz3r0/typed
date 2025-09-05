package mime

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestResolver_ResolveWithConst(t *testing.T) {
	src := `
package test

func Handler(c echo.Context) error {
	return c.Blob(http.StatusOK, echo.MIMETextPlain, nil)	
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	require.NoError(t, err)

	r, err := NewResolver()
	require.NoError(t, err)

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		contentType, err := r.Resolve(call.Args[1])
		require.NoError(t, err)
		require.Equal(t, echo.MIMETextPlain, contentType)
		return true
	})
}

func TestResolver_ResolveWithString(t *testing.T) {
	src := `
package test

func Handler(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/json", nil)	
}
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	require.NoError(t, err)

	r, err := NewResolver()
	require.NoError(t, err)

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		contentType, err := r.Resolve(call.Args[1])
		require.NoError(t, err)
		require.Equal(t, "application/json", contentType)
		return true
	})
}
