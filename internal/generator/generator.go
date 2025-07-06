package generator

import (
	"bytes"
	"fmt"
	"github.com/d1vbyz3r0/typed/internal/parser"
	"go/format"
	"golang.org/x/exp/maps"
	"golang.org/x/tools/go/packages"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	_ "embed"
)

//go:embed spec_gen.tmpl
var scriptTemplate string

type TemplateArgs struct {
	ApiPrefix              *string
	Types                  []string
	Imports                []string
	Enums                  map[string][]any
	Title                  string
	Version                string
	Servers                []Server
	HandlersPkgs           []HandlersConfig
	RoutesProviderCtor     string
	GenerateLib            bool
	LibPkg                 string
	SpecPath               string
	HandlerProcessingHooks []string
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

	return &Generator{
		cfg:    cfg,
		parser: p,
	}, nil
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

	slog.Debug("built packages load patterns", "patterns", patterns)

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}

	types := make(map[string]struct{})
	imports := map[string]struct{}{
		g.cfg.Input.RoutesProviderPkg: {},
	}
	enums := make(map[string][]any)

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, err := range pkg.Errors {
				slog.Error("failed to process package", "path", pkg.PkgPath, "error", err)
			}
			continue
		}

		res, err := g.parser.Parse(pkg, parser.ParseAllModels(), parser.ParseEnums())
		if err != nil {
			slog.Error("failed to parse package", "path", pkg.PkgPath)
			continue
		}

		for k, enum := range res.Enums {
			enums[k] = enum
		}

		for _, h := range res.Handlers {
			if h.Request != nil {
				if h.Request.BindModel != "" {
					types[h.Request.BindModel] = struct{}{}
					imports[h.Request.BindModelPkg] = struct{}{}
				}
			}

			for _, responses := range h.Responses {
				for _, resp := range responses {
					if resp.TypeName != "" {
						types[resp.TypeName] = struct{}{}
					}

					if resp.TypePkgPath != "" {
						imports[resp.TypePkgPath] = struct{}{}
					}
				}
			}

		}

		for _, model := range g.filterModels(res.AdditionalModels) {
			imports[model.PkgPath] = struct{}{}
			types[model.Name] = struct{}{}
		}
	}

	validImports := make([]string, 0)
	for imp := range imports {
		if imp != "" {
			validImports = append(validImports, imp)
		}
	}

	tmpl := template.Must(template.New("spec").Parse(scriptTemplate))
	result := bytes.NewBuffer(make([]byte, 0, len(scriptTemplate)))

	err = tmpl.Execute(result, TemplateArgs{
		ApiPrefix:              g.cfg.Input.ApiPrefix,
		Types:                  maps.Keys(types),
		Imports:                validImports,
		Enums:                  enums,
		Title:                  g.cfg.Input.Title,
		Version:                g.cfg.Input.Version,
		Servers:                g.cfg.Input.Servers,
		HandlersPkgs:           g.cfg.Input.Handlers,
		RoutesProviderCtor:     g.buildCtorCall(),
		GenerateLib:            g.cfg.GenerateLib,
		LibPkg:                 g.cfg.LibPkg,
		SpecPath:               g.cfg.Output.SpecPath,
		HandlerProcessingHooks: g.cfg.ProcessingHooks,
	})
	if err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	formatted, err := format.Source(result.Bytes())
	if err != nil {
		return fmt.Errorf("run formatter on generated code: %w", err)
	}

	path := filepath.Dir(g.cfg.Output.Path)
	err = os.MkdirAll(path, 0644)
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

func (g *Generator) filterModels(models []parser.Model) []parser.Model {
	res := make([]parser.Model, 0, len(models))
	hasIncludeFilter := len(g.cfg.Input.IncludeModels) > 0
	for _, model := range models {
		if hasIncludeFilter && !slices.Contains(g.cfg.Input.IncludeModels, model.Name) {
			continue
		}

		if slices.Contains(g.cfg.Input.ExcludeModels, model.Name) {
			continue
		}

		res = append(res, model)
	}

	return res
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

func (g *Generator) buildCtorCall() string {
	parts := strings.Split(g.cfg.Input.RoutesProviderPkg, "/")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1] + "." + g.cfg.Input.RoutesProviderCtor + "()"
}
