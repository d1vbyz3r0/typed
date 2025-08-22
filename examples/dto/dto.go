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
	ID     uuid.UUID `json:"id" xml:"id"`
	Name   string    `json:"name" xml:"name"`
	Age    int       `json:"age" xml:"age"`
	Status Status    `json:"status" xml:"status"`
}

type Form struct {
	Q         string                  `query:"q"`
	PathParam uuid.UUID               `param:"pathParam"`
	Timestamp time.Time               `form:"Timestamp"`
	File      *multipart.FileHeader   `form:"File"`
	Name      *string                 `form:"Name"`
	Age       int                     `form:"Age"`
	FileArray []*multipart.FileHeader `form:"FileArray"`
}

type FormUploadResp struct {
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	Token     string    `json:"token"`
	Timestamp time.Time `json:"timestamp"`
	Filename  string    `json:"filename"`
}

type ShouldBeExcluded struct {
}
