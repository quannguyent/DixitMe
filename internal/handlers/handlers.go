package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"dixitme/internal/database"
	"dixitme/internal/game"
	"dixitme/internal/models"
	"dixitme/internal/seeder"
	"dixitme/internal/storage"

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

// Card Management Handlers

// @Summary Upload card image
// @Description Upload an image for a card to MinIO storage
// @Tags cards
// @Accept multipart/form-data
// @Produce json
// @Param card_id path int true "Card ID"
// @Param image formData file true "Card image file"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /cards/{card_id}/image [post]
func UploadCardImage(c *gin.Context) {
	cardID, err := strconv.Atoi(c.Param("card_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID"})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}
	defer file.Close()

	// Upload to MinIO
	minioClient := storage.GetClient()
	if minioClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Storage not configured"})
		return
	}

	imageURL, err := minioClient.UploadCardImage(cardID, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	// Update card in database
	db := database.GetDB()
	var card models.Card
	if err := db.First(&card, cardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
		return
	}

	card.ImageURL = imageURL
	if err := db.Save(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update card"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"image_url": imageURL,
		"card_id":   cardID,
	})
}

// @Summary Create new card
// @Description Create a new card with optional tags
// @Tags cards
// @Accept json
// @Produce json
// @Param card body CreateCardRequest true "Card data"
// @Success 201 {object} models.Card
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /cards [post]
func CreateCard(c *gin.Context) {
	var req CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Create card
	card := models.Card{
		Title:       req.Title,
		Description: req.Description,
		Extension:   req.Extension,
		IsActive:    true,
	}

	if err := db.Create(&card).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create card"})
		return
	}

	// Add tags if provided
	if len(req.TagIDs) > 0 {
		for _, tagID := range req.TagIDs {
			cardTag := models.CardTag{
				CardID: card.ID,
				TagID:  tagID,
				Weight: 1.0,
			}
			db.Create(&cardTag)
		}
	}

	// Generate MinIO URL
	minioClient := storage.GetClient()
	if minioClient != nil {
		card.ImageURL = minioClient.GetCardImageURL(card.ID, card.Extension)
		db.Save(&card)
	}

	c.JSON(http.StatusCreated, card)
}

// @Summary Get card with tags
// @Description Get a card by ID with its associated tags
// @Tags cards
// @Produce json
// @Param card_id path int true "Card ID"
// @Success 200 {object} CardWithTagsResponse
// @Failure 404 {object} map[string]interface{}
// @Router /cards/{card_id} [get]
func GetCardWithTags(c *gin.Context) {
	cardID, err := strconv.Atoi(c.Param("card_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID"})
		return
	}

	db := database.GetDB()

	var card models.Card
	if err := db.Preload("Tags.Tag").First(&card, cardID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
		return
	}

	// Build response with tags
	tags := make([]TagResponse, 0, len(card.Tags))
	for _, cardTag := range card.Tags {
		tags = append(tags, TagResponse{
			ID:       cardTag.Tag.ID,
			Name:     cardTag.Tag.Name,
			Slug:     cardTag.Tag.Slug,
			Category: cardTag.Tag.Category,
			Color:    cardTag.Tag.Color,
			Weight:   cardTag.Weight,
		})
	}

	response := CardWithTagsResponse{
		ID:          card.ID,
		ImageURL:    card.ImageURL,
		Title:       card.Title,
		Description: card.Description,
		Extension:   card.Extension,
		IsActive:    card.IsActive,
		Tags:        tags,
		CreatedAt:   card.CreatedAt,
		UpdatedAt:   card.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary List cards with filtering
// @Description Get a list of cards with optional tag filtering
// @Tags cards
// @Produce json
// @Param tags query string false "Comma-separated tag IDs for filtering"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} CardsListResponse
// @Router /cards [get]
func ListCards(c *gin.Context) {
	db := database.GetDB()

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	tagsParam := c.Query("tags")

	offset := (page - 1) * limit

	query := db.Model(&models.Card{}).Where("is_active = ?", true)

	// Filter by tags if provided
	if tagsParam != "" {
		tagIDs := strings.Split(tagsParam, ",")
		query = query.Joins("JOIN card_tag_relations ON cards.id = card_tag_relations.card_id").
			Where("card_tag_relations.tag_id IN ?", tagIDs)
	}

	var total int64
	query.Count(&total)

	var cards []models.Card
	if err := query.Preload("Tags.Tag").Offset(offset).Limit(limit).Find(&cards).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cards"})
		return
	}

	// Build response
	cardResponses := make([]CardWithTagsResponse, 0, len(cards))
	for _, card := range cards {
		tags := make([]TagResponse, 0, len(card.Tags))
		for _, cardTag := range card.Tags {
			tags = append(tags, TagResponse{
				ID:       cardTag.Tag.ID,
				Name:     cardTag.Tag.Name,
				Slug:     cardTag.Tag.Slug,
				Category: cardTag.Tag.Category,
				Color:    cardTag.Tag.Color,
				Weight:   cardTag.Weight,
			})
		}

		cardResponses = append(cardResponses, CardWithTagsResponse{
			ID:          card.ID,
			ImageURL:    card.ImageURL,
			Title:       card.Title,
			Description: card.Description,
			Extension:   card.Extension,
			IsActive:    card.IsActive,
			Tags:        tags,
			CreatedAt:   card.CreatedAt,
			UpdatedAt:   card.UpdatedAt,
		})
	}

	response := CardsListResponse{
		Cards: cardResponses,
		Pagination: PaginationResponse{
			Page:  page,
			Limit: limit,
			Total: total,
			Pages: (total + int64(limit) - 1) / int64(limit),
		},
	}

	c.JSON(http.StatusOK, response)
}

