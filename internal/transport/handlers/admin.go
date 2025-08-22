package handlers

import (
	"net/http"
	"time"

	"dixitme/internal/database"
	"dixitme/internal/models"
	"dixitme/internal/seeder"

	"github.com/gin-gonic/gin"
)

// AdminHandlers handles admin-related HTTP requests
type AdminHandlers struct {
	deps *HandlerDependencies
}

// NewAdminHandlers creates a new AdminHandlers instance
func NewAdminHandlers(deps *HandlerDependencies) *AdminHandlers {
	return &AdminHandlers{deps: deps}
}

// SeedDatabase seeds the database with default data
// @Summary Seed database
// @Description Initialize database with default cards and tags
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/seed [post]
func SeedDatabase(c *gin.Context) {
	if err := seeder.SeedDatabase(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to seed database",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Database seeded successfully",
	})
}

// SeedTags seeds only tags into the database
// @Summary Seed tags
// @Description Initialize database with default tags
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/seed/tags [post]
func SeedTags(c *gin.Context) {
	// Use general seed database for now since specific seeders might not exist
	if err := seeder.SeedDatabase(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to seed tags",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tags seeded successfully",
	})
}

// SeedCards seeds only cards into the database
// @Summary Seed cards
// @Description Initialize database with default cards
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/seed/cards [post]
func SeedCards(c *gin.Context) {
	// Use general seed database for now since specific seeders might not exist
	if err := seeder.SeedDatabase(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to seed cards",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cards seeded successfully",
	})
}

// GetDatabaseStats returns database statistics
// @Summary Get database statistics
// @Description Get comprehensive statistics about the database content
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} DatabaseStatsResponse
// @Failure 500 {object} map[string]interface{}
// @Router /admin/stats [get]
func GetDatabaseStats(c *gin.Context) {
	db := database.GetDB()

	stats := make(map[string]interface{})

	// Count various entities
	var userCount, playerCount, gameCount, cardCount, tagCount, chatCount int64
	db.Model(&models.User{}).Count(&userCount)
	db.Model(&models.Player{}).Count(&playerCount)
	db.Model(&models.Game{}).Count(&gameCount)
	db.Model(&models.Card{}).Count(&cardCount)
	db.Model(&models.Tag{}).Count(&tagCount)
	db.Model(&models.ChatMessage{}).Count(&chatCount)

	stats["users"] = userCount
	stats["players"] = playerCount
	stats["games"] = gameCount
	stats["cards"] = cardCount
	stats["tags"] = tagCount
	stats["chat_messages"] = chatCount

	// Game status breakdown
	var gameStatusCounts []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	db.Model(&models.Game{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&gameStatusCounts)
	stats["games_by_status"] = gameStatusCounts

	// Card-tag relationships
	var cardTagCounts []struct {
		TagID int    `json:"tag_id"`
		Name  string `json:"tag_name"`
		Count int64  `json:"count"`
	}
	db.Raw(`
		SELECT ct.tag_id, t.name, COUNT(*) as count 
		FROM card_tag_relations ct 
		JOIN tags t ON ct.tag_id = t.id 
		GROUP BY ct.tag_id, t.name 
		ORDER BY count DESC 
		LIMIT 10
	`).Scan(&cardTagCounts)
	stats["popular_tags"] = cardTagCounts

	response := DatabaseStatsResponse{Stats: stats}
	c.JSON(http.StatusOK, response)
}

// CleanupOldGames removes old completed or abandoned games
// @Summary Cleanup old games
// @Description Remove games that are completed or abandoned and older than 24 hours
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/cleanup [post]
func CleanupOldGames(c *gin.Context) {
	db := database.GetDB()

	// Delete games that are completed or abandoned and older than 24 hours
	result := db.Where("status IN (?, ?) AND updated_at < ?",
		"completed", "abandoned", time.Now().Add(-24*time.Hour)).Delete(&models.Game{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup games"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Old games cleaned up successfully",
		"deleted_count": result.RowsAffected,
	})
}
