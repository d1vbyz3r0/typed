package codes

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"strconv"
	"strings"
)

type Resolver struct {
	codes map[string]int
}

func NewResolver() (*Resolver, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
	}

	pkgs, err := packages.Load(cfg, "net/http")
	if err != nil {
		return nil, fmt.Errorf("failed to load http package: %w", err)
	}

	statusCodes := make(map[string]int)
	pkg := pkgs[0]

	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		if strings.HasPrefix(name, "Status") {
			obj := scope.Lookup(name)
			if constVal, ok := obj.(*types.Const); ok {
				val, _ := strconv.Atoi(constVal.Val().String())
				statusCodes[name] = val
			}
		}
	}

	return &Resolver{codes: statusCodes}, nil
}

func (r *Resolver) Resolve(expr ast.Expr) (int, error) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.INT {
			return 0, fmt.Errorf("unexpected lit kind %s, expected INT", e.Kind)
		}

		code, err := strconv.Atoi(e.Value)
		if err != nil {
			return 0, fmt.Errorf("atoi: %w", err)
		}
		return code, nil

	case *ast.SelectorExpr:
		x, ok := e.X.(*ast.Ident)
		if !ok {
			return 0, fmt.Errorf("unexpected identifier %s, expected Ident", e.X)
		}

		if x.Name != "http" {
			return 0, fmt.Errorf("expected const from net/http, got %s", x.Name)
		}

		code, exists := r.codes[e.Sel.Name]
		if exists {
			return code, nil
		}

		return 0, fmt.Errorf("no code found for %s", e.Sel.Name)

	default:
		return 0, fmt.Errorf("unsupported expression type %T. Expected BasicLit or SelectorExpr", e)
	}
}
