package generator

import (
	"path/filepath"
	"testing"

	"github.com/d1vbyz3r0/typed/internal/testsuite"
	"github.com/stretchr/testify/require"
)

func makeTestModule(t *testing.T) string {
	t.Helper()

	module := testsuite.NewModule(t, "example.com/project")
	module.Write(t, "internal/dto/dto.go", `
package dto

type User struct{}
type DeprecatedUser struct{}
`)

	module.Write(t, "internal/dto/auth/auth.go", `
package auth

type LoginRequest struct{}
type LoginResponse struct{}
`)

	module.Write(t, "internal/dto/auth/internal/internal.go", `
package internal

type InternalToken struct{}
`)

	module.Write(t, "internal/admin/admin.go", `
package admin

type Admin struct{}
`)

	module.Write(t, "internal/core/modules/users/user.go", `
package users

type User struct{}
`)

	return module.Root()
}

func TestGetFullPkgPathFallsBackToModulePathForDirectoryWithoutGoFiles(t *testing.T) {
	root := makeTestModule(t)

	got, err := getFullPkgPath(filepath.Join(root, "internal", "core", "modules"))

	require.NoError(t, err)
	require.Equal(t, "example.com/project/internal/core/modules", got)
}

func TestFilterSetEntryContains(t *testing.T) {
	tests := []struct {
		name      string
		basePkg   string
		recursive bool
		pkgPath   string
		want      bool
	}{
		{
			name:    "exact package",
			basePkg: "example.com/project/dto",
			pkgPath: "example.com/project/dto",
			want:    true,
		},
		{
			name:      "recursive child",
			basePkg:   "example.com/project/dto",
			recursive: true,
			pkgPath:   "example.com/project/dto/auth",
			want:      true,
		},
		{
			name:      "non-recursive child",
			basePkg:   "example.com/project/dto",
			recursive: false,
			pkgPath:   "example.com/project/dto/auth",
			want:      false,
		},
		{
			name:      "common prefix is not child",
			basePkg:   "example.com/project/dto",
			recursive: true,
			pkgPath:   "example.com/project/dtoext",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := filterSetEntry{
				basePkg:   tt.basePkg,
				recursive: tt.recursive,
			}

			require.Equal(t, tt.want, entry.contains(tt.pkgPath))
		})
	}
}
