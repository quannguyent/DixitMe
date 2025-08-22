package handlers

import (
	"net/http"
	"strings"

	"dixitme/internal/database"
	"dixitme/internal/models"

	"github.com/gin-gonic/gin"
)

// TagHandlers handles tag-related HTTP requests
type TagHandlers struct {
	deps *HandlerDependencies
}

// NewTagHandlers creates a new TagHandlers instance
func NewTagHandlers(deps *HandlerDependencies) *TagHandlers {
	return &TagHandlers{deps: deps}
}

// CreateTag creates a new tag for card categorization
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

	// Generate slug if not provided
	if req.Slug == "" {
		req.Slug = strings.ToLower(strings.ReplaceAll(req.Name, " ", "-"))
	}

	// Set default values
	if req.Color == "" {
		req.Color = "#3B82F6"
	}
	if req.Weight == 0 {
		req.Weight = 1.0
	}

	tag := models.Tag{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Color:       req.Color,
		Weight:      req.Weight,
		Category:    req.Category,
		IsActive:    true,
	}

	db := database.GetDB()
	if err := db.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag"})
		return
	}

	c.JSON(http.StatusCreated, tag)
}

// ListTags gets all available tags
// @Summary List all tags
// @Description Get a list of all available tags for card categorization
// @Tags tags
// @Produce json
// @Param category query string false "Filter by category"
// @Success 200 {object} ListTagsResponse
// @Failure 500 {object} map[string]interface{}
// @Router /tags [get]
func ListTags(c *gin.Context) {
	db := database.GetDB()

	query := db.Model(&models.Tag{}).Where("is_active = ?", true)

	// Filter by category if provided
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	var tags []models.Tag
	if err := query.Order("category, name").Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tags"})
		return
	}

	response := ListTagsResponse{Tags: tags}
	c.JSON(http.StatusOK, response)
}
