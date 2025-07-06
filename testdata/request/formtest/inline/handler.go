package inline

import (
	"github.com/labstack/echo/v4"
	"strconv"
)

func FormHandler(c echo.Context) error {
	c.FormValue("name")
	_, _ = strconv.Atoi(c.FormValue("age"))
	return nil
}
