package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthType represents the authentication method
type AuthType string

const (
	AuthTypeGuest    AuthType = "guest"
	AuthTypePassword AuthType = "password"
	AuthTypeGoogle   AuthType = "google"
)

// User represents a registered user account
type User struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email        string         `json:"email" gorm:"unique;index"`
	Username     string         `json:"username" gorm:"unique;index"`
	DisplayName  string         `json:"display_name" gorm:"not null"`
	PasswordHash string         `json:"-" gorm:"type:text"` // For password auth, hidden from JSON
	AuthType     AuthType       `json:"auth_type" gorm:"not null"`
	GoogleID     string         `json:"-" gorm:"index"` // For Google SSO, hidden from JSON
	Avatar       string         `json:"avatar"`         // Profile picture URL
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Sessions []Session `json:"-" gorm:"foreignKey:UserID"`
	Players  []Player  `json:"-" gorm:"foreignKey:UserID"`
}

// Session represents a user session (for both registered users and guests)
type Session struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"` // NULL for guest sessions
	Token     string     `json:"-" gorm:"unique;not null;index"`           // JWT token, hidden from JSON
	AuthType  AuthType   `json:"auth_type" gorm:"not null"`
	IPAddress string     `json:"ip_address"`
	UserAgent string     `json:"user_agent"`
	ExpiresAt time.Time  `json:"expires_at"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}
