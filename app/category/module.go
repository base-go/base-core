package category

import (
	"base/app/model"
	"base/core/module"

	"gorm.io/gorm"
)

type Module struct {
	module.DefaultModule
	DB *gorm.DB
}

func NewModule(db *gorm.DB) *Module {
	return &Module{
		DB: db,
	}
}

func (m *Module) Init() error {
	return nil
}

func (m *Module) Migrate() error {
	return m.DB.AutoMigrate(&model.Category{})
}

func (m *Module) GetModels() []any {
	return []any{
		&model.Category{},
	}
}
