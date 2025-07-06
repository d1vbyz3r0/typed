package handlers

import (
	"github.com/d1vbyz3r0/typed/internal/parser"
	"github.com/d1vbyz3r0/typed/internal/parser/request"
	"github.com/d1vbyz3r0/typed/internal/parser/request/path"
	"github.com/d1vbyz3r0/typed/internal/parser/request/query"
	"github.com/d1vbyz3r0/typed/internal/parser/response"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/maps"
	"reflect"
	"strings"
)

var stringType = reflect.TypeOf("")

type Handler struct {
	route      echo.Route
	handler    parser.Handler
	path       string
	pathParams []path.Param
}

func NewHandler(route echo.Route, handler parser.Handler) Handler {
	segments := strings.Split(route.Path, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			trimmed := strings.TrimPrefix(segment, ":")
			segments[i] = "{" + trimmed + "}"
		}
	}
	p := strings.Join(segments, "/")

	params := make(map[string]path.Param, len(handler.Request.PathParams))
	parts := strings.Split(route.Path, "/")

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			trimmed := strings.TrimPrefix(part, ":")
			params[trimmed] = path.Param{
				Name: trimmed,
				Type: stringType,
			}
		}
	}

	for _, param := range handler.Request.PathParams {
		params[param.Name] = path.Param{
			Name: param.Name,
			Type: param.Type,
		}
	}
	pathParams := maps.Values(params)

	return Handler{
		route:      route,
		handler:    handler,
		path:       p,
		pathParams: pathParams,
	}
}

func (h Handler) Path() string {
	return h.path
}

func (h Handler) PathParams() []path.Param {
	return h.pathParams
}

func (h Handler) QueryParams() []query.Param {
	return h.handler.Request.QueryParams
}

func (h Handler) Method() string {
	return h.route.Method
}

func (h Handler) HandlerName() string {
	return h.handler.Name
}

func (h Handler) Description() string {
	return h.handler.Doc
}

func (h Handler) Request() *request.Request {
	return h.handler.Request
}

func (h Handler) Responses() response.StatusCodeMapping {
	return h.handler.Responses
}

func (h Handler) BindModel() string {
	return h.handler.Request.BindModel
}
