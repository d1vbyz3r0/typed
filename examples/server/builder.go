package server

import "github.com/labstack/echo/v4"

type Builder struct {
	server *Server
}

func NewBuilder() *Builder {
	return &Builder{
		server: newServer(),
	}
}

func (s *Builder) Build() *Server {
	s.server.mapRoutes()
	return s.server
}

func (s *Builder) OnRouteAdded(onRouteAdded func(
	host string,
	route echo.Route,
	handler echo.HandlerFunc,
	middleware []echo.MiddlewareFunc,
)) {
	s.server.router.OnAddRouteHandler = onRouteAdded
}

func (s *Builder) ProvideRoutes() {
	s.server.mapRoutes()
}
