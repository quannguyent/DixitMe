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
	log := logger.GetLogger()

	// Start with simpler models first
	log.Info("Migrating basic models (Card, Tag)...")
	if err := DB.AutoMigrate(&models.Card{}, &models.Tag{}); err != nil {
		log.Error("Failed to migrate basic models", "error", err)
		return err
	}

	// Migrate user and authentication models
	log.Info("Migrating user and authentication models...")
	if err := DB.AutoMigrate(&models.User{}, &models.Session{}); err != nil {
		log.Error("Failed to migrate user models", "error", err)
		return err
	}

	// Migrate player model (depends on User)
	log.Info("Migrating Player model...")
	if err := DB.AutoMigrate(&models.Player{}); err != nil {
		log.Error("Failed to migrate Player model", "error", err)
		return err
	}

	// Migrate game models (depends on Player)
	log.Info("Migrating game models...")
	if err := DB.AutoMigrate(&models.Game{}, &models.GamePlayer{}, &models.GameHistory{}); err != nil {
		log.Error("Failed to migrate game models", "error", err)
		return err
	}

	// Migrate round models (depends on Game and Player)
	log.Info("Migrating round models...")
	if err := DB.AutoMigrate(&models.GameRound{}, &models.CardSubmission{}, &models.Vote{}); err != nil {
		log.Error("Failed to migrate round models", "error", err)
		return err
	}

	// Migrate chat model if it exists
	log.Info("Migrating chat models...")
	if err := DB.AutoMigrate(&models.ChatMessage{}); err != nil {
		log.Error("Failed to migrate chat models", "error", err)
		return err
	}

	log.Info("All database migrations completed successfully!")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}

// SetDB sets the database instance (used for testing)
func SetDB(database *gorm.DB) {
	DB = database
}
