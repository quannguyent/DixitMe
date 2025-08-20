package game

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"dixitme/internal/logger"
	"dixitme/internal/models"
	"dixitme/internal/services/bot"

	"github.com/google/uuid"
)

// CreateGame creates a new game with the given room code
func (m *Manager) CreateGame(roomCode string, creatorID uuid.UUID, creatorName string) (*GameState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if room code already exists
	if _, exists := m.games[roomCode]; exists {
		return nil, fmt.Errorf("room code already exists")
	}

	// Create new game state
	gameID := uuid.New()
	now := time.Now()

	// Initialize deck with all available cards (1-84 for standard Dixit)
	deck := make([]int, 84)
	for i := 0; i < 84; i++ {
		deck[i] = i + 1
	}
	// Shuffle the deck
	for i := len(deck) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}

	game := &GameState{
		ID:           gameID,
		RoomCode:     roomCode,
		Players:      make(map[uuid.UUID]*Player),
		Status:       models.GameStatusWaiting,
		RoundNumber:  0,
		MaxRounds:    999, // Will be determined by 30 points or empty deck
		Deck:         deck,
		UsedCards:    make([]int, 0),
		CreatedAt:    now,
		LastActivity: now,
	}

	// Add creator as first player
	creator := &Player{
		ID:          creatorID,
		Name:        creatorName,
		Score:       0,
		Position:    1,
		Hand:        make([]int, 0),
		IsConnected: true,
		IsActive:    true,
	}

	game.Players[creatorID] = creator

	// Store in memory
	m.games[roomCode] = game

	// Persist to database
	if err := m.persistGame(game); err != nil {
		delete(m.games, roomCode)
		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") &&
			strings.Contains(err.Error(), "games_room_code_key") {
			return nil, fmt.Errorf("room code '%s' is already taken, please try a different one", roomCode)
		}
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	// Store in Redis for scaling
	if err := m.storeGameInRedis(game); err != nil {
		logger.Error("Failed to store game in Redis", "error", err, "room_code", roomCode)
	}

	return game, nil
}

// JoinGame adds a player to an existing game
func (m *Manager) JoinGame(roomCode string, playerID uuid.UUID, playerName string) (*GameState, error) {
	m.mu.RLock()
	game, exists := m.games[roomCode]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("game not found")
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	// Update activity
	game.LastActivity = time.Now()

	if game.Status != models.GameStatusWaiting {
		return nil, fmt.Errorf("game already started")
	}

	if len(game.Players) >= 6 {
		return nil, fmt.Errorf("game is full")
	}

	// Check if player already exists
	if _, exists := game.Players[playerID]; exists {
		return nil, fmt.Errorf("player already in game")
	}

	// Create new player
	player := &Player{
		ID:          playerID,
		Name:        playerName,
		Score:       0,
		Position:    len(game.Players) + 1,
		Hand:        make([]int, 0),
		IsConnected: true,
		IsActive:    true,
	}

	game.Players[playerID] = player

	// Persist player
	if err := m.persistGamePlayer(game.ID, player); err != nil {
		delete(game.Players, playerID)
		return nil, fmt.Errorf("failed to persist player: %w", err)
	}

	// Update Redis
	if err := m.storeGameInRedis(game); err != nil {
		logger.Error("Failed to update game in Redis", "error", err, "room_code", roomCode)
	}

	// Broadcast player joined
	m.BroadcastToGame(game, MessageTypePlayerJoined, PlayerJoinedPayload{Player: player})

	// Send system message
	m.SendSystemMessage(roomCode, fmt.Sprintf("%s joined the game", playerName))

	return game, nil
}

