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

	// Skip User and Session for now, test other models
	log.Info("Testing migration with simpler models first...")

	// Try migrating Card first (it uses int PK, simpler)
	log.Info("Migrating Card table...")
	if err := DB.AutoMigrate(&models.Card{}); err != nil {
		log.Error("Failed to migrate Card", "error", err)
		return err
	}

	log.Info("Card migration successful, migrating Tag...")
	if err := DB.AutoMigrate(&models.Tag{}); err != nil {
		log.Error("Failed to migrate Tag", "error", err)
		return err
	}

	log.Info("Basic migrations successful, now trying UUID models...")
	if err := DB.AutoMigrate(&models.User{}, &models.Session{}); err != nil {
		log.Error("Failed to migrate UUID models", "error", err)
		return err
	}

	log.Info("All migrations successful!")
	return nil
}

func GetDB() *gorm.DB {
	return DB
}

// SetDB sets the database instance (used for testing)
func SetDB(database *gorm.DB) {
	DB = database
}
