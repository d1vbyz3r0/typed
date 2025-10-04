package headers

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"strconv"
)

func Handler(c echo.Context) error {
	c.Request().Header.Get("h1")
	h1, err := strconv.ParseInt(c.Request().Header.Get("h2"), 10, 64)
	_ = h1
	_ = err

	uuid.MustParse(c.Request().Header.Get("h3"))
	strconv.ParseBool(c.Request().Header.Get("h4"))
	return c.JSON(200, nil)
}
