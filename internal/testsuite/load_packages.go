package testsuite

import (
	"go/ast"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/packages"
)

const DefaultLoadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedSyntax |
	packages.NeedTypes |
	packages.NeedTypesInfo |
	packages.NeedImports

// Root returns the repository module root.
func Root(t testing.TB) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve testsuite source path")
	}

	dir := filepath.Dir(filename)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}

// FixturePath returns an absolute path below the repository testdata directory.
func FixturePath(t testing.TB, name string) string {
	t.Helper()
	root := Root(t)
	return filepath.Join(root, "testdata", filepath.FromSlash(name))
}

// LoadFixturePackage loads one package from the repository testdata directory.
func LoadFixturePackage(t testing.TB, name string) *packages.Package {
	t.Helper()
	return LoadPackage(t, FixturePath(t, name))
}

// LoadFixtureFunc loads a fixture package and returns a named function.
func LoadFixtureFunc(t testing.TB, fixture string, name string) (*packages.Package, *ast.FuncDecl) {
	t.Helper()

	pkg := LoadFixturePackage(t, fixture)
	return pkg, Func(t, pkg, name)
}

func LoadPackage(t testing.TB, dir string) *packages.Package {
	t.Helper()

	cfg := &packages.Config{
		Mode: DefaultLoadMode,
		Dir:  dir,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		t.Fatalf("packages.Load: %v", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		t.Fatalf("package load errors")
	}

	if len(pkgs) != 1 {
		t.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	return pkgs[0]
}

// Func returns a named function declaration from a loaded package.
func Func(t testing.TB, pkg *packages.Package, name string) *ast.FuncDecl {
	t.Helper()

	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if ok && fn.Name.Name == name {
				return fn
			}
		}
	}

	t.Fatalf("function %q not found in package %s", name, pkg.PkgPath)
	return nil
}
