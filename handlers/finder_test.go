package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFinder_getHandlerFullPath(t *testing.T) {
	f := &Finder{}
	_ = f
	cases := []struct {
		name string
		want string
	}{
		{
			name: "github.com/d1vbyz3r0/typed/internal/api.(*Server).mapUsers.LoginUserHandler.func1",
			want: "github.com/d1vbyz3r0/typed/internal/api.LoginUserHandler",
		},
		{
			name: "project.LoginUserHandler",
			want: "project.LoginUserHandler",
		},
		{
			name: "main.main.H.func2",
			want: "main.H",
		},
		{
			name: "github.com/d1vbyz3r0/typed/internal/api.(*Server).mapUsers.LoginUserHandler",
			want: "github.com/d1vbyz3r0/typed/internal/api.LoginUserHandler",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := f.getHandlerFullPath(echo.Route{
				Name: c.name,
			})
			require.Equal(t, c.want, got)
		})
	}
}
