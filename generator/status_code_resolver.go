package generator

import (
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"log/slog"
	"strconv"
	"strings"
)

type statusCodeResolver struct {
	httpConstants map[string]int
}

func newStatusCodeResolver() *statusCodeResolver {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax,
	}

	pkgs, err := packages.Load(cfg, "net/http")
	if err != nil {
		slog.Warn("Failed to load http package", err)
		return &statusCodeResolver{}
	}

	constants := make(map[string]int)
	pkg := pkgs[0]

	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		if strings.HasPrefix(name, "Status") {
			obj := scope.Lookup(name)
			if constVal, ok := obj.(*types.Const); ok {
				val, _ := strconv.Atoi(constVal.Val().String())
				constants[name] = val
			}
		}
	}

	return &statusCodeResolver{
		httpConstants: constants,
	}
}

func (r *statusCodeResolver) resolve(expr ast.Expr, info *types.Info) int {
	// Handle direct integer literals (200, 404, etc)
	if lit, ok := expr.(*ast.BasicLit); ok {
		if lit.Kind == token.INT {
			if code, err := strconv.Atoi(lit.Value); err == nil {
				return code
			}
		}
	}

	// Handle http.Status* constants
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		if x, ok := sel.X.(*ast.Ident); ok {
			if x.Name == "http" {
				if code, exists := r.httpConstants[sel.Sel.Name]; exists {
					return code
				}
			}
		}
	}

	return 0
}
