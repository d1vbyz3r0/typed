package generator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"text/template"

	"github.com/d1vbyz3r0/typed/common/meta"
	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser"
	"github.com/d1vbyz3r0/typed/logging"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"

	_ "embed"
)

//go:embed spec_gen.tmpl
var scriptTemplate string

type TemplateArgs struct {
	ApiPrefix              *string
	Types                  []*typing.Type
	Imports                []*importMapping
	Title                  string
	Version                string
	Servers                []Server
	HandlersPkgs           []HandlersConfig
	RoutesProviderCtorName string
	RoutesProviderPkgAlias string
	PackageName            string
	IsMain                 bool
	SpecPath               string
	HandlerProcessingHooks []string
	Concurrency            int
	AliasNamer             typing.NamerFunc
	Debug                  bool
}

type Generator struct {
	cfg    Config
	parser *parser.Parser
}

func New(cfg Config) (*Generator, error) {
	p, err := parser.New()
	if err != nil {
		return nil, fmt.Errorf("init parser: %w", err)
	}

	g := &Generator{
		cfg:    cfg,
		parser: p,
	}

	return g, nil
}

func (g *Generator) Generate() error {
	cfg := &packages.Config{
		Mode: packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedName,
	}

	patterns, err := g.buildLoadPatterns()
	if err != nil {
		return fmt.Errorf("build load patterns: %w", err)
	}

	logging.Debug("built packages load patterns", "patterns", patterns)

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("failed to load one or more packages")
	}

	if g.cfg.Concurrency <= 0 {
		g.cfg.Concurrency = len(pkgs)
	}

	var (
		mtx     sync.Mutex
		eg      errgroup.Group
		results []parser.Result
	)

	eg.SetLimit(g.cfg.Concurrency)

	for _, pkg := range pkgs {
		eg.Go(func() error {
			// TODO: determine if should parse all models and enums based on current package path and filters
			res, err := g.parser.Parse(pkg, parser.ParseAllModels(), parser.ParseEnums())
			if err != nil {
				return fmt.Errorf("parse pkg %s: %w", pkg.PkgPath, err)
			}

			mtx.Lock()
			results = append(results, res)
			mtx.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	_imports, _types, err := g.processParserResults(results)
	if err != nil {
		return fmt.Errorf("process parser results: %w", err)
	}

	return g.execTemplate(_imports, _types)
}

func (g *Generator) processParserResults(results []parser.Result) ([]*importMapping, []*typing.Type, error) {
	// required to guarantee order of results between runs, so import mappings become stable
	slices.SortFunc(results, func(a, b parser.Result) int {
		if a.PkgPath < b.PkgPath {
			return -1
		} else if a.PkgPath > b.PkgPath {
			return 1
		}
		return 0
	})

	initialImports := initialMapping()
	if g.cfg.Output.IsMain() {
		processImport("github.com/d1vbyz3r0/typed/handlers", initialImports)
		processImport("log/slog", initialImports)
		processImport("os", initialImports)
		processImport(g.cfg.Input.RoutesProviderPkg, initialImports)
		if g.cfg.Debug {
			processImport("github.com/d1vbyz3r0/typed/logging", initialImports)
		}
	}

	_imports, err := createImportMappings(results, initialImports)
	if err != nil {
		return nil, nil, fmt.Errorf("create import mappings: %w", err)
	}

	_types, err := collectTypes(results)
	if err != nil {
		return nil, nil, fmt.Errorf("collect types: %w", err)
	}

	_types, err = g.filterModels(_types)
	if err != nil {
		return nil, nil, fmt.Errorf("filter models: %w", err)
	}

	return _imports, _types, nil
}

func (g *Generator) execTemplate(_imports []*importMapping, _types []*typing.Type) error {
	resolveAlias := aliasNamer(_imports)
	tmpl := template.Must(template.
		New("spec").
		Funcs(map[string]any{
			"typeToString":     typing.ToString,
			"typeTreeToString": typing.TypeTreeToString,
			"lastSegment":      meta.GetPkgName,
			"resolveAlias": func(t *typing.Type) string {
				pkg, _ := resolveAlias(t)
				return pkg
			},
		}).
		Parse(scriptTemplate),
	)

	var routesProviderPkgAlias string
	if g.cfg.Output.IsMain() {
		var ok bool
		routesProviderPkgAlias, ok = lookupAlias(_imports, g.cfg.Input.RoutesProviderPkg)
		if !ok {
			return fmt.Errorf("alias for routes provider package not found in imports mapping")
		}
	}

	var result bytes.Buffer
	err := tmpl.Execute(&result, TemplateArgs{
		ApiPrefix:              g.cfg.Input.ApiPrefix,
		Types:                  _types,
		Imports:                _imports,
		Title:                  g.cfg.Input.Title,
		Version:                g.cfg.Input.Version,
		Servers:                g.cfg.Input.Servers,
		HandlersPkgs:           g.cfg.Input.Handlers,
		RoutesProviderCtorName: g.cfg.Input.RoutesProviderCtor,
		RoutesProviderPkgAlias: routesProviderPkgAlias,
		PackageName:            g.cfg.Output.Package(),
		IsMain:                 g.cfg.Output.IsMain(),
		SpecPath:               g.cfg.Output.SpecPath,
		HandlerProcessingHooks: g.cfg.ProcessingHooks,
		Concurrency:            g.cfg.Concurrency,
		AliasNamer:             resolveAlias,
		Debug:                  g.cfg.Debug,
	})
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	src := result.Bytes()
	formatted, err := imports.Process("generated.go", src, nil)
	if err != nil {
		return fmt.Errorf("run formatter on generated code: %w", err)
	}

	path := filepath.Dir(g.cfg.Output.Path)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	f, err := os.Create(g.cfg.Output.Path)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(formatted)
	if err != nil {
		return fmt.Errorf("write output file: %w", err)
	}

	return nil
}

func (g *Generator) filterModels(models []*typing.Type) ([]*typing.Type, error) {
	filterSet, err := newFilterSet(g.cfg.Input.Models, getFullPkgPath)
	if err != nil {
		return nil, fmt.Errorf("new filter set: %w", err)
	}
	allowed := filterSet.FilterTypes(models)
	return allowed, nil
}

func (g *Generator) buildLoadPatterns() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get current working directory: %v", err)
	}

	var patterns []string
	for _, h := range g.cfg.Input.Handlers {
		pattern := h.Path
		if h.Recursive {
			pattern = filepath.Join(cwd, pattern, "...")
		}
		patterns = append(patterns, pattern)
	}

	for _, m := range g.cfg.Input.Models {
		pattern := m.Path
		if m.Recursive {
			pattern = filepath.Join(cwd, pattern, "...")
		}
		patterns = append(patterns, pattern)
	}

	return patterns, nil
}
