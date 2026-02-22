package dto

import (
	"github.com/google/uuid"
	"mime/multipart"
	"time"
)

type Status string

const (
	StatusActive   = Status("active")
	StatusInactive = Status("inactive")
)

type User struct {
	ID     uuid.UUID `json:"id" xml:"id" validate:"required"`
	Name   string    `json:"name" xml:"name" validate:"required,min=2,max=80"`
	Age    int       `json:"age" xml:"age" validate:"gte=0,lte=130"`
	Status Status    `json:"status" xml:"status" validate:"required"`
}

type Form struct {
	Q         string                  `query:"q" validate:"omitempty,min=2,max=64"`
	PathParam uuid.UUID               `param:"pathParam" validate:"required"`
	Timestamp time.Time               `form:"Timestamp" validate:"required"`
	File      *multipart.FileHeader   `form:"File"`
	Name      *string                 `form:"Name" validate:"omitempty,min=2,max=64"`
	Age       int                     `form:"Age" validate:"gte=0,lte=130"`
	FileArray []*multipart.FileHeader `form:"FileArray"`
	// HeaderParam uuid.UUID               `header:"headerParam"`
}

type FormUploadResp struct {
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	Token     string    `json:"token"`
	Timestamp time.Time `json:"timestamp"`
	Filename  string    `json:"filename"`
}

type UserContact struct {
	Email   string  `json:"email" validate:"required,email"`
	Website *string `json:"website,omitempty" validate:"omitempty,uri"`
}

type UserAddress struct {
	CountryCode string `json:"countryCode" validate:"required,len=2,alpha"`
	City        string `json:"city" validate:"required,min=2,max=64"`
	PostCode    string `json:"postCode" validate:"required,alphanum"`
}

type CreateValidatedUserRequest struct {
	Name     string            `json:"name" validate:"required,min=3,max=64"`
	Age      int               `json:"age" validate:"gte=18,lte=120"`
	Role     string            `json:"role" validate:"required,oneof=admin member guest"`
	Contact  UserContact       `json:"contact" validate:"required"`
	Address  UserAddress       `json:"address" validate:"required"`
	Skills   []string          `json:"skills" validate:"required,min=1,dive,required,alpha"`
	Metadata map[string]string `json:"metadata,omitempty" validate:"omitempty,dive,keys,printascii,endkeys,printascii"`
}

type BulkCreateUsersRequest struct {
	RequestID string                       `json:"requestId" validate:"required,uuid4"`
	Users     []CreateValidatedUserRequest `json:"users" validate:"required,min=1,dive,required"`
	DryRun    bool                         `json:"dryRun"`
}

type UpdateUserStatusRequest struct {
	Status Status  `json:"status" validate:"required"`
	Reason *string `json:"reason,omitempty" validate:"omitempty,min=3,max=128"`
	Notify bool    `json:"notify"`
}

type SearchUsersRequest struct {
	Query           *string `query:"query" validate:"omitempty,min=2,max=64"`
	Page            int     `query:"page" validate:"gte=1"`
	Limit           int     `query:"limit" validate:"gte=1,lte=100"`
	SortBy          string  `query:"sortBy" validate:"omitempty,oneof=name age status"`
	IncludeInactive bool    `query:"includeInactive"`
}

type BulkCreateUsersResponse struct {
	RequestID string      `json:"requestId"`
	Created   int         `json:"created"`
	UserIDs   []uuid.UUID `json:"userIds"`
}

type ErrorResponse struct {
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type ShouldBeExcluded struct {
}
