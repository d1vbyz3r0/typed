package server

import (
	"github.com/labstack/echo/v4"
)

//go:generate typed -config ../typed.yaml
//go:generate go run ../gen/spec.go

type Server struct {
	router *echo.Echo
}

func newServer() *Server {
	e := echo.New()
	return &Server{
		router: e,
	}
}

func (s *Server) mapRoutes() {
	api := s.router.Group("/api/v1")

	json := api.Group("/json")
	json.GET("/:id", getUserJSON)
	json.GET("/pretty", getUserJSONPretty)
	json.GET("/blob", getUserJSONBlob)

	xml := api.Group("/xml")
	xml.GET("/", getUserXML())
	xml.GET("/pretty", getUserXMLPretty())
	xml.GET("/blob", getUserXMLBlob())

	blobs := api.Group("/blobs")
	blobs.GET("/string", getString)
	blobs.GET("/blob", getBlob)
	blobs.GET("/stream", getStream)

	nocontent := api.Group("/nocontent")
	nocontent.GET("/redirect", redirectSomewhere)
	nocontent.DELETE("/resource/:id", deleteResource)

	forms := api.Group("/forms")
	h := FormsHandler{}
	forms.POST("/inline", h.inlineForm)
	forms.POST("/struct/:pathParam", h.structForm)

	validation := api.Group("/validation")
	vh := ValidationHandler{}
	validation.POST("/users", vh.createUser)
	validation.POST("/users/bulk", vh.bulkCreate)
	validation.PATCH("/users/:id/status", vh.updateStatus)
	validation.GET("/users", vh.search)
}

func (s *Server) Start(addr string) error {
	return s.router.Start(addr)
}
