package postgres

import (
	models "dixitme/internal/data/models"
	"dixitme/internal/game/domain"
	"gorm.io/gorm"
)

// BotRepository implements bot.DataProvider
type BotRepository struct {
	db *gorm.DB
}

// NewBotRepository creates a new BotRepository
func NewBotRepository(db *gorm.DB) *BotRepository {
	return &BotRepository{db: db}
}

// GetCardTags retrieves tags for a card
func (r *BotRepository) GetCardTags(cardID int) ([]domain.Tag, error) {
	var cardTags []models.CardTag
	err := r.db.Preload("Tag").Where("card_id = ?", cardID).Find(&cardTags).Error
	if err != nil {
		return nil, err
	}

	tags := make([]domain.Tag, 0, len(cardTags))
	for _, ct := range cardTags {
		tags = append(tags, domain.Tag{
			ID:       ct.Tag.ID,
			Name:     ct.Tag.Name,
			Slug:     ct.Tag.Slug,
			Category: ct.Tag.Category,
			Color:    ct.Tag.Color,
			Weight:   ct.Tag.Weight * ct.Weight, // Combine tag weight and relation weight
		})
	}

	return tags, nil
}

// GetCard retrieves card details
func (r *BotRepository) GetCard(cardID int) (*domain.Card, error) {
	var dbCard models.Card
	err := r.db.First(&dbCard, cardID).Error
	if err != nil {
		return nil, err
	}

	return &domain.Card{
		ID:          dbCard.ID,
		ImageURL:    dbCard.ImageURL,
		Title:       dbCard.Title,
		Description: dbCard.Description,
		Extension:   dbCard.Extension,
		IsActive:    dbCard.IsActive,
		// Tags not loaded here by default unless needed, logic above loads tags separately
	}, nil
}
