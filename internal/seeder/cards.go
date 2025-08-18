package seeder

import (
	"fmt"

	"dixitme/internal/database"
	"dixitme/internal/logger"
	"dixitme/internal/models"
	"dixitme/internal/storage"
)

// CardData represents a card definition for seeding
type CardData struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Extension   string   `json:"extension"`
	Tags        []string `json:"tags"`
}

// TagData represents a tag definition for seeding
type TagData struct {
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Color       string  `json:"color"`
	Weight      float64 `json:"weight"`
}

// GetDefaultTags returns the predefined tags for card categorization
func GetDefaultTags() []TagData {
	return []TagData{
		// Emotion tags
		{Name: "Happy", Slug: "happy", Description: "Joyful, cheerful, positive emotions", Category: "emotion", Color: "#FCD34D", Weight: 1.0},
		{Name: "Sad", Slug: "sad", Description: "Melancholy, sorrow, grief", Category: "emotion", Color: "#60A5FA", Weight: 1.0},
		{Name: "Angry", Slug: "angry", Description: "Rage, fury, aggression", Category: "emotion", Color: "#F87171", Weight: 1.0},
		{Name: "Fear", Slug: "fear", Description: "Scary, frightening, terror", Category: "emotion", Color: "#6B7280", Weight: 1.2},
		{Name: "Love", Slug: "love", Description: "Romance, affection, caring", Category: "emotion", Color: "#F472B6", Weight: 1.1},
		{Name: "Surprise", Slug: "surprise", Description: "Shock, amazement, wonder", Category: "emotion", Color: "#A78BFA", Weight: 1.0},
		{Name: "Peaceful", Slug: "peaceful", Description: "Calm, serene, tranquil", Category: "emotion", Color: "#34D399", Weight: 1.0},

		// Nature tags
		{Name: "Forest", Slug: "forest", Description: "Trees, woodland, jungle", Category: "nature", Color: "#059669", Weight: 1.0},
		{Name: "Ocean", Slug: "ocean", Description: "Sea, waves, marine life", Category: "nature", Color: "#0EA5E9", Weight: 1.0},
		{Name: "Mountain", Slug: "mountain", Description: "Peaks, hills, rocky terrain", Category: "nature", Color: "#78716C", Weight: 1.0},
		{Name: "Sky", Slug: "sky", Description: "Clouds, heavens, aerial views", Category: "nature", Color: "#38BDF8", Weight: 1.0},
		{Name: "Desert", Slug: "desert", Description: "Sand, arid landscapes", Category: "nature", Color: "#FDE047", Weight: 1.0},
		{Name: "Garden", Slug: "garden", Description: "Flowers, plants, cultivation", Category: "nature", Color: "#22C55E", Weight: 1.0},
		{Name: "Storm", Slug: "storm", Description: "Lightning, rain, tempest", Category: "nature", Color: "#64748B", Weight: 1.1},

		// Fantasy tags
		{Name: "Magic", Slug: "magic", Description: "Spells, enchantment, supernatural", Category: "fantasy", Color: "#8B5CF6", Weight: 1.3},
		{Name: "Dragon", Slug: "dragon", Description: "Mythical dragons and serpents", Category: "fantasy", Color: "#DC2626", Weight: 1.4},
		{Name: "Fairy", Slug: "fairy", Description: "Fairies, sprites, mystical beings", Category: "fantasy", Color: "#EC4899", Weight: 1.2},
		{Name: "Wizard", Slug: "wizard", Description: "Mages, sorcerers, magic users", Category: "fantasy", Color: "#7C3AED", Weight: 1.3},
		{Name: "Castle", Slug: "castle", Description: "Fortresses, palaces, kingdoms", Category: "fantasy", Color: "#6B7280", Weight: 1.1},
		{Name: "Treasure", Slug: "treasure", Description: "Gold, gems, valuable items", Category: "fantasy", Color: "#F59E0B", Weight: 1.2},
		{Name: "Quest", Slug: "quest", Description: "Adventure, journey, mission", Category: "fantasy", Color: "#10B981", Weight: 1.2},

		// Animal tags
		{Name: "Bird", Slug: "bird", Description: "Flying creatures, feathers", Category: "animal", Color: "#06B6D4", Weight: 1.0},
		{Name: "Cat", Slug: "cat", Description: "Felines, domestic cats", Category: "animal", Color: "#F97316", Weight: 1.0},
		{Name: "Dog", Slug: "dog", Description: "Canines, loyal companions", Category: "animal", Color: "#EAB308", Weight: 1.0},
		{Name: "Wild Animal", Slug: "wild-animal", Description: "Untamed creatures", Category: "animal", Color: "#84CC16", Weight: 1.1},
		{Name: "Fish", Slug: "fish", Description: "Aquatic creatures", Category: "animal", Color: "#0EA5E9", Weight: 1.0},
		{Name: "Horse", Slug: "horse", Description: "Equines, noble steeds", Category: "animal", Color: "#A3A3A3", Weight: 1.0},

		// Human activity tags
		{Name: "Dance", Slug: "dance", Description: "Movement, rhythm, celebration", Category: "activity", Color: "#F472B6", Weight: 1.0},
		{Name: "Music", Slug: "music", Description: "Songs, instruments, melody", Category: "activity", Color: "#A855F7", Weight: 1.0},
		{Name: "Art", Slug: "art", Description: "Painting, creativity, expression", Category: "activity", Color: "#EC4899", Weight: 1.0},
		{Name: "Sport", Slug: "sport", Description: "Athletics, competition, games", Category: "activity", Color: "#EF4444", Weight: 1.0},
		{Name: "Reading", Slug: "reading", Description: "Books, knowledge, learning", Category: "activity", Color: "#8B5CF6", Weight: 1.0},
		{Name: "Cooking", Slug: "cooking", Description: "Food preparation, culinary arts", Category: "activity", Color: "#F97316", Weight: 1.0},

		// Object tags
		{Name: "Key", Slug: "key", Description: "Unlocking, secrets, access", Category: "object", Color: "#FCD34D", Weight: 1.1},
		{Name: "Mirror", Slug: "mirror", Description: "Reflection, self-image", Category: "object", Color: "#E5E7EB", Weight: 1.2},
		{Name: "Clock", Slug: "clock", Description: "Time, schedules, deadlines", Category: "object", Color: "#6B7280", Weight: 1.1},
		{Name: "Book", Slug: "book", Description: "Knowledge, stories, wisdom", Category: "object", Color: "#7C2D12", Weight: 1.0},
		{Name: "Candle", Slug: "candle", Description: "Light, warmth, ambiance", Category: "object", Color: "#FDE047", Weight: 1.0},
		{Name: "Crown", Slug: "crown", Description: "Royalty, leadership, power", Category: "object", Color: "#F59E0B", Weight: 1.3},

		// Abstract concepts
		{Name: "Dream", Slug: "dream", Description: "Sleep visions, aspirations", Category: "abstract", Color: "#C084FC", Weight: 1.2},
		{Name: "Memory", Slug: "memory", Description: "Recollections, nostalgia", Category: "abstract", Color: "#93C5FD", Weight: 1.1},
		{Name: "Hope", Slug: "hope", Description: "Optimism, future possibilities", Category: "abstract", Color: "#FDE047", Weight: 1.1},
		{Name: "Freedom", Slug: "freedom", Description: "Liberation, independence", Category: "abstract", Color: "#22D3EE", Weight: 1.2},
		{Name: "Mystery", Slug: "mystery", Description: "Unknown, puzzling, enigmatic", Category: "abstract", Color: "#6366F1", Weight: 1.3},
		{Name: "Balance", Slug: "balance", Description: "Harmony, equilibrium", Category: "abstract", Color: "#10B981", Weight: 1.1},

		// Time and season tags
		{Name: "Night", Slug: "night", Description: "Darkness, moon, stars", Category: "time", Color: "#1E293B", Weight: 1.0},
		{Name: "Day", Slug: "day", Description: "Sunlight, brightness, morning", Category: "time", Color: "#FDE047", Weight: 1.0},
		{Name: "Winter", Slug: "winter", Description: "Snow, cold, hibernation", Category: "time", Color: "#E0E7FF", Weight: 1.0},
		{Name: "Spring", Slug: "spring", Description: "Growth, renewal, flowers", Category: "time", Color: "#BBF7D0", Weight: 1.0},
		{Name: "Summer", Slug: "summer", Description: "Heat, vacation, abundance", Category: "time", Color: "#FEF3C7", Weight: 1.0},
		{Name: "Autumn", Slug: "autumn", Description: "Harvest, falling leaves", Category: "time", Color: "#FED7AA", Weight: 1.0},
	}
}

