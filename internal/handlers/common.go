package handlers

import (
	"net/http"

	"dixitme/internal/database"
	"dixitme/internal/game"

	"github.com/gin-gonic/gin"
)

// HealthCheck performs a health check of the service
// @Summary Health check
// @Description Check the health status of the API and its dependencies
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	db := database.GetDB()
	sqlDB, err := db.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "unhealthy",
			"error":  "Failed to get database instance",
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "unhealthy",
			"error":  "Database connection failed",
		})
		return
	}

	// Check active games
	manager := game.GetManager()
	activeGamesCount := manager.GetActiveGamesCount()

	c.JSON(http.StatusOK, gin.H{
		"status":       "healthy",
		"active_games": activeGamesCount,
		"timestamp":    c.GetHeader("X-Request-ID"),
	})
}

// CORSMiddleware provides CORS support
func CORSMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}
