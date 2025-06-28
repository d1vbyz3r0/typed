package mime

import (
	"fmt"
	"go/ast"
	"go/types"
	"golang.org/x/tools/go/packages"
	"strconv"
	"strings"

	_ "github.com/labstack/echo/v4"
)

type Resolver struct {
	contentTypes map[string]string
}

func NewResolver() (*Resolver, error) {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
	}

	pkgs, err := packages.Load(cfg, "github.com/labstack/echo/v4")
	if err != nil {
		return nil, fmt.Errorf("failed to load github.com/labstack/echo/v4 package: %w", err)
	}

	contentTypes := make(map[string]string)
	pkg := pkgs[0]

	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		if strings.HasPrefix(name, "MIME") {
			obj := scope.Lookup(name)
			if constVal, ok := obj.(*types.Const); ok {
				val := constVal.Val().String()
				contentTypes[name] = val
			}
		}
	}

	return &Resolver{contentTypes: contentTypes}, nil
}

func (r *Resolver) Resolve(expr ast.Expr) (string, error) {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return "", fmt.Errorf("expected selector expr, got %T", expr)
	}

	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", fmt.Errorf("expected identifier, got %T", sel)
	}

	if x.Name != "echo" {
		return "", fmt.Errorf("expected const from echo, got %s", x.Name)
	}

	ct, exists := r.contentTypes[sel.Sel.Name]
	if exists {
		res, err := strconv.Unquote(ct)
		if err != nil {
			return "", fmt.Errorf("unquote constant: %w", err)
		}

		return res, nil
	}

	return "", fmt.Errorf("content type not found for %s", sel.Sel.Name)
}
