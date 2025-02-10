package generator

import (
	"fmt"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	_ "embed"
)

//go:embed spec_builder.tmpl
var specBuilderTemplate string

type TemplateArgs struct {
	Types        map[string]string
	Imports      []string
	Title        string
	Version      string
	ServerUrl    string
	HandlersPkgs []HandlersConfig
	ServerInit   string
	ServerPkg    string
	SpecPath     string
}

// SearchPatterns is map, where key is packages pattern and val is regex search pattern
type SearchPatterns map[string]*regexp.Regexp

func (sp SearchPatterns) Patterns() []string {
	var patterns []string
	for pattern := range sp {
		patterns = append(patterns, pattern)
	}

	return patterns
}

type Generator struct {
	cfg *Config
}

func New(cfg *Config) *Generator {
	return &Generator{
		cfg: cfg,
	}
}

func (g *Generator) Generate() error {
	cfg := &packages.Config{
		Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo,
	}

	sp, err := g.buildPatterns()
	if err != nil {
		return fmt.Errorf("build patterns: %w", err)
	}

	pkgs, err := packages.Load(cfg, sp.Patterns()...)
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}

	_types := make(map[string]string)
	imports := make(map[string]struct{})

	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			pattern, ok := sp[pkg.PkgPath]
			if ok && pattern != nil {
				matched := pattern.MatchString(name)
				if !matched {
					continue
				}
			}

			obj := scope.Lookup(name)
			if _, ok := obj.Type().Underlying().(*types.Struct); ok {
				path := obj.Pkg().Path()
				if _, ok := imports[path]; !ok {
					imports[path] = struct{}{}
				}

				pkgName := obj.Pkg().Name()

				// Key format: "package_name.TypeName"
				typeKey := pkgName + "." + name
				_types[typeKey] = pkgName + "." + obj.Name()
			}
		}

		if pkg.Errors != nil && len(pkg.Errors) > 0 {
			err := fmt.Errorf("%w", pkg.Errors[0])
			for i := 1; i < len(pkg.Errors); i++ {
				err = fmt.Errorf("%w: %w", err, pkg.Errors[i])
			}

			return fmt.Errorf("analyze package %s: %w", pkg.String(), err)
		}
	}

	// Filter out empty imports
	validImports := make([]string, 0)
	for imp := range imports {
		if imp != "" {
			validImports = append(validImports, imp)
		}
	}

	tmpl := template.Must(template.New("spec").Parse(specBuilderTemplate))
	f, err := os.Create(g.cfg.Output.Path)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, TemplateArgs{
		Title:        g.cfg.Input.Title,
		Version:      g.cfg.Input.Version,
		ServerUrl:    g.cfg.Input.ServerUrl,
		Types:        _types,
		Imports:      validImports,
		HandlersPkgs: g.cfg.Input.Handlers,
		ServerInit:   g.buildCtorCall(),
		ServerPkg:    g.cfg.Input.RoutesProviderPkg,
		SpecPath:     g.cfg.Output.SpecPath,
	})
}

func (g *Generator) buildPatterns() (SearchPatterns, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get current working directory: %v", err)
	}

	patterns := make(SearchPatterns)
	for _, m := range g.cfg.Input.Models {
		pattern := m.Path
		if m.Recursive {
			pattern = filepath.Join(cwd, pattern, "...")
		}

		var r *regexp.Regexp
		if m.Filter != nil {
			r, err = regexp.Compile(*m.Filter)
			if err != nil {
				return nil, fmt.Errorf("bad regex: %w", err)
			}
		}

		patterns[pattern] = r
	}

	return patterns, nil
}

func (g *Generator) buildCtorCall() string {
	parts := strings.Split(g.cfg.Input.RoutesProviderPkg, "/")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1] + "." + g.cfg.Input.RoutesProviderCtor + "()"
}
