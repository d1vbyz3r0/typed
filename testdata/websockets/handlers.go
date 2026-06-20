package websockets

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket/wsjson"
	"github.com/labstack/echo"

	coder "github.com/coder/websocket"
	gorilla "github.com/gorilla/websocket"
	xnet "golang.org/x/net/websocket"
)

var (
	upgrader = gorilla.Upgrader{}
)

func XNetWebsocket(c echo.Context) error {
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

func Regular(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{})
}
