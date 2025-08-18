package auth

import (
	"net/http"
	"strings"

	"dixitme/internal/database"
	"dixitme/internal/logger"
	"dixitme/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	AuthContextKey  = "auth_user"
	TokenContextKey = "auth_token"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(jwtService *JWTService, required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)

		if token == "" {
			if required {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Authentication required",
					"code":  "AUTH_REQUIRED",
				})
				c.Abort()
				return
			}
			// Continue without auth for optional endpoints
			c.Next()
			return
		}

		// Validate token
		userInfo, err := jwtService.ExtractUserInfo(token)
		if err != nil {
			if required {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid token",
					"code":  "INVALID_TOKEN",
				})
				c.Abort()
				return
			}
			// Log but continue for optional auth
			logger.GetLogger().Warn("Invalid token in optional auth", "error", err)
			c.Next()
			return
		}

		// Verify session is still active
		db := database.GetDB()
		var session models.Session
		if err := db.Where("id = ? AND is_active = ? AND expires_at > NOW()",
			userInfo.SessionID, true).First(&session).Error; err != nil {
			if required {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Session expired or invalid",
					"code":  "SESSION_INVALID",
				})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// Add user info to context
		c.Set(AuthContextKey, userInfo)
		c.Set(TokenContextKey, token)

		c.Next()
	}
}

// RequireAuth creates middleware that requires authentication
func RequireAuth(jwtService *JWTService) gin.HandlerFunc {
	return AuthMiddleware(jwtService, true)
}

// OptionalAuth creates middleware with optional authentication
func OptionalAuth(jwtService *JWTService) gin.HandlerFunc {
	return AuthMiddleware(jwtService, false)
}

// RequireAdmin creates middleware that requires admin privileges
func RequireAdmin(jwtService *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First require auth
		RequireAuth(jwtService)(c)
		if c.IsAborted() {
			return
		}

		// Check if user has admin privileges
		userInfo, exists := GetUserFromContext(c)
		if !exists || userInfo.AuthType == models.AuthTypeGuest {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Admin access required",
				"code":  "ADMIN_REQUIRED",
			})
			c.Abort()
			return
		}

		// Additional admin check could be added here
		// For now, any registered user can access admin endpoints
		// In production, you'd want a proper admin role system

		c.Next()
	}
}

// GuestOrAuth allows both guest and authenticated access
func GuestOrAuth(jwtService *JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)

		if token == "" {
			// No token provided - continue as guest
			c.Next()
			return
		}

		// Token provided - validate it
		userInfo, err := jwtService.ExtractUserInfo(token)
		if err != nil {
			// Invalid token - reject
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Valid token - add to context
		c.Set(AuthContextKey, userInfo)
		c.Set(TokenContextKey, token)

		c.Next()
	}
}

// extractToken extracts JWT token from Authorization header or cookie
func extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Try cookie as fallback
	if cookie, err := c.Cookie("auth_token"); err == nil {
		return cookie
	}

	return ""
}

// GetUserFromContext extracts user info from gin context
func GetUserFromContext(c *gin.Context) (*UserInfo, bool) {
	userInfo, exists := c.Get(AuthContextKey)
	if !exists {
		return nil, false
	}

	user, ok := userInfo.(*UserInfo)
	return user, ok
}

// GetTokenFromContext extracts token from gin context
func GetTokenFromContext(c *gin.Context) (string, bool) {
	token, exists := c.Get(TokenContextKey)
	if !exists {
		return "", false
	}

	tokenStr, ok := token.(string)
	return tokenStr, ok
}

// IsAuthenticated checks if request is authenticated
func IsAuthenticated(c *gin.Context) bool {
	_, exists := GetUserFromContext(c)
	return exists
}

// IsGuest checks if request is from a guest (no authentication)
func IsGuest(c *gin.Context) bool {
	userInfo, exists := GetUserFromContext(c)
	if !exists {
		return true // No auth info = guest
	}
	return userInfo.AuthType == models.AuthTypeGuest
}

// GetPlayerID extracts player ID for game operations
func GetPlayerID(c *gin.Context) uuid.UUID {
	userInfo, exists := GetUserFromContext(c)
	if !exists {
		// For guests, generate or get player ID from query/body
		playerIDStr := c.Query("player_id")
		if playerIDStr == "" {
			// Try to get from request body if it's a POST/PUT
			var requestBody map[string]interface{}
			if c.ContentType() == "application/json" {
				c.ShouldBindJSON(&requestBody)
				if playerID, ok := requestBody["player_id"].(string); ok {
					playerIDStr = playerID
				}
			}
		}

		if playerIDStr != "" {
			if playerID, err := uuid.Parse(playerIDStr); err == nil {
				return playerID
			}
		}

		// Generate new UUID for new guest
		return uuid.New()
	}

	// For authenticated users, use session ID as player ID
	// This ensures consistency across sessions
	return userInfo.SessionID
}
