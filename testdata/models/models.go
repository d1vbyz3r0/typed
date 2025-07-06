package models

import "github.com/google/uuid"

type UserRole string

const (
	UserRoleAdmin UserRole = "Admin"
	UserRoleUser  UserRole = "User"
)

type User struct {
	Id   int      `json:"id"`
	Name string   `json:"name"`
	Role UserRole `json:"role"`
}

type PathParams struct {
	UUID uuid.UUID `param:"uuid"`
	Name string    `param:"name"`
	Id   int       `param:"id"`
	Role UserRole  `param:"role"`
}

type QueryParams struct {
	UUID uuid.UUID `query:"uuid"`
	Name string    `query:"name"`
	Id   *int      `query:"id"`
	Role UserRole  `query:"role"`
}
