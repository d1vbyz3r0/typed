package allmodels

import (
	"github.com/labstack/echo/v4"
	"mime/multipart"
	"net/http"
	"strconv"
)

type Form struct {
	Name string          `form:"name"`
	File *multipart.File `form:"file"`
}

type Error struct {
	Msg string
}

type Result struct {
	Id int64
}

type User struct {
	Name string `json:"name" form:"name"`
	Age  int    `json:"age" form:"age"`
}

// Handler 1
func Handler(c echo.Context) error {
	var req Form
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusInternalServerError, "Bad Request")
	}

	if c.QueryParam("x") != "y" {
		return c.XML(http.StatusBadRequest, Error{Msg: "error"})
	}

	b, _ := strconv.ParseBool(c.QueryParam("json"))
	if b {
		return c.JSON(http.StatusBadRequest, Error{Msg: "error"})
	}

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	res := Result{Id: id}
	return c.JSON(http.StatusOK, res)
}

// OtherHandler is other handler
func OtherHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		var req User
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		return c.NoContent(http.StatusOK)
	}
}
