package handlers

import (
	"net/http"

	"dixitme/internal/database"
	"dixitme/internal/models"
	"dixitme/internal/services/game"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GameHandlers handles game-related HTTP requests
type GameHandlers struct {
	deps *HandlerDependencies
}

// NewGameHandlers creates a new GameHandlers instance
func NewGameHandlers(deps *HandlerDependencies) *GameHandlers {
	return &GameHandlers{deps: deps}
}

// GetGame gets a game by room code
// @Summary Get game by room code
// @Description Get game information and check if it's currently live
// @Tags games
// @Accept json
// @Produce json
// @Param room_code path string true "Game room code"
// @Success 200 {object} GetGameResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /games/{room_code} [get]
func (h *GameHandlers) GetGame(c *gin.Context) {
	var req GetGameRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var dbGame models.Game
	if err := db.Preload("Players.Player").Preload("Rounds").First(&dbGame, "room_code = ?", req.RoomCode).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	// Check if game is live (exists in memory)
	liveGame := h.deps.GameService.GetGame(req.RoomCode)
	isLive := liveGame != nil

	response := GetGameResponse{
		Game:   &dbGame,
		IsLive: isLive,
	}

	c.JSON(http.StatusOK, response)
}

// GetGames gets all games
// @Summary Get all games
// @Description Get a list of all games with their current status
// @Tags games
// @Accept json
// @Produce json
// @Success 200 {object} GetGamesResponse
// @Failure 500 {object} map[string]string
// @Router /games [get]
func (h *GameHandlers) GetGames(c *gin.Context) {
	db := database.GetDB()
	var games []models.Game

	// Get all games with their players and rounds
	if err := db.Preload("Players.Player").Preload("Rounds").Find(&games).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch games"})
		return
	}

	response := GetGamesResponse{Games: games}
	c.JSON(http.StatusOK, response)
}

// AddBotToGame adds a bot to an existing game
// @Summary Add bot to game
// @Description Add an AI bot player to an existing game
// @Tags games
// @Accept json
// @Produce json
// @Param bot body AddBotRequest true "Bot information"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/games/add-bot [post]
func (h *GameHandlers) AddBotToGame(c *gin.Context) {
	var req AddBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate bot level
	validLevels := map[string]bool{"easy": true, "medium": true, "hard": true}
	if req.BotLevel == "" {
		req.BotLevel = "medium" // Default level
	}
	if !validLevels[req.BotLevel] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot level. Must be easy, medium, or hard"})
		return
	}

	// Get game and add bot
	liveGame := h.deps.GameService.GetGame(req.RoomCode)
	if liveGame == nil {
		// Game not in memory, try to load from database
		db := database.GetDB()
		var dbGame models.Game
		if err := db.Where("room_code = ? AND status IN ?", req.RoomCode, []string{"waiting", "in_progress"}).
			Preload("Players.Player").
			First(&dbGame).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
			return
		}

		// Try to load the game into memory using the game manager
		gameManager := h.deps.GameService.(*game.Manager)
		gameState := gameManager.LoadGameFromDatabase(&dbGame)
		if gameState == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load game"})
			return
		}
		liveGame = gameState
	}

	// Check if game is in waiting state (can only add bots before game starts)
	if liveGame.Status != "waiting" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot add bots to a game that has already started"})
		return
	}

	// Check if game has space for more players
	if len(liveGame.Players) >= 6 { // Max 6 players in Dixit
		c.JSON(http.StatusBadRequest, gin.H{"error": "Game is full (maximum 6 players)"})
		return
	}

	// Add bot to game
	_, err := h.deps.GameService.AddBot(req.RoomCode, req.BotLevel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Bot added successfully",
		"bot_level": req.BotLevel,
		"room_code": req.RoomCode,
	})
}

