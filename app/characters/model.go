package characters

import (
	"gorm.io/gorm"
)

type Character struct {
	gorm.Model
	Name string `json:"name" gorm:"column:name"`
	Description string `json:"description" gorm:"column:description"`
	Gender string `json:"gender" gorm:"column:gender"`
	Main_character bool `json:"main_character" gorm:"column:main_character"`
}

type CreateRequest struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Gender string `json:"gender"`
	Main_character bool `json:"main_character"`
}

type CreateResponse struct {
	gorm.Model
	Name string `json:"name"`
	Description string `json:"description"`
	Gender string `json:"gender"`
	Main_character bool `json:"main_character"`
}

type UpdateRequest struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Gender string `json:"gender"`
	Main_character bool `json:"main_character"`
}

type UpdateResponse struct {
	gorm.Model
	Name string `json:"name"`
	Description string `json:"description"`
	Gender string `json:"gender"`
	Main_character bool `json:"main_character"`
}