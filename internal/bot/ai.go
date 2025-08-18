package bot

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"dixitme/internal/database"
	"dixitme/internal/logger"
	"dixitme/internal/models"

	"github.com/google/uuid"
)

// BotDifficulty represents different bot AI levels
type BotDifficulty string

const (
	BotEasy   BotDifficulty = "easy"
	BotMedium BotDifficulty = "medium"
	BotHard   BotDifficulty = "hard"
)

// BotPlayer represents an AI bot player
type BotPlayer struct {
	ID         uuid.UUID     `json:"id"`
	Name       string        `json:"name"`
	Difficulty BotDifficulty `json:"difficulty"`
	GameID     uuid.UUID     `json:"game_id"`
	Hand       []int         `json:"hand"` // Card IDs in bot's hand
}

// CardScore represents a card with its calculated score for selection
type CardScore struct {
	CardID int     `json:"card_id"`
	Score  float64 `json:"score"`
	Tags   []Tag   `json:"tags"`
}

// Tag represents a simple tag structure for bot logic
type Tag struct {
	Name     string  `json:"name"`
	Weight   float64 `json:"weight"`
	Category string  `json:"category"`
}

// BotManager manages all bot players
type BotManager struct {
	bots map[uuid.UUID]*BotPlayer
}

var botManager *BotManager

// GetBotManager returns the singleton bot manager
func GetBotManager() *BotManager {
	if botManager == nil {
		botManager = &BotManager{
			bots: make(map[uuid.UUID]*BotPlayer),
		}
	}
	return botManager
}

// CreateBot creates a new bot player
func (bm *BotManager) CreateBot(name string, difficulty BotDifficulty) *BotPlayer {
	bot := &BotPlayer{
		ID:         uuid.New(),
		Name:       name,
		Difficulty: difficulty,
		Hand:       make([]int, 0),
	}

	bm.bots[bot.ID] = bot
	logger.Info("Bot created", "bot_id", bot.ID, "name", name, "difficulty", difficulty)

	return bot
}

// GetBot returns a bot by ID
func (bm *BotManager) GetBot(botID uuid.UUID) *BotPlayer {
	return bm.bots[botID]
}

// SelectCardForClue selects the best card for a given clue using heuristic analysis
func (bp *BotPlayer) SelectCardForClue(clue string) (int, error) {
	if len(bp.Hand) == 0 {
		return 0, fmt.Errorf("bot has no cards in hand")
	}

	cardScores := make([]CardScore, 0, len(bp.Hand))

	// Score each card based on the clue
	for _, cardID := range bp.Hand {
		score := bp.calculateCardScore(cardID, clue)
		tags := bp.getCardTags(cardID)
		cardScores = append(cardScores, CardScore{
			CardID: cardID,
			Score:  score,
			Tags:   tags,
		})
	}

	// Select card based on difficulty level
	selectedCardID := bp.selectCardByDifficulty(cardScores)

	logger.Debug("Bot selected card for clue",
		"bot_id", bp.ID,
		"clue", clue,
		"card_id", selectedCardID,
		"difficulty", bp.Difficulty)

	return selectedCardID, nil
}

// SelectCardAsStoryteller selects a card when the bot is the storyteller
func (bp *BotPlayer) SelectCardAsStoryteller() (int, string, error) {
	if len(bp.Hand) == 0 {
		return 0, "", fmt.Errorf("bot has no cards in hand")
	}

	// For storyteller, we want to pick a card that's not too obvious or too obscure
	cardScores := make([]CardScore, 0, len(bp.Hand))

	for _, cardID := range bp.Hand {
		score := bp.calculateStorytellerScore(cardID)
		tags := bp.getCardTags(cardID)
		cardScores = append(cardScores, CardScore{
			CardID: cardID,
			Score:  score,
			Tags:   tags,
		})
	}

	selectedCardID := bp.selectCardByDifficulty(cardScores)
	clue := bp.generateClueForCard(selectedCardID)

	logger.Debug("Bot selected storyteller card",
		"bot_id", bp.ID,
		"card_id", selectedCardID,
		"clue", clue,
		"difficulty", bp.Difficulty)

	return selectedCardID, clue, nil
}

