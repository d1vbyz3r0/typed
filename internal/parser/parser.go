package parser

import (
	"fmt"
	"go/ast"

	"github.com/d1vbyz3r0/typed/common/meta"
	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser/enums"
	"github.com/d1vbyz3r0/typed/internal/parser/request"
	"github.com/d1vbyz3r0/typed/internal/parser/response"
	"github.com/d1vbyz3r0/typed/internal/parser/response/codes"
	"github.com/d1vbyz3r0/typed/internal/parser/response/mime"
	"github.com/d1vbyz3r0/typed/logging"
	"golang.org/x/tools/go/packages"
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
	Handlers []Handler
	// AdditionalModels will contain array of all type declarations and structs used in c.Bind() if ParseAllModels was provided as opt.
	// It can contain duplicates, it's up to you to deduplicate them.
	AdditionalModels []*typing.Type
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

// Parse parses package and returns all found enums and handlers
func (p *Parser) Parse(pkg *packages.Package, opts ...ParseOpt) (Result, error) {
	parseOpts := new(parserOpts)
	for _, opt := range opts {
		opt(parseOpts)
	}

	var result Result
	for _, file := range pkg.Syntax {
		if parseOpts.parseEnums {
			foundEnums, err := enums.Extract(pkg.Types, file, pkg.TypesInfo)
			if err != nil {
				return Result{}, fmt.Errorf("extract enums: %w", err)
			}
			result.AdditionalModels = append(result.AdditionalModels, foundEnums...)
		}

		ast.Inspect(file, func(n ast.Node) bool {
			decl, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if !isEchoHandler(decl) && !isWrapperFunction(decl) {
				return true
			}

			logging.Debug("found echo handler", "pkg", pkg, "filename", file.Name, "name", decl.Name.Name)

			req := request.New(decl, pkg.TypesInfo, parseOpts.RequestParseOpts()...)
			responses := response.NewStatusCodeMapping(decl, p.codesResolver, p.mimeResolver, pkg.TypesInfo)

			if parseOpts.parseAllModels {
				if req.ModelType != nil {
					result.AdditionalModels = append(result.AdditionalModels, req.ModelType)
				}

				for _, resp := range responses {
					for _, r := range resp {
						if r.ModelType == nil {
							continue
						}
						result.AdditionalModels = append(result.AdditionalModels, r.ModelType)
					}
				}
			}

			h := Handler{
				Doc:       meta.GetFuncDocumentation(decl),
				Name:      decl.Name.Name,
				Pkg:       pkg.PkgPath,
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
					logging.Warn("object not found in scope", "pkg", pkg.PkgPath, "name", name)
					continue
				}

				if !obj.Exported() {
					logging.Debug("skipping non-exported object, can't be a model", "pkg", pkg.PkgPath, "name", name)
					continue
				}

				if typing.IsFunc(obj.Type()) {
					logging.Debug("skipping function object, can't be a model", "pkg", pkg.PkgPath, "name", name)
					continue
				}

				if typing.IsInterface(obj.Type()) {
					logging.Debug("skipping interface object, can't be a model", "pkg", pkg.PkgPath, "name", name)
					continue
				}

				if typing.IsConstOrGlobal(obj) {
					logging.Debug("skipping const/global object, can't be a model", "pkg", pkg.PkgPath, "name", name)
					continue
				}

				model, err := typing.NewType(obj.Type())
				if err != nil {
					logging.Error("failed to create typing.Type from type object", "err", err)
					continue
				}

				result.AdditionalModels = append(result.AdditionalModels, model)
			}
		}
	}

	return result, nil
}
