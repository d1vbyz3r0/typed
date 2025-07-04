package multiple

import "github.com/labstack/echo/v4"

type Data struct {
	Id   int    `json:"id" xml:"id" form:"id"`
	Name string `json:"name" xml:"name" form:"name"`
}

func Handler(c echo.Context) error {
	var data Data
	c.Bind(&data)
	return nil
}
