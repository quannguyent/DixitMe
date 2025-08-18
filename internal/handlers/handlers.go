package handlers

import (
	"net/http"
	"strconv"

	"dixitme/internal/database"
	"dixitme/internal/game"
	"dixitme/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreatePlayerRequest represents the request to create a new player
type CreatePlayerRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreatePlayerResponse represents the response for player creation
type CreatePlayerResponse struct {
	Player *models.Player `json:"player"`
}

// GetGameRequest represents the request to get game info
type GetGameRequest struct {
	RoomCode string `uri:"room_code" binding:"required"`
}

// GetGameResponse represents the response for getting game info
type GetGameResponse struct {
	Game   *models.Game `json:"game"`
	IsLive bool         `json:"is_live"`
}

// GetGamesResponse represents the response for getting all games
type GetGamesResponse struct {
	Games []models.Game `json:"games"`
}

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

// GetGame gets game information by room code
// @Summary Get game by room code
// @Description Get game information and status by room code
// @Tags games
// @Accept json
// @Produce json
// @Param room_code path string true "Room Code"
// @Success 200 {object} GetGameResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /games/{room_code} [get]
func GetGame(c *gin.Context) {
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

	// Check if game is live (in memory)
	manager := game.GetManager()
	liveGame := manager.GetGame(req.RoomCode)
	isLive := liveGame != nil

	c.JSON(http.StatusOK, GetGameResponse{
		Game:   &dbGame,
		IsLive: isLive,
	})
}

// GetGames gets all games with pagination
// @Summary List all games
// @Description Get a paginated list of all games with optional status filter
// @Tags games
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Game status filter" Enums(waiting,in_progress,completed,abandoned)
// @Success 200 {object} GetGamesResponse
// @Failure 500 {object} map[string]string
// @Router /games [get]
func GetGames(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	db := database.GetDB()
	offset := (page - 1) * limit

	query := db.Preload("Players.Player")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var games []models.Game
	if err := query.Limit(limit).Offset(offset).Find(&games).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch games"})
		return
	}

	c.JSON(http.StatusOK, GetGamesResponse{Games: games})
}

// GetGameHistory gets game history for a player
// @Summary Get player game history
// @Description Get paginated game history for a specific player
// @Tags players
// @Accept json
// @Produce json
// @Param player_id path string true "Player ID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string][]models.GameHistory
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /player/{player_id}/history [get]
func GetGameHistory(c *gin.Context) {
	playerIDStr := c.Param("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	db := database.GetDB()
	var gameHistories []models.GameHistory
	if err := db.Preload("Game").Preload("Winner").
		Joins("JOIN game_players ON game_histories.game_id = game_players.game_id").
		Where("game_players.player_id = ?", playerID).
		Limit(limit).Offset(offset).
		Find(&gameHistories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch game history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"game_histories": gameHistories})
}

// GetPlayerStats gets statistics for a player
// @Summary Get player statistics
// @Description Get comprehensive statistics for a specific player
// @Tags players
// @Accept json
// @Produce json
// @Param player_id path string true "Player ID" format(uuid)
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /player/{player_id}/stats [get]
func GetPlayerStats(c *gin.Context) {
	playerIDStr := c.Param("player_id")
	playerID, err := uuid.Parse(playerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid player ID"})
		return
	}

	db := database.GetDB()

	// Get total games played
	var totalGames int64
	db.Model(&models.GamePlayer{}).Where("player_id = ?", playerID).Count(&totalGames)

	// Get games won
	var gamesWon int64
	db.Model(&models.GameHistory{}).Where("winner_id = ?", playerID).Count(&gamesWon)

	// Get average score
	var avgScore float64
	db.Model(&models.GamePlayer{}).
		Where("player_id = ?", playerID).
		Select("AVG(score)").
		Scan(&avgScore)

	// Get total score
	var totalScore int64
	db.Model(&models.GamePlayer{}).
		Where("player_id = ?", playerID).
		Select("SUM(score)").
		Scan(&totalScore)

	stats := gin.H{
		"total_games": totalGames,
		"games_won":   gamesWon,
		"win_rate":    float64(gamesWon) / float64(totalGames) * 100,
		"avg_score":   avgScore,
		"total_score": totalScore,
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// HealthCheck endpoint for monitoring
// @Summary Health check
// @Description Check the health status of the API and its dependencies
// @Tags system
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

// GetCards returns the list of available cards (placeholder implementation)
// @Summary Get available cards
// @Description Get the list of all available cards in the game
// @Tags cards
// @Accept json
// @Produce json
// @Success 200 {object} map[string][]map[string]interface{}
// @Router /cards [get]
func GetCards(c *gin.Context) {
	// In a real implementation, you'd load this from a file or database
	// For now, we'll return a simple structure representing cards
	cards := make([]gin.H, 84) // Dixit typically has 84 cards per expansion
	for i := 0; i < 84; i++ {
		cards[i] = gin.H{
			"id":  i + 1,
			"url": "/cards/" + strconv.Itoa(i+1) + ".jpg",
		}
	}

	c.JSON(http.StatusOK, gin.H{"cards": cards})
}

// CORS middleware
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
