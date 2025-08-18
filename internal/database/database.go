package database

import (
	"log"

	"dixitme/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Initialize(databaseURL string) {
	var err error

	DB, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := migrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	log.Println("Database connection established and migrations completed")
}

func migrate() error {
	return DB.AutoMigrate(
		&models.Player{},
		&models.Game{},
		&models.GamePlayer{},
		&models.GameRound{},
		&models.CardSubmission{},
		&models.Vote{},
		&models.GameHistory{},
	)
}

func GetDB() *gorm.DB {
	return DB
}
