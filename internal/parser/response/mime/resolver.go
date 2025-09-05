package mime

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	_ "github.com/labstack/echo/v4"
	"golang.org/x/tools/go/packages"
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

// Resolve content type as string (ex: "text/plain") or as echo constant (ex: echo.MIMETextPlain)
func (r *Resolver) Resolve(expr ast.Expr) (string, error) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.STRING {
			return "", fmt.Errorf("expected string literal, got %s", e.Kind)
		}

		contentType, err := strconv.Unquote(e.Value)
		if err != nil {
			return "", fmt.Errorf("unquote: %w", err)
		}

		return contentType, nil

	case *ast.SelectorExpr:
		x, ok := e.X.(*ast.Ident)
		if !ok {
			return "", fmt.Errorf("expected identifier, got %T", e.X)
		}

		if x.Name != "echo" {
			return "", fmt.Errorf("expected const from echo, got %s", x.Name)
		}

		ct, exists := r.contentTypes[e.Sel.Name]
		if exists {
			res, err := strconv.Unquote(ct)
			if err != nil {
				return "", fmt.Errorf("unquote constant: %w", err)
			}
			return res, nil
		}

		return "", fmt.Errorf("no content type for %s", e.Sel.Name)

	default:
		return "", fmt.Errorf("unsupported expression type %T. Expected BasicLit or SelectorExpr", e)
	}
}
