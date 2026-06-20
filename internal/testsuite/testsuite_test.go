package testsuite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFixturePath(t *testing.T) {
	path := FixturePath(t, "parser/c1")
	if !filepath.IsAbs(path) {
		t.Fatalf("fixture path is not absolute: %s", path)
	}
	if _, err := os.Stat(filepath.Join(path, "c1.go")); err != nil {
		t.Fatalf("fixture file: %v", err)
	}
}

func TestLoadFixtureFunc(t *testing.T) {
	pkg, fn := LoadFixtureFunc(t, "parser/c1", "Handler")
	if pkg.PkgPath != "github.com/d1vbyz3r0/typed/testdata/parser/c1" {
		t.Fatalf("unexpected package path: %s", pkg.PkgPath)
	}
	if fn.Name.Name != "Handler" {
		t.Fatalf("unexpected function: %s", fn.Name.Name)
	}
}

func TestModule(t *testing.T) {
	module := NewModule(t, "example.com/test")
	module.Write(t, "sample/sample.go", `
package sample

type Value struct{}
`)

	pkg := LoadPackage(t, module.Path("sample"))
	if pkg.PkgPath != "example.com/test/sample" {
		t.Fatalf("unexpected package path: %s", pkg.PkgPath)
	}
}
