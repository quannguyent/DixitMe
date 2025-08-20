package auth

import (
	"testing"
	"time"

	"dixitme/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateToken(t *testing.T) {
	jwtService := NewJWTService("test-secret-key")

	tests := []struct {
		name        string
		user        *models.User
		session     *models.Session
		guestName   string
		expectError bool
	}{
		{
			name: "Generate token for registered user",
			user: &models.User{
				ID:          uuid.New(),
				Email:       "test@example.com",
				DisplayName: "Test User",
			},
			session: &models.Session{
				ID:       uuid.New(),
				AuthType: models.AuthTypePassword,
			},
			expectError: false,
		},
		{
			name: "Generate token for guest user",
			user: nil,
			session: &models.Session{
				ID:       uuid.New(),
				AuthType: models.AuthTypeGuest,
			},
			guestName:   "Guest Player",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtService.GenerateToken(tt.session, tt.user, tt.guestName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Validate the generated token
				claims, err := jwtService.ValidateToken(token)
				require.NoError(t, err)
				assert.Equal(t, tt.session.ID, claims.SessionID)
				assert.Equal(t, tt.session.AuthType, claims.AuthType)

				if tt.user != nil {
					assert.Equal(t, &tt.user.ID, claims.UserID)
					assert.Equal(t, tt.user.Email, claims.Email)
					assert.Equal(t, tt.user.DisplayName, claims.Name)
				} else {
					assert.Nil(t, claims.UserID)
					assert.Equal(t, tt.guestName, claims.Name)
				}
			}
		})
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	jwtService := NewJWTService("test-secret-key")

	// Generate a valid token
	user := &models.User{
		ID:          uuid.New(),
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
	session := &models.Session{
		ID:       uuid.New(),
		AuthType: models.AuthTypePassword,
	}

	validToken, err := jwtService.GenerateToken(session, user, "")
	require.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "Valid token",
			token:       validToken,
			expectError: false,
		},
		{
			name:        "Invalid token",
			token:       "invalid.jwt.token",
			expectError: true,
		},
		{
			name:        "Empty token",
			token:       "",
			expectError: true,
		},
		{
			name:        "Malformed token",
			token:       "not.a.jwt",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtService.ValidateToken(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, session.ID, claims.SessionID)
				assert.Equal(t, user.ID, *claims.UserID)
			}
		})
	}
}

func TestJWTService_ExtractUserInfo(t *testing.T) {
	jwtService := NewJWTService("test-secret-key")

	// Create test data
	userID := uuid.New()
	sessionID := uuid.New()
	user := &models.User{
		ID:          userID,
		Email:       "test@example.com",
		DisplayName: "Test User",
	}
	session := &models.Session{
		ID:       sessionID,
		AuthType: models.AuthTypePassword,
	}

	token, err := jwtService.GenerateToken(session, user, "")
	require.NoError(t, err)

	userInfo, err := jwtService.ExtractUserInfo(token)
	assert.NoError(t, err)
	assert.NotNil(t, userInfo)
	assert.Equal(t, userID, *userInfo.UserID)
	assert.Equal(t, sessionID, userInfo.SessionID)
	assert.Equal(t, "test@example.com", userInfo.Email)
	assert.Equal(t, "Test User", userInfo.Name)
	assert.Equal(t, models.AuthTypePassword, userInfo.AuthType)
}

func TestJWTService_TokenExpiration(t *testing.T) {
	jwtService := NewJWTService("test-secret-key")

	session := &models.Session{
		ID:       uuid.New(),
		AuthType: models.AuthTypeGuest,
	}

	// Generate token
	token, err := jwtService.GenerateToken(session, nil, "Guest")
	require.NoError(t, err)

	// Validate immediately (should work)
	claims, err := jwtService.ValidateToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)

	// Check expiration time (should be 24 hours from now)
	expectedExpiry := time.Now().Add(24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time
	assert.WithinDuration(t, expectedExpiry, actualExpiry, 5*time.Second)
}

func TestJWTService_DifferentSecrets(t *testing.T) {
	jwtService1 := NewJWTService("secret1")
	jwtService2 := NewJWTService("secret2")

	session := &models.Session{
		ID:       uuid.New(),
		AuthType: models.AuthTypeGuest,
	}

	// Generate token with first service
	token, err := jwtService1.GenerateToken(session, nil, "Guest")
	require.NoError(t, err)

	// Try to validate with second service (different secret)
	_, err = jwtService2.ValidateToken(token)
	assert.Error(t, err, "Token generated with one secret should not be valid with another secret")

	// Validate with correct service
	_, err = jwtService1.ValidateToken(token)
	assert.NoError(t, err, "Token should be valid with the correct secret")
}
