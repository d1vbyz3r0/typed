package jsontest

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type JsonDTO struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func Handler(c echo.Context) error {
	var dto JsonDTO
	if err := c.Bind(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, dto)
}
