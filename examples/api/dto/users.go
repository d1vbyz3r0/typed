package dto

type User struct {
	Id    int      `json:"id"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Role  UserRole `json:"role"`
}

type UsersFilter struct {
	Search *string   `query:"search"`
	Limit  uint64    `query:"limit"`
	Offset uint64    `query:"offset"`
	Sort   SortOrder `query:"sort"`
}
