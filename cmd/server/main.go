package main

import (
	"log"

	"dixitme/internal/config"
	"dixitme/internal/database"
	"dixitme/internal/handlers"
	"dixitme/internal/redis"
	websocketHandler "dixitme/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize database
	database.Initialize(cfg.DatabaseURL)

	// Initialize Redis
	redis.Initialize(cfg.RedisURL)

	// Create Gin router
	r := gin.Default()

	// Add CORS middleware
	r.Use(handlers.CORSMiddleware())

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

		// Card routes
		api.GET("/cards", handlers.GetCards)
	}

	// WebSocket endpoint
	r.GET("/ws", websocketHandler.HandleWebSocket)

	// Serve static files (for card images in development)
	r.Static("/cards", "./assets/cards")
	r.Static("/static", "./web/build/static")
	r.StaticFile("/", "./web/build/index.html")

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
