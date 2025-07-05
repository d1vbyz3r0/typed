package handlers

import (
	"fmt"
	"github.com/d1vbyz3r0/typed/internal/parser"
	"github.com/labstack/echo/v4"
	"golang.org/x/tools/go/packages"
	"log/slog"
	"regexp"
	"strings"
)

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

// TODO: pass packages patterns to load
func (f *Finder) Find() error {
	cfg := &packages.Config{
		Mode: packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo,
	}

	pkgs, err := packages.Load(cfg)
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
			fullHandlerPath := h.Pkg + "." + h.Name
			f.handlers[fullHandlerPath] = h
		}
	}

	return nil
}

func (f *Finder) Match(routes []echo.Route) []Handler {
	res := make([]Handler, 0, len(routes))
	for _, route := range routes {
		fullPath := f.getHandlerFullPath(route)
		h, ok := f.handlers[fullPath]
		if !ok {
			slog.Warn("matched handler not found, skipping", "path", fullPath)
			continue
		}

		res = append(res, Handler{
			route:   route,
			handler: h,
		})
	}

	return res
}

func (f *Finder) getHandlerFullPath(route echo.Route) string {
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
