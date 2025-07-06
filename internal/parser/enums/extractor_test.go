package enums

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go/parser"
	"go/token"
	"testing"
)

func TestEnumExtractor_extractFromFile(t *testing.T) {
	src := `
package test

type Role string
type Status int

const (
	RoleAdmin = Role("admin")
	RoleUser  = Role("user")
	RoleGuest = Role("guest")
)

const (
	StatusNew Status = 1
	StatusDone Status = 2
)
`

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", src, 0)
	require.NoError(t, err)

	res, err := Extract("test", file)
	require.NoError(t, err)

	cases := map[string][]any{
		"test.Role":   {"admin", "user", "guest"},
		"test.Status": {1, 2},
	}

	assert.Equal(t, res, cases)
}