// GetBotStats gets statistics about bot performance
// @Summary Get bot statistics
// @Description Get statistics about AI bot performance and usage
// @Tags bots
// @Accept json
// @Produce json
// @Success 200 {object} BotStatsResponse
// @Failure 500 {object} map[string]interface{}
// @Router /bots/stats [get]
func (h *GameHandlers) GetBotStats(c *gin.Context) {
	db := database.GetDB()

	var totalBots int64
	db.Model(&models.Player{}).Where("type = ?", "bot").Count(&totalBots)

	var activeBots int64
	db.Raw(`
		SELECT COUNT(DISTINCT p.id) 
		FROM players p 
		JOIN game_players gp ON p.id = gp.player_id 
		JOIN games g ON gp.game_id = g.id 
		WHERE p.type = 'bot' AND g.status = 'in_progress'
	`).Scan(&activeBots)

	// Bot performance by level
	var botsByLevel []struct {
		Level string `json:"level"`
		Count int64  `json:"count"`
	}
	db.Model(&models.Player{}).
		Select("bot_level as level, COUNT(*) as count").
		Where("type = ?", "bot").
		Group("bot_level").
		Scan(&botsByLevel)

	botLevelMap := make(map[string]int64)
	for _, bot := range botsByLevel {
		botLevelMap[bot.Level] = bot.Count
	}

	// Bot win rates (simplified calculation)
	var botPerformance []struct {
		Level   string  `json:"level"`
		WinRate float64 `json:"win_rate"`
	}
	db.Raw(`
		SELECT p.bot_level as level, 
		       COALESCE(AVG(CASE WHEN gh.winner_id = p.id THEN 1.0 ELSE 0.0 END), 0) * 100 as win_rate
		FROM players p 
		LEFT JOIN game_histories gh ON p.id = gh.winner_id 
		WHERE p.type = 'bot' AND p.bot_level IS NOT NULL
		GROUP BY p.bot_level
	`).Scan(&botPerformance)

	performanceMap := make(map[string]interface{})
	for _, perf := range botPerformance {
		performanceMap[perf.Level] = map[string]float64{
			"win_rate": perf.WinRate,
		}
	}

	stats := BotStatsResponse{
		TotalBots:      totalBots,
		ActiveBots:     activeBots,
		BotsByLevel:    botLevelMap,
		BotPerformance: performanceMap,
	}

	c.JSON(http.StatusOK, stats)
}

// RemovePlayerFromGame removes a player or bot from an existing game
// @Summary Remove player from game
// @Description Remove a player or bot from an existing game (only in waiting status)
// @Tags games
// @Accept json
// @Produce json
// @Param player body RemovePlayerRequest true "Player removal information"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/games/remove-player [delete]
func (h *GameHandlers) RemovePlayerFromGame(c *gin.Context) {
	var req RemovePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse player ID
	playerID, err := uuid.Parse(req.PlayerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID format"})
		return
	}

	// Get game and remove player
	liveGame := h.deps.GameService.GetGame(req.RoomCode)
	if liveGame == nil {
		// Game not in memory, try to load from database
		db := database.GetDB()
		var dbGame models.Game
		if err := db.Where("room_code = ? AND status IN ?", req.RoomCode, []string{"waiting", "in_progress"}).
			Preload("Players.Player").
			First(&dbGame).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
			return
		}

		// Try to load the game into memory using the game manager
		gameManager := h.deps.GameService.(*game.Manager)
		gameState := gameManager.LoadGameFromDatabase(&dbGame)
		if gameState == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load game"})
			return
		}
		liveGame = gameState
	}

	// Check if game is in waiting state (can only remove players before game starts)
	if liveGame.Status != "waiting" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove players from a game that has already started"})
		return
	}

	// Check if player exists in the game
	player, exists := liveGame.Players[playerID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found in game"})
		return
	}

	// Remove player using the game service
	gameManager := h.deps.GameService.(*game.Manager)
	_, err = gameManager.RemovePlayer(req.RoomCode, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Player removed successfully",
		"room_code":   req.RoomCode,
		"player_id":   req.PlayerID,
		"player_name": player.Name,
		"is_bot":      player.IsBot,
	})
}

// LeaveGame allows a player to leave a game
// @Summary Leave game
// @Description Allow a player to leave a game they are currently in
// @Tags games
// @Accept json
// @Produce json
// @Param leave body LeaveGameRequest true "Leave game information"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/games/leave [post]
func (h *GameHandlers) LeaveGame(c *gin.Context) {
	var req LeaveGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get player ID from context (set by auth middleware)
	playerIDInterface, exists := c.Get("player_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Player ID not found in context"})
		return
	}

	playerID, ok := playerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid player ID format"})
		return
	}

	// Use the game service to leave the game
	gameManager := h.deps.GameService.(*game.Manager)
	updatedGame, err := gameManager.LeaveGame(req.RoomCode, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Successfully left the game",
		"room_code":    req.RoomCode,
		"player_count": len(updatedGame.Players),
	})
}

// DeleteGame allows a lobby manager to delete an entire game
// @Summary Delete game
// @Description Delete an entire game (only allowed by lobby manager/creator)
// @Tags games
// @Accept json
// @Produce json
// @Param room_code path string true "Room code"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/games/{room_code} [delete]
func (h *GameHandlers) DeleteGame(c *gin.Context) {
	roomCode := c.Param("room_code")
	if roomCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Room code is required"})
		return
	}

	// Get player ID from context (set by auth middleware)
	playerIDInterface, exists := c.Get("player_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Player ID not found in context"})
		return
	}

	playerID, ok := playerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid player ID format"})
		return
	}

	// Use the game service to delete the game
	gameManager := h.deps.GameService.(*game.Manager)
	err := gameManager.DeleteGame(roomCode, playerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Game deleted successfully",
		"room_code": roomCode,
	})
}
