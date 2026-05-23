package testsuite

import (
	"testing"

	"golang.org/x/tools/go/packages"
)

func LoadPackage(t *testing.T, dir string) *packages.Package {
	t.Helper()

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedImports,
		Dir: dir,
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
