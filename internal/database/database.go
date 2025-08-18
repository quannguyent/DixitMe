package database

import (
	"dixitme/internal/logger"
	"dixitme/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func Initialize(databaseURL string) {
	var err error
	log := logger.GetLogger()

	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})

	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		panic(err)
	}

	// Run migrations
	if err := migrate(); err != nil {
		log.Error("Failed to run migrations", "error", err)
		panic(err)
	}

	log.Info("Database connection established and migrations completed")
}

func migrate() error {
	return DB.AutoMigrate(
		&models.Player{},
		&models.Game{},
		&models.GamePlayer{},
		&models.GameRound{},
		&models.CardSubmission{},
		&models.Vote{},
		&models.Card{},
		&models.Tag{},
		&models.CardTag{},
		&models.GameHistory{},
		&models.ChatMessage{},
	)
}

func GetDB() *gorm.DB {
	return DB
}
