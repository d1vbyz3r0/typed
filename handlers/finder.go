package handlers

import (
	"fmt"
	"github.com/d1vbyz3r0/typed/internal/parser"
	"github.com/labstack/echo/v4"
	"golang.org/x/tools/go/packages"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"
)

type SearchPattern struct {
	Path      string
	Recursive bool
}

type Finder struct {
	parser   *parser.Parser
	handlers map[string]parser.Handler
}

func NewFinder() (*Finder, error) {
	p, err := parser.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	return &Finder{
		parser:   p,
		handlers: make(map[string]parser.Handler),
	}, nil
}

func (f *Finder) Find(patterns []SearchPattern) error {
	cfg := &packages.Config{
		Mode: packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedName,
	}

	pkgs, err := packages.Load(cfg, f.buildSearchPatterns(patterns)...)
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, err := range pkg.Errors {
				slog.Error("failed to process package", "path", pkg.PkgPath, "error", err)
			}
			continue
		}

		res, err := f.parser.Parse(pkg, parser.ParseInlineForms(), parser.ParseInlinePathParams(), parser.ParseInlineQueryParams())
		if err != nil {
			slog.Error("failed to parse package", "path", pkg.PkgPath)
			continue
		}

		for _, h := range res.Handlers {
			//fullHandlerPath := h.Pkg + "." + h.Name
			v, ok := f.handlers[h.Name]
			if ok {
				// workaround...
				slog.Warn(
					"handler already found in map, use unique names for your handlers",
					"old_pkg", v.Pkg,
					"new_pkg", h.Pkg,
					"handler", h.Name,
				)
			}

			f.handlers[h.Name] = h
		}
	}

	return nil
}

func (f *Finder) Match(routes []echo.Route) []Handler {
	res := make([]Handler, 0, len(routes))
	for _, route := range routes {
		//fullPath := f.getHandlerFullPath(route)
		handlerName := f.getHandlerName(route)
		h, ok := f.handlers[handlerName]
		if !ok {
			slog.Warn("matched handler not found, skipping", "handler", handlerName)
			continue
		}

		res = append(res, NewHandler(route, h))
	}

	return res
}

func (f *Finder) getHandlerFullPath(route echo.Route) string {
	// TODO: not working in case of direct echo.HandlerFunc usages, last pkg skipped :)
	closureRegexp := regexp.MustCompile(`^func\d+(\.\d+)?$`)
	slashParts := strings.Split(route.Name, "/")
	hasDots := strings.Contains(route.Name, ".")
	if len(slashParts) == 1 && !hasDots {
		return route.Name
	}

	pkgPath := strings.Join(slashParts[:len(slashParts)-1], "/")
	last := slashParts[len(slashParts)-1]

	dotIdx := strings.IndexByte(last, '.')
	if dotIdx != -1 {
		if pkgPath == "" {
			pkgPath = last[:dotIdx]
		} else {
			pkgPath += "/" + last[:dotIdx]
		}
	}

	dotParts := strings.Split(last, ".")

	var handler string
	if len(dotParts) >= 2 {
		lastPart := dotParts[len(dotParts)-1]
		secondLast := dotParts[len(dotParts)-2]
		if closureRegexp.MatchString(lastPart) {
			handler = secondLast
		} else {
			handler = lastPart
		}
	} else {
		handler = dotParts[0]
	}

	return pkgPath + "." + handler
}

func (f *Finder) getHandlerName(route echo.Route) string {
	// Example route name: "xxx/internal/api.(*Server).mapUsers.LoginUserHandler.func1"
	parts := strings.Split(route.Name, ".")

	// Get the package and handler name
	// For wrapper handlers (with .funcN suffix), we need to take the part before .funcN
	handlerName := parts[len(parts)-1]
	if strings.HasPrefix(handlerName, "func") {
		handlerName = parts[len(parts)-2]
	}

	return handlerName
}

func (f *Finder) buildSearchPatterns(patterns []SearchPattern) []string {
	res := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		p := pattern.Path
		if pattern.Recursive {
			p = filepath.Join(p, "...")
		}
		res = append(res, p)
	}

	return res
}
