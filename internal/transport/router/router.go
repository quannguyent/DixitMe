package router

import (
	"dixitme/internal/logger"
	"dixitme/internal/services/auth"
	"dixitme/internal/transport/handlers"
	websocketHandler "dixitme/internal/transport/websocket"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RouterDependencies holds all the dependencies needed to set up routes
type RouterDependencies struct {
	AuthHandlers   *auth.AuthHandlers
	JWTService     *auth.JWTService
	GameHandlers   *handlers.GameHandlers
	PlayerHandlers *handlers.PlayerHandlers
	CardHandlers   *handlers.CardHandlers
	TagHandlers    *handlers.TagHandlers
	AdminHandlers  *handlers.AdminHandlers
	ChatHandlers   *handlers.ChatHandlers
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
		setupPlayerRoutes(api, deps)
		setupGameRoutes(api, deps)
		setupCardRoutes(api, deps)
		setupTagRoutes(api, deps)
		setupBotRoutes(api, deps)
		setupAdminRoutes(api, deps)
		setupChatRoutes(api, deps)
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
func setupPlayerRoutes(api *gin.RouterGroup, deps *RouterDependencies) {
	// Player routes (allow both auth and guest)
	playerGroup := api.Group("/players")
	playerGroup.Use(auth.GuestOrAuth(deps.JWTService))
	{
		playerGroup.POST("", handlers.CreatePlayer)
		playerGroup.GET("/:id", handlers.GetPlayer)
	}

	// Player stats routes (separate to avoid route conflicts)
	playerStatsGroup := api.Group("/player")
	playerStatsGroup.Use(auth.GuestOrAuth(deps.JWTService))
	{
		playerStatsGroup.GET("/:player_id/stats", handlers.GetPlayerStats)
		playerStatsGroup.GET("/:player_id/history", handlers.GetGameHistory)
	}
}

// setupGameRoutes configures game management routes
func setupGameRoutes(api *gin.RouterGroup, deps *RouterDependencies) {
	gameGroup := api.Group("/games")
	gameGroup.Use(auth.GuestOrAuth(deps.JWTService))
	{
		gameGroup.GET("", deps.GameHandlers.GetGames)
		gameGroup.GET("/:room_code", deps.GameHandlers.GetGame)
		gameGroup.POST("/add-bot", deps.GameHandlers.AddBotToGame)
	}
}

// setupCardRoutes configures card management routes
func setupCardRoutes(api *gin.RouterGroup, deps *RouterDependencies) {
	cardsGroup := api.Group("/cards")
	{
		// Public card routes
		cardsGroup.GET("", handlers.ListCards)
		cardsGroup.GET("/legacy", handlers.GetCards)
		cardsGroup.GET("/:card_id", handlers.GetCardWithTags)

		// Protected card routes (auth required)
		cardsGroup.POST("", auth.RequireAuth(deps.JWTService), handlers.CreateCard)
		cardsGroup.POST("/:card_id/image", auth.RequireAuth(deps.JWTService), handlers.UploadCardImage)
	}
}

// setupTagRoutes configures tag management routes
func setupTagRoutes(api *gin.RouterGroup, deps *RouterDependencies) {
	tagsGroup := api.Group("/tags")
	{
		tagsGroup.GET("", handlers.ListTags)                                      // Public
		tagsGroup.POST("", auth.RequireAuth(deps.JWTService), handlers.CreateTag) // Auth required
	}
}

// setupBotRoutes configures bot management routes
func setupBotRoutes(api *gin.RouterGroup, deps *RouterDependencies) {
	botGroup := api.Group("/bots")
	{
		botGroup.GET("/stats", deps.GameHandlers.GetBotStats) // Public
	}
}

// setupAdminRoutes configures admin routes (all require authentication)
func setupAdminRoutes(api *gin.RouterGroup, deps *RouterDependencies) {
	adminGroup := api.Group("/admin")
	adminGroup.Use(auth.RequireAuth(deps.JWTService)) // All admin routes require auth
	{
		adminGroup.POST("/seed", handlers.SeedDatabase)
		adminGroup.POST("/seed/tags", handlers.SeedTags)
		adminGroup.POST("/seed/cards", handlers.SeedCards)
		adminGroup.GET("/stats", handlers.GetDatabaseStats)
		adminGroup.POST("/cleanup", handlers.CleanupOldGames)
	}
}

// setupChatRoutes configures chat routes
func setupChatRoutes(api *gin.RouterGroup, deps *RouterDependencies) {
	chatGroup := api.Group("/chat")
	chatGroup.Use(auth.GuestOrAuth(deps.JWTService))
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
