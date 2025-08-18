package auth

import (
	"fmt"
	"time"

	"dixitme/internal/database"
	"dixitme/internal/logger"
	"dixitme/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/oauth2/v2"
	"gorm.io/gorm"
)

// AuthService handles authentication operations
type AuthService struct {
	jwtService *JWTService
	db         *gorm.DB
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtService *JWTService) *AuthService {
	return &AuthService{
		jwtService: jwtService,
		db:         database.GetDB(),
	}
}

// RegisterWithPassword registers a new user with email/password
func (a *AuthService) RegisterWithPassword(email, username, displayName, password string) (*models.User, error) {
	// Validate input
	if email == "" || username == "" || displayName == "" || password == "" {
		return nil, fmt.Errorf("all fields are required")
	}

	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters long")
	}

	// Check if user already exists
	var existingUser models.User
	if err := a.db.Where("email = ? OR username = ?", email, username).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("user with this email or username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := models.User{
		ID:           uuid.New(),
		Email:        email,
		Username:     username,
		DisplayName:  displayName,
		PasswordHash: string(hashedPassword),
		AuthType:     models.AuthTypePassword,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := a.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.GetLogger().Info("User registered with password", "user_id", user.ID, "email", email, "username", username)
	return &user, nil
}

// LoginWithPassword authenticates user with email/password
func (a *AuthService) LoginWithPassword(emailOrUsername, password string, ipAddress, userAgent string) (*models.User, *models.Session, string, error) {
	// Find user
	var user models.User
	if err := a.db.Where("(email = ? OR username = ?) AND auth_type = ? AND is_active = ?",
		emailOrUsername, emailOrUsername, models.AuthTypePassword, true).First(&user).Error; err != nil {
		return nil, nil, "", fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, "", fmt.Errorf("invalid credentials")
	}

	// Create session
	session, token, err := a.createSession(&user, models.AuthTypePassword, ipAddress, userAgent)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	user.LastLoginAt = &session.CreatedAt
	a.db.Save(&user)

	logger.GetLogger().Info("User logged in with password", "user_id", user.ID, "session_id", session.ID)
	return &user, session, token, nil
}

// LoginWithGoogle authenticates user with Google OAuth token
func (a *AuthService) LoginWithGoogle(googleAccessToken string, ipAddress, userAgent string) (*models.User, *models.Session, string, error) {
	// For now, we'll simulate user info - in production you'd verify the token properly
	// This is a simplified implementation for demo purposes
	// You would use the googleAccessToken to verify with Google's API
	userInfo := &oauth2.Userinfo{
		Id:      "google_user_" + googleAccessToken[:8], // Simulate user ID from token
		Email:   "user@example.com",
		Name:    "Google User",
		Picture: "https://example.com/avatar.jpg",
	}

	// Check if user exists
	var user models.User
	err := a.db.Where("google_id = ? OR (email = ? AND auth_type = ?)",
		userInfo.Id, userInfo.Email, models.AuthTypeGoogle).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// Create new user
		user = models.User{
			ID:          uuid.New(),
			Email:       userInfo.Email,
			Username:    generateUsernameFromEmail(userInfo.Email),
			DisplayName: userInfo.Name,
			AuthType:    models.AuthTypeGoogle,
			GoogleID:    userInfo.Id,
			Avatar:      userInfo.Picture,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := a.db.Create(&user).Error; err != nil {
			return nil, nil, "", fmt.Errorf("failed to create user: %w", err)
		}

		logger.GetLogger().Info("New user registered with Google", "user_id", user.ID, "email", user.Email)
	} else if err != nil {
		return nil, nil, "", fmt.Errorf("database error: %w", err)
	} else {
		// Update existing user info
		user.DisplayName = userInfo.Name
		user.Avatar = userInfo.Picture
		user.UpdatedAt = time.Now()
		a.db.Save(&user)
	}

	// Create session
	session, token, err := a.createSession(&user, models.AuthTypeGoogle, ipAddress, userAgent)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	user.LastLoginAt = &session.CreatedAt
	a.db.Save(&user)

	logger.GetLogger().Info("User logged in with Google", "user_id", user.ID, "session_id", session.ID)
	return &user, session, token, nil
}

// CreateGuestSession creates a session for guest users
func (a *AuthService) CreateGuestSession(guestName, ipAddress, userAgent string) (*models.Session, string, error) {
	if guestName == "" {
		guestName = "Guest " + uuid.New().String()[:8]
	}

	session, token, err := a.createSession(nil, models.AuthTypeGuest, ipAddress, userAgent)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create guest session: %w", err)
	}

	logger.GetLogger().Info("Guest session created", "session_id", session.ID, "guest_name", guestName)
	return session, token, nil
}

// RefreshSession refreshes an existing session
func (a *AuthService) RefreshSession(sessionID uuid.UUID) (*models.Session, string, error) {
	var session models.Session
	if err := a.db.Preload("User").Where("id = ? AND is_active = ? AND expires_at > NOW()",
		sessionID, true).First(&session).Error; err != nil {
		return nil, "", fmt.Errorf("session not found or expired")
	}

	// Extend session
	session.ExpiresAt = time.Now().Add(24 * time.Hour)
	session.UpdatedAt = time.Now()
	a.db.Save(&session)

	// Generate new token
	var guestName string
	if session.User == nil {
		guestName = "Guest " + session.ID.String()[:8]
	}

	token, err := a.jwtService.GenerateToken(&session, session.User, guestName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return &session, token, nil
}

// Logout invalidates a session
func (a *AuthService) Logout(sessionID uuid.UUID) error {
	result := a.db.Model(&models.Session{}).Where("id = ?", sessionID).Update("is_active", false)
	if result.Error != nil {
		return fmt.Errorf("failed to logout: %w", result.Error)
	}

	logger.GetLogger().Info("User logged out", "session_id", sessionID)
	return nil
}

// GetUserByID retrieves a user by ID
func (a *AuthService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := a.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

// GetSessionByID retrieves a session by ID
func (a *AuthService) GetSessionByID(sessionID uuid.UUID) (*models.Session, error) {
	var session models.Session
	if err := a.db.Preload("User").Where("id = ? AND is_active = ?", sessionID, true).First(&session).Error; err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}
	return &session, nil
}

// CleanupExpiredSessions removes expired sessions
func (a *AuthService) CleanupExpiredSessions() error {
	result := a.db.Where("expires_at < NOW()").Delete(&models.Session{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		logger.GetLogger().Info("Cleaned up expired sessions", "count", result.RowsAffected)
	}

	return nil
}

// Helper functions

func (a *AuthService) createSession(user *models.User, authType models.AuthType, ipAddress, userAgent string) (*models.Session, string, error) {
	session := models.Session{
		ID:        uuid.New(),
		AuthType:  authType,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hours
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if user != nil {
		session.UserID = &user.ID
	}

	if err := a.db.Create(&session).Error; err != nil {
		return nil, "", err
	}

	// Generate JWT token
	var guestName string
	if user == nil {
		guestName = "Guest " + session.ID.String()[:8]
	}

	token, err := a.jwtService.GenerateToken(&session, user, guestName)
	if err != nil {
		return nil, "", err
	}

	// Update session with token
	session.Token = token
	a.db.Save(&session)

	return &session, token, nil
}

func generateUsernameFromEmail(email string) string {
	// Simple username generation from email
	at := 0
	for i, char := range email {
		if char == '@' {
			at = i
			break
		}
	}
	if at > 0 {
		return email[:at] + "_" + uuid.New().String()[:4]
	}
	return "user_" + uuid.New().String()[:8]
}