// Tag Management Handlers

// @Summary Create new tag
// @Description Create a new tag for card categorization
// @Tags tags
// @Accept json
// @Produce json
// @Param tag body CreateTagRequest true "Tag data"
// @Success 201 {object} models.Tag
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /tags [post]
func CreateTag(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Generate slug from name
	slug := strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))

	tag := models.Tag{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Color:       req.Color,
		Weight:      req.Weight,
		Category:    req.Category,
		IsActive:    true,
	}

	if err := db.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag"})
		return
	}

	c.JSON(http.StatusCreated, tag)
}

// @Summary List all tags
// @Description Get all active tags grouped by category
// @Tags tags
// @Produce json
// @Success 200 {object} TagsListResponse
// @Router /tags [get]
func ListTags(c *gin.Context) {
	db := database.GetDB()

	var tags []models.Tag
	if err := db.Where("is_active = ?", true).Order("category, name").Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}

	// Group by category
	categories := make(map[string][]models.Tag)
	for _, tag := range tags {
		category := tag.Category
		if category == "" {
			category = "general"
		}
		categories[category] = append(categories[category], tag)
	}

	response := TagsListResponse{
		Tags:       tags,
		Categories: categories,
	}

	c.JSON(http.StatusOK, response)
}

// Bot Management Handlers

// @Summary Add bot to game
// @Description Add an AI bot player to an existing game
// @Tags bots
// @Accept json
// @Produce json
// @Param request body AddBotRequest true "Bot configuration"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /games/add-bot [post]
func AddBotToGame(c *gin.Context) {
	var req AddBotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate bot level
	validLevels := []string{"easy", "medium", "hard"}
	validLevel := false
	for _, level := range validLevels {
		if req.BotLevel == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bot level. Must be easy, medium, or hard"})
		return
	}

	// Add bot to game
	manager := game.GetManager()
	gameState, err := manager.AddBot(req.RoomCode, req.BotLevel)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Count bots in game
	botCount := 0
	for _, player := range gameState.Players {
		if player.IsBot {
			botCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"bot_added":     true,
		"bot_level":     req.BotLevel,
		"total_bots":    botCount,
		"total_players": len(gameState.Players),
		"room_code":     req.RoomCode,
	})
}

// @Summary Get bot statistics
// @Description Get performance statistics for bots
// @Tags bots
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /bots/stats [get]
func GetBotStats(c *gin.Context) {
	db := database.GetDB()

	// Get bot game statistics
	var stats struct {
		TotalBotGames int64   `json:"total_bot_games"`
		EasyBotWins   int64   `json:"easy_bot_wins"`
		MediumBotWins int64   `json:"medium_bot_wins"`
		HardBotWins   int64   `json:"hard_bot_wins"`
		AvgBotScore   float64 `json:"avg_bot_score"`
	}

	// Count games with bots
	db.Model(&models.GamePlayer{}).
		Joins("JOIN players ON game_players.player_id = players.id").
		Where("players.type = ?", models.PlayerTypeBot).
		Count(&stats.TotalBotGames)

	// Count wins by difficulty
	db.Raw(`
		SELECT COUNT(*) 
		FROM game_histories gh 
		JOIN players p ON gh.winner_id = p.id 
		WHERE p.type = ? AND p.bot_level = ?
	`, models.PlayerTypeBot, "easy").Scan(&stats.EasyBotWins)

	db.Raw(`
		SELECT COUNT(*) 
		FROM game_histories gh 
		JOIN players p ON gh.winner_id = p.id 
		WHERE p.type = ? AND p.bot_level = ?
	`, models.PlayerTypeBot, "medium").Scan(&stats.MediumBotWins)

	db.Raw(`
		SELECT COUNT(*) 
		FROM game_histories gh 
		JOIN players p ON gh.winner_id = p.id 
		WHERE p.type = ? AND p.bot_level = ?
	`, models.PlayerTypeBot, "hard").Scan(&stats.HardBotWins)

	// Calculate average bot score
	db.Raw(`
		SELECT AVG(gp.score) 
		FROM game_players gp 
		JOIN players p ON gp.player_id = p.id 
		WHERE p.type = ?
	`, models.PlayerTypeBot).Scan(&stats.AvgBotScore)

	c.JSON(http.StatusOK, stats)
}

