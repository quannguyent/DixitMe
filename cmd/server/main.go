// @title DixitMe API
// @version 1.0
// @description API for DixitMe - Online Dixit Card Game
// @description This API provides endpoints for managing players, games, and real-time gameplay through WebSocket connections.
// @contact.name DixitMe API Support
// @contact.email support@dixitme.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
package main

import (
	_ "dixitme/docs" // Import docs for swagger
	"dixitme/internal/auth"
	"dixitme/internal/bot"
	"dixitme/internal/config"
	"dixitme/internal/database"
	"dixitme/internal/logger"
	"dixitme/internal/redis"
	"dixitme/internal/router"
	"dixitme/internal/seeder"
	"dixitme/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
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

	log.Info("Server starting", "port", cfg.Port, "gin_mode", cfg.GinMode)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Error("Failed to start server", "error", err)
	}
}
