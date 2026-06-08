package handlers

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/d1vbyz3r0/typed/internal/parser"
	"github.com/d1vbyz3r0/typed/logging"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/go/packages"
)

type SearchPattern struct {
	Path      string
	Recursive bool
}

type EchoRoute struct {
	Route       echo.Route
	HandlerFunc echo.HandlerFunc
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

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current working directory: %v", err)
	}

	logging.Debug("detected current working directory", cwd)

	cfg := &packages.Config{
		Dir: cwd,
		Mode: packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedName,
	}

	sp, err := f.buildSearchPatterns(patterns)
	if err != nil {
		return fmt.Errorf("build search patterns: %w", err)
	}

	logging.Debug("loaded search patterns", "patterns", sp)

	pkgs, err := packages.Load(cfg, sp...)
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}

	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("failed to load packages")
	}

	var (
		mtx sync.Mutex
		eg  errgroup.Group
	)

	eg.SetLimit(findOpts.concurrency)

	for _, pkg := range pkgs {
		eg.Go(func() error {
			res, err := f.parser.Parse(
				pkg,
				parser.ParseInlineForms(),
				parser.ParseInlinePathParams(),
				parser.ParseInlineQueryParams(),
				parser.ParseInlineHeaders(),
			)
			if err != nil {
				return fmt.Errorf("failed to parse pkg %s: %w", pkg.PkgPath, err)
			}

			for _, h := range res.Handlers {
				key := h.Pkg + "." + h.Name
				mtx.Lock()
				if _, ok := f.handlers[key]; !ok {
					f.handlers[key] = h
				}
				mtx.Unlock()
			}
			return nil
		})
	}

	return eg.Wait()
}

func (f *Finder) Match(routes []EchoRoute) []Handler {
	res := make([]Handler, 0, len(routes))
	for _, route := range routes {
		handlerName := f.getHandlerName(route.Route)
		handlerPkg := funcPackagePath(route.HandlerFunc)
		key := handlerPkg + "." + handlerName
		h, ok := f.handlers[key]
		if !ok {
			logging.Warn("matched handler not found, skipping", "pkg", handlerPkg, "handler", handlerName)
			continue
		}
		res = append(res, NewHandler(route.Route, route.Middlewares, h))
	}

	return res
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
	res := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		p := pattern.Path
		if pattern.Recursive {
			p = path.Join(p, "...")
			if p[0] != '.' {
				p = "./" + p
			}
		}
		res = append(res, p)
	}

	return res, nil
}

func funcPackagePath(fn any) string {
	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		panic(fmt.Sprintf("not a function: %s", v.Kind()))
	}

	pc := v.Pointer()
	f := runtime.FuncForPC(pc)
	if f == nil {
		return ""
	}

	// ex: github.com/acme/project/pkg/service.(*Handler).Serve
	name := f.Name()
	lastSlash := strings.LastIndex(name, "/")
	lastDot := strings.Index(name[lastSlash+1:], ".")
	if lastDot == -1 {
		return ""
	}
	return name[:lastSlash+1+lastDot]
}