// GetDefaultCards returns the predefined card data for the Dixit game
func GetDefaultCards() []CardData {
	return []CardData{
		// Nature and landscape cards (1-20)
		{ID: 1, Title: "Ancient Oak", Description: "A massive oak tree with twisted branches reaching toward the sky", Extension: ".jpg", Tags: []string{"forest", "nature", "ancient"}},
		{ID: 2, Title: "Rose Garden", Description: "A blooming garden filled with red and pink roses", Extension: ".jpg", Tags: []string{"garden", "love", "beauty"}},
		{ID: 3, Title: "Mountain Lake", Description: "A serene lake reflecting snow-capped peaks", Extension: ".jpg", Tags: []string{"mountain", "peaceful", "water"}},
		{ID: 4, Title: "Forest Creatures", Description: "Woodland animals gathering in a mystical forest", Extension: ".jpg", Tags: []string{"forest", "wild-animal", "magic"}},
		{ID: 5, Title: "Stormy Sky", Description: "Dark clouds with lightning illuminating the horizon", Extension: ".jpg", Tags: []string{"sky", "storm", "power"}},
		{ID: 6, Title: "Desert Oasis", Description: "A hidden oasis with palm trees in endless sand dunes", Extension: ".jpg", Tags: []string{"desert", "hope", "hidden"}},
		{ID: 7, Title: "Ocean Waves", Description: "Powerful waves crashing against rocky cliffs", Extension: ".jpg", Tags: []string{"ocean", "power", "nature"}},
		{ID: 8, Title: "Sunset Valley", Description: "Golden sunset painting a peaceful valley", Extension: ".jpg", Tags: []string{"peaceful", "day", "beauty"}},
		{ID: 9, Title: "Winter Forest", Description: "Snow-covered trees in silent woodland", Extension: ".jpg", Tags: []string{"winter", "forest", "peaceful"}},
		{ID: 10, Title: "Spring Meadow", Description: "Wildflowers blooming in a green meadow", Extension: ".jpg", Tags: []string{"spring", "garden", "renewal"}},
		{ID: 11, Title: "Desert Mirage", Description: "Shimmering heat waves creating illusions", Extension: ".jpg", Tags: []string{"desert", "mystery", "illusion"}},
		{ID: 12, Title: "Tropical Jungle", Description: "Dense rainforest with exotic plants", Extension: ".jpg", Tags: []string{"forest", "wild", "abundance"}},
		{ID: 13, Title: "Cave of Wonders", Description: "Mysterious cave with glowing crystals", Extension: ".jpg", Tags: []string{"mystery", "magic", "hidden"}},
		{ID: 14, Title: "Waterfall", Description: "Cascading water falling into a crystal pool", Extension: ".jpg", Tags: []string{"water", "power", "peaceful"}},
		{ID: 15, Title: "Endless Plains", Description: "Vast grasslands stretching to the horizon", Extension: ".jpg", Tags: []string{"freedom", "vast", "peaceful"}},
		{ID: 16, Title: "Lightning Storm", Description: "Electric bolts splitting the dark sky", Extension: ".jpg", Tags: []string{"storm", "power", "fear"}},
		{ID: 17, Title: "Rainbow Bridge", Description: "Colorful arc spanning across the clouds", Extension: ".jpg", Tags: []string{"hope", "beauty", "bridge"}},
		{ID: 18, Title: "Volcanic Peak", Description: "Molten lava flowing from an active volcano", Extension: ".jpg", Tags: []string{"fire", "power", "danger"}},
		{ID: 19, Title: "Ice Palace", Description: "Crystalline structures in a frozen landscape", Extension: ".jpg", Tags: []string{"winter", "magic", "castle"}},
		{ID: 20, Title: "Garden Maze", Description: "Intricate hedge maze with hidden paths", Extension: ".jpg", Tags: []string{"garden", "mystery", "puzzle"}},

		// Fantasy and magic cards (21-40)
		{ID: 21, Title: "Fire Dragon", Description: "Majestic dragon breathing flames", Extension: ".jpg", Tags: []string{"dragon", "fire", "power"}},
		{ID: 22, Title: "Enchanted Castle", Description: "Mystical fortress floating in the clouds", Extension: ".jpg", Tags: []string{"castle", "magic", "fantasy"}},
		{ID: 23, Title: "Wise Wizard", Description: "Ancient mage with a long beard and staff", Extension: ".jpg", Tags: []string{"wizard", "wisdom", "magic"}},
		{ID: 24, Title: "Unicorn", Description: "Pure white unicorn in a moonlit glade", Extension: ".jpg", Tags: []string{"magic", "purity", "rare"}},
		{ID: 25, Title: "Forest Fairy", Description: "Delicate fairy dancing among flowers", Extension: ".jpg", Tags: []string{"fairy", "garden", "magic"}},
		{ID: 26, Title: "Crystal Ball", Description: "Mystical orb showing swirling visions", Extension: ".jpg", Tags: []string{"magic", "mystery", "future"}},
		{ID: 27, Title: "Portal", Description: "Glowing doorway to another dimension", Extension: ".jpg", Tags: []string{"magic", "mystery", "travel"}},
		{ID: 28, Title: "Phoenix Rising", Description: "Mythical bird emerging from flames", Extension: ".jpg", Tags: []string{"fire", "rebirth", "magic"}},
		{ID: 29, Title: "Griffin", Description: "Majestic creature with eagle and lion parts", Extension: ".jpg", Tags: []string{"magic", "power", "flight"}},
		{ID: 30, Title: "Mermaid", Description: "Beautiful sea maiden swimming with dolphins", Extension: ".jpg", Tags: []string{"ocean", "magic", "beauty"}},
		{ID: 31, Title: "Knight", Description: "Brave warrior in shining armor", Extension: ".jpg", Tags: []string{"courage", "quest", "honor"}},
		{ID: 32, Title: "Elf", Description: "Graceful forest dweller with pointed ears", Extension: ".jpg", Tags: []string{"forest", "magic", "nature"}},
		{ID: 33, Title: "Spell Casting", Description: "Magical energy swirling around a caster", Extension: ".jpg", Tags: []string{"magic", "power", "mystery"}},
		{ID: 34, Title: "Treasure Chest", Description: "Overflowing chest of gold and jewels", Extension: ".jpg", Tags: []string{"treasure", "wealth", "discovery"}},
		{ID: 35, Title: "Magic Potion", Description: "Bubbling brew in ornate bottles", Extension: ".jpg", Tags: []string{"magic", "mystery", "transformation"}},
		{ID: 36, Title: "Epic Quest", Description: "Heroes setting out on an adventure", Extension: ".jpg", Tags: []string{"quest", "courage", "journey"}},
		{ID: 37, Title: "Shadow Demon", Description: "Dark entity lurking in the shadows", Extension: ".jpg", Tags: []string{"fear", "darkness", "magic"}},
		{ID: 38, Title: "Guardian Angel", Description: "Celestial being with glowing wings", Extension: ".jpg", Tags: []string{"protection", "light", "hope"}},
		{ID: 39, Title: "Excalibur", Description: "Legendary sword embedded in stone", Extension: ".jpg", Tags: []string{"power", "legend", "destiny"}},
		{ID: 40, Title: "Enchanted Forest", Description: "Magical woodland with glowing mushrooms", Extension: ".jpg", Tags: []string{"forest", "magic", "mystery"}},

		// Human emotions and activities (41-60)
		{ID: 41, Title: "Celebration", Description: "Joyful people dancing and celebrating", Extension: ".jpg", Tags: []string{"happy", "dance", "community"}},
		{ID: 42, Title: "Loving Embrace", Description: "Two figures in a tender embrace", Extension: ".jpg", Tags: []string{"love", "comfort", "connection"}},
		{ID: 43, Title: "Lonely Figure", Description: "Solitary person sitting in contemplation", Extension: ".jpg", Tags: []string{"sad", "solitude", "reflection"}},
		{ID: 44, Title: "Angry Storm", Description: "Person standing defiantly against wind", Extension: ".jpg", Tags: []string{"angry", "storm", "defiance"}},
		{ID: 45, Title: "Tears of Sorrow", Description: "Figure weeping under falling rain", Extension: ".jpg", Tags: []string{"sad", "grief", "rain"}},
		{ID: 46, Title: "Moment of Surprise", Description: "Person with wide eyes and open mouth", Extension: ".jpg", Tags: []string{"surprise", "shock", "discovery"}},
		{ID: 47, Title: "Deep Thought", Description: "Contemplative figure with hand on chin", Extension: ".jpg", Tags: []string{"thinking", "wisdom", "meditation"}},
		{ID: 48, Title: "Dancer", Description: "Graceful figure in mid-dance movement", Extension: ".jpg", Tags: []string{"dance", "grace", "expression"}},
		{ID: 49, Title: "Peaceful Sleep", Description: "Serene person sleeping under stars", Extension: ".jpg", Tags: []string{"peaceful", "night", "dreams"}},
		{ID: 50, Title: "Runner", Description: "Athletic figure racing toward the horizon", Extension: ".jpg", Tags: []string{"sport", "speed", "determination"}},
		{ID: 51, Title: "Mountain Climber", Description: "Adventurer scaling a steep cliff", Extension: ".jpg", Tags: []string{"mountain", "courage", "achievement"}},
		{ID: 52, Title: "Swimmer", Description: "Person diving into crystal clear water", Extension: ".jpg", Tags: []string{"water", "freedom", "sport"}},
		{ID: 53, Title: "Flying Dream", Description: "Figure soaring through cloudy skies", Extension: ".jpg", Tags: []string{"dream", "freedom", "flight"}},
		{ID: 54, Title: "Artist", Description: "Painter creating a masterpiece", Extension: ".jpg", Tags: []string{"art", "creativity", "expression"}},
		{ID: 55, Title: "Musician", Description: "Person playing a beautiful melody", Extension: ".jpg", Tags: []string{"music", "harmony", "emotion"}},
		{ID: 56, Title: "Reader", Description: "Figure absorbed in an ancient book", Extension: ".jpg", Tags: []string{"reading", "knowledge", "discovery"}},
		{ID: 57, Title: "Chef", Description: "Cook preparing a delicious feast", Extension: ".jpg", Tags: []string{"cooking", "creativity", "nourishment"}},
		{ID: 58, Title: "Gardener", Description: "Person tending to a beautiful garden", Extension: ".jpg", Tags: []string{"garden", "growth", "care"}},
		{ID: 59, Title: "Meditation", Description: "Figure in peaceful lotus position", Extension: ".jpg", Tags: []string{"peaceful", "balance", "inner"}},
		{ID: 60, Title: "Helper", Description: "Person extending a helping hand", Extension: ".jpg", Tags: []string{"kindness", "support", "community"}},

		// Objects and symbols (61-80)
		{ID: 61, Title: "Ancient Clock", Description: "Ornate timepiece with roman numerals", Extension: ".jpg", Tags: []string{"clock", "time", "ancient"}},
		{ID: 62, Title: "Magic Mirror", Description: "Ornate mirror showing mysterious reflections", Extension: ".jpg", Tags: []string{"mirror", "magic", "truth"}},
		{ID: 63, Title: "Golden Key", Description: "Elaborate key that unlocks secrets", Extension: ".jpg", Tags: []string{"key", "mystery", "access"}},
		{ID: 64, Title: "Ancient Tome", Description: "Leather-bound book of forgotten knowledge", Extension: ".jpg", Tags: []string{"book", "wisdom", "secrets"}},
		{ID: 65, Title: "Glowing Candle", Description: "Single candle illuminating the darkness", Extension: ".jpg", Tags: []string{"candle", "light", "hope"}},
		{ID: 66, Title: "Compass", Description: "Navigation tool pointing true north", Extension: ".jpg", Tags: []string{"direction", "journey", "guidance"}},
		{ID: 67, Title: "Telescope", Description: "Instrument for gazing at distant stars", Extension: ".jpg", Tags: []string{"discovery", "sky", "exploration"}},
		{ID: 68, Title: "Stone Bridge", Description: "Ancient bridge spanning a deep ravine", Extension: ".jpg", Tags: []string{"bridge", "connection", "journey"}},
		{ID: 69, Title: "Mysterious Door", Description: "Ornate doorway leading to unknown places", Extension: ".jpg", Tags: []string{"mystery", "opportunity", "unknown"}},
		{ID: 70, Title: "Spiral Staircase", Description: "Winding stairs ascending into light", Extension: ".jpg", Tags: []string{"ascension", "progress", "spiral"}},
		{ID: 71, Title: "Crystal Crown", Description: "Magnificent crown of pure crystal", Extension: ".jpg", Tags: []string{"crown", "power", "clarity"}},
		{ID: 72, Title: "Hourglass", Description: "Sand timer measuring precious moments", Extension: ".jpg", Tags: []string{"time", "patience", "flow"}},
		{ID: 73, Title: "Feather", Description: "Single white feather floating in air", Extension: ".jpg", Tags: []string{"lightness", "freedom", "purity"}},
		{ID: 74, Title: "Mask", Description: "Ornate carnival mask hiding identity", Extension: ".jpg", Tags: []string{"mystery", "identity", "hidden"}},
		{ID: 75, Title: "Scales", Description: "Balance scales weighing truth and lies", Extension: ".jpg", Tags: []string{"balance", "justice", "truth"}},
		{ID: 76, Title: "Heart", Description: "Glowing heart radiating warm light", Extension: ".jpg", Tags: []string{"love", "emotion", "warmth"}},
		{ID: 77, Title: "Chain", Description: "Golden chain connecting two realms", Extension: ".jpg", Tags: []string{"connection", "bond", "unity"}},
		{ID: 78, Title: "Balloon", Description: "Colorful balloon floating toward freedom", Extension: ".jpg", Tags: []string{"freedom", "joy", "flight"}},
		{ID: 79, Title: "Puzzle", Description: "Incomplete jigsaw with missing pieces", Extension: ".jpg", Tags: []string{"puzzle", "incomplete", "search"}},
		{ID: 80, Title: "Violin", Description: "Beautiful stringed instrument", Extension: ".jpg", Tags: []string{"music", "elegance", "emotion"}},

		// Abstract concepts (81-84)
		{ID: 81, Title: "Infinite Spiral", Description: "Never-ending spiral of light and energy", Extension: ".jpg", Tags: []string{"infinity", "mystery", "energy"}},
		{ID: 82, Title: "Perfect Balance", Description: "Yin and yang in harmonious unity", Extension: ".jpg", Tags: []string{"balance", "harmony", "unity"}},
		{ID: 83, Title: "Chaos Swirls", Description: "Chaotic patterns of color and form", Extension: ".jpg", Tags: []string{"chaos", "confusion", "complexity"}},
		{ID: 84, Title: "Metamorphosis", Description: "Transformation from caterpillar to butterfly", Extension: ".jpg", Tags: []string{"transformation", "growth", "change"}},
	}
}

