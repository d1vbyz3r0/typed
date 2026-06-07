package generator

import (
	"testing"

	"github.com/d1vbyz3r0/typed/internal/testsuite"
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

	return module.Root()
}
