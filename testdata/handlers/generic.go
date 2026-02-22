package handlers

import (
	"github.com/d1vbyz3r0/typed/testdata/handlers/shared"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Page[T any] struct {
	Items []T
}

func GenericHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, Page[shared.Item]{})
}
