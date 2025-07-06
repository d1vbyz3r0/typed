package response

import (
	"github.com/d1vbyz3r0/typed/internal/parser/response/codes"
	"github.com/d1vbyz3r0/typed/internal/parser/response/mime"
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"testing"
)

func TestContextResponseType_ContentType(t *testing.T) {
	cr, err := codes.NewResolver()
	require.NoError(t, err)

	mr, err := mime.NewResolver()
	require.NoError(t, err)

	type fields struct {
		src   string
		codes *codes.Resolver
		mime  *mime.Resolver
	}

	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "JSON",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.JSON(http.StatusOK, nil)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    "application/json",
			wantErr: false,
		},
		{
			name: "JSONPretty",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, nil, "")
}`,
				codes: cr,
				mime:  mr,
			},
			want:    "application/json",
			wantErr: false,
		},
		{
			name: "JSONBlob",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.JSONBlob(http.StatusOK, nil)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    "application/json",
			wantErr: false,
		},
		{
			name: "XML",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.XML(http.StatusOK, nil)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    "application/xml",
			wantErr: false,
		},
		{
			name: "XMLPretty",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.XMLPretty(http.StatusOK, nil, "")
}`,
				codes: cr,
				mime:  mr,
			},
			want:    "application/xml",
			wantErr: false,
		},
		{
			name: "XMLBlob",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.XMLBlob(http.StatusOK, nil)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    "application/xml",
			wantErr: false,
		},
		{
			name: "String",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.String(http.StatusOK, "")
}`,
				codes: cr,
				mime:  mr,
			},
			want:    "text/plain",
			wantErr: false,
		},
		{
			name: "Blob",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/json", nil)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    "application/json",
			wantErr: false,
		},
		{
			name: "Redirect",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.Redirect(301)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "NoContent",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Stream",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.Stream(http.StatusOK, "application/json", nil)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    "application/json",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", tt.fields.src, parser.AllErrors)
			require.NoError(t, err)
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				ct, supported := newContextResponseType(call, cr, mr, nil)
				if !supported {
					return true
				}

				contentType, err := ct.ContentType()
				require.NoError(t, err)
				require.Equal(t, tt.want, contentType)
				return true
			})
		})
	}
}

func TestContextResponseType_StatusCode(t *testing.T) {
	cr, err := codes.NewResolver()
	require.NoError(t, err)

	mr, err := mime.NewResolver()
	require.NoError(t, err)

	type fields struct {
		src   string
		codes *codes.Resolver
		mime  *mime.Resolver
	}

	tests := []struct {
		name    string
		fields  fields
		want    int
		wantErr bool
	}{
		{
			name: "JSON",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.JSON(http.StatusOK, nil)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    http.StatusOK,
			wantErr: false,
		},
		{
			name: "JSONPretty",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.JSONPretty(http.StatusOK, nil, "")
}`,
				codes: cr,
				mime:  mr,
			},
			want:    http.StatusOK,
			wantErr: false,
		},
		{
			name: "JSONBlob",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.JSONBlob(http.StatusOK, nil)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    http.StatusOK,
			wantErr: false,
		},
		{
			name: "XML",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.XML(http.StatusOK, nil)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    http.StatusOK,
			wantErr: false,
		},
		{
			name: "XMLPretty",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.XMLPretty(http.StatusOK, nil, "")
}`,
				codes: cr,
				mime:  mr,
			},
			want:    http.StatusOK,
			wantErr: false,
		},
		{
			name: "XMLBlob",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.XMLBlob(http.StatusOK, nil)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    http.StatusOK,
			wantErr: false,
		},
		{
			name: "String",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.String(http.StatusOK, "")
}`,
				codes: cr,
				mime:  mr,
			},
			want:    http.StatusOK,
			wantErr: false,
		},
		{
			name: "Blob",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/json", nil)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    http.StatusOK,
			wantErr: false,
		},
		{
			name: "Redirect",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.Redirect(http.StatusFound)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    http.StatusFound,
			wantErr: false,
		},
		{
			name: "NoContent",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    http.StatusOK,
			wantErr: false,
		},
		{
			name: "Stream",
			fields: fields{
				src: `
package test

func Handler(c echo.Context) error {
	return c.Stream(http.StatusOK, "application/json", nil)
}`,
				codes: cr,
				mime:  mr,
			},
			want:    http.StatusOK,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", tt.fields.src, parser.AllErrors)
			require.NoError(t, err)
			ast.Inspect(file, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				ct, supported := newContextResponseType(call, cr, mr, nil)
				if !supported {
					return true
				}

				code, err := ct.StatusCode()
				require.NoError(t, err)
				require.Equal(t, tt.want, code)
				return true
			})
		})
	}
}
