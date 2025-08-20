package auth

import (
	"net/http"

	"dixitme/internal/logger"

	"github.com/gin-gonic/gin"
)

// AuthHandlers contains authentication HTTP handlers
type AuthHandlers struct {
	authService *AuthService
	jwtService  *JWTService
	enableSSO   bool
}

// NewAuthHandlers creates new authentication handlers
func NewAuthHandlers(authService *AuthService, jwtService *JWTService, enableSSO bool) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		jwtService:  jwtService,
		enableSSO:   enableSSO,
	}
}

// Request/Response types

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Username    string `json:"username" binding:"required,min=3,max=50"`
	DisplayName string `json:"display_name" binding:"required,min=1,max=100"`
	Password    string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	EmailOrUsername string `json:"email_or_username" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

type GoogleLoginRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
}

type GuestLoginRequest struct {
	Name string `json:"name,omitempty"`
}

type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

type AuthResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	User    interface{} `json:"user,omitempty"`
	Token   string      `json:"token,omitempty"`
	Type    string      `json:"type"` // "registered", "guest"
}

type UserResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email,omitempty"`
	Username    string `json:"username,omitempty"`
	DisplayName string `json:"display_name"`
	AuthType    string `json:"auth_type"`
	Avatar      string `json:"avatar,omitempty"`
}

type GuestResponse struct {
	SessionID string `json:"session_id"`
	Name      string `json:"name"`
	AuthType  string `json:"auth_type"`
}

// Authentication endpoints

// @Summary Register with email/password
// @Description Create a new user account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration data"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /auth/register [post]
func (h *AuthHandlers) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.RegisterWithPassword(req.Email, req.Username, req.DisplayName, req.Password)
	if err != nil {
		statusCode := http.StatusBadRequest
		if err.Error() == "user with this email or username already exists" {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	userResp := UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AuthType:    string(user.AuthType),
		Avatar:      user.Avatar,
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Success: true,
		Message: "User registered successfully",
		User:    userResp,
		Type:    "registered",
	})
}

// @Summary Login with email/password
// @Description Authenticate user with email/username and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 401 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, session, token, err := h.authService.LoginWithPassword(
		req.EmailOrUsername,
		req.Password,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userResp := UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AuthType:    string(user.AuthType),
		Avatar:      user.Avatar,
	}

	// Set token as HTTP-only cookie
	c.SetCookie("auth_token", token, 86400, "/", "", false, true) // 24 hours

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Login successful",
		User:    userResp,
		Token:   token,
		Type:    "registered",
	})

	logger.GetLogger().Info("User logged in", "user_id", user.ID, "session_id", session.ID)
}

// @Summary Login with Google OAuth
// @Description Authenticate user with Google OAuth access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body GoogleLoginRequest true "Google access token"
// @Success 200 {object} AuthResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /auth/google [post]
func (h *AuthHandlers) GoogleLogin(c *gin.Context) {
	// Check if SSO is enabled
	if !h.enableSSO {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "SSO authentication is temporarily disabled",
			"code":  "SSO_DISABLED",
		})
		return
	}

	var req GoogleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, session, token, err := h.authService.LoginWithGoogle(
		req.AccessToken,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	userResp := UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AuthType:    string(user.AuthType),
		Avatar:      user.Avatar,
	}

	// Set token as HTTP-only cookie
	c.SetCookie("auth_token", token, 86400, "/", "", false, true) // 24 hours

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Google login successful",
		User:    userResp,
		Token:   token,
		Type:    "registered",
	})

	logger.GetLogger().Info("User logged in with Google", "user_id", user.ID, "session_id", session.ID)
}

// @Summary Guest login
// @Description Create a guest session without registration
// @Tags auth
// @Accept json
// @Produce json
// @Param request body GuestLoginRequest true "Guest name (optional)"
// @Success 200 {object} AuthResponse
// @Failure 500 {object} map[string]interface{}
// @Router /auth/guest [post]
func (h *AuthHandlers) GuestLogin(c *gin.Context) {
	var req GuestLoginRequest
	c.ShouldBindJSON(&req) // Optional request body

	session, token, err := h.authService.CreateGuestSession(
		req.Name,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	guestName := req.Name
	if guestName == "" {
		guestName = "Guest " + session.ID.String()[:8]
	}

	guestResp := GuestResponse{
		SessionID: session.ID.String(),
		Name:      guestName,
		AuthType:  string(session.AuthType),
	}

	// Set token as HTTP-only cookie
	c.SetCookie("auth_token", token, 86400, "/", "", false, true) // 24 hours

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Guest session created",
		User:    guestResp,
		Token:   token,
		Type:    "guest",
	})

	logger.GetLogger().Info("Guest session created", "session_id", session.ID)
}

// @Summary Refresh token
// @Description Refresh authentication token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Token to refresh"
// @Success 200 {object} AuthResponse
// @Failure 401 {object} map[string]interface{}
// @Router /auth/refresh [post]
func (h *AuthHandlers) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extract session ID from token
	userInfo, err := h.jwtService.ExtractUserInfo(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	session, newToken, err := h.authService.RefreshSession(userInfo.SessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Set new token as cookie
	c.SetCookie("auth_token", newToken, 86400, "/", "", false, true)

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Token refreshed",
		Token:   newToken,
		Type:    string(session.AuthType),
	})
}

// @Summary Logout
// @Description Logout user and invalidate session
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *AuthHandlers) Logout(c *gin.Context) {
	userInfo, exists := GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	if err := h.authService.Logout(userInfo.SessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	// Clear cookie
	c.SetCookie("auth_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logged out successfully",
	})
}

// @Summary Get current user
// @Description Get current authenticated user information
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/me [get]
func (h *AuthHandlers) GetCurrentUser(c *gin.Context) {
	userInfo, exists := GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	if userInfo.AuthType == "guest" {
		guestResp := GuestResponse{
			SessionID: userInfo.SessionID.String(),
			Name:      userInfo.Name,
			AuthType:  string(userInfo.AuthType),
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"user":    guestResp,
			"type":    "guest",
		})
		return
	}

	// Get full user data for registered users
	if userInfo.UserID == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user data"})
		return
	}

	user, err := h.authService.GetUserByID(*userInfo.UserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	userResp := UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		AuthType:    string(user.AuthType),
		Avatar:      user.Avatar,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    userResp,
		"type":    "registered",
	})
}

// @Summary Validate token
// @Description Validate authentication token
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /auth/validate [get]
func (h *AuthHandlers) ValidateToken(c *gin.Context) {
	userInfo, exists := GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"valid": false,
			"error": "Invalid or missing token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"session_id": userInfo.SessionID,
		"auth_type":  userInfo.AuthType,
		"name":       userInfo.Name,
	})
}

// @Summary Get authentication status
// @Description Get current authentication configuration and available methods
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/status [get]
func (h *AuthHandlers) GetAuthStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"methods": gin.H{
			"password": true,
			"google":   h.enableSSO,
			"guest":    true,
		},
		"sso_enabled": h.enableSSO,
		"version":     "1.0",
	})
}
