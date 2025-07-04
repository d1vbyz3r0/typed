package xmltest

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type XMLDto struct {
	Name string `xml:"name"`
	Age  int    `xml:"age"`
}

func XmlHandler(c echo.Context) error {
	var dto XMLDto
	if err := c.Bind(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, dto)
}
