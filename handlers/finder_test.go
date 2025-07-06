package handlers

import (
	"github.com/d1vbyz3r0/typed/examples/api"
	"github.com/d1vbyz3r0/typed/internal/parser"
	"github.com/d1vbyz3r0/typed/internal/parser/request"
	"github.com/d1vbyz3r0/typed/internal/parser/request/path"
	"github.com/d1vbyz3r0/typed/internal/parser/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
	"log/slog"
	"net/http"
	"os"
	"reflect"
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

func TestFinder_Find(t *testing.T) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	f, err := NewFinder()
	require.NoError(t, err)

	err = f.Find([]SearchPattern{
		{
			Path:      "../examples/api",
			Recursive: true,
		},
	})
	require.NoError(t, err)

	want := map[string]parser.Handler{
		"github.com/d1vbyz3r0/typed/examples/api/handlers.GetUser": {
			Doc:  "GetUser will return user by id",
			Name: "GetUser",
			Pkg:  "github.com/d1vbyz3r0/typed/examples/api/handlers",
			Request: &request.Request{
				BindModel:          "",
				BindModelPkg:       "",
				ContentTypeMapping: request.ContentTypeMapping{},
				PathParams: []path.Param{
					{
						Name: "userId",
						Type: reflect.TypeOf(int(0)),
					},
				},
				QueryParams: nil,
			},
			Responses: response.StatusCodeMapping{
				http.StatusBadRequest: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.Error",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
				http.StatusInternalServerError: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.Error",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
				http.StatusOK: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.User",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
			},
		},
		"github.com/d1vbyz3r0/typed/examples/api/handlers.GetUsers": {
			Doc:  "",
			Name: "GetUsers",
			Pkg:  "github.com/d1vbyz3r0/typed/examples/api/handlers",
			Request: &request.Request{
				BindModel:          "dto.UsersFilter",
				BindModelPkg:       "github.com/d1vbyz3r0/typed/examples/api/dto",
				ContentTypeMapping: request.ContentTypeMapping{},
				PathParams:         nil,
				QueryParams:        nil,
			},
			Responses: response.StatusCodeMapping{
				http.StatusBadRequest: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.Error",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
				http.StatusInternalServerError: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.Error",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
				http.StatusOK: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "[]dto.User",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
			},
		},
		"github.com/d1vbyz3r0/typed/examples/api/handlers.CreateUser": {
			Doc:  "",
			Name: "CreateUser",
			Pkg:  "github.com/d1vbyz3r0/typed/examples/api/handlers",
			Request: &request.Request{
				BindModel:    "dto.User",
				BindModelPkg: "github.com/d1vbyz3r0/typed/examples/api/dto",
				ContentTypeMapping: request.ContentTypeMapping{
					echo.MIMEApplicationJSON: {},
				},
				PathParams:  nil,
				QueryParams: nil,
			},
			Responses: response.StatusCodeMapping{
				http.StatusBadRequest: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.Error",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
				http.StatusInternalServerError: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.Error",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
				http.StatusOK: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.User",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
			},
		},
		"github.com/d1vbyz3r0/typed/examples/api/handlers.ReturningMap": {
			Doc:  "",
			Name: "ReturningMap",
			Pkg:  "github.com/d1vbyz3r0/typed/examples/api/handlers",
			Request: &request.Request{
				BindModel:          "",
				BindModelPkg:       "",
				ContentTypeMapping: request.ContentTypeMapping{},
				PathParams:         nil,
				QueryParams:        nil,
			},
			Responses: response.StatusCodeMapping{
				http.StatusOK: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "map[string][]dto.User",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
			},
		},
	}

	for k, got := range f.handlers {
		require.Equal(t, want[k].Pkg, got.Pkg)
		require.Equal(t, want[k].Name, got.Name)
		require.Equal(t, want[k].Responses, got.Responses)
		require.Equal(t, want[k].Doc, got.Doc)
		require.Equal(t, want[k].Request, got.Request)
		require.Equal(t, want[k].Responses, got.Responses)

	}
}