// Chat Handlers

// @Summary Send chat message
// @Description Send a chat message in a game
// @Tags chat
// @Accept json
// @Produce json
// @Param request body SendChatRequest true "Chat message data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /chat/send [post]
func SendChatMessage(c *gin.Context) {
	var req SendChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	manager := game.GetManager()
	err := manager.SendChatMessage(req.RoomCode, req.PlayerID, req.Message, req.MessageType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Chat message sent",
	})
}

// @Summary Get chat history
// @Description Get chat messages for a game and phase
// @Tags chat
// @Produce json
// @Param room_code query string true "Room code"
// @Param phase query string false "Game phase (lobby, voting, all)" default(all)
// @Param limit query int false "Number of messages to return" default(50)
// @Success 200 {object} ChatHistoryResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /chat/history [get]
func GetChatHistory(c *gin.Context) {
	roomCode := c.Query("room_code")
	if roomCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_code is required"})
		return
	}

	phase := c.DefaultQuery("phase", "all")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	manager := game.GetManager()
	messages, err := manager.GetChatHistory(roomCode, phase, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response := ChatHistoryResponse{
		Messages: messages,
		Phase:    phase,
		Count:    len(messages),
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get chat statistics
// @Description Get chat statistics for a game
// @Tags chat
// @Produce json
// @Param room_code query string true "Room code"
// @Success 200 {object} ChatStatsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /chat/stats [get]
func GetChatStats(c *gin.Context) {
	roomCode := c.Query("room_code")
	if roomCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "room_code is required"})
		return
	}

	db := database.GetDB()

	// Get the game ID
	var game models.Game
	if err := db.Where("room_code = ?", roomCode).First(&game).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	var stats ChatStatsResponse

	// Total messages
	db.Model(&models.ChatMessage{}).Where("game_id = ? AND is_visible = ?", game.ID, true).Count(&stats.TotalMessages)

	// Messages by phase
	phaseStats := make(map[string]int64)
	phases := []string{"lobby", "storytelling", "submitting", "voting", "scoring"}
	for _, phase := range phases {
		var count int64
		db.Model(&models.ChatMessage{}).Where("game_id = ? AND phase = ? AND is_visible = ?", game.ID, phase, true).Count(&count)
		phaseStats[phase] = count
	}
	stats.MessagesByPhase = phaseStats

	// Messages by player
	playerStats := make(map[string]PlayerChatStats)
	var playerMessages []struct {
		PlayerID     string `json:"player_id"`
		PlayerName   string `json:"player_name"`
		MessageCount int64  `json:"message_count"`
	}

	db.Raw(`
		SELECT 
			p.id as player_id,
			p.name as player_name,
			COUNT(cm.id) as message_count
		FROM chat_messages cm
		JOIN players p ON cm.player_id = p.id
		WHERE cm.game_id = ? AND cm.is_visible = ? AND cm.message_type = 'chat'
		GROUP BY p.id, p.name
		ORDER BY message_count DESC
	`, game.ID, true).Scan(&playerMessages)

	for _, pm := range playerMessages {
		playerStats[pm.PlayerID] = PlayerChatStats{
			PlayerName:   pm.PlayerName,
			MessageCount: pm.MessageCount,
		}
	}
	stats.MessagesByPlayer = playerStats

	c.JSON(http.StatusOK, stats)
}

// Database Seeding Handlers

// @Summary Seed database with default data
// @Description Seeds the database with default tags and cards
// @Tags admin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/seed [post]
func SeedDatabase(c *gin.Context) {
	if err := seeder.SeedDatabase(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get counts
	db := database.GetDB()
	var tagCount, cardCount int64
	db.Model(&models.Tag{}).Count(&tagCount)
	db.Model(&models.Card{}).Count(&cardCount)

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Database seeded successfully",
		"tags_count":  tagCount,
		"cards_count": cardCount,
	})
}

// @Summary Seed only tags
// @Description Seeds the database with default tags only
// @Tags admin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/seed/tags [post]
func SeedTags(c *gin.Context) {
	if err := seeder.SeedTagsOnly(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get count
	db := database.GetDB()
	var tagCount int64
	db.Model(&models.Tag{}).Count(&tagCount)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Tags seeded successfully",
		"tags_count": tagCount,
	})
}

