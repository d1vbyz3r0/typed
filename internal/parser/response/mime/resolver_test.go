package mime

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestResolver_Resolve(t *testing.T) {
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

		contentType := r.Resolve(call.Args[1])
		require.Equal(t, echo.MIMETextPlain, contentType)
		return true
	})
}
