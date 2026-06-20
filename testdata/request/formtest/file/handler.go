package file

import (
	"mime/multipart"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Form struct {
	Name string                `form:"name"`
	File *multipart.FileHeader `form:"file"`
}

func Handler(c echo.Context) error {
	var dto Form
	if err := c.Bind(&dto); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, dto)
}
