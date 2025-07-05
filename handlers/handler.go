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
	route   echo.Route
	handler parser.Handler
}

func (h Handler) Path() string {
	segments := strings.Split(h.route.Path, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			trimmed := strings.TrimPrefix(segment, ":")
			segments[i] = "{" + trimmed + "}"
		}
	}

	return strings.Join(segments, "/")
}

func (h Handler) PathParams() []path.Param {
	params := make(map[string]path.Param, len(h.handler.Request.PathParams))
	parts := strings.Split(h.route.Path, "/")

	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			trimmed := strings.TrimPrefix(part, ":")
			params[trimmed] = path.Param{
				Name: trimmed,
				Type: stringType,
			}
		}
	}

	for _, param := range h.handler.Request.PathParams {
		params[param.Name] = path.Param{
			Name: param.Name,
			Type: param.Type,
		}
	}

	return maps.Values(params)
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
