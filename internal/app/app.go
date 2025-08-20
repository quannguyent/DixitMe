// Package app contains the application initialization and dependency injection logic
package app

import (
	"dixitme/internal/config"
	"dixitme/internal/database"
	"dixitme/internal/logger"
	"dixitme/internal/redis"
	"dixitme/internal/seeder"
	"dixitme/internal/services/auth"
	"dixitme/internal/services/bot"
	"dixitme/internal/storage"
	"dixitme/internal/transport/router"

	"github.com/gin-gonic/gin"
)

// App represents the application instance with all its dependencies
type App struct {
	Router  *gin.Engine
	Config  *config.Config
	Cleanup func() // Cleanup function for graceful shutdown
}

// NewApp creates and initializes a new application instance
func NewApp() (*App, error) {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.InitLogger(cfg.Logger)
	log := logger.GetLogger()

	log.Info("Starting DixitMe server", "version", "1.0")

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize database
	database.Initialize(cfg.DatabaseURL)

	// Initialize Redis
	redis.Initialize(cfg.RedisURL)

	// Initialize MinIO storage
	if err := storage.Initialize(cfg.MinIO); err != nil {
		log.Error("Failed to initialize MinIO", "error", err)
		// Continue without MinIO - fallback to local storage
	}

	// Initialize bot system
	bot.Initialize()

	// Seed database with default data
	if err := seeder.SeedDatabase(); err != nil {
		log.Error("Failed to seed database", "error", err)
		// Continue without seeding - not critical for startup
	}

	// Initialize authentication services
	jwtService := auth.NewJWTService(cfg.Auth.JWTSecret)
	authService := auth.NewAuthService(jwtService)
	authHandlers := auth.NewAuthHandlers(authService, jwtService, cfg.Auth.EnableSSO)

	// Setup router with dependencies
	routerDeps := &router.RouterDependencies{
		AuthHandlers: authHandlers,
		JWTService:   jwtService,
	}
	r := router.SetupRouter(routerDeps)

	// Create cleanup function
	cleanup := func() {
		log.Info("Shutting down application...")

		// Close database connection
		if db := database.GetDB(); db != nil {
			if sqlDB, err := db.DB(); err == nil {
				sqlDB.Close()
			}
		}

		// Close Redis connection
		redis.Close()

		log.Info("Application shutdown complete")
	}

	return &App{
		Router:  r,
		Config:  cfg,
		Cleanup: cleanup,
	}, nil
}

// Run starts the application server
func (a *App) Run() error {
	log := logger.GetLogger()

	log.Info("Server starting",
		"port", a.Config.Port,
		"gin_mode", a.Config.GinMode,
	)

	return a.Router.Run(":" + a.Config.Port)
}
