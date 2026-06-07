package testsuite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type Module struct {
	root string
}

// NewModule creates a temporary Go module for tests that exercise go/packages.
func NewModule(t testing.TB, modulePath string) *Module {
	t.Helper()

	m := &Module{root: t.TempDir()}
	m.Write(t, "go.mod", "module "+modulePath+"\n\ngo 1.25\n")
	return m
}

func (m *Module) Root() string {
	return m.root
}

func (m *Module) Path(name string) string {
	return filepath.Join(m.root, filepath.FromSlash(name))
}

func (m *Module) Write(t testing.TB, name string, content string) {
	t.Helper()

	path := m.Path(name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("create directory %s: %v", filepath.Dir(path), err)
	}

	content = strings.TrimSpace(content) + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
