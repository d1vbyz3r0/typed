package server

import (
	"github.com/labstack/echo/v4"
)

//go:generate go run ../../cmd/typed/typed.go -config ../configs/standalone.yaml
//go:generate go run ../standalone-example/spec.go

//go:generate go run ../../cmd/typed/typed.go -config ../configs/lib.yaml
//go:generate go run ../lib-example/gen.go

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

	ws := api.Group("/ws")
	ws.GET("/xnet", XNetWebsocketHandler())
	ws.GET("/gorilla", GorillaWebsocket)
	ws.GET("/coder", CoderWebsocket)
}

func (s *Server) Start(addr string) error {
	return s.router.Start(addr)
}
