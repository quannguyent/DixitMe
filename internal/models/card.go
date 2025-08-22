package models

import (
	"time"
)

// Card represents a game card with tags for categorization and bot AI
type Card struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	ImageURL    string    `json:"image_url" gorm:"not null"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Extension   string    `json:"extension" gorm:"default:'.jpg'"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Tags []CardTag `json:"tags" gorm:"many2many:card_tag_relations;"`
}

// Tag represents a categorization tag that can be applied to cards
type Tag struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"unique;not null;index"`
	Slug        string    `json:"slug" gorm:"unique;not null;index"`
	Description string    `json:"description"`
	Color       string    `json:"color" gorm:"default:'#3B82F6'"` // Hex color for UI
	Weight      float64   `json:"weight" gorm:"default:1.0"`      // For weighted random selection
	Category    string    `json:"category"`                       // Group tags by category
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Cards []Card `json:"cards" gorm:"many2many:card_tag_relations;"`
}

// CardTag represents the many-to-many relationship between cards and tags
type CardTag struct {
	CardID int     `json:"card_id" gorm:"primaryKey"`
	TagID  int     `json:"tag_id" gorm:"primaryKey"`
	Weight float64 `json:"weight" gorm:"default:1.0"` // How strongly this tag applies to this card

	// Relationships
	Card Card `json:"card" gorm:"foreignKey:CardID"`
	Tag  Tag  `json:"tag" gorm:"foreignKey:TagID"`
}
