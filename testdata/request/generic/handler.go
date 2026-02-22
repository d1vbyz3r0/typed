package generic

import (
	"github.com/d1vbyz3r0/typed/testdata/request/generic/shared"
	"github.com/labstack/echo/v4"
)

type Page[T any] struct {
	Items []T `json:"items"`
}

func Handler(c echo.Context) error {
	var req Page[shared.User]
	if err := c.Bind(&req); err != nil {
		return err
	}

	return nil
}
