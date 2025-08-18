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
	"dixitme/internal/bot"
	"dixitme/internal/config"
	"dixitme/internal/database"
	"dixitme/internal/handlers"
	"dixitme/internal/logger"
	"dixitme/internal/redis"
	"dixitme/internal/seeder"
	"dixitme/internal/storage"
	websocketHandler "dixitme/internal/websocket"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	// Create Gin router (without default logger)
	r := gin.New()

	// Add recovery middleware
	r.Use(gin.Recovery())

	// Add our custom logger middleware
	r.Use(logger.GinLogger())

	// Add CORS middleware
	r.Use(handlers.CORSMiddleware())

	// Swagger documentation endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint
	r.GET("/health", handlers.HealthCheck)

	// API routes
	api := r.Group("/api/v1")
	{
		// Player routes
		api.POST("/players", handlers.CreatePlayer)
		api.GET("/players/:id", handlers.GetPlayer)
		api.GET("/player/:player_id/stats", handlers.GetPlayerStats)
		api.GET("/player/:player_id/history", handlers.GetGameHistory)

		// Game routes
		api.GET("/games", handlers.GetGames)
		api.GET("/games/:room_code", handlers.GetGame)

		// Card management routes
		api.POST("/cards", handlers.CreateCard)
		api.GET("/cards", handlers.ListCards)
		api.GET("/cards/:card_id", handlers.GetCardWithTags)
		api.POST("/cards/:card_id/image", handlers.UploadCardImage)

		// Legacy card route for compatibility
		api.GET("/cards/legacy", handlers.GetCards)

		// Tag management routes
		api.POST("/tags", handlers.CreateTag)
		api.GET("/tags", handlers.ListTags)

		// Bot management routes
		api.POST("/games/add-bot", handlers.AddBotToGame)
		api.GET("/bots/stats", handlers.GetBotStats)

		// Admin routes for database management
		api.POST("/admin/seed", handlers.SeedDatabase)
		api.POST("/admin/seed/tags", handlers.SeedTags)
		api.POST("/admin/seed/cards", handlers.SeedCards)
		api.GET("/admin/stats", handlers.GetDatabaseStats)

		// Chat routes
		api.POST("/chat/send", handlers.SendChatMessage)
		api.GET("/chat/history", handlers.GetChatHistory)
		api.GET("/chat/stats", handlers.GetChatStats)
	}

	// WebSocket endpoint
	r.GET("/ws", websocketHandler.HandleWebSocket)

	// Serve static files (for card images in development)
	r.Static("/cards", "./assets/cards")
	r.Static("/static", "./web/build/static")
	r.StaticFile("/", "./web/build/index.html")

	log.Info("Server starting", "port", cfg.Port, "gin_mode", cfg.GinMode)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Error("Failed to start server", "error", err)
	}
}
