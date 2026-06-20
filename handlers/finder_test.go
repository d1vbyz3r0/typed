package handlers

import (
	"testing"

	"github.com/d1vbyz3r0/typed/internal/testsuite"
	"github.com/stretchr/testify/require"
)

func genericTest[I, O any]() {}

type handler struct{}

func (handler) method() {}

func (*handler) method2() {}

func closure() func() {
	return func() {}
}

func Test_funcPackagePath(t *testing.T) {
	tests := []struct {
		name  string
		_func any
		want  string
	}{
		{
			name:  "anonymous function",
			_func: func() {},
			want:  "github.com/d1vbyz3r0/typed/handlers",
		},
		{
			name:  "generic function",
			_func: genericTest[int, string],
			want:  "github.com/d1vbyz3r0/typed/handlers",
		},
		{
			name:  "struct value method",
			_func: handler{}.method,
			want:  "github.com/d1vbyz3r0/typed/handlers",
		},
		{
			name:  "struct pointer method",
			_func: (&handler{}).method2,
			want:  "github.com/d1vbyz3r0/typed/handlers",
		},
		{
			name:  "closure",
			_func: closure(),
			want:  "github.com/d1vbyz3r0/typed/handlers",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := funcPackagePath(tc._func)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestFinder_Find(t *testing.T) {
	f, err := NewFinder()
	require.NoError(t, err)

	err = f.Find([]SearchPattern{
		{
			Path: testsuite.FixturePath(t, "parser/c1"),
		},
	})
	require.NoError(t, err)

	const pkg = "github.com/d1vbyz3r0/typed/testdata/parser/c1"
	require.Len(t, f.handlers, 2)

	handler, ok := f.handlers[pkg+".Handler"]
	require.True(t, ok)
	require.Equal(t, "Handler 1", handler.Doc)
	require.Len(t, handler.Request.PathParams, 1)
	require.Len(t, handler.Request.QueryParams, 2)
	require.Len(t, handler.Responses, 3)

	wrapper, ok := f.handlers[pkg+".OtherHandler"]
	require.True(t, ok)
	require.Equal(t, "OtherHandler is other handler", wrapper.Doc)
}
