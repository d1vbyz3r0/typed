package xmltest

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type XMLDto struct {
	Name string `xml:"name"`
	Age  int    `xml:"age"`
}

func Handler(c echo.Context) error {
	var dto XMLDto
	if err := c.Bind(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, dto)
}
