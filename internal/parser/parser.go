package parser

import (
	"fmt"
	"github.com/d1vbyz3r0/typed/internal/common/meta"
	"github.com/d1vbyz3r0/typed/internal/parser/enums"
	"github.com/d1vbyz3r0/typed/internal/parser/request"
	"github.com/d1vbyz3r0/typed/internal/parser/response"
	"github.com/d1vbyz3r0/typed/internal/parser/response/codes"
	"github.com/d1vbyz3r0/typed/internal/parser/response/mime"
	"go/ast"
	"golang.org/x/tools/go/packages"
	"log/slog"
)

// isWrapperFunction checks if func has signature: func(...) echo.HandlerFunc {}
func isWrapperFunction(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) != 1 {
		return false
	}

	result := funcDecl.Type.Results.List[0].Type
	if sel, ok := result.(*ast.SelectorExpr); ok {
		if x, ok := sel.X.(*ast.Ident); ok {
			return x.Name == "echo" && sel.Sel.Name == "HandlerFunc"
		}
	}

	return false
}

// isEchoHandler checks if func has signature of echo.HandlerFunc
func isEchoHandler(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) != 1 {
		return false
	}

	if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) != 1 {
		return false
	}

	paramType := funcDecl.Type.Params.List[0].Type
	if sel, ok := paramType.(*ast.SelectorExpr); ok {
		if x, ok := sel.X.(*ast.Ident); ok {
			if x.Name != "echo" || sel.Sel.Name != "Context" {
				return false
			}
		}
	}

	result := funcDecl.Type.Results.List[0].Type
	if ident, ok := result.(*ast.Ident); ok {
		return ident.Name == "error"
	}

	return false
}

type Handler struct {
	Doc       string
	Name      string
	Pkg       string
	Request   *request.Request
	Responses response.StatusCodeMapping
}

type Result struct {
	Enums    map[string][]any
	Handlers []Handler
}

type Parser struct {
	codesResolver *codes.Resolver
	mimeResolver  *mime.Resolver
}

func New() (*Parser, error) {
	cr, err := codes.NewResolver()
	if err != nil {
		return nil, fmt.Errorf("create codes resolver: %v", err)
	}

	mr, err := mime.NewResolver()
	if err != nil {
		return nil, fmt.Errorf("create mime resolver: %v", err)
	}

	return &Parser{
		codesResolver: cr,
		mimeResolver:  mr,
	}, nil
}

func combineEnums(dst map[string][]any, src map[string][]any) {
	for k, v := range src {
		if _, ok := dst[k]; ok {
			dst[k] = append(dst[k], v...)
		} else {
			dst[k] = v
		}
	}
}

// Parse parses package and returns all found enums and handlers
func (p *Parser) Parse(pkg *packages.Package) (Result, error) {
	result := Result{
		Enums: make(map[string][]any),
	}

	for _, file := range pkg.Syntax {
		foundEnums, err := enums.Extract(pkg.Name, file)
		if err != nil {
			return Result{}, fmt.Errorf("extract enums: %w", err)
		}

		combineEnums(result.Enums, foundEnums)

		ast.Inspect(file, func(n ast.Node) bool {
			decl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if !isEchoHandler(decl) && !isWrapperFunction(decl) {
				return true
			}

			slog.Debug("found echo handler", "pkg", pkg, "file", file.Name.String(), "name", decl.Name.Name)

			h := Handler{
				Doc:       meta.GetFuncDocumentation(decl),
				Name:      decl.Name.Name,
				Pkg:       pkg.Name,
				Request:   request.New(decl, pkg.TypesInfo),
				Responses: response.NewStatusCodeMapping(decl, p.codesResolver, p.mimeResolver, pkg.TypesInfo),
			}

			result.Handlers = append(result.Handlers, h)
			return true
		})
	}

	return result, nil
}
