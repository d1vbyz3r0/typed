package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TODO: refactor helper and move to testsuite package
func makeTestModule(t *testing.T) string {
	t.Helper()

	root := t.TempDir()

	writeFile(t, root, "go.mod", `
module example.com/project

go 1.20
`)

	writeFile(t, root, "internal/dto/dto.go", `
package dto

type User struct{}
type DeprecatedUser struct{}
`)

	writeFile(t, root, "internal/dto/auth/auth.go", `
package auth

type LoginRequest struct{}
type LoginResponse struct{}
`)

	writeFile(t, root, "internal/dto/auth/internal/internal.go", `
package internal

type InternalToken struct{}
`)

	writeFile(t, root, "internal/admin/admin.go", `
package admin

type Admin struct{}
`)

	return root
}

func writeFile(t *testing.T, root, name, content string) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(name))

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
