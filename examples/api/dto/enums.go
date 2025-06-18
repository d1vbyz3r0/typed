package dto

type SortOrder string

// You can declare values like that
const (
	ASC  = SortOrder("asc")
	DESC = SortOrder("desc")
)

type UserRole string

// Or like that
const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)
