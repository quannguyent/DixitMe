package game

import (
	"sync"
	"time"

	"dixitme/internal/database"
	"dixitme/internal/logger"
	"dixitme/internal/models"
	redisClient "dixitme/internal/redis"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var globalManager *Manager
var managerOnce sync.Once

// Global connection registry within the game package to avoid import cycles
var (
	playerConnections = make(map[uuid.UUID]*websocket.Conn)
	connectionsMutex  sync.RWMutex
)

// Manager manages all active games
type Manager struct {
	games           map[string]*GameState
	mu              sync.RWMutex
	cleanupInterval time.Duration
	inactiveTimeout time.Duration
	stopCleanup     chan bool

	// Injected dependencies
	db          *gorm.DB
	redisClient *redis.Client
}

// NewManager creates a new game manager instance with injected dependencies
func NewManager(db *gorm.DB, redisClient *redis.Client) *Manager {
	manager := &Manager{
		games:           make(map[string]*GameState),
		cleanupInterval: 2 * time.Minute,  // Check every 2 minutes
		inactiveTimeout: 10 * time.Minute, // This will be dynamic based on room state
		stopCleanup:     make(chan bool),
		db:              db,
		redisClient:     redisClient,
	}
	// Load active games from database
	go manager.loadActiveGamesFromDatabase()
	// Start the cleanup goroutine
	go manager.startCleanupService()
	return manager
}

// FullGameService combines all game-related services
// This is what most components will depend on
type FullGameService interface {
	GameService
	GamePlayService
	BotService
	ChatService
	GameCleanupService
	GameBroadcastService
	GamePersistenceService
}

// GetManager returns the singleton game manager (for backward compatibility)
// This is used by WebSocket handlers until proper DI is implemented there
func GetManager() *Manager {
	managerOnce.Do(func() {
		// For backward compatibility, import the database/redis packages here
		// This is not ideal but maintains compatibility for WebSocket handlers
		db := database.GetDB()
		redisConn := redisClient.GetClient()
		globalManager = NewManager(db, redisConn)
	})
	return globalManager
}

// Connection management functions
func RegisterPlayerConnection(playerID uuid.UUID, conn *websocket.Conn) {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()
	playerConnections[playerID] = conn
	log := logger.GetLogger()
	log.Info("Registered player connection", "player_id", playerID, "total_connections", len(playerConnections))
}

func UnregisterPlayerConnection(playerID uuid.UUID) {
	connectionsMutex.Lock()
	defer connectionsMutex.Unlock()
	delete(playerConnections, playerID)
}

func GetPlayerConnection(playerID uuid.UUID) *websocket.Conn {
	connectionsMutex.RLock()
	defer connectionsMutex.RUnlock()
	conn := playerConnections[playerID]
	log := logger.GetLogger()
	log.Debug("Retrieved player connection", "player_id", playerID, "found", conn != nil)
	return conn
}

// loadActiveGamesFromDatabase loads active/waiting games from database into memory
func (m *Manager) loadActiveGamesFromDatabase() {
	log := logger.GetLogger()
	log.Info("Loading active games from database...")

	var dbGames []models.Game
	// Load games that are still active (waiting or in_progress)
	if err := m.db.Where("status IN ?", []string{"waiting", "in_progress"}).
		Preload("Players.Player").
		Find(&dbGames).Error; err != nil {
		log.Error("Failed to load games from database", "error", err)
		return
	}

	loadedCount := 0
	for _, dbGame := range dbGames {
		// Convert database game to in-memory GameState
		gameState := m.convertDBGameToGameState(&dbGame)
		if gameState != nil {
			m.mu.Lock()
			m.games[dbGame.RoomCode] = gameState
			m.mu.Unlock()
			loadedCount++
		}
	}

	log.Info("Successfully loaded games from database", "count", loadedCount)
}

// convertDBGameToGameState converts a database Game model to in-memory GameState
func (m *Manager) convertDBGameToGameState(dbGame *models.Game) *GameState {
	log := logger.GetLogger()

	// Convert database players to in-memory players
	players := make(map[uuid.UUID]*Player)
	for _, dbGamePlayer := range dbGame.Players {
		player := &Player{
			ID:          dbGamePlayer.Player.ID,
			Name:        dbGamePlayer.Player.Name,
			IsBot:       dbGamePlayer.Player.Type == models.PlayerTypeBot,
			IsActive:    dbGamePlayer.IsActive,
			Score:       dbGamePlayer.Score,
			Hand:        make([]int, 0), // Will be loaded separately if needed
			IsConnected: false,          // Will be set when WebSocket connects
			BotLevel:    dbGamePlayer.Player.BotLevel,
			Position:    dbGamePlayer.Position,
		}
		players[dbGamePlayer.Player.ID] = player
	}

	// Create GameState
	gameState := &GameState{
		ID:           dbGame.ID,
		RoomCode:     dbGame.RoomCode,
		Status:       dbGame.Status,
		Players:      players,
		CreatedAt:    dbGame.CreatedAt,
		LastActivity: time.Now(), // Set to now since we're loading it
		// Initialize other fields with defaults
		CurrentRound: nil,
		RoundNumber:  0,
		MaxRounds:    10, // Default value
		Deck:         make([]int, 0),
		UsedCards:    make([]int, 0),
	}

	log.Debug("Converted database game to in-memory state",
		"room_code", dbGame.RoomCode,
		"status", dbGame.Status,
		"player_count", len(players))

	return gameState
}

// LoadGameFromDatabase loads a single game from database into memory (public method for handlers)
func (m *Manager) LoadGameFromDatabase(dbGame *models.Game) *GameState {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if the game already exists in memory
	if existingGame, exists := m.games[dbGame.RoomCode]; exists {
		// Game already loaded, return existing state to preserve connections
		return existingGame
	}

	// Convert database game to in-memory state
	gameState := m.convertDBGameToGameState(dbGame)
	if gameState != nil {
		m.games[dbGame.RoomCode] = gameState

		log := logger.GetLogger()
		log.Info("Loaded game from database into memory",
			"room_code", dbGame.RoomCode,
			"status", dbGame.Status)
	}
	return gameState
}
