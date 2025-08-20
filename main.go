package main

import (
	appmodules "base/app"
	coremodules "base/core/app"
	"base/core/config"
	"base/core/database"
	"base/core/email"
	"base/core/emitter"
	"base/core/logger"
	"base/core/module"
	"base/core/router"
	"base/core/storage"
	"base/core/swagger"
	_ "base/core/translation"
	"base/core/websocket"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// @title Base API
// @version 2.0.0
// @description This is the API documentation for Base Framework, change this on main.go
// @host http://localhost:8100
// @BasePath /api
// @schemes http https
// @produces json
// @consumes json

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Api-Key
// @description API Key for authentication

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your token with the prefix "Bearer "

// DeletedAt is a type definition for GORM's soft delete functionality
type DeletedAt gorm.DeletedAt

// Time represents a time.Time
type Time time.Time

// App represents the Base application with simplified initialization
type App struct {
	config      *config.Config
	db          *database.Database
	router      *router.Router
	logger      logger.Logger
	emitter     *emitter.Emitter
	storage     *storage.ActiveStorage
	emailSender email.Sender
	swagger     *swagger.SwaggerService
	wsHub       *websocket.Hub

	// State
	running bool
}

// New creates a new Base application instance
func New() *App {
	return &App{}
}

// Start initializes and starts the application
func (app *App) Start() error {
	return app.
		loadEnvironment().
		initConfig().
		initLogger().
		initDatabase().
		initInfrastructure().
		initRouter().
		autoDiscoverModules().
		setupRoutes().
		displayServerInfo().
		run()
}

// loadEnvironment loads environment variables
func (app *App) loadEnvironment() *App {
	if err := godotenv.Load(); err != nil {
		// Non-fatal - continue without .env file
	}
	return app
}

// initConfig initializes configuration
func (app *App) initConfig() *App {
	app.config = config.NewConfig()
	return app
}

// initLogger initializes the logger
func (app *App) initLogger() *App {
	logConfig := logger.Config{
		Environment: app.config.Env,
		LogPath:     "logs",
		Level:       "debug",
	}

	log, err := logger.NewLogger(logConfig)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	app.logger = log
	app.logger.Info("🚀 Starting Base Framework",
		logger.String("version", app.config.Version),
		logger.String("environment", app.config.Env))

	return app
}

// initDatabase initializes the database connection
func (app *App) initDatabase() *App {
	db, err := database.InitDB(app.config)
	if err != nil {
		app.logger.Error("Failed to initialize database", logger.String("error", err.Error()))
		panic(fmt.Sprintf("Database initialization failed: %v", err))
	}

	app.db = db
	app.logger.Info("✅ Database initialized")
	return app
}

// initInfrastructure initializes core infrastructure components
func (app *App) initInfrastructure() *App {
	// Initialize emitter
	app.emitter = &emitter.Emitter{}

	// Initialize storage
	storageConfig := storage.Config{
		Provider:  app.config.StorageProvider,
		Path:      app.config.StoragePath,
		BaseURL:   app.config.StorageBaseURL,
		APIKey:    app.config.StorageAPIKey,
		APISecret: app.config.StorageAPISecret,
		Endpoint:  app.config.StorageEndpoint,
		Bucket:    app.config.StorageBucket,
		CDN:       app.config.CDN,
	}

	activeStorage, err := storage.NewActiveStorage(app.db.DB, storageConfig)
	if err != nil {
		app.logger.Error("Failed to initialize storage", logger.String("error", err.Error()))
		panic(fmt.Sprintf("Storage initialization failed: %v", err))
	}
	app.storage = activeStorage

	// Initialize email sender (non-fatal)
	emailSender, err := email.NewSender(app.config)
	if err != nil {
		app.logger.Warn("Email sender initialization failed - continuing without email functionality",
			logger.String("error", err.Error()))
		app.emailSender = nil
	} else {
		app.emailSender = emailSender
	}

	app.logger.Info("✅ Infrastructure initialized")
	return app
}

// initRouter initializes the router with middleware
func (app *App) initRouter() *App {
	app.router = router.New()
	app.setupMiddleware()
	app.setupStaticRoutes()
	app.setupSwagger()
	app.initWebSocket()

	app.logger.Info("✅ Router initialized")
	return app
}

// setupMiddleware configures all middleware
func (app *App) setupMiddleware() {
	// Recovery middleware
	app.router.Use(func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			defer func() {
				if r := recover(); r != nil {
					app.logger.Error("Panic recovered", logger.Any("panic", r))
					c.JSON(500, map[string]any{"error": "Internal server error"})
				}
			}()
			return next(c)
		}
	})

	// Request logging middleware
	app.router.Use(func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			start := time.Now()
			err := next(c)

			app.logger.Info("Request",
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.Int("status", c.Writer.Status()),
				logger.Duration("duration", time.Since(start)),
				logger.String("ip", c.ClientIP()),
			)
			return err
		}
	})

	// CORS middleware
	app.router.Use(func(next router.HandlerFunc) router.HandlerFunc {
		return func(c *router.Context) error {
			c.SetHeader("Access-Control-Allow-Origin", "*")
			c.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.SetHeader("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Api-Key, Base-Orgid")
			c.SetHeader("Access-Control-Expose-Headers", "Content-Length, Content-Type")
			c.SetHeader("Access-Control-Allow-Credentials", "true")
			c.SetHeader("Access-Control-Max-Age", "43200")

			if c.Request.Method == "OPTIONS" {
				return c.NoContent()
			}

			return next(c)
		}
	})
}