// SeedDatabase populates the database with default tags and cards
func SeedDatabase() error {
	db := database.GetDB()
	log := logger.GetLogger()

	// Check if data already exists
	var tagCount, cardCount int64
	db.Model(&models.Tag{}).Count(&tagCount)
	db.Model(&models.Card{}).Count(&cardCount)

	if tagCount > 0 || cardCount > 0 {
		log.Info("Database already seeded", "tags", tagCount, "cards", cardCount)
		return nil
	}

	log.Info("Starting database seeding...")

	// Seed tags first
	tags := GetDefaultTags()
	tagMap := make(map[string]int) // tag slug to ID mapping

	for _, tagData := range tags {
		tag := models.Tag{
			Name:        tagData.Name,
			Slug:        tagData.Slug,
			Description: tagData.Description,
			Category:    tagData.Category,
			Color:       tagData.Color,
			Weight:      tagData.Weight,
			IsActive:    true,
		}

		if err := db.Create(&tag).Error; err != nil {
			return fmt.Errorf("failed to create tag %s: %w", tagData.Name, err)
		}

		tagMap[tagData.Slug] = tag.ID
		log.Debug("Created tag", "name", tag.Name, "id", tag.ID)
	}

	log.Info("Tags seeded successfully", "count", len(tags))

	// Seed cards
	cards := GetDefaultCards()
	minioClient := storage.GetClient()

	for _, cardData := range cards {
		// Create card
		card := models.Card{
			ID:          cardData.ID,
			Title:       cardData.Title,
			Description: cardData.Description,
			Extension:   cardData.Extension,
			IsActive:    true,
		}

		// Generate image URL
		if minioClient != nil {
			card.ImageURL = minioClient.GetCardImageURL(cardData.ID, cardData.Extension)
		} else {
			// Fallback to local static files
			card.ImageURL = fmt.Sprintf("/cards/%d%s", cardData.ID, cardData.Extension)
		}

		if err := db.Create(&card).Error; err != nil {
			return fmt.Errorf("failed to create card %d: %w", cardData.ID, err)
		}

		// Create card-tag relationships
		for _, tagSlug := range cardData.Tags {
			if tagID, exists := tagMap[tagSlug]; exists {
				cardTag := models.CardTag{
					CardID: card.ID,
					TagID:  tagID,
					Weight: 1.0, // Default weight
				}

				if err := db.Create(&cardTag).Error; err != nil {
					log.Error("Failed to create card-tag relationship", "card_id", card.ID, "tag_id", tagID, "error", err)
					// Continue with other relationships
				}
			} else {
				log.Warn("Tag not found for card", "card_id", card.ID, "tag_slug", tagSlug)
			}
		}

		log.Debug("Created card", "title", card.Title, "id", card.ID, "tags", len(cardData.Tags))
	}

	log.Info("Database seeding completed successfully",
		"tags_created", len(tags),
		"cards_created", len(cards))

	return nil
}