// VoteForCard selects which card to vote for
func (bp *BotPlayer) VoteForCard(submittedCardIDs []int, clue string, storytellerCardID int) (int, error) {
	if len(submittedCardIDs) == 0 {
		return 0, fmt.Errorf("no cards to vote for")
	}

	cardScores := make([]CardScore, 0, len(submittedCardIDs))

	// Score each submitted card based on how well it matches the clue
	for _, cardID := range submittedCardIDs {
		// Don't vote for our own card (find it in our hand)
		isOwnCard := false
		for _, handCardID := range bp.Hand {
			if cardID == handCardID {
				isOwnCard = true
				break
			}
		}
		if isOwnCard {
			continue
		}

		score := bp.calculateCardScore(cardID, clue)
		tags := bp.getCardTags(cardID)
		cardScores = append(cardScores, CardScore{
			CardID: cardID,
			Score:  score,
			Tags:   tags,
		})
	}

	if len(cardScores) == 0 {
		return 0, fmt.Errorf("no valid cards to vote for")
	}

	selectedCardID := bp.selectCardByDifficulty(cardScores)

	logger.Debug("Bot voted for card",
		"bot_id", bp.ID,
		"clue", clue,
		"voted_card_id", selectedCardID,
		"difficulty", bp.Difficulty)

	return selectedCardID, nil
}

// calculateCardScore calculates how well a card matches a clue using tag analysis
func (bp *BotPlayer) calculateCardScore(cardID int, clue string) float64 {
	score := 0.0

	// Get card tags
	tags := bp.getCardTags(cardID)

	// Parse clue into keywords
	clueWords := strings.Fields(strings.ToLower(clue))

	// Score based on tag matching
	for _, tag := range tags {
		tagName := strings.ToLower(tag.Name)
		tagWords := strings.Fields(tagName)

		// Direct tag name matches
		for _, clueWord := range clueWords {
			for _, tagWord := range tagWords {
				if clueWord == tagWord {
					score += 3.0 * tag.Weight
				} else if strings.Contains(tagWord, clueWord) || strings.Contains(clueWord, tagWord) {
					score += 1.5 * tag.Weight
				}
			}
		}

		// Category-based scoring
		switch tag.Category {
		case "emotion":
			if containsEmotionalWords(clueWords) {
				score += 2.0 * tag.Weight
			}
		case "nature":
			if containsNatureWords(clueWords) {
				score += 2.0 * tag.Weight
			}
		case "action":
			if containsActionWords(clueWords) {
				score += 2.0 * tag.Weight
			}
		}
	}

	// Add semantic scoring based on card description
	card := bp.getCardDetails(cardID)
	if card.Description != "" {
		descScore := bp.calculateSemanticScore(card.Description, clue)
		score += descScore
	}

	// Add randomness to prevent predictable behavior
	randomFactor := rand.Float64() * 0.5
	score += randomFactor

	return score
}

// calculateStorytellerScore calculates how good a card is for being a storyteller card
func (bp *BotPlayer) calculateStorytellerScore(cardID int) float64 {
	score := 0.0

	tags := bp.getCardTags(cardID)

	// Prefer cards with multiple diverse tags (more interpretable)
	score += float64(len(tags)) * 0.5

	// Category diversity bonus
	categories := make(map[string]bool)
	for _, tag := range tags {
		categories[tag.Category] = true
	}
	score += float64(len(categories)) * 0.3

	// Tag weight factor
	totalWeight := 0.0
	for _, tag := range tags {
		totalWeight += tag.Weight
	}
	if len(tags) > 0 {
		avgWeight := totalWeight / float64(len(tags))
		score += avgWeight
	}

	// Difficulty-based adjustments
	switch bp.Difficulty {
	case BotEasy:
		// Easy bots prefer more obvious cards (higher weighted tags)
		score += totalWeight * 0.2
	case BotMedium:
		// Medium bots prefer balanced cards
		if len(tags) >= 2 && len(tags) <= 4 {
			score += 1.0
		}
	case BotHard:
		// Hard bots prefer subtle cards with complex tag combinations
		if len(categories) > 2 {
			score += 2.0
		}
		if len(tags) > 3 {
			score += 1.0
		}
	}

	// Add randomness
	score += rand.Float64() * 1.5

	return score
}

