package handlers

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/d1vbyz3r0/typed/internal/parser"
	"github.com/labstack/echo/v4"
	"golang.org/x/tools/go/packages"
)

type SearchPattern struct {
	Path      string
	Recursive bool
}

type EchoRoute struct {
	Route       echo.Route
	Middlewares []echo.MiddlewareFunc
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

func (f *Finder) Find(patterns []SearchPattern, opts ...FinderOpt) error {
	findOpts := newFinderOpts()
	for _, opt := range opts {
		opt(findOpts)
	}

	cfg := &packages.Config{
		Mode: packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedName,
	}

	sp, err := f.buildSearchPatterns(patterns)
	if err != nil {
		return fmt.Errorf("build search patterns: %w", err)
	}

	slog.Debug("loaded search patterns", "patterns", sp)

	pkgs, err := packages.Load(cfg, sp...)
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}

	var (
		mtx sync.Mutex
		wg  sync.WaitGroup
		sem = make(chan struct{}, findOpts.concurrency)
	)

	for _, pkg := range pkgs {
		wg.Add(1)
		sem <- struct{}{}
		go func(pkg *packages.Package) {
			defer func() {
				wg.Done()
				<-sem
			}()

			if len(pkg.Errors) > 0 {
				for _, err := range pkg.Errors {
					slog.Error("failed to process package", "path", pkg.PkgPath, "error", err)
				}
				return
			}

			res, err := f.parser.Parse(pkg, parser.ParseInlineForms(), parser.ParseInlinePathParams(), parser.ParseInlineQueryParams())
			if err != nil {
				slog.Error("failed to parse package", "path", pkg.PkgPath)
				return
			}

			for _, h := range res.Handlers {
				//fullHandlerPath := h.Pkg + "." + h.Name
				mtx.Lock()
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
				mtx.Unlock()
			}
		}(pkg)
	}

	wg.Wait()
	return nil
}

func (f *Finder) Match(routes []EchoRoute) []Handler {
	res := make([]Handler, 0, len(routes))
	for _, route := range routes {
		//fullPath := f.getHandlerFullPath(route)
		handlerName := f.getHandlerName(route.Route)
		h, ok := f.handlers[handlerName]
		if !ok {
			slog.Warn("matched handler not found, skipping", "handler", handlerName)
			continue
		}

		res = append(res, NewHandler(route.Route, route.Middlewares, h))
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

	// struct methods seems to have format <func-name>-fm
	idx := strings.Index(handlerName, "-")
	if idx != -1 {
		handlerName = handlerName[:idx]
	}

	return handlerName
}

func (f *Finder) buildSearchPatterns(patterns []SearchPattern) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get current working directory: %v", err)
	}

	res := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		p := pattern.Path
		if pattern.Recursive {
			p = filepath.Join(p, "...")
		}
		res = append(res, filepath.Join(cwd, p))
	}

	return res, nil
}
