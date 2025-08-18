package main

import (
	"flag"
	"fmt"
	"os"

	"dixitme/internal/config"
	"dixitme/internal/database"
	"dixitme/internal/logger"
	"dixitme/internal/seeder"
	"dixitme/internal/storage"
)

func main() {
	// Define flags
	var (
		tagsOnly  = flag.Bool("tags", false, "Seed only tags")
		cardsOnly = flag.Bool("cards", false, "Seed only cards")
		force     = flag.Bool("force", false, "Force reseed (delete existing data)")
		help      = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		fmt.Println("DixitMe Database Seeder")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/seed/main.go [options]")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -tags      Seed only tags")
		fmt.Println("  -cards     Seed only cards")
		fmt.Println("  -force     Force reseed (delete existing data)")
		fmt.Println("  -help      Show this help message")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run cmd/seed/main.go                 # Seed everything")
		fmt.Println("  go run cmd/seed/main.go -tags           # Seed only tags")
		fmt.Println("  go run cmd/seed/main.go -cards          # Seed only cards")
		fmt.Println("  go run cmd/seed/main.go -force          # Force complete reseed")
		return
	}

	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.InitLogger(cfg.Logger)
	log := logger.GetLogger()

	log.Info("Starting database seeding...")

	// Initialize database
	database.Initialize(cfg.DatabaseURL)

	// Initialize MinIO (optional)
	if err := storage.Initialize(cfg.MinIO); err != nil {
		log.Warn("MinIO not available, using local file paths", "error", err)
	}

	db := database.GetDB()

	if *force {
		log.Warn("Force mode: deleting existing data...")

		// Delete in correct order due to foreign key constraints
		if err := db.Exec("DELETE FROM card_tag_relations").Error; err != nil {
			log.Error("Failed to delete card-tag relations", "error", err)
			os.Exit(1)
		}

		if err := db.Exec("DELETE FROM cards").Error; err != nil {
			log.Error("Failed to delete cards", "error", err)
			os.Exit(1)
		}

		if err := db.Exec("DELETE FROM tags").Error; err != nil {
			log.Error("Failed to delete tags", "error", err)
			os.Exit(1)
		}

		log.Info("Existing data deleted")
	}

	// Perform seeding based on flags
	if *tagsOnly {
		log.Info("Seeding tags only...")
		if err := seeder.SeedTagsOnly(); err != nil {
			log.Error("Failed to seed tags", "error", err)
			os.Exit(1)
		}
	} else if *cardsOnly {
		log.Info("Seeding cards only...")
		if err := seeder.SeedCardsOnly(); err != nil {
			log.Error("Failed to seed cards", "error", err)
			os.Exit(1)
		}
	} else {
		log.Info("Seeding complete database...")
		if err := seeder.SeedDatabase(); err != nil {
			log.Error("Failed to seed database", "error", err)
			os.Exit(1)
		}
	}

	// Show final statistics
	var tagCount, cardCount, relationCount int64
	db.Table("tags").Count(&tagCount)
	db.Table("cards").Count(&cardCount)
	db.Table("card_tag_relations").Count(&relationCount)

	log.Info("Seeding completed successfully!",
		"tags", tagCount,
		"cards", cardCount,
		"relationships", relationCount)

	fmt.Println()
	fmt.Printf("✅ Database seeded successfully!\n")
	fmt.Printf("   - Tags: %d\n", tagCount)
	fmt.Printf("   - Cards: %d\n", cardCount)
	fmt.Printf("   - Tag relationships: %d\n", relationCount)
	fmt.Println()
}
