package inlinefiles

import (
	"github.com/labstack/echo/v4"
)

func FormHandler(c echo.Context) error {
	c.FormValue("name")
	_, _ = c.FormFile("file")
	return nil
}
