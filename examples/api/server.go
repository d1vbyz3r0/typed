package api

import (
	"github.com/d1vbyz3r0/typed/examples/api/handlers"
	"github.com/labstack/echo/v4"
)

//go:generate typed -config ../typed.yaml
//go:generate go run ../gen/spec_gen.go
type server struct {
	router *echo.Echo
	addr   string
}

func (s *server) mapHandlers() {
	api := s.router.Group("/api/v1")
	users := api.Group("/users")
	users.GET("/:userId", handlers.GetUser(nil))
	users.GET("/", handlers.GetUsers(nil))
	users.POST("/", handlers.CreateUser(nil))
}
