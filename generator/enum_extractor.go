package generator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

type enumExtractor struct {
	Enums map[string][]any
}

func newEnumExtractor() *enumExtractor {
	return &enumExtractor{
		Enums: make(map[string][]any),
	}
}

func (e *enumExtractor) extractFromFile(pkgName string, filename string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}

		for _, spec := range genDecl.Specs {
			valSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			exprCount := len(valSpec.Values)
			for i := range valSpec.Names {
				var expr ast.Expr
				if exprCount == 1 {
					expr = valSpec.Values[0]
				} else if i < exprCount {
					expr = valSpec.Values[i]
				}

				if expr == nil {
					continue
				}

				switch v := expr.(type) {
				case *ast.BasicLit:
					if valSpec.Type != nil {
						if ident, ok := valSpec.Type.(*ast.Ident); ok {
							lit, err := parseLiteral(v.Value)
							if err != nil {
								return err
							}

							k := pkgName + "." + ident.Name
							e.Enums[k] = append(e.Enums[k], lit)
						}
					}

				case *ast.CallExpr:
					if funIdent, ok := v.Fun.(*ast.Ident); ok && len(v.Args) == 1 {
						if lit, ok := v.Args[0].(*ast.BasicLit); ok {
							parsedLit, err := parseLiteral(lit.Value)
							if err != nil {
								return err
							}

							k := pkgName + "." + funIdent.Name
							e.Enums[k] = append(e.Enums[k], parsedLit)
						}
					}
				}
			}
		}
	}

	return nil
}

func parseLiteral(lit string) (any, error) {
	if i, err := strconv.Atoi(lit); err == nil {
		return i, nil
	}
	if f, err := strconv.ParseFloat(lit, 64); err == nil {
		return f, nil
	}
	if s, err := strconv.Unquote(lit); err == nil {
		return s, nil
	}
	return nil, fmt.Errorf("unknown literal: %s", lit)
}