// @Summary Seed only cards
// @Description Seeds the database with default cards only (requires existing tags)
// @Tags admin
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/seed/cards [post]
func SeedCards(c *gin.Context) {
	if err := seeder.SeedCardsOnly(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get count
	db := database.GetDB()
	var cardCount int64
	db.Model(&models.Card{}).Count(&cardCount)

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "Cards seeded successfully",
		"cards_count": cardCount,
	})
}

// @Summary Get database statistics
// @Description Get counts of tags, cards, and card-tag relationships
// @Tags admin
// @Produce json
// @Success 200 {object} DatabaseStatsResponse
// @Router /admin/stats [get]
func GetDatabaseStats(c *gin.Context) {
	db := database.GetDB()

	var stats DatabaseStatsResponse
	db.Model(&models.Tag{}).Count(&stats.TagsCount)
	db.Model(&models.Card{}).Count(&stats.CardsCount)
	db.Model(&models.CardTag{}).Count(&stats.CardTagsCount)
	db.Model(&models.Player{}).Count(&stats.PlayersCount)
	db.Model(&models.Game{}).Count(&stats.GamesCount)

	// Get tags by category
	var tags []models.Tag
	db.Where("is_active = ?", true).Find(&tags)

	categoryStats := make(map[string]int64)
	for _, tag := range tags {
		categoryStats[tag.Category]++
	}
	stats.TagsByCategory = categoryStats

	// Get cards with most tags
	var cardTagCounts []struct {
		CardID   int    `json:"card_id"`
		Title    string `json:"title"`
		TagCount int64  `json:"tag_count"`
	}

	db.Raw(`
		SELECT c.id as card_id, c.title, COUNT(ct.tag_id) as tag_count
		FROM cards c
		LEFT JOIN card_tag_relations ct ON c.id = ct.card_id
		WHERE c.is_active = true
		GROUP BY c.id, c.title
		ORDER BY tag_count DESC
		LIMIT 10
	`).Scan(&cardTagCounts)

	// Convert to interface slice
	topCards := make([]interface{}, len(cardTagCounts))
	for i, card := range cardTagCounts {
		topCards[i] = card
	}
	stats.TopTaggedCards = topCards

	c.JSON(http.StatusOK, stats)
}

// @Summary Cleanup old games
// @Description Remove completed and abandoned games older than 24 hours
// @Tags admin
// @Produce json
// @Success 200 {object} map[string]interface{}
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

// Request/Response types

type CreateCardRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Extension   string `json:"extension"`
	TagIDs      []int  `json:"tag_ids"`
}

type CreateTagRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	Weight      float64 `json:"weight"`
	Category    string  `json:"category" binding:"required"`
}

type AddBotRequest struct {
	RoomCode string `json:"room_code" binding:"required"`
	BotLevel string `json:"bot_level" binding:"required"`
}

type TagResponse struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	Category string  `json:"category"`
	Color    string  `json:"color"`
	Weight   float64 `json:"weight"`
}

type CardWithTagsResponse struct {
	ID          int           `json:"id"`
	ImageURL    string        `json:"image_url"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Extension   string        `json:"extension"`
	IsActive    bool          `json:"is_active"`
	Tags        []TagResponse `json:"tags"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type CardsListResponse struct {
	Cards      []CardWithTagsResponse `json:"cards"`
	Pagination PaginationResponse     `json:"pagination"`
}

type TagsListResponse struct {
	Tags       []models.Tag            `json:"tags"`
	Categories map[string][]models.Tag `json:"categories"`
}

type PaginationResponse struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
	Pages int64 `json:"pages"`
}

type DatabaseStatsResponse struct {
	TagsCount      int64            `json:"tags_count"`
	CardsCount     int64            `json:"cards_count"`
	CardTagsCount  int64            `json:"card_tags_count"`
	PlayersCount   int64            `json:"players_count"`
	GamesCount     int64            `json:"games_count"`
	TagsByCategory map[string]int64 `json:"tags_by_category"`
	TopTaggedCards []interface{}    `json:"top_tagged_cards"`
}

// Chat request/response types
type SendChatRequest struct {
	RoomCode    string    `json:"room_code" binding:"required"`
	PlayerID    uuid.UUID `json:"player_id" binding:"required"`
	Message     string    `json:"message" binding:"required"`
	MessageType string    `json:"message_type,omitempty"` // chat, emote
}

type ChatHistoryResponse struct {
	Messages []game.ChatMessagePayload `json:"messages"`
	Phase    string                    `json:"phase"`
	Count    int                       `json:"count"`
}

type ChatStatsResponse struct {
	TotalMessages    int64                      `json:"total_messages"`
	MessagesByPhase  map[string]int64           `json:"messages_by_phase"`
	MessagesByPlayer map[string]PlayerChatStats `json:"messages_by_player"`
}

type PlayerChatStats struct {
	PlayerName   string `json:"player_name"`
	MessageCount int64  `json:"message_count"`
}
