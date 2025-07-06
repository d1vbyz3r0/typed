package typed

import "github.com/labstack/echo/v4"

type RoutesProvider interface {
	OnRouteAdded(func(
		host string,
		route echo.Route,
		handler echo.HandlerFunc,
		middleware []echo.MiddlewareFunc,
	))
	ProvideRoutes()
}