// selectCardByDifficulty selects a card based on bot difficulty using weighted random
func (bp *BotPlayer) selectCardByDifficulty(cardScores []CardScore) int {
	if len(cardScores) == 0 {
		return 0
	}

	// Sort cards by score (highest first)
	for i := 0; i < len(cardScores)-1; i++ {
		for j := i + 1; j < len(cardScores); j++ {
			if cardScores[i].Score < cardScores[j].Score {
				cardScores[i], cardScores[j] = cardScores[j], cardScores[i]
			}
		}
	}

	switch bp.Difficulty {
	case BotEasy:
		// Easy: 70% chance best card, 20% second best, 10% random
		randVal := rand.Float64()
		if randVal < 0.7 {
			return cardScores[0].CardID
		} else if randVal < 0.9 && len(cardScores) > 1 {
			return cardScores[1].CardID
		} else {
			return cardScores[rand.Intn(len(cardScores))].CardID
		}

	case BotMedium:
		// Medium: 50% best, 30% second best, 20% weighted random
		randVal := rand.Float64()
		if randVal < 0.5 {
			return cardScores[0].CardID
		} else if randVal < 0.8 && len(cardScores) > 1 {
			return cardScores[1].CardID
		} else {
			return bp.weightedRandomSelection(cardScores)
		}

	case BotHard:
		// Hard: More strategic, uses weighted random based on scores
		return bp.weightedRandomSelection(cardScores)

	default:
		return cardScores[0].CardID
	}
}

// weightedRandomSelection selects a card using weighted random based on scores
func (bp *BotPlayer) weightedRandomSelection(cardScores []CardScore) int {
	if len(cardScores) == 0 {
		return 0
	}

	// Calculate total weight
	totalWeight := 0.0
	for _, cs := range cardScores {
		totalWeight += math.Max(cs.Score, 0.1) // Minimum weight of 0.1
	}

	// Random selection
	random := rand.Float64() * totalWeight
	currentWeight := 0.0

	for _, cs := range cardScores {
		currentWeight += math.Max(cs.Score, 0.1)
		if random <= currentWeight {
			return cs.CardID
		}
	}

	// Fallback to first card
	return cardScores[0].CardID
}

// generateClueForCard generates a clue for a given card
func (bp *BotPlayer) generateClueForCard(cardID int) string {
	tags := bp.getCardTags(cardID)

	if len(tags) == 0 {
		// Fallback generic clues
		clues := []string{"Mystery", "Adventure", "Dream", "Journey", "Magic"}
		return clues[rand.Intn(len(clues))]
	}

	// Select tag-based clue
	switch bp.Difficulty {
	case BotEasy:
		// Direct tag name or close variant
		selectedTag := tags[rand.Intn(len(tags))]
		return bp.directTagClue(selectedTag.Name)

	case BotMedium:
		// Abstract the tag name
		selectedTag := tags[rand.Intn(len(tags))]
		return bp.abstractTagClue(selectedTag.Name, selectedTag.Category)

	case BotHard:
		// Creative combination or metaphorical clue
		return bp.creativeClue(tags)

	default:
		selectedTag := tags[rand.Intn(len(tags))]
		return selectedTag.Name
	}
}

// Helper methods for clue generation
func (bp *BotPlayer) directTagClue(tagName string) string {
	// Return tag name or simple variant
	variants := map[string][]string{
		"happy":   {"Joy", "Cheerful", "Bright"},
		"sad":     {"Sorrow", "Melancholy", "Blue"},
		"nature":  {"Natural", "Wild", "Green"},
		"animal":  {"Creature", "Beast", "Living"},
		"fantasy": {"Magical", "Mystical", "Enchanted"},
		"water":   {"Liquid", "Flow", "Ocean"},
		"fire":    {"Flame", "Heat", "Burning"},
		"ancient": {"Old", "Historic", "Past"},
		"modern":  {"New", "Contemporary", "Future"},
	}

	if vars, exists := variants[strings.ToLower(tagName)]; exists {
		return vars[rand.Intn(len(vars))]
	}

	return tagName
}

func (bp *BotPlayer) abstractTagClue(tagName, category string) string {
	abstractions := map[string][]string{
		"emotion": {"Feeling", "Mood", "Spirit", "Heart"},
		"nature":  {"Organic", "Wild", "Pure", "Life"},
		"action":  {"Movement", "Energy", "Force", "Power"},
		"time":    {"Moment", "Era", "Forever", "Now"},
		"space":   {"Vast", "Distance", "Journey", "Path"},
	}

	if abs, exists := abstractions[category]; exists {
		return abs[rand.Intn(len(abs))]
	}

	return bp.directTagClue(tagName)
}