// AddBot adds a bot player to an existing game
func (m *Manager) AddBot(roomCode string, botLevel string) (*GameState, error) {
	m.mu.RLock()
	game, exists := m.games[roomCode]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("game not found")
	}

	game.Lock()
	defer game.Unlock()

	if game.Status != models.GameStatusWaiting {
		return nil, fmt.Errorf("cannot add bot to game in progress")
	}

	if len(game.Players) >= 6 {
		return nil, fmt.Errorf("game is full")
	}

	// Create bot player
	botNames := bot.GetBotNames()
	botName := botNames[rand.Intn(len(botNames))]

	// Ensure unique bot name
	for {
		nameExists := false
		for _, player := range game.Players {
			if player.Name == botName {
				nameExists = true
				break
			}
		}
		if !nameExists {
			break
		}
		botName = botNames[rand.Intn(len(botNames))]
	}

	botID := uuid.New()

	// Create bot in bot manager
	botManager := bot.GetBotManager()
	botPlayer := botManager.CreateBot(botName, bot.BotDifficulty(botLevel))
	botPlayer.SetGameID(game.ID)

	// Create game player
	player := &Player{
		ID:          botID,
		Name:        botName,
		Score:       0,
		Position:    len(game.Players),
		Hand:        make([]int, 0),
		Connection:  nil,
		IsConnected: true, // Bots are always "connected"
		IsActive:    true,
		IsBot:       true,
		BotLevel:    botLevel,
	}

	game.Players[botID] = player

	// Persist bot player to database
	dbPlayer := &models.Player{
		ID:       botID,
		Name:     botName,
		Type:     models.PlayerTypeBot,
		BotLevel: botLevel,
	}

	if err := m.persistPlayer(dbPlayer); err != nil {
		delete(game.Players, botID)
		return nil, fmt.Errorf("failed to persist bot player: %w", err)
	}

	if err := m.persistGamePlayer(game.ID, player); err != nil {
		delete(game.Players, botID)
		return nil, fmt.Errorf("failed to persist bot game player: %w", err)
	}

	// Update Redis
	if err := m.storeGameInRedis(game); err != nil {
		logger.Error("Failed to update game in Redis", "error", err, "room_code", roomCode)
	}

	// Broadcast bot joined
	m.BroadcastToGame(game, MessageTypePlayerJoined, PlayerJoinedPayload{Player: player})

	// Send system message
	m.SendSystemMessage(roomCode, fmt.Sprintf("Bot %s (%s difficulty) joined the game", botName, botLevel))

	logger.Info("Bot added to game", "bot_id", botID, "bot_name", botName, "bot_level", botLevel, "room_code", roomCode)

	return game, nil
}

// StartGame starts a game if conditions are met
func (m *Manager) StartGame(roomCode string, playerID uuid.UUID) error {
	m.mu.RLock()
	game, exists := m.games[roomCode]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("game not found")
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	// Update activity
	game.LastActivity = time.Now()

	// Check if player is in the game
	if _, exists := game.Players[playerID]; !exists {
		return fmt.Errorf("player not in game")
	}

	// Check if game can start (minimum 3 players)
	if len(game.Players) < 3 {
		return fmt.Errorf("need at least 3 players to start")
	}

	if game.Status != models.GameStatusWaiting {
		return fmt.Errorf("game already started")
	}

	// Initialize game
	game.Status = models.GameStatusInProgress

	// Deal cards to players
	m.dealCards(game)

	// Start first round
	if err := m.startNewRound(game); err != nil {
		return fmt.Errorf("failed to start first round: %w", err)
	}

	// Update database
	if err := m.updateGameStatus(game.ID, models.GameStatusInProgress); err != nil {
		return fmt.Errorf("failed to update game status: %w", err)
	}

	// Broadcast game started
	m.BroadcastToGame(game, MessageTypeGameStarted, GameStartedPayload{GameState: game})

	// Send system message
	m.SendSystemMessage(roomCode, "Game started! Let the storytelling begin!")

	return nil
}

// Helper methods

func (m *Manager) GetGame(roomCode string) *GameState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.games[roomCode]
}

func (m *Manager) getGame(roomCode string) *GameState {
	return m.GetGame(roomCode)
}

func (m *Manager) GetActiveGamesCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.games)
}
