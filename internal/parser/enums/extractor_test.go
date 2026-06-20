package enums

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/stretchr/testify/assert"
)

func parseAndCheck(t *testing.T, src string) (*types.Package, *ast.File, *types.Info) {
	t.Helper()

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse file: %v", err)
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	conf := types.Config{}
	pkg, err := conf.Check("test", fset, []*ast.File{file}, info)
	if err != nil {
		t.Fatalf("type check: %v", err)
	}

	return pkg, file, info
}

func TestExtractEnums(t *testing.T) {
	src := `
package test

type Role string

const (
	RoleAdmin = Role("admin")
	RoleUser  = Role("user")
	RoleGuest = Role("guest")
)

type Status int

const (
	StatusNew  Status = 1
	StatusDone Status = 2
)
`

	pkg, file, _types := parseAndCheck(t, src)
	enums, err := Extract(pkg, file, _types)
	if err != nil {
		t.Fatalf("ExtractEnums: %v", err)
	}

	if len(enums) != 2 {
		t.Fatalf("expected 2 enums, got %d: %#v", len(enums), enums)
	}

	got := map[string]*typing.Type{}
	for _, enum := range enums {
		got[enum.Name()] = enum
	}

	assertEnum(t, got["Role"], "Role", []any{"admin", "user", "guest"})
	assertEnum(t, got["Status"], "Status", []any{int64(1), int64(2)})
}

func TestExtractEnums_IgnoresUntypedAndNonEnumConsts(t *testing.T) {
	src := `
package test

const Plain = "plain"

type Role string

const (
	RoleAdmin = Role("admin")
	Other     = "not enum"
)

type Config struct{}
`

	pkg, file, _types := parseAndCheck(t, src)
	enums, err := Extract(pkg, file, _types)
	if err != nil {
		t.Fatalf("ExtractEnums: %v", err)
	}

	if len(enums) != 1 {
		t.Fatalf("expected 1 enum, got %d: %#v", len(enums), enums)
	}

	assertEnum(t, enums[0], "Role", []any{"admin"})
}

func TestExtractEnums_SupportsImplicitConstType(t *testing.T) {
	src := `
package test

type Status int

const (
	StatusNew Status = iota
	StatusDone
	StatusArchived
)
`

	pkg, file, _types := parseAndCheck(t, src)
	enums, err := Extract(pkg, file, _types)
	if err != nil {
		t.Fatalf("ExtractEnums: %v", err)
	}

	if len(enums) != 1 {
		t.Fatalf("expected 1 enum, got %d: %#v", len(enums), enums)
	}

	assertEnum(t, enums[0], "Status", []any{int64(0), int64(1), int64(2)})
}

func assertEnum(t *testing.T, typ *typing.Type, name string, values []any) {
	t.Helper()

	if typ == nil {
		t.Fatalf("enum %q not found", name)
	}

	if typ.Kind() != typing.TypeKindEnum {
		t.Fatalf("%s: expected kind enum, got %v", name, typ.Kind())
	}

	if typ.Name() != name {
		t.Fatalf("expected enum name %q, got %q", name, typ.Name())
	}

	got := typ.EnumValues()
	if len(got) != len(values) {
		t.Fatalf("%s: expected values %#v, got %#v", name, values, got)
	}

	for i := range values {
		assert.Equal(t, values[i], got[i], "%s: value[%d]: expected %#v, got %#v", name, i, values[i], got[i])
	}
}