func (bp *BotPlayer) creativeClue(tags []Tag) string {
	// Combine multiple tags creatively
	if len(tags) >= 2 {
		creativeClues := []string{
			"Whispered Secrets", "Dancing Shadows", "Silent Thunder",
			"Frozen Fire", "Liquid Stone", "Flying Roots",
			"Backwards Tomorrow", "Invisible Light", "Heavy Air",
			"Sweet Sorrow", "Bright Darkness", "Quiet Storm",
		}
		return creativeClues[rand.Intn(len(creativeClues))]
	}

	// Single tag creative interpretation
	metaphors := []string{
		"Between Worlds", "Hidden Truth", "Lost Memory",
		"Distant Echo", "Forgotten Dream", "Silent Song",
		"Broken Circle", "Empty Fullness", "Gentle Chaos",
	}

	return metaphors[rand.Intn(len(metaphors))]
}

// Helper methods for tag and semantic analysis
func (bp *BotPlayer) getCardTags(cardID int) []Tag {
	db := database.GetDB()

	var cardTags []models.CardTag
	err := db.Preload("Tag").Where("card_id = ?", cardID).Find(&cardTags).Error
	if err != nil {
		logger.Error("Failed to get card tags", "error", err, "card_id", cardID)
		return []Tag{}
	}

	tags := make([]Tag, 0, len(cardTags))
	for _, ct := range cardTags {
		tags = append(tags, Tag{
			Name:     ct.Tag.Name,
			Weight:   ct.Tag.Weight * ct.Weight, // Combine tag weight and relation weight
			Category: ct.Tag.Category,
		})
	}

	return tags
}

func (bp *BotPlayer) getCardDetails(cardID int) models.Card {
	db := database.GetDB()

	var card models.Card
	err := db.First(&card, cardID).Error
	if err != nil {
		logger.Error("Failed to get card details", "error", err, "card_id", cardID)
		return models.Card{}
	}

	return card
}

func (bp *BotPlayer) calculateSemanticScore(description, clue string) float64 {
	descWords := strings.Fields(strings.ToLower(description))
	clueWords := strings.Fields(strings.ToLower(clue))

	score := 0.0
	for _, clueWord := range clueWords {
		for _, descWord := range descWords {
			if clueWord == descWord {
				score += 2.0
			} else if strings.Contains(descWord, clueWord) || strings.Contains(clueWord, descWord) {
				score += 1.0
			}
		}
	}

	return score
}

// Helper functions for semantic analysis
func containsEmotionalWords(words []string) bool {
	emotionalWords := []string{"happy", "sad", "angry", "fear", "joy", "love", "hate", "surprise", "disgust", "calm", "excited"}
	for _, word := range words {
		for _, emotional := range emotionalWords {
			if strings.Contains(word, emotional) {
				return true
			}
		}
	}
	return false
}

func containsNatureWords(words []string) bool {
	natureWords := []string{"tree", "flower", "mountain", "ocean", "forest", "river", "sky", "earth", "wind", "fire", "water", "green", "wild"}
	for _, word := range words {
		for _, nature := range natureWords {
			if strings.Contains(word, nature) {
				return true
			}
		}
	}
	return false
}

func containsActionWords(words []string) bool {
	actionWords := []string{"run", "jump", "fly", "dance", "fight", "move", "walk", "swim", "climb", "fall", "rise", "break", "build"}
	for _, word := range words {
		for _, action := range actionWords {
			if strings.Contains(word, action) {
				return true
			}
		}
	}
	return false
}

// UpdateHand updates the bot's hand
func (bp *BotPlayer) UpdateHand(cardIDs []int) {
	bp.Hand = cardIDs
	logger.Debug("Bot hand updated", "bot_id", bp.ID, "hand_size", len(cardIDs))
}

// SetGameID sets the game ID for the bot
func (bp *BotPlayer) SetGameID(gameID uuid.UUID) {
	bp.GameID = gameID
}

// GetBotNames returns a list of predefined bot names
func GetBotNames() []string {
	return []string{
		"Alice AI", "Bob Bot", "Charlie CPU", "Diana Digital",
		"Echo Engine", "Felix Algorithm", "Grace GPU", "Hugo Heuristic",
		"Iris Intelligence", "Jack Neural", "Kara Quantum", "Leo Logic",
		"Maya Machine", "Nova Network", "Oscar Optimizer", "Pixel AI",
		"Quinn Query", "Ruby Runtime", "Sam Synthetic", "Tera Tech",
	}
}

// Initialize sets up the bot system
func Initialize() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Initialize bot manager
	botManager = &BotManager{
		bots: make(map[uuid.UUID]*BotPlayer),
	}

	logger.Info("Bot system initialized")
}
