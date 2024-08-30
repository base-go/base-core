package users

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string `json:"name" gorm:"column:name"`
	Email string `json:"email" gorm:"column:email"`
	Password string `json:"password" gorm:"column:password"`
	Avatar string `json:"avatar" gorm:"column:avatar"`
}

type CreateRequest struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
	Avatar string `json:"avatar"`
}

type CreateResponse struct {
	gorm.Model
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
	Avatar string `json:"avatar"`
}

type UpdateRequest struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
	Avatar string `json:"avatar"`
}

type UpdateResponse struct {
	gorm.Model
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
	Avatar string `json:"avatar"`
}