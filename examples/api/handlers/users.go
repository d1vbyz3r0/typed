package handlers

import (
	"context"
	"github.com/d1vbyz3r0/typed/examples/api/dto"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

// Just for example purpose it's here...
type UsersService interface {
	GetUserById(ctx context.Context, userId int) (dto.User, error)
	GetUsers(ctx context.Context, f dto.UsersFilter) ([]dto.User, error)
	CreateUser(ctx context.Context, user dto.User) (dto.User, error)
}

func GetUser(srv UsersService) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId, err := strconv.Atoi(c.Param("userId"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
		}

		user, err := srv.GetUserById(c.Request().Context(), userId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		}

		return c.JSON(http.StatusOK, user)
	}
}

func GetUsers(srv UsersService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var q dto.UsersFilter
		if err := c.Bind(&q); err != nil {
			return c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
		}

		users, err := srv.GetUsers(c.Request().Context(), q)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		}

		return c.JSON(http.StatusOK, users)
	}
}

func CreateUser(srv UsersService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var data dto.User
		if err := c.Bind(&data); err != nil {
			return c.JSON(http.StatusBadRequest, dto.Error{Message: err.Error()})
		}

		user, err := srv.CreateUser(c.Request().Context(), data)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, dto.Error{Message: err.Error()})
		}

		return c.JSON(http.StatusOK, user)
	}
}

func ReturningMap(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string][]dto.User{})
}
