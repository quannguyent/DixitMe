package game

import (
	"sync"
	"time"
)

// Manager manages all active games
type Manager struct {
	games           map[string]*GameState
	mu              sync.RWMutex
	cleanupInterval time.Duration
	inactiveTimeout time.Duration
	stopCleanup     chan bool
}

var gameManager *Manager

// GetManager returns the singleton game manager
func GetManager() *Manager {
	if gameManager == nil {
		gameManager = &Manager{
			games:           make(map[string]*GameState),
			cleanupInterval: 2 * time.Minute,  // Check every 2 minutes
			inactiveTimeout: 10 * time.Minute, // This will be dynamic based on room state
			stopCleanup:     make(chan bool),
		}
		// Start the cleanup goroutine
		go gameManager.startCleanupService()
	}
	return gameManager
}