// setupStaticRoutes configures static file serving
func (app *App) setupStaticRoutes() {
	app.router.Static("/static", "./static")
	app.router.Static("/storage", "./storage")
	app.router.Static("/docs", "./docs")
}

// setupSwagger initializes swagger documentation
func (app *App) setupSwagger() {
	app.swagger = swagger.NewSwaggerService(app.config)

	if app.config.SwaggerEnabled {
		app.swagger.RegisterRoutes(app.router)
		app.logger.Info("✅ Swagger documentation enabled at /swagger/")
	} else {
		app.logger.Info("⏩ Swagger documentation disabled via SWAGGER_ENABLED=false")
	}
}

// initWebSocket initializes the WebSocket hub if enabled
func (app *App) initWebSocket() {
	if !app.config.WebSocketEnabled {
		app.logger.Info("⏩ WebSocket disabled via WS_ENABLED=false")
		return
	}
	
	app.wsHub = websocket.InitWebSocketModule(app.router.Group("/api"))
	app.logger.Info("✅ WebSocket hub initialized")
}

// autoDiscoverModules automatically discovers and registers modules
func (app *App) autoDiscoverModules() *App {
	app.registerCoreModules()
	app.discoverAndRegisterAppModules()

	app.logger.Info("✅ Modules auto-discovered and registered")
	return app
}

// registerCoreModules registers core framework modules
func (app *App) registerCoreModules() {
	// Create dependencies for core modules
	deps := module.Dependencies{
		DB:          app.db.DB,
		Router:      app.router.Group("/api"),
		Logger:      app.logger,
		Emitter:     app.emitter,
		Storage:     app.storage,
		EmailSender: app.emailSender,
		Config:      app.config,
	}

	// Initialize core modules via orchestrator to ensure proper init/migrate/routes
	// Auth-related modules will mount under /api/auth
	initializer := module.NewInitializer(app.logger)
	authRouter := app.router.Group("/api/auth")
	coreProvider := coremodules.NewCoreModules()
	orchestrator := module.NewCoreOrchestrator(initializer, coreProvider, authRouter)

	initialized, err := orchestrator.InitializeCoreModules(deps)
	if err != nil {
		app.logger.Error("Failed to initialize core modules", logger.String("error", err.Error()))
	}

	app.logger.Info("✅ Core modules registered", logger.Int("count", len(initialized)))
}

// discoverAndRegisterAppModules registers application modules using the app provider
func (app *App) discoverAndRegisterAppModules() {
	// Create dependencies for app modules
	deps := module.Dependencies{
		DB:          app.db.DB,
		Router:      app.router.Group("/api"),
		Logger:      app.logger,
		Emitter:     app.emitter,
		Storage:     app.storage,
		EmailSender: app.emailSender,
		Config:      app.config,
	}

	// Use app module provider (like core modules)
	appProvider := appmodules.NewAppModules()
	modules := appProvider.GetAppModules(deps)

	if len(modules) == 0 {
		app.logger.Info("No app modules found")
		return
	}

	app.logger.Info("✅ App modules loaded", logger.Int("count", len(modules)))
	app.initializeModules(modules, deps)
}

