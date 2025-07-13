package enums

import (
	"fmt"
	"go/ast"
	"go/token"
	"log/slog"
	"strconv"
)

func Extract(pkg string, file *ast.File) (map[string][]any, error) {
	res := make(map[string][]any)
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
								return nil, err
							}

							k := pkg + "." + ident.Name
							res[k] = append(res[k], lit)
							slog.Debug("added enum type", "name", k)
						}
					}

				case *ast.CallExpr:
					if funIdent, ok := v.Fun.(*ast.Ident); ok && len(v.Args) == 1 {
						if lit, ok := v.Args[0].(*ast.BasicLit); ok {
							parsedLit, err := parseLiteral(lit.Value)
							if err != nil {
								return nil, err
							}

							k := pkg + "." + funIdent.Name
							res[k] = append(res[k], parsedLit)
							slog.Debug("added enum type", "name", k)
						}
					}
				}
			}
		}
	}

	return res, nil
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

	if b, err := strconv.ParseBool(lit); err == nil {
		return b, nil
	}

	return nil, fmt.Errorf("unknown literal: %s", lit)
}
