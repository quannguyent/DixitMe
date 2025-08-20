package auth

import (
	"testing"

	"dixitme/internal/models"
	"dixitme/internal/testutils"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_RegisterWithPassword(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Mock the database
	cleanup := testutils.MockDatabase(db)
	defer cleanup()

	jwtService := NewJWTService("test-secret")
	authService := NewAuthService(jwtService)

	tests := []struct {
		name         string
		email        string
		username     string
		displayName  string
		password     string
		expectError  bool
		errorMessage string
		setupFunc    func()
	}{
		{
			name:        "Valid registration",
			email:       "test@example.com",
			username:    "testuser",
			displayName: "Test User",
			password:    "password123",
			expectError: false,
		},
		{
			name:         "Empty email",
			email:        "",
			username:     "testuser",
			displayName:  "Test User",
			password:     "password123",
			expectError:  true,
			errorMessage: "all fields are required",
		},
		{
			name:         "Short password",
			email:        "test@example.com",
			username:     "testuser",
			displayName:  "Test User",
			password:     "short",
			expectError:  true,
			errorMessage: "password must be at least 8 characters long",
		},
		{
			name:        "Duplicate email",
			email:       "duplicate@example.com",
			username:    "newuser",
			displayName: "New User",
			password:    "password123",
			expectError: true,
			setupFunc: func() {
				// Create existing user
				existingUser := &models.User{
					ID:       uuid.New(),
					Email:    "duplicate@example.com",
					Username: "existinguser",
					AuthType: models.AuthTypePassword,
				}
				db.Create(existingUser)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			user, err := authService.RegisterWithPassword(tt.email, tt.username, tt.displayName, tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
				assert.Equal(t, tt.username, user.Username)
				assert.Equal(t, tt.displayName, user.DisplayName)
				assert.Equal(t, models.AuthTypePassword, user.AuthType)
				assert.True(t, user.IsActive)
				assert.NotEmpty(t, user.PasswordHash)

				// Verify password was hashed
				err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(tt.password))
				assert.NoError(t, err, "Password should be properly hashed")

				// Verify user was saved to database
				var dbUser models.User
				err = db.Where("email = ?", tt.email).First(&dbUser).Error
				assert.NoError(t, err)
				assert.Equal(t, user.ID, dbUser.ID)
			}
		})
	}
}

func TestAuthService_LoginWithPassword(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	cleanup := testutils.MockDatabase(db)
	defer cleanup()

	jwtService := NewJWTService("test-secret")
	authService := NewAuthService(jwtService)

	// Create a test user
	password := "testpassword123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	testUser := &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		Username:     "testuser",
		DisplayName:  "Test User",
		PasswordHash: string(hashedPassword),
		AuthType:     models.AuthTypePassword,
		IsActive:     true,
	}
	db.Create(testUser)

	tests := []struct {
		name            string
		emailOrUsername string
		password        string
		expectError     bool
		errorMessage    string
	}{
		{
			name:            "Valid login with email",
			emailOrUsername: "test@example.com",
			password:        password,
			expectError:     false,
		},
		{
			name:            "Valid login with username",
			emailOrUsername: "testuser",
			password:        password,
			expectError:     false,
		},
		{
			name:            "Invalid password",
			emailOrUsername: "test@example.com",
			password:        "wrongpassword",
			expectError:     true,
			errorMessage:    "invalid credentials",
		},
		{
			name:            "Non-existent user",
			emailOrUsername: "nonexistent@example.com",
			password:        password,
			expectError:     true,
			errorMessage:    "invalid credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, session, token, err := authService.LoginWithPassword(tt.emailOrUsername, tt.password, "127.0.0.1", "test-agent")

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
				assert.Nil(t, session)
				assert.Empty(t, token)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotNil(t, session)
				assert.NotEmpty(t, token)

				assert.Equal(t, testUser.ID, user.ID)
				assert.Equal(t, testUser.Email, user.Email)
				assert.Equal(t, models.AuthTypePassword, session.AuthType)
				assert.True(t, session.IsActive)

				// Verify session was created in database
				var dbSession models.Session
				err = db.Where("id = ?", session.ID).First(&dbSession).Error
				assert.NoError(t, err)

				// Verify JWT token is valid
				userInfo, err := jwtService.ExtractUserInfo(token)
				assert.NoError(t, err)
				assert.Equal(t, user.ID, *userInfo.UserID)
				assert.Equal(t, session.ID, userInfo.SessionID)
			}
		})
	}
}

func TestAuthService_CreateGuestSession(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	cleanup := testutils.MockDatabase(db)
	defer cleanup()

	jwtService := NewJWTService("test-secret")
	authService := NewAuthService(jwtService)

	tests := []struct {
		name      string
		guestName string
	}{
		{
			name:      "Guest with name",
			guestName: "Guest Player",
		},
		{
			name:      "Guest with empty name",
			guestName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, token, err := authService.CreateGuestSession(tt.guestName, "127.0.0.1", "test-agent")

			assert.NoError(t, err)
			assert.NotNil(t, session)
			assert.NotEmpty(t, token)

			assert.Equal(t, models.AuthTypeGuest, session.AuthType)
			assert.True(t, session.IsActive)
			assert.Nil(t, session.UserID)

			// Verify session was created in database
			var dbSession models.Session
			err = db.Where("id = ?", session.ID).First(&dbSession).Error
			assert.NoError(t, err)

			// Verify JWT token is valid
			userInfo, err := jwtService.ExtractUserInfo(token)
			assert.NoError(t, err)
			assert.Nil(t, userInfo.UserID)
			assert.Equal(t, session.ID, userInfo.SessionID)
			assert.Equal(t, models.AuthTypeGuest, userInfo.AuthType)

			expectedName := tt.guestName
			if expectedName == "" {
				expectedName = "Guest"
			}
			assert.Equal(t, expectedName, userInfo.Name)
		})
	}
}

func TestAuthService_ValidateSession(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	cleanup := testutils.MockDatabase(db)
	defer cleanup()

	jwtService := NewJWTService("test-secret")
	authService := NewAuthService(jwtService)

	// Create test session
	session := testutils.CreateTestSession(t, db, nil)

	tests := []struct {
		name        string
		sessionID   uuid.UUID
		expectValid bool
	}{
		{
			name:        "Valid session",
			sessionID:   session.ID,
			expectValid: true,
		},
		{
			name:        "Non-existent session",
			sessionID:   uuid.New(),
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := authService.ValidateSession(tt.sessionID)
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	db := testutils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	cleanup := testutils.MockDatabase(db)
	defer cleanup()

	jwtService := NewJWTService("test-secret")
	authService := NewAuthService(jwtService)

	// Create test session
	session := testutils.CreateTestSession(t, db, nil)
	require.True(t, session.IsActive)

	// Logout
	err := authService.Logout(session.ID)
	assert.NoError(t, err)

	// Verify session is deactivated
	var dbSession models.Session
	err = db.Where("id = ?", session.ID).First(&dbSession).Error
	assert.NoError(t, err)
	assert.False(t, dbSession.IsActive)
}
