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

	// Initialize authentication services
	jwtService := auth.NewJWTService(cfg.Auth.JWTSecret)
	authService := auth.NewAuthService(jwtService)
	authHandlers := auth.NewAuthHandlers(authService, jwtService, cfg.Auth.EnableSSO)

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
		// Authentication routes (public)
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/register", authHandlers.Register)
			authGroup.POST("/login", authHandlers.Login)
			authGroup.POST("/google", authHandlers.GoogleLogin)
			authGroup.POST("/guest", authHandlers.GuestLogin)
			authGroup.POST("/refresh", authHandlers.RefreshToken)
			authGroup.GET("/status", authHandlers.GetAuthStatus)

			// Protected auth routes
			authGroup.POST("/logout", auth.RequireAuth(jwtService), authHandlers.Logout)
			authGroup.GET("/me", auth.RequireAuth(jwtService), authHandlers.GetCurrentUser)
			authGroup.GET("/validate", auth.RequireAuth(jwtService), authHandlers.ValidateToken)
		}

		// Player routes (allow both auth and guest)
		playerGroup := api.Group("/players")
		playerGroup.Use(auth.GuestOrAuth(jwtService))
		{
			playerGroup.POST("", handlers.CreatePlayer)
			playerGroup.GET("/:id", handlers.GetPlayer)
			playerGroup.GET("/:player_id/stats", handlers.GetPlayerStats)
			playerGroup.GET("/:player_id/history", handlers.GetGameHistory)
		}

		// Game routes (allow both auth and guest)
		gameGroup := api.Group("/games")
		gameGroup.Use(auth.GuestOrAuth(jwtService))
		{
			gameGroup.GET("", handlers.GetGames)
			gameGroup.GET("/:room_code", handlers.GetGame)
			gameGroup.POST("/add-bot", handlers.AddBotToGame)
		}

		// Card management routes (require auth for modifications, allow guest for reading)
		cardsGroup := api.Group("/cards")
		{
			cardsGroup.GET("", handlers.ListCards)                                                     // Public
			cardsGroup.GET("/legacy", handlers.GetCards)                                               // Public
			cardsGroup.GET("/:card_id", handlers.GetCardWithTags)                                      // Public
			cardsGroup.POST("", auth.RequireAuth(jwtService), handlers.CreateCard)                     // Auth required
			cardsGroup.POST("/:card_id/image", auth.RequireAuth(jwtService), handlers.UploadCardImage) // Auth required
		}

		// Tag management routes
		tagsGroup := api.Group("/tags")
		{
			tagsGroup.GET("", handlers.ListTags)                                 // Public
			tagsGroup.POST("", auth.RequireAuth(jwtService), handlers.CreateTag) // Auth required
		}

		// Bot management routes
		botGroup := api.Group("/bots")
		{
			botGroup.GET("/stats", handlers.GetBotStats) // Public
		}

		// Admin routes (require auth)
		adminGroup := api.Group("/admin")
		adminGroup.Use(auth.RequireAuth(jwtService)) // All admin routes require auth
		{
			adminGroup.POST("/seed", handlers.SeedDatabase)
			adminGroup.POST("/seed/tags", handlers.SeedTags)
			adminGroup.POST("/seed/cards", handlers.SeedCards)
			adminGroup.GET("/stats", handlers.GetDatabaseStats)
		}

		// Chat routes (require session - guest or auth)
		chatGroup := api.Group("/chat")
		chatGroup.Use(auth.GuestOrAuth(jwtService))
		{
			chatGroup.POST("/send", handlers.SendChatMessage)
			chatGroup.GET("/history", handlers.GetChatHistory)
			chatGroup.GET("/stats", handlers.GetChatStats)
		}
	}

	// WebSocket endpoint (with optional authentication)
	r.GET("/ws", websocketHandler.HandleWebSocketWithAuth(jwtService))

	// Serve static files (for card images in development)
	r.Static("/cards", "./assets/cards")
	r.Static("/static", "./web/build/static")
	r.StaticFile("/", "./web/build/index.html")

	log.Info("Server starting", "port", cfg.Port, "gin_mode", cfg.GinMode)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Error("Failed to start server", "error", err)
	}
}
