package users

import (
	"base/core/layout"
	"base/core/logger"
	"base/core/module"
	"base/core/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserModule struct {
	module.DefaultModule
	DB            *gorm.DB
	Controller    *UserController
	ApiController *UserAPIController
	Service       *UserService
	Logger        logger.Logger
	ActiveStorage *storage.ActiveStorage
}

func NewUserModule(
	db *gorm.DB,
	webRouter *gin.RouterGroup,
	apiRouter *gin.RouterGroup,
	logger logger.Logger,
	activeStorage *storage.ActiveStorage,
	layoutEngine *layout.Engine,
) module.Module {
	// Initialize service with active storage
	service := NewUserService(db, logger, activeStorage)
	controller := NewUserController(service, logger, layoutEngine)
	apiController := NewUserAPIController(service, logger)

	usersModule := &UserModule{
		DB:            db,
		Controller:    controller,
		ApiController: apiController,
		Service:       service,
		Logger:        logger,
		ActiveStorage: activeStorage,
	}

	return usersModule
}

// Routes implements the standard module interface for web routes
func (m *UserModule) Routes(webRouter *gin.RouterGroup) {
	// Setup web routes for user module
	m.Controller.Routes(webRouter)
}

// APIRoutes implements the standard module interface for API routes
func (m *UserModule) APIRoutes(apiRouter *gin.RouterGroup) {
	// Setup API routes for user module
	m.ApiController.Routes(apiRouter)
}

func (m *UserModule) Migrate() error {
	err := m.DB.AutoMigrate(&User{})
	if err != nil {
		m.Logger.Error("Migration failed", logger.String("error", err.Error()))
		return err
	}
	return nil
}

func (m *UserModule) GetModels() []any {
	return []any{
		&User{},
	}
}

func (m *UserModule) GetModelNames() []string {
	models := m.GetModels()
	names := make([]string, len(models))
	for i, model := range models {
		names[i] = m.DB.Model(model).Statement.Table
	}
	return names
}