func TestFinder_Match(t *testing.T) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	f, err := NewFinder()
	require.NoError(t, err)

	err = f.Find([]SearchPattern{
		{
			Path:      "../examples/api",
			Recursive: true,
		},
	})
	require.NoError(t, err)

	server := api.Server{
		Router: echo.New(),
	}
	server.MapHandlers()

	routes := make(map[string]EchoRoute)
	for _, r := range server.Router.Routes() {
		route := *r
		p := f.getHandlerName(route)
		routes[p] = EchoRoute{Route: route}
	}

	expectedFindRes := map[string]parser.Handler{
		"GetUser": {
			Doc:  "GetUser will return user by id",
			Name: "GetUser",
			Pkg:  "github.com/d1vbyz3r0/typed/examples/api/handlers",
			Request: &request.Request{
				BindModel:          "",
				BindModelPkg:       "",
				ContentTypeMapping: request.ContentTypeMapping{},
				PathParams: []path.Param{
					{
						Name: "userId",
						Type: reflect.TypeOf(int(0)),
					},
				},
				QueryParams: nil,
			},
			Responses: response.StatusCodeMapping{
				http.StatusBadRequest: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.Error",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
				http.StatusInternalServerError: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.Error",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
				http.StatusOK: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.User",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
			},
		},
		"GetUsers": {
			Doc:  "",
			Name: "GetUsers",
			Pkg:  "github.com/d1vbyz3r0/typed/examples/api/handlers",
			Request: &request.Request{
				BindModel:          "dto.UsersFilter",
				BindModelPkg:       "github.com/d1vbyz3r0/typed/examples/api/dto",
				ContentTypeMapping: request.ContentTypeMapping{},
				PathParams:         nil,
				QueryParams:        nil,
			},
			Responses: response.StatusCodeMapping{
				http.StatusBadRequest: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.Error",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
				http.StatusInternalServerError: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.Error",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
				http.StatusOK: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "[]dto.User",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
			},
		},
		"CreateUser": {
			Doc:  "",
			Name: "CreateUser",
			Pkg:  "github.com/d1vbyz3r0/typed/examples/api/handlers",
			Request: &request.Request{
				BindModel:    "dto.User",
				BindModelPkg: "github.com/d1vbyz3r0/typed/examples/api/dto",
				ContentTypeMapping: request.ContentTypeMapping{
					echo.MIMEApplicationJSON: {},
				},
				PathParams:  nil,
				QueryParams: nil,
			},
			Responses: response.StatusCodeMapping{
				http.StatusBadRequest: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.Error",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
				http.StatusInternalServerError: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.Error",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
				http.StatusOK: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "dto.User",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
			},
		},
		"ReturningMap": {
			Doc:  "",
			Name: "ReturningMap",
			Pkg:  "github.com/d1vbyz3r0/typed/examples/api/handlers",
			Request: &request.Request{
				BindModel:          "",
				BindModelPkg:       "",
				ContentTypeMapping: request.ContentTypeMapping{},
				PathParams:         nil,
				QueryParams:        nil,
			},
			Responses: response.StatusCodeMapping{
				http.StatusOK: {
					{
						ContentType: echo.MIMEApplicationJSON,
						TypeName:    "map[string][]dto.User",
						TypePkgPath: "github.com/d1vbyz3r0/typed/examples/api/dto",
					},
				},
			},
		},
	}
	want := make([]Handler, 0, len(expectedFindRes))
	for k, v := range expectedFindRes {
		r, ok := routes[k]
		require.True(t, ok, "key: %s", k)

		want = append(want, Handler{
			route:   r.Route,
			handler: v,
		})
	}

	v := maps.Values(routes)
	handlers := f.Match(v)
	require.ElementsMatch(t, want, handlers)
}
