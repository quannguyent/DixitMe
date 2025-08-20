package handlers

import (
	"net/http"
	"strconv"

	"dixitme/internal/database"
	"dixitme/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreatePlayer creates a new player
// @Summary Create a new player
// @Description Create a new player with a given name
// @Tags players
// @Accept json
// @Produce json
// @Param player body CreatePlayerRequest true "Player information"
// @Success 201 {object} CreatePlayerResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /players [post]
func CreatePlayer(c *gin.Context) {
	var req CreatePlayerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	player := &models.Player{
		ID:   uuid.New(),
		Name: req.Name,
	}

	db := database.GetDB()
	if err := db.Create(player).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create player"})
		return
	}

	c.JSON(http.StatusCreated, CreatePlayerResponse{Player: player})
}

// GetPlayer gets a player by ID
// @Summary Get player by ID
// @Description Get player information by player ID
// @Tags players
// @Accept json
// @Produce json
// @Param id path string true "Player ID" format(uuid)
// @Success 200 {object} map[string]models.Player
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /players/{id} [get]
func GetPlayer(c *gin.Context) {
	playerIDStr := c.Param("id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}

	var player models.Player
	db := database.GetDB()
	if err := db.First(&player, "id = ?", playerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"player": player})
}

// GetPlayerStats gets statistics for a specific player
// @Summary Get player statistics
// @Description Get detailed statistics for a player including games played, win rate, etc.
// @Tags players
// @Accept json
// @Produce json
// @Param player_id path string true "Player ID" format(uuid)
// @Success 200 {object} PlayerStatsResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /player/{player_id}/stats [get]
func GetPlayerStats(c *gin.Context) {
	playerIDStr := c.Param("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}

	db := database.GetDB()

	// Check if player exists
	var player models.Player
	if err := db.First(&player, "id = ?", playerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}

	// Get total games played
	var totalGames int64
	db.Model(&models.GamePlayer{}).Where("player_id = ?", playerID).Count(&totalGames)

	// Get games won (assume winner is player with highest score)
	var gamesWon int64
	db.Raw(`
		SELECT COUNT(*) FROM game_histories gh
		WHERE gh.winner_id = ?
	`, playerID).Scan(&gamesWon)

	// Calculate win rate
	var winRate float64
	if totalGames > 0 {
		winRate = float64(gamesWon) / float64(totalGames) * 100
	}

	// Get average and total score
	var avgScore, totalScore float64
	db.Model(&models.GamePlayer{}).
		Where("player_id = ?", playerID).
		Select("AVG(score) as avg_score, SUM(score) as total_score").
		Row().Scan(&avgScore, &totalScore)

	// Get games as storyteller (simplified - would need more complex query for actual implementation)
	var gamesAsStoryteller int64
	db.Model(&models.GameRound{}).Where("storyteller_id = ?", playerID).Count(&gamesAsStoryteller)

	stats := PlayerStatsResponse{
		PlayerID:           playerID.String(),
		TotalGames:         totalGames,
		GamesWon:           gamesWon,
		WinRate:            winRate,
		AverageScore:       avgScore,
		TotalScore:         int64(totalScore),
		FavoriteRole:       "storyteller", // Simplified
		GamesAsStoryteller: gamesAsStoryteller,
	}

	c.JSON(http.StatusOK, stats)
}

// GetGameHistory gets the game history for a specific player
// @Summary Get player's game history
// @Description Get a list of games played by a specific player
// @Tags players
// @Accept json
// @Produce json
// @Param player_id path string true "Player ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of games per page" default(10)
// @Success 200 {object} GameHistoryResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /player/{player_id}/history [get]
func GetGameHistory(c *gin.Context) {
	playerIDStr := c.Param("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	db := database.GetDB()

	// Check if player exists
	var player models.Player
	if err := db.First(&player, "id = ?", playerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}

	// Get total count
	var total int64
	db.Model(&models.GameHistory{}).
		Joins("JOIN game_players gp ON gp.game_id = game_histories.game_id").
		Where("gp.player_id = ?", playerID).
		Count(&total)

	// Get game history with pagination
	var gameHistories []models.GameHistory
	err = db.Preload("Game").Preload("Winner").
		Joins("JOIN game_players gp ON gp.game_id = game_histories.game_id").
		Where("gp.player_id = ?", playerID).
		Order("game_histories.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&gameHistories).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch game history"})
		return
	}

	response := GameHistoryResponse{
		Games: gameHistories,
		Total: total,
		Page:  page,
		Limit: limit,
	}

	c.JSON(http.StatusOK, response)
}
