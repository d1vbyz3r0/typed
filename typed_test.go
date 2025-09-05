package typed

import (
	"fmt"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
