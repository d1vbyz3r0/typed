package headers

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"strconv"
)

func Handler(c echo.Context) error {
	c.Request().Header.Get("h1")
	strconv.ParseInt(c.Request().Header.Get("h2"), 10, 64)
	uuid.MustParse(c.Request().Header.Get("h3"))
	strconv.ParseBool(c.Request().Header.Get("h4"))
	return nil
}
