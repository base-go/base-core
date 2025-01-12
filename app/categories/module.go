package categories

import (
    "base/core/module"
    "base/app/models"
    "base/core/logger"
    "base/core/emitter"
    "base/core/storage"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type Module struct {
    module.DefaultModule
    DB         *gorm.DB
    Controller *CategoryController
    Service    *CategoryService
}

func NewCategoryModule(db *gorm.DB, router *gin.RouterGroup, log logger.Logger, emitter *emitter.Emitter, activeStorage *storage.ActiveStorage) module.Module {
    // Register attachment configuration for category.image
    activeStorage.RegisterAttachment("category", storage.AttachmentConfig{
        MaxFileSize:       10 * 1024 * 1024, // 10MB
        AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif"},
        Field:            "image",
    })

    service := NewCategoryService(db, activeStorage)
    controller := NewCategoryController(service)

    m := &Module{
        DB:         db,
        Controller: controller,
        Service:    service,
    }

    m.Routes(router)
    m.Migrate()

    return m
}

func (m *Module) Routes(router *gin.RouterGroup) {
    m.Controller.Routes(router)
}

func (m *Module) Migrate() error {
    return m.DB.AutoMigrate(&models.Category{})
}

func (m *Module) GetModels() []interface{} {
    return []interface{}{&models.Category{}}
}
