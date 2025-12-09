package postgres

import (
	"dixitme/internal/data/models"
	"dixitme/internal/platform/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func migrate() error {
	return DB.AutoMigrate(
		&model.User{},
		&model.Session{},
		&model.Player{},
		&model.Game{},
		&model.GamePlayer{},
		&model.GameRound{},
		&model.CardSubmission{},
		&model.Vote{},
		&model.Card{},
		&model.Tag{},
		&model.CardTag{},
		&model.GameHistory{},
		&model.ChatMessage{},
	)
}

// Open returns a new gorm DB connection without mutating global state.
func Open(databaseURL string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
}

// Initialize opens a DB connection, runs migrations, and sets the package global for legacy callers.
func Initialize(databaseURL string) *gorm.DB {
	log := logger.GetLogger()

	db, err := Open(databaseURL)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		panic(err)
	}

	DB = db

	if err := migrate(); err != nil {
		log.Error("Failed to run migrations", "error", err)
		panic(err)
	}

	log.Info("Database connection established and migrations completed")
	return DB
}

func GetDB() *gorm.DB {
	return DB
}
