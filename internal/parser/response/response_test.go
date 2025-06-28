package response

import (
	"github.com/d1vbyz3r0/typed/internal/parser/response/codes"
	"github.com/d1vbyz3r0/typed/internal/parser/response/mime"
	"github.com/stretchr/testify/require"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"net/http"
	"testing"
)

func TestStatusCodeMapping_extractResponses(t *testing.T) {
	cr, err := codes.NewResolver()
	require.NoError(t, err)

	mr, err := mime.NewResolver()
	require.NoError(t, err)

	t1 := "handlers.Example"
	t2 := "map[string]any"
	t3 := "[]map[int]handlers.Example"

	tests := []struct {
		name string
		want StatusCodeMapping
	}{
		{
			name: "json response",
			want: StatusCodeMapping{
				http.StatusOK: []Response{
					{
						ContentType: "application/json",
						TypeName:    &t1,
					},
					{
						ContentType: "application/xml",
						TypeName:    &t2,
					},
				},
				http.StatusBadRequest: []Response{
					{
						ContentType: "application/json",
						TypeName:    &t3,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkgs, err := packages.Load(&packages.Config{
				Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
			}, "../../../testdata/handlers/")
			require.NoError(t, err)

			pkg := pkgs[0]
			for _, file := range pkg.Syntax {
				ast.Inspect(file, func(n ast.Node) bool {
					funcDecl, ok := n.(*ast.FuncDecl)
					if !ok {
						return true
					}

					if funcDecl.Name.Name != "Handler" {
						return true
					}

					mapping := NewStatusCodeMapping(funcDecl, cr, mr, pkg.TypesInfo)
					for k, v := range mapping {
						want := tt.want[k]
						require.ElementsMatch(t, want, v)
					}

					return true
				})
			}
		})
	}
}
