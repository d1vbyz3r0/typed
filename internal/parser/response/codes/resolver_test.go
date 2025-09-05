package codes

import (
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
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

		code, err := r.Resolve(call.Args[0])
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, code)
		return true
	})
}
