package typing

import (
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func Test_IsParamUsage(t *testing.T) {
	cases := []struct {
		Name      string
		Src       string
		FuncName  string
		ParamName string
		Want      bool
	}{
		{
			Name: "FormValue with strconv.Atoi",
			Src: `
package test
func Test(c echo.Context) {
	val, _ := strconv.Atoi(c.FormValue("val"))
}
`,
			FuncName:  "FormValue",
			ParamName: "val",
			Want:      true,
		},
		{
			Name: "QueryParam with wrong uuid.MustParse usage",
			Src: `
package test
func Test(c echo.Context) {
	x := c.QueryParam("val")
	uuid.MustParse(x)
}
`,
			FuncName:  "QueryParam",
			ParamName: "val",
			Want:      false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", tc.Src, parser.AllErrors)
			require.NoError(t, err)

			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				isUsage := IsParamUsage(call, tc.FuncName, tc.ParamName)
				require.Equal(t, tc.Want, isUsage)
				return false
			})
		})
	}
}
