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

type Model struct {
	Name    string
	PkgPath string
}

type Result struct {
	Enums    map[string][]any
	Handlers []Handler
	// AdditionalModels will contain array of all type declarations and structs used in c.Bind(), if ParseAllModels was provided as opt.
	//It can contain duplicates, it's up to you to deduplicate them
	AdditionalModels []Model
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
func (p *Parser) Parse(pkg *packages.Package, opts ...ParseOpt) (Result, error) {
	parseOpts := new(parserOpts)
	for _, opt := range opts {
		opt(parseOpts)
	}

	result := Result{
		Enums: make(map[string][]any),
	}

	for _, file := range pkg.Syntax {
		if parseOpts.parseEnums {
			foundEnums, err := enums.Extract(pkg.Name, file)
			if err != nil {
				return Result{}, fmt.Errorf("extract enums: %w", err)
			}
			combineEnums(result.Enums, foundEnums)
		}

		ast.Inspect(file, func(n ast.Node) bool {
			decl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if !isEchoHandler(decl) && !isWrapperFunction(decl) {
				return true
			}

			slog.Debug("found echo handler", "pkg", pkg, "file", file.Name, "name", decl.Name.Name)

			req := request.New(decl, pkg.TypesInfo, parseOpts.RequestParseOpts()...)
			responses := response.NewStatusCodeMapping(decl, p.codesResolver, p.mimeResolver, pkg.TypesInfo)

			if parseOpts.parseAllModels {
				result.AdditionalModels = append(result.AdditionalModels, Model{
					Name:    req.BindModel,
					PkgPath: req.BindModelPkg,
				})

				for _, resp := range responses {
					for _, r := range resp {
						result.AdditionalModels = append(result.AdditionalModels, Model{
							Name:    r.TypeName,
							PkgPath: r.TypePkgPath,
						})
					}
				}
			}

			h := Handler{
				Doc:       meta.GetFuncDocumentation(decl),
				Name:      decl.Name.Name,
				Pkg:       pkg.Name,
				Request:   req,
				Responses: responses,
			}

			result.Handlers = append(result.Handlers, h)
			return true
		})

		if parseOpts.parseAllModels {
			scope := pkg.Types.Scope()
			for _, name := range scope.Names() {
				obj := scope.Lookup(name)
				if obj == nil {
					slog.Warn("object not found in scope", "pkg", pkg.PkgPath, "name", name)
					continue
				}

				if !obj.Exported() {
					slog.Debug("skipping non-exported object", "pkg", pkg.PkgPath, "name", name)
					continue
				}

				pkgName := obj.Pkg().Name()
				pkgPath := obj.Pkg().Path()
				result.AdditionalModels = append(result.AdditionalModels, Model{
					Name:    pkgName + "." + name,
					PkgPath: pkgPath,
				})
			}
		}
	}

	return result, nil
}
