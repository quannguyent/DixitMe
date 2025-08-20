package websocket

import (
	"dixitme/internal/services/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// extractTokenFromWebSocket extracts JWT token from WebSocket connection
func extractTokenFromWebSocket(c *gin.Context) string {
	// Try query parameter first (for WebSocket connections)
	if token := c.Query("token"); token != "" {
		return token
	}

	// Try Authorization header
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			return authHeader[7:]
		}
	}

	// Try cookie
	if cookie, err := c.Cookie("auth_token"); err == nil {
		return cookie
	}

	return ""
}

// extractPlayerInfo extracts player information from authentication context
func extractPlayerInfo(c *gin.Context, jwtService *auth.JWTService) (uuid.UUID, *auth.UserInfo, error) {
	var playerID uuid.UUID
	var userInfo *auth.UserInfo

	// Try to extract authentication info
	token := extractTokenFromWebSocket(c)
	if token != "" {
		if info, err := jwtService.ExtractUserInfo(token); err == nil {
			userInfo = info
			playerID = info.SessionID // Use session ID as player ID for consistency
			return playerID, userInfo, nil
		}
	}

	// If no valid auth, fall back to legacy behavior
	playerIDStr := c.Query("player_id")
	if playerIDStr != "" {
		parsedID, err := uuid.Parse(playerIDStr)
		if err != nil {
			return uuid.Nil, nil, err
		}
		playerID = parsedID
	} else {
		playerID = uuid.New()
	}

	return playerID, nil, nil
}
