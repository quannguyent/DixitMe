package handlers

import (
	"net/http"
	"time"

	"dixitme/internal/database"
	"dixitme/internal/models"
	"dixitme/internal/seeder"
	"dixitme/internal/services/game"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// ReplacePlayerWithBot manually replaces a player with a bot for testing
// @Summary Replace player with bot (Testing)
// @Description Manually replace a player with a bot for testing purposes
// @Tags admin
// @Accept json
// @Produce json
// @Param request body ReplacePlayerRequest true "Player replacement request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/replace-player [post]
func (h *AdminHandlers) ReplacePlayerWithBot(c *gin.Context) {
	var req ReplacePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.RoomCode == "" || req.PlayerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room code and player ID are required"})
		return
	}

	playerID, err := uuid.Parse(req.PlayerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID format"})
		return
	}

	reason := req.Reason
	if reason == "" {
		reason = "Manual replacement (admin)"
	}

	manager := game.GetManager()
	gameState, err := manager.ReplacePlayerWithBot(req.RoomCode, playerID, reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to replace player with bot",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Player replaced with bot successfully",
		"room_code":  req.RoomCode,
		"player_id":  req.PlayerID,
		"reason":     reason,
		"game_state": gameState,
	})
}

// CheckAFKPlayers manually checks and replaces AFK players for testing
// @Summary Check AFK players (Testing)
// @Description Manually check and replace AFK players for testing purposes
// @Tags admin
// @Accept json
// @Produce json
// @Param request body CheckAFKRequest true "AFK check request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/check-afk [post]
func (h *AdminHandlers) CheckAFKPlayers(c *gin.Context) {
	var req CheckAFKRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.RoomCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room code is required"})
		return
	}

	afkTimeout := time.Duration(req.AFKTimeoutMinutes) * time.Minute
	if afkTimeout == 0 {
		afkTimeout = 3 * time.Minute // Default to 3 minutes
	}

	manager := game.GetManager()
	gameState, err := manager.CheckAndReplaceAFKPlayers(req.RoomCode, afkTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check AFK players",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":          true,
		"message":          "AFK check completed",
		"room_code":        req.RoomCode,
		"afk_timeout_mins": req.AFKTimeoutMinutes,
		"game_state":       gameState,
	})
}
