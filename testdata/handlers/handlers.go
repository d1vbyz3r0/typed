package handlers

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type Example struct {
	X string
	Y int
}

func Handler(c echo.Context) error {
	if c.QueryParam("x") != "y" {
		return c.JSON(http.StatusBadRequest, []map[int]Example{})
	}

	if c.QueryParam("xml") != "true" {
		return c.XML(http.StatusOK, map[string]any{})
	}

	return c.JSON(http.StatusOK, Example{})
}
