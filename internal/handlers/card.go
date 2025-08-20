package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"dixitme/internal/database"
	"dixitme/internal/models"
	"dixitme/internal/storage"

	"github.com/gin-gonic/gin"
)

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

// UploadCardImage uploads an image for a card to MinIO storage
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

// CreateCard creates a new card with optional tags
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

// GetCardWithTags gets a card by ID with its associated tags
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

// ListCards gets a list of cards with optional tag filtering
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
