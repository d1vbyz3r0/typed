package generics

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type Page[I any] struct {
	Items []I    `json:"items"`
	Total uint64 `json:"total"`
}

type InstantiatedGeneric struct {
	Pages []Page[int] `json:"pages"`
}

func Handler(c echo.Context) error {
	var req Page[any]
	c.Bind(&req)
	return c.JSON(http.StatusOK, Page[string]{})
}
