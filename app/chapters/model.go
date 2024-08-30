package chapters

import (
	"gorm.io/gorm"
)

type Chapter struct {
	gorm.Model
	Title string `json:"title" gorm:"column:title"`
	Description string `json:"description" gorm:"column:description"`
	Cover string `json:"cover" gorm:"column:cover"`
	Pg_rating string `json:"pg_rating" gorm:"column:pg_rating"`
	User_id int `json:"user_id" gorm:"column:user_id"`
	Slug string `json:"slug" gorm:"column:slug"`
}

type CreateRequest struct {
	Title string `json:"title"`
	Description string `json:"description"`
	Cover string `json:"cover"`
	Pg_rating string `json:"pg_rating"`
	User_id int `json:"user_id"`
	Slug string `json:"slug"`
}

type CreateResponse struct {
	gorm.Model
	Title string `json:"title"`
	Description string `json:"description"`
	Cover string `json:"cover"`
	Pg_rating string `json:"pg_rating"`
	User_id int `json:"user_id"`
	Slug string `json:"slug"`
}

type UpdateRequest struct {
	Title string `json:"title"`
	Description string `json:"description"`
	Cover string `json:"cover"`
	Pg_rating string `json:"pg_rating"`
	User_id int `json:"user_id"`
	Slug string `json:"slug"`
}

type UpdateResponse struct {
	gorm.Model
	Title string `json:"title"`
	Description string `json:"description"`
	Cover string `json:"cover"`
	Pg_rating string `json:"pg_rating"`
	User_id int `json:"user_id"`
	Slug string `json:"slug"`
}