package typed

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTagOperation(t *testing.T) {
	tests := []struct {
		ApiPrefix string
		Route     string
	}{
		{
			ApiPrefix: "/api/v1",
			Route:     "/api/v1/tasks/test/",
		},
		{
			ApiPrefix: "/api/v1/",
			Route:     "/api/v1/tasks/test",
		},
		{
			ApiPrefix: "/api/v1",
			Route:     "/api/v1/tasks/",
		},
		{
			ApiPrefix: "/api/v1/",
			Route:     "/api/v1/tasks",
		},
	}

	expected := []string{"Tasks"}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			op := openapi3.NewOperation()
			err := TagOperation(op, tc.Route, tc.ApiPrefix)
			require.NoError(t, err)
			assert.Equal(t, expected, op.Tags)
		})
	}
}

func TestNormalizePathParams(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple path param",
			input:    "/api/users/:id",
			expected: "/api/users/{id}",
		},
		{
			name:     "multiple path params",
			input:    "/api/v1/tags/:tagId/pin/:taskId",
			expected: "/api/v1/tags/{tagId}/pin/{taskId}",
		},
		{
			name:     "no path params",
			input:    "/api/health",
			expected: "/api/health",
		},
		{
			name:     "root path",
			input:    "/",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePathParams(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizePathParams() = %v, want %v", got, tt.expected)
			}
		})
	}
}
