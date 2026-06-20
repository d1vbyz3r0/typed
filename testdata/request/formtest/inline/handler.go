package inline

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

func Handler(c echo.Context) error {
	c.FormValue("name")
	_, _ = strconv.Atoi(c.FormValue("age"))
	return nil
}
