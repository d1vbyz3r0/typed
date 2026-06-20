package generator

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/d1vbyz3r0/typed/common/typing"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/packages"
)

type typeFilter struct {
	include []compiledEntry
	exclude []compiledEntry
}

// newTypeFilter compiles all filter entries from the config.
// baseImportPath is the resolved import path for ModelConfig.Path.
func newTypeFilter(cfg ModelsConfig) (*typeFilter, error) {
	include, err := compileEntries(cfg.IncludeModels)
	if err != nil {
		return nil, fmt.Errorf("compile include filters: %w", err)
	}
	exclude, err := compileEntries(cfg.ExcludeModels)
	if err != nil {
		return nil, fmt.Errorf("compile exclude filters: %w", err)
	}
	return &typeFilter{
		include: include,
		exclude: exclude,
	}, nil
}

// matchesAny returns true if the type matches at least one entry (OR semantics).
func (f *typeFilter) matchesAny(t *typing.Type, entries []compiledEntry, basePkg string) bool {
	for i := range entries {
		if entries[i].matches(t, basePkg) {
			return true
		}
	}
	return false
}

// Allow returns true if the type should be processed.
//
// Logic:
//  1. If include is non-empty then type must match at least one include entry.
//  2. If type matches any exclude entry then should be rejected.
func (f *typeFilter) Allow(t *typing.Type, basePkg string) bool {
	if len(f.include) > 0 && !f.matchesAny(t, f.include, basePkg) {
		return false
	}
	return !f.matchesAny(t, f.exclude, basePkg)
}

type filterSet struct {
	entries []filterSetEntry
}

type filterSetEntry struct {
	basePkg   string
	recursive bool
	filter    *typeFilter
}

func (e *filterSetEntry) contains(pkgPath string) bool {
	if pkgPath == e.basePkg {
		return true
	}
	return e.recursive && strings.HasPrefix(pkgPath, e.basePkg+"/")
}

func newFilterSet(
	models []ModelsConfig,
	resolveImportPath func(path string) (string, error),
) (*filterSet, error) {
	entries := make([]filterSetEntry, 0, len(models))
	for _, m := range models {
		basePkg, err := resolveImportPath(m.Path)
		if err != nil {
			return nil, fmt.Errorf("resolve import path for %s: %w", m.Path, err)
		}

		f, err := newTypeFilter(m)
		if err != nil {
			return nil, fmt.Errorf("model %s: %w", m.Path, err)
		}

		entries = append(entries, filterSetEntry{
			basePkg:   basePkg,
			recursive: m.Recursive,
			filter:    f,
		})
	}

	return &filterSet{entries: entries}, nil
}

// Allow returs true if provided type should be processed
// Filter is chosen based on longest basePkg (longest prefix match) to work proerply with nested paths
// If no ModelConfig pretends on type - it will be skipped (not ours type)
func (fs *filterSet) Allow(t *typing.Type) bool {
	entry := fs.findEntry(t.Pkg())
	if entry == nil {
		// not ours or basic, so allowed
		return true
	}
	return entry.filter.Allow(t, entry.basePkg)
}

func (fs *filterSet) FilterTypes(types []*typing.Type) []*typing.Type {
	result := make([]*typing.Type, 0, len(types))
	for _, t := range types {
		if t.Kind() == typing.TypeKindBasic {
			result = append(result, t)
			continue
		}

		if fs.Allow(t) {
			result = append(result, t)
		}
	}
	return result
}

// findEntry searches for longest basePkg, which is prefix for pkgPath
func (fs *filterSet) findEntry(pkgPath string) *filterSetEntry {
	var best *filterSetEntry
	for i := range fs.entries {
		e := &fs.entries[i]
		if !e.contains(pkgPath) {
			continue
		}

		if best == nil || len(e.basePkg) > len(best.basePkg) {
			best = e
		}
	}
	return best
}

type compiledEntry struct {
	name       *regexp.Regexp // matches Type.Name()
	importPath *regexp.Regexp // matches Type.Pkg()
	pkg        *regexp.Regexp // matches path.Base(Type.Pkg())
	relPath    *regexp.Regexp // matches Type.Pkg() relative to baseImportPath
}

// matches returns true if ALL non-nil fields match the type (AND semantics).
func (e *compiledEntry) matches(t *typing.Type, baseImportPath string) bool {
	if e.name != nil && !e.name.MatchString(t.Name()) {
		return false
	}
	if e.importPath != nil && !e.importPath.MatchString(t.Pkg()) {
		return false
	}
	if e.pkg != nil && !e.pkg.MatchString(path.Base(t.Pkg())) {
		return false
	}
	if e.relPath != nil {
		rel := strings.TrimPrefix(t.Pkg(), baseImportPath)
		rel = strings.TrimPrefix(rel, "/")
		if !e.relPath.MatchString(rel) {
			return false
		}
	}
	return true
}

func compileEntries(entries []ModelFilter) ([]compiledEntry, error) {
	result := make([]compiledEntry, len(entries))
	for i, e := range entries {
		ce, err := compileEntry(e)
		if err != nil {
			return nil, fmt.Errorf("entry[%d]: %w", i, err)
		}
		result[i] = ce
	}
	return result, nil
}

func compileEntry(e ModelFilter) (compiledEntry, error) {
	var (
		ce  compiledEntry
		err error
	)

	if e.Name != "" {
		if ce.name, err = regexp.Compile(e.Name); err != nil {
			return ce, fmt.Errorf("name %q: %w", e.Name, err)
		}
	}
	if e.ImportPath != "" {
		if ce.importPath, err = regexp.Compile(e.ImportPath); err != nil {
			return ce, fmt.Errorf("import-path %q: %w", e.ImportPath, err)
		}
	}
	if e.Pkg != "" {
		if ce.pkg, err = regexp.Compile(e.Pkg); err != nil {
			return ce, fmt.Errorf("pkg %q: %w", e.Pkg, err)
		}
	}
	if e.Path != "" {
		if ce.relPath, err = regexp.Compile(e.Path); err != nil {
			return ce, fmt.Errorf("path %q: %w", e.Path, err)
		}
	}
	return ce, nil
}

func getFullPkgPath(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("get absolute path: %w", err)
	}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles,
		Dir:  abs,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return moduleImportPath(abs)
	}

	if packagesContainErrors(pkgs) {
		return moduleImportPath(abs)
	}

	if len(pkgs) == 0 {
		return moduleImportPath(abs)
	}

	if len(pkgs) > 1 {
		return "", fmt.Errorf("expected 1 package, got %d", len(pkgs))
	}

	return pkgs[0].PkgPath, nil
}

func packagesContainErrors(pkgs []*packages.Package) bool {
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return true
		}
	}
	return false
}

func moduleImportPath(dir string) (string, error) {
	moduleRoot, modulePath, err := findModule(dir)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(moduleRoot, dir)
	if err != nil {
		return "", fmt.Errorf("get path relative to module root: %w", err)
	}
	if rel == "." {
		return modulePath, nil
	}
	return modulePath + "/" + filepath.ToSlash(rel), nil
}

func findModule(dir string) (string, string, error) {
	for {
		modPath := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(modPath)
		if err == nil {
			modulePath := modfile.ModulePath(data)
			if modulePath == "" {
				return "", "", fmt.Errorf("module path not found in %s", modPath)
			}
			return dir, modulePath, nil
		}
		if !os.IsNotExist(err) {
			return "", "", fmt.Errorf("read %s: %w", modPath, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", "", fmt.Errorf("go.mod not found for %s", dir)
		}
		dir = parent
	}
}
