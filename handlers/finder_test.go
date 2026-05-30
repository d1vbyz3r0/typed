package handlers

import (
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/d1vbyz3r0/typed/common/typing"
	"github.com/d1vbyz3r0/typed/internal/parser"
	"github.com/d1vbyz3r0/typed/internal/parser/request"
	"github.com/d1vbyz3r0/typed/internal/parser/request/path"
	"github.com/d1vbyz3r0/typed/internal/parser/response"
	"github.com/labstack/echo/v4"
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
				ContentTypeMapping: request.ContentTypeMapping{},
				PathParams: []path.Param{
					{
						Name: "userId",
						Type: reflect.TypeFor[int](),
					},
				},
				QueryParams: nil,
			},
			Responses: response.StatusCodeMapping{
				http.StatusBadRequest: {
					{
						ModelType:   typing.Named("github.com/d1vbyz3r0/typed/examples/api/dto", "Error"),
						ContentType: echo.MIMEApplicationJSON,
					},
				},
				http.StatusInternalServerError: {
					{
						ModelType:   typing.Named("github.com/d1vbyz3r0/typed/examples/api/dto", "Error"),
						ContentType: echo.MIMEApplicationJSON,
					},
				},
				http.StatusOK: {
					{
						ModelType:   typing.Named("github.com/d1vbyz3r0/typed/examples/api/dto", "User"),
						ContentType: echo.MIMEApplicationJSON,
					},
				},
			},
		},
		"github.com/d1vbyz3r0/typed/examples/api/handlers.GetUsers": {
			Doc:  "",
			Name: "GetUsers",
			Pkg:  "github.com/d1vbyz3r0/typed/examples/api/handlers",
			Request: &request.Request{
				ModelType:          typing.Named("github.com/d1vbyz3r0/typed/examples/api/dto", "UsersFilter"),
				ContentTypeMapping: request.ContentTypeMapping{},
				PathParams:         nil,
				QueryParams:        nil,
			},
			Responses: response.StatusCodeMapping{
				http.StatusBadRequest: {
					{
						ModelType:   typing.Named("github.com/d1vbyz3r0/typed/examples/api/dto", "Error"),
						ContentType: echo.MIMEApplicationJSON,
					},
				},
				http.StatusInternalServerError: {
					{
						ModelType:   typing.Named("github.com/d1vbyz3r0/typed/examples/api/dto", "Error"),
						ContentType: echo.MIMEApplicationJSON,
					},
				},
				http.StatusOK: {
					{
						ModelType:   typing.Named("github.com/d1vbyz3r0/typed/examples/api/dto", "User"),
						ContentType: echo.MIMEApplicationJSON,
					},
				},
			},
		},
		"github.com/d1vbyz3r0/typed/examples/api/handlers.CreateUser": {
			Doc:  "",
			Name: "CreateUser",
			Pkg:  "github.com/d1vbyz3r0/typed/examples/api/handlers",
			Request: &request.Request{
				ModelType: typing.Named("github.com/d1vbyz3r0/typed/examples/api/dto", "User"),
				ContentTypeMapping: request.ContentTypeMapping{
					echo.MIMEApplicationJSON: {},
				},
				PathParams:  nil,
				QueryParams: nil,
			},
			Responses: response.StatusCodeMapping{
				http.StatusBadRequest: {
					{
						ModelType:   typing.Named("github.com/d1vbyz3r0/typed/examples/api/dto", "Error"),
						ContentType: echo.MIMEApplicationJSON,
					},
				},
				http.StatusInternalServerError: {
					{
						ModelType:   typing.Named("github.com/d1vbyz3r0/typed/examples/api/dto", "Error"),
						ContentType: echo.MIMEApplicationJSON,
					},
				},
				http.StatusOK: {
					{
						ModelType:   typing.Named("github.com/d1vbyz3r0/typed/examples/api/dto", "User"),
						ContentType: echo.MIMEApplicationJSON,
					},
				},
			},
		},
		"github.com/d1vbyz3r0/typed/examples/api/handlers.ReturningMap": {
			Doc:  "",
			Name: "ReturningMap",
			Pkg:  "github.com/d1vbyz3r0/typed/examples/api/handlers",
			Request: &request.Request{
				ContentTypeMapping: request.ContentTypeMapping{},
				PathParams:         nil,
				QueryParams:        nil,
			},
			Responses: response.StatusCodeMapping{
				http.StatusOK: {
					{
						ModelType:   typing.Named("github.com/d1vbyz3r0/typed/examples/api/dto", "User"),
						ContentType: echo.MIMEApplicationJSON,
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
