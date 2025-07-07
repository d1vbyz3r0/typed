package dto

type User struct {
	Id    int      `json:"id" form:"id"`
	Name  string   `json:"name" form:"name"`
	Email string   `json:"email" form:"email"`
	Role  UserRole `json:"role" form:"role"`
}

type UsersFilter struct {
	Search *string   `query:"search"`
	Limit  uint64    `query:"limit"`
	Offset uint64    `query:"offset"`
	Sort   SortOrder `query:"sort"`
}
