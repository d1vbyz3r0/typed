package handlers

import (
	"fmt"
	"os"
	"path/filepath"
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

	cfg := &packages.Config{
		Mode: packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedFiles |
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

	fillCache(pkgs)

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
					logging.Debug("saved handler to map", "pkg", h.Pkg, "name", h.Name)
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
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get current working directory: %v", err)
	}

	logging.Debug("detected current working directory", cwd)

	res := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		p := pattern.Path
		if pattern.Recursive {
			p = filepath.Join(p, "...")
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(cwd, p)
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
	fun := runtime.FuncForPC(pc)
	if fun == nil {
		return ""
	}

	name := fun.Name()
	if strings.Contains(name, ".func") ||
		strings.Contains(name, ").") && strings.Contains(name, ".func") {
		logging.Debug("using fallback to packagePathFromFileLine since handler is closure", "name", name)
		if pkg := packagePathFromFileLine(fun, pc); pkg != "" {
			logging.Debug("resolved original closure package path", "pkg", pkg)
			return pkg
		}
	}

	return packagePathFromFuncName(name)
}

func packagePathFromFileLine(fun *runtime.Func, pc uintptr) string {
	file, _ := fun.FileLine(pc)
	absFile, _ := filepath.Abs(file)

	pkg, ok := lookupPkgByFile(absFile)
	if ok {
		logging.Debug("package cache hit", "file", absFile, "pkg", pkg.PkgPath)
		return pkg.PkgPath
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles,
		Dir:  filepath.Dir(file),
	}

	pkgs, err := packages.Load(cfg, ".")
	if packages.PrintErrors(pkgs) > 0 {
		return ""
	}

	if err != nil || len(pkgs) == 0 {
		return ""
	}

	for _, pkg := range pkgs {
		for _, f := range pkg.GoFiles {
			absGoFile, _ := filepath.Abs(f)
			if sameFile(absFile, absGoFile) {
				logging.Debug("package cache miss, adding", "file", absFile, "pkg", pkg.PkgPath)
				putToCache(absGoFile, pkg)
				return pkg.PkgPath
			}
		}
	}

	return ""
}

func packagePathFromFuncName(funcName string) string {
	// ex: github.com/acme/project/pkg/service.(*Handler).Serve
	lastSlash := strings.LastIndex(funcName, "/")
	lastDot := strings.Index(funcName[lastSlash+1:], ".")
	if lastDot == -1 {
		return ""
	}
	return funcName[:lastSlash+1+lastDot]
}
