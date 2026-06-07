package response

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"net/http"
	"reflect"
	"testing"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser/headers"
	"github.com/d1vbyz3r0/typed/internal/parser/response/codes"
	"github.com/d1vbyz3r0/typed/internal/parser/response/mime"
	"github.com/d1vbyz3r0/typed/internal/testsuite"
	"github.com/stretchr/testify/require"
)

func TestStatusCodeMapping_extractResponses(t *testing.T) {
	cr, err := codes.NewResolver()
	require.NoError(t, err)

	mr, err := mime.NewResolver()
	require.NoError(t, err)

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
						ModelType:   typing.Named("github.com/d1vbyz3r0/typed/testdata/handlers", "Example"),
					},
				},
				http.StatusBadRequest: []Response{
					{
						ContentType: "application/json",
						ModelType: typing.Slice(typing.Map(
							typing.Basic("int"),
							typing.Named("github.com/d1vbyz3r0/typed/testdata/handlers", "Example"),
						)),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := testsuite.LoadFixturePackage(t, "handlers")
			fn := testsuite.Func(t, pkg, "Handler")
			mapping := NewStatusCodeMapping(fn, cr, mr, pkg.TypesInfo)
			require.Equal(t, len(tt.want), len(mapping))
			for status, want := range tt.want {
				require.ElementsMatch(t, want, mapping[status])
			}
		})
	}
}

func parseFunc(t *testing.T, src string) (*ast.FuncDecl, *types.Info, token.Pos) {
	t.Helper()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	_, err = conf.Check("test", fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatalf("types check: %v", err)
	}

	var fn *ast.FuncDecl
	for _, decl := range f.Decls {
		if fdecl, ok := decl.(*ast.FuncDecl); ok && fdecl.Name.Name == "Handler" {
			fn = fdecl
			break
		}
	}
	if fn == nil {
		t.Fatal("Handler not found")
	}

	var returnPos token.Pos
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if ret, ok := n.(*ast.ReturnStmt); ok {
			returnPos = ret.Pos()
			return false
		}
		return true
	})
	if returnPos == token.NoPos {
		t.Fatal("return not found")
	}

	return fn, info, returnPos
}

func TestFindHeaders(t *testing.T) {
	src := `
package test
import "net/http"

type Resp struct {
	Header http.Header
}
type Ctx struct{}
func (c *Ctx) Response() *Resp { return &Resp{} }

func Handler(c *Ctx) error {
	c.Response().Header.Set("X-Test", "v")
	return nil
}
`
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	conf := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	_, err = conf.Check("test", fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatalf("types check: %v", err)
	}

	var fn *ast.FuncDecl
	for _, decl := range f.Decls {
		if fdecl, ok := decl.(*ast.FuncDecl); ok && fdecl.Name.Name == "Handler" {
			fn = fdecl
			break
		}
	}
	if fn == nil {
		t.Fatal("Handler not found")
	}

	var returnPos token.Pos
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if ret, ok := n.(*ast.ReturnStmt); ok {
			returnPos = ret.Pos()
			return false
		}
		return true
	})
	if returnPos == token.NoPos {
		t.Fatal("return not found")
	}

	got := findHeaders(fn, returnPos, info)
	want := []headers.Header{
		{Name: "X-Test", Type: stringType, Required: false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected result:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestFindHeaders_MultipleCases(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		expect []headers.Header
	}{
		{
			name: "basic literal",
			code: `
package test
import "net/http"
type Resp struct{ Header http.Header }
type Ctx struct{}
func (c *Ctx) Response() *Resp { return &Resp{} }
func Handler(c *Ctx) error {
	c.Response().Header.Set("X-Test", "v")
	return nil
}
`,
			expect: []headers.Header{
				{Name: "X-Test", Type: stringType, Required: false},
			},
		},
		{
			name: "const header name",
			code: `
package test
import "net/http"
const XHeader = "X-Const"
type Resp struct{ Header http.Header }
type Ctx struct{}
func (c *Ctx) Response() *Resp { return &Resp{} }
func Handler(c *Ctx) error {
	c.Response().Header.Add(XHeader, "a")
	return nil
}
`,
			expect: []headers.Header{
				{Name: "X-Const", Type: stringType, Required: false},
			},
		},
		{
			name: "multiple headers before return",
			code: `
package test
import "net/http"
type Resp struct{ Header http.Header }
type Ctx struct{}
func (c *Ctx) Response() *Resp { return &Resp{} }
func Handler(c *Ctx) error {
	c.Response().Header.Set("X-One", "a")
	c.Response().Header.Add("X-Two", "b")
	return nil
}
`,
			expect: []headers.Header{
				{Name: "X-One", Type: stringType, Required: false},
				{Name: "X-Two", Type: stringType, Required: false},
			},
		},
		{
			name: "header inside if branch",
			code: `
package test
import "net/http"
type Resp struct{ Header http.Header }
type Ctx struct{}
func (c *Ctx) Response() *Resp { return &Resp{} }
func Handler(c *Ctx) error {
	if true {
		c.Response().Header.Set("X-If", "1")
	}
	c.Response().Header.Add("X-End", "2")
	return nil
}
`,
			expect: []headers.Header{
				{Name: "X-If", Type: stringType, Required: false},
				{Name: "X-End", Type: stringType, Required: false},
			},
		},
		{
			name: "duplicate header calls",
			code: `
package test
import "net/http"
type Resp struct{ Header http.Header }
type Ctx struct{}
func (c *Ctx) Response() *Resp { return &Resp{} }
func Handler(c *Ctx) error {
	c.Response().Header.Set("X-Same", "1")
	c.Response().Header.Set("X-Same", "2")
	return nil
}
`,
			expect: []headers.Header{
				{Name: "X-Same", Type: stringType, Required: false},
				{Name: "X-Same", Type: stringType, Required: false},
			},
		},
		{
			name: "headers in multiple return paths",
			code: `
package test
import "net/http"
type Resp struct{ Header http.Header }
type Ctx struct{}
func (c *Ctx) Response() *Resp { return &Resp{} }
func Handler(c *Ctx) error {
	if false {
		c.Response().Header.Set("X-400", "a")
		return nil
	}
	c.Response().Header.Add("X-200", "b")
	return nil
}
`,
			expect: []headers.Header{
				{Name: "X-400", Type: stringType, Required: false},
				{Name: "X-200", Type: stringType, Required: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, info, returnPos := parseFunc(t, tt.code)
			got := findHeaders(fn, returnPos, info)
			if !reflect.DeepEqual(got, tt.expect) {
				t.Errorf("unexpected headers:\n got: %#v\nwant: %#v", got, tt.expect)
			}
		})
	}
}
