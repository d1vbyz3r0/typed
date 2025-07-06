package nofile

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type Form struct {
	Name string `form:"name"`
	Age  int    `form:"age"`
}

func FormHandler(c echo.Context) error {
	var dto Form
	if err := c.Bind(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, dto)
}
