package auth

import (
	"errors"
	"time"

	"dixitme/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the claims in our JWT token
type JWTClaims struct {
	UserID    *uuid.UUID      `json:"user_id,omitempty"` // NULL for guests
	SessionID uuid.UUID       `json:"session_id"`
	AuthType  models.AuthType `json:"auth_type"`
	Name      string          `json:"name"`
	Email     string          `json:"email,omitempty"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	secretKey []byte
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
	}
}

// GenerateToken generates a JWT token for a user or guest session
func (j *JWTService) GenerateToken(session *models.Session, user *models.User, guestName string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // 24 hours expiry

	var userID *uuid.UUID
	var email, name string

	if user != nil {
		userID = &user.ID
		email = user.Email
		name = user.DisplayName
	} else {
		name = guestName
	}

	claims := JWTClaims{
		UserID:    userID,
		SessionID: session.ID,
		AuthType:  session.AuthType,
		Name:      name,
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "dixitme",
			Subject:   session.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check if token has expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// RefreshToken generates a new token with extended expiry
func (j *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Extend expiry by 24 hours
	now := time.Now()
	claims.ExpiresAt = jwt.NewNumericDate(now.Add(24 * time.Hour))
	claims.IssuedAt = jwt.NewNumericDate(now)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ExtractUserInfo extracts user information from token claims
func (j *JWTService) ExtractUserInfo(tokenString string) (*UserInfo, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
		AuthType:  claims.AuthType,
		Name:      claims.Name,
		Email:     claims.Email,
	}, nil
}

// UserInfo represents extracted user information from JWT
type UserInfo struct {
	UserID    *uuid.UUID      `json:"user_id,omitempty"`
	SessionID uuid.UUID       `json:"session_id"`
	AuthType  models.AuthType `json:"auth_type"`
	Name      string          `json:"name"`
	Email     string          `json:"email,omitempty"`
}
