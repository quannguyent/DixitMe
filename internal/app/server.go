package app

import (
	"fmt"

	_ "dixitme/docs" // swagger docs
	"dixitme/internal/auth"
	authhttp "dixitme/internal/auth/http"
	"dixitme/internal/config"
	"dixitme/internal/data"
	"dixitme/internal/data/postgres"
	redisstore "dixitme/internal/data/redis"
	"dixitme/internal/data/seeder"
	"dixitme/internal/game/bot"
	"dixitme/internal/game/core"
	handlers "dixitme/internal/game/http"
	ws "dixitme/internal/game/ws"
	"dixitme/internal/platform/logger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// NewServer builds the HTTP server with all dependencies wired.
// This acts as the composition root to keep entrypoints thin and adapters isolated.
func NewServer(cfg *config.Config) (*gin.Engine, error) {
	logger.InitLogger(cfg.Logger)
	log := logger.GetLogger()

	gin.SetMode(cfg.GinMode)

	// Infrastructure
	store := data.MustStore(cfg)
	db := store.DB
	redisClient := store.Redis

	// Core services
	gameRepo := postgres.NewGameRepository(db)
	gameCache := redisstore.NewGameCache(redisClient)
	jwtService := auth.NewJWTService(cfg.Auth.JWTSecret)
	hub := ws.NewHub(jwtService, redisClient)
	manager := game.NewManager(gameRepo, gameCache, hub)
	hub.SetManager(manager)
	handlers.InitManager(manager)

	// Bot system uses DB-backed data provider
	botRepo := postgres.NewBotRepository(db)
	bot.Initialize(botRepo)

	authService := auth.NewAuthService(jwtService, db)
	authHandlers := authhttp.NewAuthHandlers(authService, jwtService, cfg.Auth.EnableSSO)
	authhttp.InitStore(store)

	// Seed database with defaults (non-fatal)
	if err := seeder.SeedDatabase(); err != nil {
		log.Warn("Database seeding skipped", "error", err)
	}

	// HTTP server and middleware
	r := gin.New()
	r.Use(gin.Recovery(), logger.GinLogger(), handlers.CORSMiddleware())

	// Swagger and health
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/health", handlers.HealthCheck)

	// API routes
	api := r.Group("/api/v1")

	// Authentication routes
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", authHandlers.Register)
		authGroup.POST("/login", authHandlers.Login)
		authGroup.POST("/google", authHandlers.GoogleLogin)
		authGroup.POST("/guest", authHandlers.GuestLogin)
		authGroup.POST("/refresh", authHandlers.RefreshToken)
		authGroup.GET("/status", authHandlers.GetAuthStatus)

		authGroup.POST("/logout", authhttp.RequireAuth(jwtService), authHandlers.Logout)
		authGroup.GET("/me", authhttp.RequireAuth(jwtService), authHandlers.GetCurrentUser)
		authGroup.GET("/validate", authhttp.RequireAuth(jwtService), authHandlers.ValidateToken)
	}

	// Player routes
	playerGroup := api.Group("/players")
	playerGroup.Use(authhttp.GuestOrAuth(jwtService))
	{
		playerGroup.POST("", handlers.CreatePlayer)
		playerGroup.GET("/:id", handlers.GetPlayer)
	}

	// Player stats routes
	playerStatsGroup := api.Group("/player")
	playerStatsGroup.Use(authhttp.GuestOrAuth(jwtService))
	{
		playerStatsGroup.GET("/:player_id/stats", handlers.GetPlayerStats)
		playerStatsGroup.GET("/:player_id/history", handlers.GetGameHistory)
	}

	// Game routes
	gameGroup := api.Group("/games")
	gameGroup.Use(authhttp.GuestOrAuth(jwtService))
	{
		gameGroup.GET("", handlers.GetGames)
		gameGroup.GET("/:room_code", handlers.GetGame)
		gameGroup.POST("/add-bot", handlers.AddBotToGame)
	}

	// Card management routes
	cardsGroup := api.Group("/cards")
	{
		cardsGroup.GET("", handlers.ListCards)
		cardsGroup.GET("/legacy", handlers.GetCards)
		cardsGroup.GET("/:card_id", handlers.GetCardWithTags)
		cardsGroup.POST("", authhttp.RequireAuth(jwtService), handlers.CreateCard)
		cardsGroup.POST("/:card_id/image", authhttp.RequireAuth(jwtService), handlers.UploadCardImage)
	}

	// Tag management routes
	tagsGroup := api.Group("/tags")
	{
		tagsGroup.GET("", handlers.ListTags)
		tagsGroup.POST("", authhttp.RequireAuth(jwtService), handlers.CreateTag)
	}

	// Bot routes
	botGroup := api.Group("/bots")
	{
		botGroup.GET("/stats", handlers.GetBotStats)
	}

	// Admin routes
	adminGroup := api.Group("/admin")
	adminGroup.Use(authhttp.RequireAuth(jwtService))
	{
		adminGroup.POST("/seed", handlers.SeedDatabase)
		adminGroup.POST("/seed/tags", handlers.SeedTags)
		adminGroup.POST("/seed/cards", handlers.SeedCards)
		adminGroup.GET("/stats", handlers.GetDatabaseStats)
		adminGroup.POST("/cleanup", handlers.CleanupOldGames)
	}

	// Chat routes
	chatGroup := api.Group("/chat")
	chatGroup.Use(authhttp.GuestOrAuth(jwtService))
	{
		chatGroup.GET("/history", handlers.GetChatHistory)
		chatGroup.GET("/stats", handlers.GetChatStats)
	}

	// WebSocket endpoint with optional auth support
	r.GET("/ws", hub.HandleWebSocketWithAuth)

	// Static assets
	r.Static("/cards", "./assets/cards")
	r.Static("/static", "./web/build/static")
	r.StaticFile("/", "./web/build/index.html")

	return r, nil
}

// Run starts the server on the configured port.
func Run(cfg *config.Config) error {
	server, err := NewServer(cfg)
	if err != nil {
		return err
	}
	if err := server.Run(":" + cfg.Port); err != nil {
		return fmt.Errorf("server failed to start: %w", err)
	}
	return nil
}
