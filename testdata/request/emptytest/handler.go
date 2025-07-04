package emptytest

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type NoTags struct {
	Name string
	Age  int
}

func NoTagsHandler(c echo.Context) error {
	var dto NoTags
	if err := c.Bind(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, dto)
}
