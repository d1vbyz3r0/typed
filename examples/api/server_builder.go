package api

import (
	"github.com/labstack/echo/v4"
)

type ServerBuilder struct {
	server     *server
	listenAddr string
}

func NewServerBuilder() *ServerBuilder {
	return &ServerBuilder{server: &server{
		router: echo.New(),
		addr:   "",
	}}
}

func (b *ServerBuilder) clone() *ServerBuilder {
	return &ServerBuilder{
		server:     b.server,
		listenAddr: b.listenAddr,
	}
}

func (b *ServerBuilder) WithAddr(addr string) *ServerBuilder {
	b = b.clone()
	b.server.addr = addr
	return b
}

func (b *ServerBuilder) OnRouteAdded(cb func(string, echo.Route, echo.HandlerFunc, []echo.MiddlewareFunc)) {
	b.server.router.OnAddRouteHandler = cb
}

func (b *ServerBuilder) ProvideRoutes() {
	b.server.mapHandlers()
}
