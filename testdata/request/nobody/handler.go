package nobody

import "github.com/labstack/echo/v4"

type Body struct{}

func Handler(c echo.Context) error {
	var body Body
	c.Bind(&body)
	return nil
}
