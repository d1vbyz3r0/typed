package generator

import (
	"github.com/labstack/echo/v4"
	"strconv"
	"strings"
)

type HandlerType int

const (
	DirectHandler  HandlerType = iota + 1 // func(echo.Context) error
	WrapperHandler                        // func(...) echo.HandlerFunc
)

type RouteInfo struct {
	Path        string
	Method      string
	HandlerName string
	HandlerType HandlerType
	Handler     *HandlerInfo // Add reference to discovered handler
	PathParams  []*PathParam
}

// RouteAnalyzer supposed to combine AST analysis results with known routes, extracted as []*echo.Route
type RouteAnalyzer struct {
	routes []*RouteInfo
	//handlersBasePkg string
}

func NewRouteAnalyzer() *RouteAnalyzer {
	return &RouteAnalyzer{
		routes: make([]*RouteInfo, 0),
		//handlersBasePkg: handlersBasePkg,
	}
}

func (ra *RouteAnalyzer) Routes() []*RouteInfo {
	return ra.routes
}

func (ra *RouteAnalyzer) CollectRoutes(routes []*echo.Route) error {
	for _, route := range routes {
		info := &RouteInfo{
			Path:        route.Path,
			Method:      route.Method,
			HandlerName: route.Name,
			HandlerType: classifyHandler(route.Name),
			PathParams:  extractPathParams(route.Path),
		}
		ra.routes = append(ra.routes, info)
	}
	return nil
}

func classifyHandler(name string) HandlerType {
	parts := strings.Split(name, ".")
	lastPart := parts[len(parts)-1]

	// Check if ends with .funcN where N is any number
	if strings.HasPrefix(lastPart, "func") && len(lastPart) > 4 {
		_, err := strconv.Atoi(lastPart[4:])
		if err == nil {
			return WrapperHandler
		}
	}
	return DirectHandler
}

func (ra *RouteAnalyzer) MatchHandlers(handlers map[string]*HandlerInfo) {
	logger.Info("Matching routes with handlers", "count", len(ra.routes))

	for _, route := range ra.routes {
		logger.Debug("Processing route", "method", route.Method, "path", route.Path, "handler_name", route.HandlerName)
		// Example route name: "xxx/internal/api.(*Server).mapUsers.LoginUserHandler.func1"
		parts := strings.Split(route.HandlerName, ".")

		// Get the package and handler name
		// For wrapper handlers (with .funcN suffix), we need to take the part before .funcN
		handlerName := parts[len(parts)-1]
		if strings.HasPrefix(handlerName, "func") {
			handlerName = parts[len(parts)-2]
		}

		// Find matching handler in discovered handlers
		for key, handler := range handlers {
			if strings.HasSuffix(key, handlerName) {
				route.Handler = handler
				route.HandlerType = getHandlerType(handler)
				for _, param := range route.PathParams {
					param.Type = paramTypeFromContext(handler.Node, param.Name)
				}

				logger.Debug("Matched handler", "key", key)
				break
			}
		}

		if route.Handler == nil {
			logger.Warn("No handler found for route", "method", route.Method, "path", route.Path)
		}
	}

	logger.Info("Handler matching completed")
}

func getHandlerType(handler *HandlerInfo) HandlerType {
	if handler.IsWrapper {
		return WrapperHandler
	}
	return DirectHandler
}