// initializeModules initializes a collection of modules
func (app *App) initializeModules(modules map[string]module.Module, deps module.Dependencies) {
	initializer := module.NewInitializer(app.logger)
	initializedModules := initializer.Initialize(modules, deps)

	app.logger.Info("✅ Module initialization complete",
		logger.Int("total", len(modules)),
		logger.Int("initialized", len(initializedModules)))
}

// setupRoutes sets up basic system routes
func (app *App) setupRoutes() *App {
	// Health check
	app.router.GET("/health", func(c *router.Context) error {
		return c.JSON(200, map[string]any{
			"status":  "ok",
			"version": app.config.Version,
		})
	})

	// Root endpoint
	app.router.GET("/", func(c *router.Context) error {
		return c.JSON(200, map[string]any{
			"message": "pong",
			"version": app.config.Version,
			"swagger": "/swagger/index.html",
		})
	})

	return app
}

// displayServerInfo shows server startup information
func (app *App) displayServerInfo() *App {
	localIP := app.getLocalIP()
	port := app.config.ServerPort

	fmt.Printf("\n🎉 Base Framework Ready!\n\n")
	fmt.Printf("📍 Server URLs:\n")
	fmt.Printf("   • Local:   http://localhost%s\n", port)
	fmt.Printf("   • Network: http://%s%s\n", localIP, port)
	fmt.Printf("\n📚 Documentation:\n")
	fmt.Printf("   • Swagger: http://localhost%s/swagger/index.html\n", port)
	fmt.Printf("\n")

	return app
}

// getLocalIP gets the local network IP address
func (app *App) getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

// run starts the HTTP server
func (app *App) run() error {
	app.running = true
	port := app.config.ServerPort

	app.logger.Info("🌐 Server starting",
		logger.String("port", port))

	err := app.router.Run(port)
	if err != nil {
		// Check if it's an "address already in use" error
		if strings.Contains(err.Error(), "bind: address already in use") {
			app.logger.Error("❌ Server failed to start - Port already in use",
				logger.String("port", port),
				logger.String("error", err.Error()))
			return fmt.Errorf("port %s is already in use. Please:\n  • Stop any other servers running on this port\n  • Change the SERVER_PORT in your .env file\n  • Use a different port with: export SERVER_PORT=:8101", port)
		}
		// For other network errors, provide a generic helpful message
		app.logger.Error("❌ Server failed to start",
			logger.String("error", err.Error()))
		return fmt.Errorf("server failed to start: %w", err)
	}
	return nil
}

// Graceful shutdown (future enhancement)
func (app *App) Stop() error {
	if !app.running {
		return nil
	}

	app.logger.Info("🛑 Shutting down gracefully...")
	app.running = false
	return nil
}

// GenerateSwaggerDocs generates static Swagger documentation files
func (app *App) GenerateSwaggerDocs(outputDir string) error {
	// Initialize only what's needed for Swagger generation
	app.loadEnvironment().initConfig()

	// Initialize minimal logger for output
	logConfig := logger.Config{
		Environment: app.config.Env,
		LogPath:     "logs",
		Level:       "info",
	}

	log, err := logger.NewLogger(logConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	app.logger = log

	// Initialize swagger service
	swaggerService := swagger.NewSwaggerService(app.config)

	// Generate static files
	return swaggerService.GenerateStaticFiles(outputDir)
}

func main() {
	// Parse command line flags
	var generateDocs = flag.Bool("generate-docs", false, "Generate static Swagger documentation files")
	var docsOutput = flag.String("docs-output", "docs", "Output directory for generated documentation")
	flag.Parse()

	// Initialize the Base application
	app := New()

	// Handle documentation generation
	if *generateDocs {
		fmt.Println("🔧 Generating static Swagger documentation...")
		if err := app.GenerateSwaggerDocs(*docsOutput); err != nil {
			fmt.Printf("\n❌ Documentation generation failed:\n%v\n\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Documentation generated successfully!")
		return
	}

	// Normal application startup
	if err := app.Start(); err != nil {
		// Print user-friendly error message instead of panicking
		fmt.Printf("\n❌ Application failed to start:\n%v\n\n", err)
		os.Exit(1)
	}
}
