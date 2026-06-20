package server

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"examples/dto"

	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	coder "github.com/coder/websocket"
	gorilla "github.com/gorilla/websocket"
	xnet "golang.org/x/net/websocket"
)

var (
	upgrader = gorilla.Upgrader{}
)

// You can use regular echo.HandlerFunc

// Also docstrings are supported to document your handlers
func getUserJSON(c echo.Context) error {
	id, _ := uuid.Parse(c.Param("id"))
	req := c.Request()
	req.Header.Get("User-ID")

	c.Response().Header().Set("X-Custom-Header1", "1")
	c.Response().Header().Add("X-Custom-Header2", "2")
	return c.JSON(http.StatusOK, dto.User{
		ID:     id,
		Name:   "Alice",
		Age:    30,
		Status: dto.StatusActive,
	})
}

func getUserJSONPretty(c echo.Context) error {
	u := dto.User{
		ID:     uuid.New(),
		Name:   "Bob",
		Age:    25,
		Status: dto.StatusInactive,
	}
	return c.JSONPretty(http.StatusOK, u, "  ")
}

func getUserJSONBlob(c echo.Context) error {
	data := []byte(`{"message": "this is raw json blob"}`)
	return c.JSONBlob(http.StatusOK, data)
}

// Or "scoped" to pass dependencies around

func getUserXML( /*deps here*/ ) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.QueryParam("message")
		u := dto.User{
			ID:     uuid.New(),
			Name:   "Charlie",
			Age:    40,
			Status: dto.StatusActive,
		}
		return c.XML(http.StatusOK, u)
	}
}

func getUserXMLPretty() echo.HandlerFunc {
	return func(c echo.Context) error {
		buf := &bytes.Buffer{}
		enc := xml.NewEncoder(buf)
		enc.Indent("", "  ")
		_ = enc.Encode(dto.User{
			ID:     uuid.New(),
			Name:   "Diana",
			Age:    50,
			Status: dto.StatusInactive,
		})
		return c.XMLBlob(http.StatusOK, buf.Bytes())
	}
}

func getUserXMLBlob() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := []byte(`<User><id>uuid</id><name>Eva</name><age>60</age><status>active</status></User>`)
		return c.XMLBlob(http.StatusOK, data)
	}
}

func getString(c echo.Context) error {
	return c.String(http.StatusOK, "Plain text response")
}

func getBlob(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/octet-stream", []byte("binary blob here"))
}

func getStream(c echo.Context) error {
	r := io.NopCloser(strings.NewReader("streamed data"))
	return c.Stream(http.StatusOK, echo.MIMETextPlain, r)
}

func redirectSomewhere(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/json/"+uuid.New().String())
}

func deleteResource(c echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

// And also handlers can be declared as struct methods

type FormsHandler struct{}

func (h FormsHandler) inlineForm(c echo.Context) error {
	name := c.FormValue("name")
	active, err := strconv.ParseBool(c.FormValue("active"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	timestamp, err := time.Parse(time.RFC3339, c.FormValue("timestamp"))
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}
	defer src.Close()

	res, err := os.Create(filepath.Join("uploads", file.Filename))
	_, err = io.Copy(res, src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	token := c.QueryParam("token")
	resp := dto.FormUploadResp{
		Name:      name,
		Active:    active,
		Token:     token,
		Timestamp: timestamp,
		Filename:  file.Filename,
	}

	return c.JSON(http.StatusOK, resp)
}

func (h FormsHandler) structForm(c echo.Context) error {
	var req dto.Form
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": err.Error(),
		})
	}

	if req.File != nil {
		f, err := req.File.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}
		defer f.Close()

		res, err := os.Create(filepath.Join("uploads", req.File.Filename))
		_, err = io.Copy(res, f)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
	}

	for _, file := range req.FileArray {
		f, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}
		defer f.Close()

		res, err := os.Create(filepath.Join("uploads", file.Filename))
		_, err = io.Copy(res, f)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"name":           req.Name,
		"age":            req.Age,
		"file_name":      req.File.Filename,
		"file_array_len": len(req.FileArray),
	})
}

// handler with golang.org/x/net/websocket usage
func XNetWebsocketHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		xnet.Handler(func(ws *xnet.Conn) {
			defer ws.Close()
			for {
				// Write
				if err := xnet.Message.Send(ws, "Hello, Client!"); err != nil {
					c.Logger().Error("failed to write WS message", "error", err)
				}

				// Read
				msg := ""
				if err := xnet.Message.Receive(ws, &msg); err != nil {
					c.Logger().Error("failed to write WS message", "error", err)
				}
				fmt.Printf("%s\n", msg)
			}
		}).ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

// handler with github.com/gorilla/websocket usage
func GorillaWebsocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	for {
		// Write
		err := ws.WriteMessage(gorilla.TextMessage, []byte("Hello, Client!"))
		if err != nil {
			c.Logger().Error("failed to write WS message", "error", err)
		}

		// Read
		_, msg, err := ws.ReadMessage()
		if err != nil {
			c.Logger().Error("failed to read WS message", "error", err)
		}
		fmt.Printf("%s\n", msg)
	}
}

// handler with github.com/coder/websocket usage
func CoderWebsocket(c echo.Context) error {
	conn, err := coder.Accept(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer conn.CloseNow()

	// Set the context as needed. Use of r.Context() is not recommended
	// to avoid surprising behavior (see http.Hijacker).
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var v any
	err = wsjson.Read(ctx, conn, &v)
	if err != nil {
		return err
	}

	c.Logger().Printf("received: %v", v)

	conn.Close(coder.StatusNormalClosure, "")
	return nil
}
