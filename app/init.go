package app

import (
	// MODULE_IMPORT_MARKER - Do not remove this comment because it's used by the CLI to add new module imports

	"base/core/config"
	"base/core/database"
	"base/core/emitter"
	"base/core/logger"
	"base/core/module"
	"base/core/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	DB      *gorm.DB
	Router  *gin.Engine
	Log     logger.Logger
	Emitter *emitter.Emitter
	Storage *storage.ActiveStorage
	Modules []module.Module
}

// NewApp creates and initializes a new App instance
func NewApp(cfg *config.Config) (*App, error) {
	// Initialize logger
	logConfig := logger.Config{
		Environment: "development",
		LogPath:     "logs",
		Level:       "debug",
	}
	log, err := logger.NewLogger(logConfig)
	if err != nil {
		return nil, err
	}
	// Initialize router
	router := gin.Default()
	// Initialize emitter
	emitter := &emitter.Emitter{}
	// Initialize database (you'll need to implement this)
	db, err := initDB(cfg)
	if err != nil {
		return nil, err
	}
	// Initialize storage
	storageConfig := storage.Config{
		Provider:  cfg.StorageProvider,
		Path:      cfg.StoragePath,
		BaseURL:   cfg.StorageBaseURL,
		APIKey:    cfg.StorageAPIKey,
		APISecret: cfg.StorageAPISecret,
		Endpoint:  cfg.StorageEndpoint,
		Bucket:    cfg.StorageBucket,
		CDN:       cfg.CDN,
	}
	activeStorage, err := storage.NewActiveStorage(db, storageConfig)
	if err != nil {
		return nil, err
	}
	app := &App{
		DB:      db,
		Router:  router,
		Log:     log,
		Emitter: emitter,
		Storage: activeStorage,
		Modules: make([]module.Module, 0),
	}
	// Initialize modules
	moduleInitializer := &AppModuleInitializer{
		DB:      db,
		Router:  router.Group("/api"),
		Logger:  log,
		Emitter: emitter,
		Storage: activeStorage,
	}
	app.Modules = moduleInitializer.InitializeModules(db)
	return app, nil
}

// AppModuleInitializer holds all dependencies needed for app module initialization
type AppModuleInitializer struct {
	DB      *gorm.DB
	Router  *gin.RouterGroup
	Logger  logger.Logger
	Emitter *emitter.Emitter
	Storage *storage.ActiveStorage
}

// InitializeModules initializes all application modules
func (a *AppModuleInitializer) InitializeModules(db *gorm.DB) []module.Module {
	var modules []module.Module
	// Initialize modules
	moduleMap := a.getModules(db)
	// Register and initialize each module
	for name, mod := range moduleMap {
		if err := module.RegisterModule(name, mod); err != nil {
			a.Logger.Error("Failed to register module",
				logger.String("module", name),
				logger.String("error", err.Error()))
			continue
		}
		if err := mod.Init(); err != nil {
			a.Logger.Error("Failed to initialize module",
				logger.String("module", name),
				logger.String("error", err.Error()))
			continue
		}
		if err := mod.Migrate(); err != nil {
			a.Logger.Error("Failed to migrate module",
				logger.String("module", name),
				logger.String("error", err.Error()))
			continue
		}
		modules = append(modules, mod)
	}
	return modules
}

// getModules returns a map of module name to module instance
func (a *AppModuleInitializer) getModules(db *gorm.DB) map[string]module.Module {
	modules := make(map[string]module.Module)
	// Define the module initializers directly
	moduleInitializers := map[string]func(*gorm.DB, *gin.RouterGroup, logger.Logger, *emitter.Emitter, *storage.ActiveStorage) module.Module{

		// MODULE_INITIALIZER_MARKER - Do not remove this comment because it's used by the CLI to add new module initializers
	}

	// Initialize and register each module
	for name, initializer := range moduleInitializers {
		modules[name] = initializer(db, a.Router, a.Logger, a.Emitter, a.Storage)
	}

	return modules
}

// initDB initializes the database connection
func initDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := database.InitDB(cfg)
	if err != nil {
		return nil, err
	}
	return db.DB, nil
}
