package handlers

import (
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"
)

var (
	mtx   sync.RWMutex
	cache = make(map[string]*packages.Package)
)

func lookupPkgByFile(filePath string) (*packages.Package, bool) {
	mtx.RLock()
	defer mtx.RUnlock()

	k := makeCacheKey(filePath)
	v, ok := cache[k]
	return v, ok
}

func putToCache(filePath string, pkg *packages.Package) {
	mtx.Lock()
	defer mtx.Unlock()

	k := makeCacheKey(filePath)
	cache[k] = pkg
}

func fillCache(pkgs []*packages.Package) {
	mtx.Lock()
	defer mtx.Unlock()
	for _, pkg := range pkgs {
		for _, file := range pkg.GoFiles {
			k := makeCacheKey(file)
			cache[k] = pkg
		}
	}
}

func makeCacheKey(p string) string {
	filePath := filepath.ToSlash(filepath.Clean(p))
	if runtime.GOOS == "windows" {
		filePath = strings.ToLower(filePath)
	}
	return filePath
}

func sameFile(a string, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)

	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}

	return a == b
}
