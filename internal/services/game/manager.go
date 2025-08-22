package game

import (
	"sync"
	"time"

	"dixitme/internal/database"
	redisClient "dixitme/internal/redis"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var globalManager *Manager
var managerOnce sync.Once

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
