package handlers

import (
	"dixitme/internal/services/auth"
	"dixitme/internal/services/game"
)

// HandlerDependencies holds all the dependencies that handlers need
type HandlerDependencies struct {
	AuthService auth.AuthenticationService
	GameService game.FullGameService
	JWTService  *auth.JWTService
}

// NewHandlerDependencies creates a new HandlerDependencies instance
func NewHandlerDependencies(
	authService auth.AuthenticationService,
	gameService game.FullGameService,
	jwtService *auth.JWTService,
) *HandlerDependencies {
	return &HandlerDependencies{
		AuthService: authService,
		GameService: gameService,
		JWTService:  jwtService,
	}
}
