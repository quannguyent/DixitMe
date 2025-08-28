package handlers

import (
	"net/http"
	"time"

	"dixitme/internal/database"
	"dixitme/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ChatHandlers handles chat-related HTTP requests
type ChatHandlers struct {
	deps *HandlerDependencies
}

// NewChatHandlers creates a new ChatHandlers instance
func NewChatHandlers(deps *HandlerDependencies) *ChatHandlers {
	return &ChatHandlers{deps: deps}
}

// SendChatMessage sends a chat message to a game
// @Summary Send chat message
// @Description Send a chat message to a specific game
// @Tags chat
// @Accept json
// @Produce json
// @Param message body SendChatMessageRequest true "Chat message data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /chat/send [post]
func SendChatMessage(c *gin.Context) {
	var req SendChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Get player ID from authenticated session
	// For now, using a placeholder
	playerID := uuid.New() // This should come from auth context

	gameID, err := uuid.Parse(req.GameID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	// Set default message type
	if req.MessageType == "" {
		req.MessageType = "chat"
	}

	chatMessage := models.ChatMessage{
		GameID:      gameID,
		PlayerID:    &playerID,
		Message:     req.Message,
		MessageType: req.MessageType,
		Phase:       "lobby", // TODO: Get actual game phase
		IsVisible:   true,
	}

	db := database.GetDB()
	if err := db.Create(&chatMessage).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Message sent successfully",
	})
}

// GetChatHistory gets chat history for a game
// @Summary Get chat history
// @Description Get chat message history for a specific game
// @Tags chat
// @Accept json
// @Produce json
// @Param game_id query string true "Game ID"
// @Param limit query int false "Number of messages to retrieve" default(50)
// @Param offset query int false "Number of messages to skip" default(0)
// @Success 200 {object} ChatHistoryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /chat/history [get]
func GetChatHistory(c *gin.Context) {
	var req GetChatHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gameID, err := uuid.Parse(req.GameID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid game ID"})
		return
	}

	// Set defaults
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 50
	}

	db := database.GetDB()

	// Get total count
	var total int64
	db.Model(&models.ChatMessage{}).
		Where("game_id = ? AND is_visible = ?", gameID, true).
		Count(&total)

	// Get messages with pagination
	var messages []models.ChatMessage
	err = db.Preload("Player").
		Where("game_id = ? AND is_visible = ?", gameID, true).
		Order("created_at DESC").
		Limit(req.Limit).
		Offset(req.Offset).
		Find(&messages).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chat history"})
		return
	}

	response := ChatHistoryResponse{
		Messages: messages,
		Total:    total,
	}

	c.JSON(http.StatusOK, response)
}

// GetChatStats gets chat statistics
// @Summary Get chat statistics
// @Description Get statistics about chat usage across all games
// @Tags chat
// @Accept json
// @Produce json
// @Success 200 {object} ChatStatsResponse
// @Failure 500 {object} map[string]interface{}
// @Router /chat/stats [get]
func GetChatStats(c *gin.Context) {
	db := database.GetDB()

	var totalMessages int64
	db.Model(&models.ChatMessage{}).Count(&totalMessages)

	var activeGames int64
	db.Model(&models.Game{}).Where("status = ?", "in_progress").Count(&activeGames)

	var messagesToday int64
	today := time.Now().Truncate(24 * time.Hour)
	db.Model(&models.ChatMessage{}).
		Where("created_at >= ?", today).
		Count(&messagesToday)

	var messagesThisWeek int64
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday()))
	weekStart = weekStart.Truncate(24 * time.Hour)
	db.Model(&models.ChatMessage{}).
		Where("created_at >= ?", weekStart).
		Count(&messagesThisWeek)

	stats := ChatStatsResponse{
		TotalMessages:    totalMessages,
		ActiveGames:      activeGames,
		MessagesToday:    messagesToday,
		MessagesThisWeek: messagesThisWeek,
	}

	c.JSON(http.StatusOK, stats)
}
