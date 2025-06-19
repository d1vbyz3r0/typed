package generator

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestEnumExtractor_extractFromFile(t *testing.T) {
	const src = `
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

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_enum.go")
	if err := os.WriteFile(testFile, []byte(src), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	extractor := newEnumExtractor()
	if err := extractor.extractFromFile("test", testFile); err != nil {
		t.Fatalf("extractFromFile failed: %v", err)
	}

	cases := map[string][]any{
		"test.Role":   {"admin", "user", "guest"},
		"test.Status": {1, 2},
	}

	assert.Equal(t, extractor.Enums, cases)
}