// SeedCardsOnly seeds only the cards (assumes tags already exist)
func SeedCardsOnly() error {
	db := database.GetDB()
	log := logger.GetLogger()

	// Check if cards already exist
	var cardCount int64
	db.Model(&models.Card{}).Count(&cardCount)

	if cardCount > 0 {
		log.Info("Cards already exist", "count", cardCount)
		return nil
	}

	// Get tag mapping
	var tags []models.Tag
	db.Find(&tags)
	tagMap := make(map[string]int)
	for _, tag := range tags {
		tagMap[tag.Slug] = tag.ID
	}

	if len(tagMap) == 0 {
		return fmt.Errorf("no tags found in database. Please seed tags first")
	}

	// Seed cards
	cards := GetDefaultCards()
	minioClient := storage.GetClient()

	for _, cardData := range cards {
		// Create card
		card := models.Card{
			ID:          cardData.ID,
			Title:       cardData.Title,
			Description: cardData.Description,
			Extension:   cardData.Extension,
			IsActive:    true,
		}

		// Generate image URL
		if minioClient != nil {
			card.ImageURL = minioClient.GetCardImageURL(cardData.ID, cardData.Extension)
		} else {
			card.ImageURL = fmt.Sprintf("/cards/%d%s", cardData.ID, cardData.Extension)
		}

		if err := db.Create(&card).Error; err != nil {
			return fmt.Errorf("failed to create card %d: %w", cardData.ID, err)
		}

		// Create card-tag relationships
		for _, tagSlug := range cardData.Tags {
			if tagID, exists := tagMap[tagSlug]; exists {
				cardTag := models.CardTag{
					CardID: card.ID,
					TagID:  tagID,
					Weight: 1.0,
				}
				db.Create(&cardTag)
			}
		}
	}

	log.Info("Cards seeded successfully", "count", len(cards))
	return nil
}

// SeedTagsOnly seeds only the tags
func SeedTagsOnly() error {
	db := database.GetDB()
	log := logger.GetLogger()

	// Check if tags already exist
	var tagCount int64
	db.Model(&models.Tag{}).Count(&tagCount)

	if tagCount > 0 {
		log.Info("Tags already exist", "count", tagCount)
		return nil
	}

	// Seed tags
	tags := GetDefaultTags()

	for _, tagData := range tags {
		tag := models.Tag{
			Name:        tagData.Name,
			Slug:        tagData.Slug,
			Description: tagData.Description,
			Category:    tagData.Category,
			Color:       tagData.Color,
			Weight:      tagData.Weight,
			IsActive:    true,
		}

		if err := db.Create(&tag).Error; err != nil {
			return fmt.Errorf("failed to create tag %s: %w", tagData.Name, err)
		}
	}

	log.Info("Tags seeded successfully", "count", len(tags))
	return nil
}
