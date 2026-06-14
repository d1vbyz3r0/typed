package generator

import (
	"fmt"
	"maps"
	"slices"

	"github.com/d1vbyz3r0/typed/common/meta"
	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser"
)

type importMapping struct {
	lastIdx int
	Alias   string
	Pkg     string
}

func createImportMappings(
	results []parser.Result,
	initialImports map[string]*importMapping,
) ([]*importMapping, error) {
	imports := maps.Clone(initialImports)
	if imports == nil {
		imports = make(map[string]*importMapping)
	}

	processType := func(n *typing.Type) {
		pkg := n.Pkg()
		if pkg == "" {
			return
		}
		processImport(pkg, imports)
	}

	for _, result := range results {
		for _, handler := range result.Handlers {
			processImport(handler.Pkg, imports)
			req := handler.Request
			if req != nil {
				if err := typing.Traverse(req.ModelType, processType); err != nil {
					return nil, fmt.Errorf("traverse %s: %w", req.ModelType, err)
				}
			}

			for _, responses := range handler.Responses {
				for _, resp := range responses {
					if err := typing.Traverse(resp.ModelType, processType); err != nil {
						return nil, fmt.Errorf("traverse %s: %w", resp.ModelType, err)
					}
				}
			}
		}

		for _, model := range result.AdditionalModels {
			if err := typing.Traverse(model, processType); err != nil {
				return nil, fmt.Errorf("traverse %s: %w", model, err)
			}
		}
	}

	res := slices.Collect(maps.Values(imports))
	slices.SortFunc(res, func(a, b *importMapping) int {
		if a.Alias < b.Alias {
			return -1
		} else if a.Alias > b.Alias {
			return 1
		}
		return 0
	})

	return res, nil
}

func initialMapping() map[string]*importMapping {
	// static imports, key is last pkg part, value is actual mapping
	return map[string]*importMapping{
		"typed":    {Alias: "typed", Pkg: "github.com/d1vbyz3r0/typed"},
		"typing":   {Alias: "typing", Pkg: "github.com/d1vbyz3r0/typed/common/typing"},
		"openapi3": {Alias: "openapi3", Pkg: "github.com/getkin/kin-openapi/openapi3"},
	}
}

func processImport(pkg string, imports map[string]*importMapping) {
	if pkg == "" {
		return
	}

	pkgName := meta.GetPkgName(pkg)
	if imp, ok := imports[pkgName]; ok {
		if imp.Pkg == pkg {
			return
		}

		alias := fmt.Sprintf("%s%d", pkgName, imp.lastIdx)
		for {
			prev, ok := imports[alias]
			if !ok {
				break
			}

			if prev.Pkg == pkg {
				// alias already created
				return
			}

			imp.lastIdx++
			alias = fmt.Sprintf("%s%d", pkgName, imp.lastIdx)
		}

		imports[alias] = &importMapping{
			Alias: alias,
			Pkg:   pkg,
		}
		return
	}

	imports[pkgName] = &importMapping{
		Alias:   pkgName,
		Pkg:     pkg,
		lastIdx: 1,
	}
}

func lookupAlias(imports []*importMapping, pkg string) (string, bool) {
	for _, imp := range imports {
		if imp.Pkg == pkg {
			return imp.Alias, true
		}
	}
	return "", false
}

func aliasNamer(imports []*importMapping) typing.NamerFunc {
	return func(t *typing.Type) (string, string) {
		pkg := t.Pkg()
		if pkg == "" {
			return "", t.Name()
		}

		alias, ok := lookupAlias(imports, pkg)
		if !ok {
			panic("alias not found for pkg: " + pkg)
		}

		return alias, t.Name()
	}
}
