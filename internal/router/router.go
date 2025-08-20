package router

import (
	"dixitme/internal/auth"
	"dixitme/internal/handlers"
	"dixitme/internal/logger"
	websocketHandler "dixitme/internal/websocket"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RouterDependencies holds all the dependencies needed to set up routes
type RouterDependencies struct {
	AuthHandlers *auth.AuthHandlers
	JWTService   *auth.JWTService
}

// SetupRouter creates and configures the Gin router with all routes
func SetupRouter(deps *RouterDependencies) *gin.Engine {
	// Create Gin router (without default logger)
	r := gin.New()

	// Add recovery middleware
	r.Use(gin.Recovery())

	// Add our custom logger middleware
	r.Use(logger.GinLogger())

	// Add CORS middleware
	r.Use(handlers.CORSMiddleware())

	// Setup all routes
	setupSwaggerRoutes(r)
	setupHealthRoutes(r)
	setupAPIRoutes(r, deps)
	setupWebSocketRoutes(r, deps.JWTService)
	setupStaticRoutes(r)

	return r
}

// setupSwaggerRoutes configures Swagger documentation endpoints
func setupSwaggerRoutes(r *gin.Engine) {
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// setupHealthRoutes configures health check endpoints
func setupHealthRoutes(r *gin.Engine) {
	r.GET("/health", handlers.HealthCheck)
}

// setupAPIRoutes configures all API routes under /api/v1
func setupAPIRoutes(r *gin.Engine, deps *RouterDependencies) {
	api := r.Group("/api/v1")
	{
		setupAuthRoutes(api, deps)
		setupPlayerRoutes(api, deps.JWTService)
		setupGameRoutes(api, deps.JWTService)
		setupCardRoutes(api, deps.JWTService)
		setupTagRoutes(api, deps.JWTService)
		setupBotRoutes(api)
		setupAdminRoutes(api, deps.JWTService)
		setupChatRoutes(api, deps.JWTService)
	}
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(api *gin.RouterGroup, deps *RouterDependencies) {
	authGroup := api.Group("/auth")
	{
		// Public auth routes
		authGroup.POST("/register", deps.AuthHandlers.Register)
		authGroup.POST("/login", deps.AuthHandlers.Login)
		authGroup.POST("/google", deps.AuthHandlers.GoogleLogin)
		authGroup.POST("/guest", deps.AuthHandlers.GuestLogin)
		authGroup.POST("/refresh", deps.AuthHandlers.RefreshToken)
		authGroup.GET("/status", deps.AuthHandlers.GetAuthStatus)

		// Protected auth routes
		authGroup.POST("/logout", auth.RequireAuth(deps.JWTService), deps.AuthHandlers.Logout)
		authGroup.GET("/me", auth.RequireAuth(deps.JWTService), deps.AuthHandlers.GetCurrentUser)
		authGroup.GET("/validate", auth.RequireAuth(deps.JWTService), deps.AuthHandlers.ValidateToken)
	}
}

// setupPlayerRoutes configures player management routes
func setupPlayerRoutes(api *gin.RouterGroup, jwtService *auth.JWTService) {
	// Player routes (allow both auth and guest)
	playerGroup := api.Group("/players")
	playerGroup.Use(auth.GuestOrAuth(jwtService))
	{
		playerGroup.POST("", handlers.CreatePlayer)
		playerGroup.GET("/:id", handlers.GetPlayer)
	}

	// Player stats routes (separate to avoid route conflicts)
	playerStatsGroup := api.Group("/player")
	playerStatsGroup.Use(auth.GuestOrAuth(jwtService))
	{
		playerStatsGroup.GET("/:player_id/stats", handlers.GetPlayerStats)
		playerStatsGroup.GET("/:player_id/history", handlers.GetGameHistory)
	}
}

// setupGameRoutes configures game management routes
func setupGameRoutes(api *gin.RouterGroup, jwtService *auth.JWTService) {
	gameGroup := api.Group("/games")
	gameGroup.Use(auth.GuestOrAuth(jwtService))
	{
		gameGroup.GET("", handlers.GetGames)
		gameGroup.GET("/:room_code", handlers.GetGame)
		gameGroup.POST("/add-bot", handlers.AddBotToGame)
	}
}

// setupCardRoutes configures card management routes
func setupCardRoutes(api *gin.RouterGroup, jwtService *auth.JWTService) {
	cardsGroup := api.Group("/cards")
	{
		// Public card routes
		cardsGroup.GET("", handlers.ListCards)
		cardsGroup.GET("/legacy", handlers.GetCards)
		cardsGroup.GET("/:card_id", handlers.GetCardWithTags)

		// Protected card routes (auth required)
		cardsGroup.POST("", auth.RequireAuth(jwtService), handlers.CreateCard)
		cardsGroup.POST("/:card_id/image", auth.RequireAuth(jwtService), handlers.UploadCardImage)
	}
}

// setupTagRoutes configures tag management routes
func setupTagRoutes(api *gin.RouterGroup, jwtService *auth.JWTService) {
	tagsGroup := api.Group("/tags")
	{
		tagsGroup.GET("", handlers.ListTags)                                 // Public
		tagsGroup.POST("", auth.RequireAuth(jwtService), handlers.CreateTag) // Auth required
	}
}

// setupBotRoutes configures bot management routes
func setupBotRoutes(api *gin.RouterGroup) {
	botGroup := api.Group("/bots")
	{
		botGroup.GET("/stats", handlers.GetBotStats) // Public
	}
}

// setupAdminRoutes configures admin routes (all require authentication)
func setupAdminRoutes(api *gin.RouterGroup, jwtService *auth.JWTService) {
	adminGroup := api.Group("/admin")
	adminGroup.Use(auth.RequireAuth(jwtService)) // All admin routes require auth
	{
		adminGroup.POST("/seed", handlers.SeedDatabase)
		adminGroup.POST("/seed/tags", handlers.SeedTags)
		adminGroup.POST("/seed/cards", handlers.SeedCards)
		adminGroup.GET("/stats", handlers.GetDatabaseStats)
		adminGroup.POST("/cleanup", handlers.CleanupOldGames)
	}
}

// setupChatRoutes configures chat routes
func setupChatRoutes(api *gin.RouterGroup, jwtService *auth.JWTService) {
	chatGroup := api.Group("/chat")
	chatGroup.Use(auth.GuestOrAuth(jwtService))
	{
		chatGroup.POST("/send", handlers.SendChatMessage)
		chatGroup.GET("/history", handlers.GetChatHistory)
		chatGroup.GET("/stats", handlers.GetChatStats)
	}
}

// setupWebSocketRoutes configures WebSocket endpoints
func setupWebSocketRoutes(r *gin.Engine, jwtService *auth.JWTService) {
	r.GET("/ws", websocketHandler.HandleWebSocketWithAuth(jwtService))
}

// setupStaticRoutes configures static file serving
func setupStaticRoutes(r *gin.Engine) {
	// Serve static files (for card images in development)
	r.Static("/cards", "./assets/cards")
	r.Static("/static", "./web/build/static")
	r.StaticFile("/", "./web/build/index.html")
}
