package game

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"dixitme/internal/bot"
	"dixitme/internal/database"
	"dixitme/internal/logger"
	"dixitme/internal/models"
	"dixitme/internal/redis"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

// CreateGame creates a new game with the given room code
func (m *Manager) CreateGame(roomCode string, creatorID uuid.UUID, creatorName string) (*GameState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if room code already exists
	if _, exists := m.games[roomCode]; exists {
		return nil, fmt.Errorf("room code already exists")
	}

	// Create new game state
	gameID := uuid.New()
	now := time.Now()
	
	// Initialize deck with all available cards (1-84 for standard Dixit)
	deck := make([]int, 84)
	for i := 0; i < 84; i++ {
		deck[i] = i + 1
	}
	// Shuffle the deck
	for i := len(deck) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		deck[i], deck[j] = deck[j], deck[i]
	}
	
	game := &GameState{
		ID:           gameID,
		RoomCode:     roomCode,
		Players:      make(map[uuid.UUID]*Player),
		Status:       models.GameStatusWaiting,
		RoundNumber:  0,
		MaxRounds:    999, // Will be determined by 30 points or empty deck
		Deck:         deck,
		UsedCards:    make([]int, 0),
		CreatedAt:    now,
		LastActivity: now,
	}

	// Add creator as first player
	creator := &Player{
		ID:          creatorID,
		Name:        creatorName,
		Score:       0,
		Position:    1,
		Hand:        make([]int, 0),
		IsConnected: true,
		IsActive:    true,
	}

	game.Players[creatorID] = creator

	// Store in memory
	m.games[roomCode] = game

	// Persist to database
	if err := m.persistGame(game); err != nil {
		delete(m.games, roomCode)
		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") &&
			strings.Contains(err.Error(), "games_room_code_key") {
			return nil, fmt.Errorf("room code '%s' is already taken, please try a different one", roomCode)
		}
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	// Store in Redis for scaling
	if err := m.storeGameInRedis(game); err != nil {
		logger.Error("Failed to store game in Redis", "error", err, "room_code", roomCode)
	}

	return game, nil
}

// JoinGame adds a player to an existing game
func (m *Manager) JoinGame(roomCode string, playerID uuid.UUID, playerName string) (*GameState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	game, exists := m.games[roomCode]
	if !exists {
		return nil, fmt.Errorf("game not found")
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	// Update activity
	game.LastActivity = time.Now()

	// Check if game is still accepting players
	if game.Status != models.GameStatusWaiting {
		return nil, fmt.Errorf("game already started")
	}

	// Check player limit (3-6 players for Dixit)
	if len(game.Players) >= 6 {
		return nil, fmt.Errorf("game is full")
	}

	// Check if player already in game
	if _, exists := game.Players[playerID]; exists {
		return nil, fmt.Errorf("player already in game")
	}

	// Add player
	player := &Player{
		ID:          playerID,
		Name:        playerName,
		Score:       0,
		Position:    len(game.Players) + 1,
		Hand:        make([]int, 0),
		IsConnected: true,
		IsActive:    true,
	}

	game.Players[playerID] = player

	// Game will end when a player reaches 30 points or deck is empty

	// Persist changes
	if err := m.persistGamePlayer(game.ID, player); err != nil {
		delete(game.Players, playerID)
		return nil, fmt.Errorf("failed to persist player: %w", err)
	}

	// Update Redis
	if err := m.storeGameInRedis(game); err != nil {
		logger.Error("Failed to update game in Redis", "error", err, "room_code", roomCode)
	}

	// Broadcast player joined
	m.BroadcastToGame(game, MessageTypePlayerJoined, PlayerJoinedPayload{Player: player})

	// Send system message
	m.SendSystemMessage(roomCode, fmt.Sprintf("%s joined the game", playerName))

	return game, nil
}

// AddBot adds a bot player to an existing game
func (m *Manager) AddBot(roomCode string, botLevel string) (*GameState, error) {
	m.mu.RLock()
	game, exists := m.games[roomCode]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("game not found")
	}

	game.Lock()
	defer game.Unlock()

	if game.Status != models.GameStatusWaiting {
		return nil, fmt.Errorf("cannot add bot to game in progress")
	}

	if len(game.Players) >= 6 {
		return nil, fmt.Errorf("game is full")
	}

	// Create bot player
	botNames := bot.GetBotNames()
	botName := botNames[rand.Intn(len(botNames))]

	// Ensure unique bot name
	for {
		nameExists := false
		for _, player := range game.Players {
			if player.Name == botName {
				nameExists = true
				break
			}
		}
		if !nameExists {
			break
		}
		botName = botNames[rand.Intn(len(botNames))]
	}

	botID := uuid.New()

	// Create bot in bot manager
	botManager := bot.GetBotManager()
	botPlayer := botManager.CreateBot(botName, bot.BotDifficulty(botLevel))
	botPlayer.SetGameID(game.ID)

	// Create game player
	player := &Player{
		ID:          botID,
		Name:        botName,
		Score:       0,
		Position:    len(game.Players),
		Hand:        make([]int, 0),
		Connection:  nil,
		IsConnected: true, // Bots are always "connected"
		IsActive:    true,
		IsBot:       true,
		BotLevel:    botLevel,
	}

	game.Players[botID] = player

	// Persist bot player to database
	dbPlayer := &models.Player{
		ID:       botID,
		Name:     botName,
		Type:     models.PlayerTypeBot,
		BotLevel: botLevel,
	}

	if err := database.GetDB().Create(dbPlayer).Error; err != nil {
		delete(game.Players, botID)
		return nil, fmt.Errorf("failed to persist bot player: %w", err)
	}

	if err := m.persistGamePlayer(game.ID, player); err != nil {
		delete(game.Players, botID)
		return nil, fmt.Errorf("failed to persist bot game player: %w", err)
	}

	// Update Redis
	if err := m.storeGameInRedis(game); err != nil {
		logger.Error("Failed to update game in Redis", "error", err, "room_code", roomCode)
	}

	// Broadcast bot joined
	m.BroadcastToGame(game, MessageTypePlayerJoined, PlayerJoinedPayload{Player: player})

	// Send system message
	m.SendSystemMessage(roomCode, fmt.Sprintf("Bot %s (%s difficulty) joined the game", botName, botLevel))

	logger.Info("Bot added to game", "bot_id", botID, "bot_name", botName, "bot_level", botLevel, "room_code", roomCode)

	return game, nil
}

// StartGame starts a game if conditions are met
func (m *Manager) StartGame(roomCode string, playerID uuid.UUID) error {
	m.mu.RLock()
	game, exists := m.games[roomCode]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("game not found")
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	// Update activity
	game.LastActivity = time.Now()

	// Check if player is in the game
	if _, exists := game.Players[playerID]; !exists {
		return fmt.Errorf("player not in game")
	}

	// Check if game can start (minimum 3 players)
	if len(game.Players) < 3 {
		return fmt.Errorf("need at least 3 players to start")
	}

	if game.Status != models.GameStatusWaiting {
		return fmt.Errorf("game already started")
	}

	// Initialize game
	game.Status = models.GameStatusInProgress

	// Deal cards to players
	m.dealCards(game)

	// Start first round
	if err := m.startNewRound(game); err != nil {
		return fmt.Errorf("failed to start first round: %w", err)
	}

	// Update database
	if err := m.updateGameStatus(game.ID, models.GameStatusInProgress); err != nil {
		return fmt.Errorf("failed to update game status: %w", err)
	}

	// Broadcast game started
	m.BroadcastToGame(game, MessageTypeGameStarted, GameStartedPayload{GameState: game})

	// Send system message
	m.SendSystemMessage(roomCode, "Game started! Let the storytelling begin!")

	return nil
}

// SubmitClue handles storyteller submitting a clue
func (m *Manager) SubmitClue(roomCode string, playerID uuid.UUID, clue string, cardID int) error {
	game := m.getGame(roomCode)
	if game == nil {
		return fmt.Errorf("game not found")
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	if game.CurrentRound == nil {
		return fmt.Errorf("no active round")
	}

	if game.CurrentRound.StorytellerID != playerID {
		return fmt.Errorf("only storyteller can submit clue")
	}

	if game.CurrentRound.Status != models.RoundStatusStorytelling {
		return fmt.Errorf("not in storytelling phase")
	}

	// Validate card is in player's hand
	player := game.Players[playerID]
	cardInHand := false
	for _, handCard := range player.Hand {
		if handCard == cardID {
			cardInHand = true
			break
		}
	}

	if !cardInHand {
		return fmt.Errorf("card not in player's hand")
	}

	// Set clue and storyteller card
	game.CurrentRound.Clue = clue
	game.CurrentRound.StorytellerCard = cardID
	game.CurrentRound.Status = models.RoundStatusSubmitting

	// Remove card from storyteller's hand and add to used cards
	for i, handCard := range player.Hand {
		if handCard == cardID {
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			game.UsedCards = append(game.UsedCards, cardID)
			break
		}
	}

	// Persist round update
	if err := m.updateRound(game.CurrentRound); err != nil {
		return fmt.Errorf("failed to update round: %w", err)
	}

	// Broadcast clue submitted
	m.BroadcastToGame(game, MessageTypeClueSubmitted, ClueSubmittedPayload{Clue: clue})

	return nil
}

// SubmitCard handles non-storyteller players submitting cards
func (m *Manager) SubmitCard(roomCode string, playerID uuid.UUID, cardID int) error {
	game := m.getGame(roomCode)
	if game == nil {
		return fmt.Errorf("game not found")
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	if game.CurrentRound == nil {
		return fmt.Errorf("no active round")
	}

	if game.CurrentRound.StorytellerID == playerID {
		return fmt.Errorf("storyteller cannot submit cards")
	}

	if game.CurrentRound.Status != models.RoundStatusSubmitting {
		return fmt.Errorf("not in card submission phase")
	}

	// Check if player already submitted
	if _, exists := game.CurrentRound.Submissions[playerID]; exists {
		return fmt.Errorf("card already submitted")
	}

	// Validate card is in player's hand
	player := game.Players[playerID]
	cardInHand := false
	for _, handCard := range player.Hand {
		if handCard == cardID {
			cardInHand = true
			break
		}
	}

	if !cardInHand {
		return fmt.Errorf("card not in player's hand")
	}

	// Add submission
	game.CurrentRound.Submissions[playerID] = &CardSubmission{
		PlayerID: playerID,
		CardID:   cardID,
	}

	// Remove card from player's hand and add to used cards
	for i, handCard := range player.Hand {
		if handCard == cardID {
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			game.UsedCards = append(game.UsedCards, cardID)
			break
		}
	}

	// Persist submission
	if err := m.persistCardSubmission(game.CurrentRound.ID, playerID, cardID); err != nil {
		return fmt.Errorf("failed to persist submission: %w", err)
	}

	// Check if all players submitted
	expectedSubmissions := len(game.Players) - 1 // Exclude storyteller
	if len(game.CurrentRound.Submissions) == expectedSubmissions {
		m.startVotingPhase(game)
	}

	// Broadcast card submitted
	m.BroadcastToGame(game, MessageTypeCardSubmitted, CardSubmittedPayload{PlayerID: playerID})

	return nil
}

// SubmitVote handles player voting
func (m *Manager) SubmitVote(roomCode string, playerID uuid.UUID, cardID int) error {
	game := m.getGame(roomCode)
	if game == nil {
		return fmt.Errorf("game not found")
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	if game.CurrentRound == nil {
		return fmt.Errorf("no active round")
	}

	if game.CurrentRound.StorytellerID == playerID {
		return fmt.Errorf("storyteller cannot vote")
	}

	if game.CurrentRound.Status != models.RoundStatusVoting {
		return fmt.Errorf("not in voting phase")
	}

	// Check if player already voted
	if _, exists := game.CurrentRound.Votes[playerID]; exists {
		return fmt.Errorf("already voted")
	}

	// Validate card is among revealed cards
	validCard := false
	for _, revealedCard := range game.CurrentRound.RevealedCards {
		if revealedCard.CardID == cardID {
			validCard = true
			break
		}
	}

	if !validCard {
		return fmt.Errorf("invalid card selection")
	}

	// Add vote
	game.CurrentRound.Votes[playerID] = &Vote{
		PlayerID: playerID,
		CardID:   cardID,
	}

	// Persist vote
	if err := m.persistVote(game.CurrentRound.ID, playerID, cardID); err != nil {
		return fmt.Errorf("failed to persist vote: %w", err)
	}

	// Check if all players voted
	expectedVotes := len(game.Players) - 1 // Exclude storyteller
	if len(game.CurrentRound.Votes) == expectedVotes {
		m.completeRound(game)
	}

	// Broadcast vote submitted
	m.BroadcastToGame(game, MessageTypeVoteSubmitted, VoteSubmittedPayload{PlayerID: playerID})

	return nil
}

// Helper methods

func (m *Manager) GetGame(roomCode string) *GameState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.games[roomCode]
}

func (m *Manager) getGame(roomCode string) *GameState {
	return m.GetGame(roomCode)
}

func (m *Manager) GetActiveGamesCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.games)
}

func (m *Manager) dealCards(game *GameState) {
	// Deal 6 cards to each player from the deck
	for _, player := range game.Players {
		player.Hand = make([]int, 0, 6) // Each player gets 6 cards
		for i := 0; i < 6 && len(game.Deck) > 0; i++ {
			// Take card from top of deck
			cardID := game.Deck[0]
			game.Deck = game.Deck[1:]
			player.Hand = append(player.Hand, cardID)
		}
	}
}

// drawCard draws one card from the deck for a player
func (m *Manager) drawCard(game *GameState, player *Player) bool {
	if len(game.Deck) == 0 {
		return false // No cards left to draw
	}
	
	// Take card from top of deck
	cardID := game.Deck[0]
	game.Deck = game.Deck[1:]
	player.Hand = append(player.Hand, cardID)
	return true
}

// refillHands draws cards for all players back to 6 cards
func (m *Manager) refillHands(game *GameState) {
	for _, player := range game.Players {
		for len(player.Hand) < 6 && len(game.Deck) > 0 {
			m.drawCard(game, player)
		}
	}
}

func (m *Manager) startNewRound(game *GameState) error {
	game.RoundNumber++

	// Determine storyteller (rotate)
	storytellerPosition := ((game.RoundNumber - 1) % len(game.Players)) + 1
	var storytellerID uuid.UUID
	for _, player := range game.Players {
		if player.Position == storytellerPosition {
			storytellerID = player.ID
			break
		}
	}

	// Create new round
	round := &Round{
		ID:            uuid.New(),
		RoundNumber:   game.RoundNumber,
		StorytellerID: storytellerID,
		Status:        models.RoundStatusStorytelling,
		Submissions:   make(map[uuid.UUID]*CardSubmission),
		Votes:         make(map[uuid.UUID]*Vote),
		CreatedAt:     time.Now(),
	}

	game.CurrentRound = round

	// Persist round
	if err := m.persistRound(game.ID, round); err != nil {
		return err
	}

	// Broadcast round started
	m.BroadcastToGame(game, MessageTypeRoundStarted, RoundStartedPayload{Round: round})

	return nil
}

func (m *Manager) startVotingPhase(game *GameState) {
	game.CurrentRound.Status = models.RoundStatusVoting

	// Prepare revealed cards (shuffle them)
	var revealedCards []RevealedCard

	// Add storyteller's card
	revealedCards = append(revealedCards, RevealedCard{
		CardID:   game.CurrentRound.StorytellerCard,
		PlayerID: game.CurrentRound.StorytellerID,
	})

	// Add submitted cards
	for playerID, submission := range game.CurrentRound.Submissions {
		revealedCards = append(revealedCards, RevealedCard{
			CardID:   submission.CardID,
			PlayerID: playerID,
		})
	}

	// Shuffle the cards
	rand.Shuffle(len(revealedCards), func(i, j int) {
		revealedCards[i], revealedCards[j] = revealedCards[j], revealedCards[i]
	})

	game.CurrentRound.RevealedCards = revealedCards

	// Broadcast voting started
	m.BroadcastToGame(game, MessageTypeVotingStarted, VotingStartedPayload{
		RevealedCards: revealedCards,
	})
}

func (m *Manager) completeRound(game *GameState) {
	game.CurrentRound.Status = models.RoundStatusScoring

	// Calculate scores
	scores := m.calculateScores(game)

	// Update player scores
	for playerID, additionalScore := range scores {
		if player, exists := game.Players[playerID]; exists {
			player.Score += additionalScore
		}
	}

	// Update vote counts for revealed cards
	for i := range game.CurrentRound.RevealedCards {
		card := &game.CurrentRound.RevealedCards[i]
		for _, vote := range game.CurrentRound.Votes {
			if vote.CardID == card.CardID {
				card.VoteCount++
			}
		}
	}

	// Broadcast round completed
	m.BroadcastToGame(game, MessageTypeRoundCompleted, RoundCompletedPayload{
		Scores:        scores,
		RevealedCards: game.CurrentRound.RevealedCards,
	})

	// Check if game should end according to Dixit rules:
	// 1. Any player reaches 30 points
	// 2. Deck is empty (no more cards to draw)
	shouldEnd := false
	var endReason string
	
	// Check for 30 points
	for _, player := range game.Players {
		if player.Score >= 30 {
			shouldEnd = true
			endReason = fmt.Sprintf("Game ended: %s reached 30 points!", player.Name)
			break
		}
	}
	
	// Check if deck is empty (can't refill hands)
	if !shouldEnd {
		// Try to refill hands - if any player can't get cards, game ends
		initialDeckSize := len(game.Deck)
		m.refillHands(game)
		
		// If deck is empty and any player has less than 6 cards, game ends
		if len(game.Deck) == 0 {
			for _, player := range game.Players {
				if len(player.Hand) < 6 {
					shouldEnd = true
					endReason = "Game ended: No more cards in deck!"
					break
				}
			}
		}
		
		// Log deck status
		if initialDeckSize != len(game.Deck) {
			logger.Info("Cards drawn after round", 
				"room_code", game.RoomCode,
				"round", game.RoundNumber,
				"cards_drawn", initialDeckSize-len(game.Deck),
				"cards_remaining", len(game.Deck))
		}
	}
	
	if shouldEnd {
		// Send end reason message
		m.SendSystemMessage(game.RoomCode, endReason)
		m.completeGame(game)
	} else {
		// Start next round after a delay
		go func() {
			time.Sleep(5 * time.Second)
			m.startNewRound(game)
		}()
	}
}

func (m *Manager) calculateScores(game *GameState) map[uuid.UUID]int {
	scores := make(map[uuid.UUID]int)
	storytellerID := game.CurrentRound.StorytellerID

	// Count votes for storyteller's card
	storytellerVotes := 0
	for _, vote := range game.CurrentRound.Votes {
		if vote.CardID == game.CurrentRound.StorytellerCard {
			storytellerVotes++
		}
	}

	totalVoters := len(game.Players) - 1 // Exclude storyteller

	// Scoring rules for Dixit
	if storytellerVotes == 0 || storytellerVotes == totalVoters {
		// All or none guessed correctly - storyteller gets 0, others get 2
		scores[storytellerID] = 0
		for playerID := range game.Players {
			if playerID != storytellerID {
				scores[playerID] = 2
			}
		}
	} else {
		// Some guessed correctly - storyteller gets 3, correct guessers get 3
		scores[storytellerID] = 3
		for _, vote := range game.CurrentRound.Votes {
			if vote.CardID == game.CurrentRound.StorytellerCard {
				scores[vote.PlayerID] = 3
			}
		}
	}

	// Additional points for votes received on your card (except storyteller's card)
	for _, vote := range game.CurrentRound.Votes {
		if vote.CardID != game.CurrentRound.StorytellerCard {
			// Find who submitted this card
			for playerID, submission := range game.CurrentRound.Submissions {
				if submission.CardID == vote.CardID {
					if _, exists := scores[playerID]; !exists {
						scores[playerID] = 0
					}
					scores[playerID] += 1
					break
				}
			}
		}
	}

	return scores
}

func (m *Manager) completeGame(game *GameState) {
	game.Status = models.GameStatusCompleted

	// Find winner (highest score)
	var winnerID uuid.UUID
	var winnerName string
	highestScore := -1
	finalScores := make(map[uuid.UUID]int)

	for playerID, player := range game.Players {
		finalScores[playerID] = player.Score
		if player.Score > highestScore {
			highestScore = player.Score
			winnerID = playerID
			winnerName = player.Name
		}
	}
	
	// Log game completion stats
	logger.Info("Game completed",
		"room_code", game.RoomCode,
		"rounds_played", game.RoundNumber,
		"winner", winnerName,
		"winning_score", highestScore,
		"cards_remaining", len(game.Deck),
		"cards_used", len(game.UsedCards))

	// Persist game completion
	if err := m.persistGameCompletion(game.ID, winnerID); err != nil {
		logger.Error("Failed to persist game completion", "error", err, "game_id", game.ID, "winner_id", winnerID)
	}

	// Broadcast game completed
	m.BroadcastToGame(game, MessageTypeGameCompleted, GameCompletedPayload{
		FinalScores: finalScores,
		Winner:      winnerID,
	})

	// Clean up game from memory after some time
	go func() {
		time.Sleep(10 * time.Minute)
		m.mu.Lock()
		delete(m.games, game.RoomCode)
		m.mu.Unlock()
	}()
}

func (m *Manager) BroadcastToGame(game *GameState, messageType MessageType, payload interface{}) {
	message := GameMessage{
		Type:    messageType,
		Payload: payload,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		logger.Error("Failed to marshal message", "error", err, "message_type", messageType)
		return
	}

	for _, player := range game.Players {
		if player.Connection != nil && player.IsConnected {
			if err := player.Connection.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
				logger.Error("Failed to send message to player", "error", err, "player_id", player.ID, "message_type", messageType)
				player.IsConnected = false
			}
		}
	}
}

// Database persistence methods (simplified)
func (m *Manager) persistGame(game *GameState) error {
	db := database.GetDB()
	dbGame := &models.Game{
		ID:           game.ID,
		RoomCode:     game.RoomCode,
		Status:       game.Status,
		CurrentRound: game.RoundNumber,
		MaxRounds:    game.MaxRounds,
		CreatedAt:    game.CreatedAt,
	}
	return db.Create(dbGame).Error
}

func (m *Manager) persistGamePlayer(gameID uuid.UUID, player *Player) error {
	db := database.GetDB()

	// First ensure player exists
	dbPlayer := &models.Player{
		ID:   player.ID,
		Name: player.Name,
	}
	db.FirstOrCreate(dbPlayer, models.Player{ID: player.ID})

	// Then create game player relationship
	gamePlayer := &models.GamePlayer{
		GameID:   gameID,
		PlayerID: player.ID,
		Score:    player.Score,
		Position: player.Position,
		IsActive: player.IsActive,
	}
	return db.Create(gamePlayer).Error
}

func (m *Manager) updateGameStatus(gameID uuid.UUID, status models.GameStatus) error {
	db := database.GetDB()
	return db.Model(&models.Game{}).Where("id = ?", gameID).Update("status", status).Error
}

func (m *Manager) persistRound(gameID uuid.UUID, round *Round) error {
	db := database.GetDB()
	dbRound := &models.GameRound{
		ID:            round.ID,
		GameID:        gameID,
		RoundNumber:   round.RoundNumber,
		StorytellerID: round.StorytellerID,
		Status:        round.Status,
		CreatedAt:     round.CreatedAt,
	}
	return db.Create(dbRound).Error
}

func (m *Manager) updateRound(round *Round) error {
	db := database.GetDB()
	return db.Model(&models.GameRound{}).Where("id = ?", round.ID).Updates(map[string]interface{}{
		"clue":             round.Clue,
		"status":           round.Status,
		"storyteller_card": round.StorytellerCard,
	}).Error
}

func (m *Manager) persistCardSubmission(roundID, playerID uuid.UUID, cardID int) error {
	db := database.GetDB()
	submission := &models.CardSubmission{
		RoundID:  roundID,
		PlayerID: playerID,
		CardID:   cardID,
	}
	return db.Create(submission).Error
}

func (m *Manager) persistVote(roundID, playerID uuid.UUID, cardID int) error {
	db := database.GetDB()
	vote := &models.Vote{
		RoundID:  roundID,
		PlayerID: playerID,
		CardID:   cardID,
	}
	return db.Create(vote).Error
}

func (m *Manager) persistGameCompletion(gameID, winnerID uuid.UUID) error {
	db := database.GetDB()

	// Update game status
	if err := db.Model(&models.Game{}).Where("id = ?", gameID).Update("status", models.GameStatusCompleted).Error; err != nil {
		return err
	}

	// Create game history
	history := &models.GameHistory{
		GameID:   gameID,
		WinnerID: winnerID,
		// Duration and TotalRounds would be calculated here
	}
	return db.Create(history).Error
}

// Redis methods for scaling
func (m *Manager) storeGameInRedis(game *GameState) error {
	client := redis.GetClient()
	ctx := context.Background()

	gameJSON, err := json.Marshal(game)
	if err != nil {
		return err
	}

	return client.Set(ctx, "game:"+game.RoomCode, gameJSON, time.Hour).Err()
}

// Bot automation methods

// ProcessBotActions handles automated bot actions for the current game phase
func (m *Manager) ProcessBotActions(game *GameState) {
	if game.CurrentRound == nil {
		return
	}

	switch game.CurrentRound.Status {
	case models.RoundStatusStorytelling:
		m.processBotStorytelling(game)
	case models.RoundStatusSubmitting:
		m.processBotSubmissions(game)
	case models.RoundStatusVoting:
		m.processBotVoting(game)
	}
}

// processBotStorytelling handles bot storytelling
func (m *Manager) processBotStorytelling(game *GameState) {
	storytellerID := game.CurrentRound.StorytellerID
	storyteller, exists := game.Players[storytellerID]

	if !exists || !storyteller.IsBot {
		return
	}

	// Bot storyteller submits clue and card
	go func() {
		// Add small delay for realism
		time.Sleep(time.Duration(2+rand.Intn(3)) * time.Second)

		botManager := bot.GetBotManager()
		botPlayer := botManager.GetBot(storytellerID)
		if botPlayer == nil {
			logger.Error("Bot player not found", "bot_id", storytellerID)
			return
		}

		// Update bot's hand
		botPlayer.UpdateHand(storyteller.Hand)

		// Bot selects card and generates clue
		selectedCard, clue, err := botPlayer.SelectCardAsStoryteller()
		if err != nil {
			logger.Error("Bot failed to select storyteller card", "error", err, "bot_id", storytellerID)
			return
		}

		// Submit clue and card
		err = m.SubmitClue(game.RoomCode, storytellerID, clue, selectedCard)
		if err != nil {
			logger.Error("Bot failed to submit clue", "error", err, "bot_id", storytellerID)
		}
	}()
}

// processBotSubmissions handles bot card submissions
func (m *Manager) processBotSubmissions(game *GameState) {
	for playerID, player := range game.Players {
		// Skip non-bots, storyteller, and players who already submitted
		if !player.IsBot || playerID == game.CurrentRound.StorytellerID {
			continue
		}
		if _, hasSubmitted := game.CurrentRound.Submissions[playerID]; hasSubmitted {
			continue
		}

		go func(botID uuid.UUID, botPlayer *Player) {
			// Add random delay for realism
			time.Sleep(time.Duration(3+rand.Intn(5)) * time.Second)

			botManager := bot.GetBotManager()
			bot := botManager.GetBot(botID)
			if bot == nil {
				logger.Error("Bot player not found", "bot_id", botID)
				return
			}

			// Update bot's hand
			bot.UpdateHand(botPlayer.Hand)

			// Bot selects card for clue
			selectedCard, err := bot.SelectCardForClue(game.CurrentRound.Clue)
			if err != nil {
				logger.Error("Bot failed to select card for clue", "error", err, "bot_id", botID)
				return
			}

			// Submit card
			err = m.SubmitCard(game.RoomCode, botID, selectedCard)
			if err != nil {
				logger.Error("Bot failed to submit card", "error", err, "bot_id", botID)
			}
		}(playerID, player)
	}
}

// processBotVoting handles bot voting
func (m *Manager) processBotVoting(game *GameState) {
	for playerID, player := range game.Players {
		// Skip non-bots, storyteller, and players who already voted
		if !player.IsBot || playerID == game.CurrentRound.StorytellerID {
			continue
		}
		if _, hasVoted := game.CurrentRound.Votes[playerID]; hasVoted {
			continue
		}

		go func(botID uuid.UUID, botPlayer *Player) {
			// Add random delay for realism
			time.Sleep(time.Duration(2+rand.Intn(4)) * time.Second)

			botManager := bot.GetBotManager()
			bot := botManager.GetBot(botID)
			if bot == nil {
				logger.Error("Bot player not found", "bot_id", botID)
				return
			}

			// Get submitted cards for voting
			submittedCards := make([]int, 0, len(game.CurrentRound.RevealedCards))
			for _, revealedCard := range game.CurrentRound.RevealedCards {
				submittedCards = append(submittedCards, revealedCard.CardID)
			}

			// Bot votes for card
			selectedCard, err := bot.VoteForCard(submittedCards, game.CurrentRound.Clue, game.CurrentRound.StorytellerCard)
			if err != nil {
				logger.Error("Bot failed to vote for card", "error", err, "bot_id", botID)
				return
			}

			// Submit vote
			err = m.SubmitVote(game.RoomCode, botID, selectedCard)
			if err != nil {
				logger.Error("Bot failed to submit vote", "error", err, "bot_id", botID)
			}
		}(playerID, player)
	}
}

// Chat functionality

// SendChatMessage handles sending chat messages in a game
func (m *Manager) SendChatMessage(roomCode string, playerID uuid.UUID, message string, messageType string) error {
	game := m.getGame(roomCode)
	if game == nil {
		return fmt.Errorf("game not found")
	}

	player, exists := game.Players[playerID]
	if !exists {
		return fmt.Errorf("player not in game")
	}

	// Validate message type
	if messageType == "" {
		messageType = "chat"
	}
	if messageType != "chat" && messageType != "emote" {
		return fmt.Errorf("invalid message type")
	}

	// Validate message content
	if len(strings.TrimSpace(message)) == 0 {
		return fmt.Errorf("message cannot be empty")
	}
	if len(message) > 500 { // Max message length
		return fmt.Errorf("message too long")
	}

	// Determine current phase
	currentPhase := "lobby"
	if game.Status == models.GameStatusInProgress && game.CurrentRound != nil {
		currentPhase = string(game.CurrentRound.Status)
	}

	// Only allow chat in lobby and voting phases
	if currentPhase != "lobby" && currentPhase != "voting" {
		return fmt.Errorf("chat not allowed in current phase")
	}

	// Create chat message
	chatMessage := models.ChatMessage{
		ID:          uuid.New(),
		GameID:      game.ID,
		PlayerID:    playerID,
		Message:     strings.TrimSpace(message),
		MessageType: messageType,
		Phase:       currentPhase,
		IsVisible:   true,
		CreatedAt:   time.Now(),
	}

	// Persist to database
	if err := m.persistChatMessage(&chatMessage); err != nil {
		return fmt.Errorf("failed to persist chat message: %w", err)
	}

	// Create payload
	payload := ChatMessagePayload{
		ID:          chatMessage.ID,
		PlayerID:    playerID,
		PlayerName:  player.Name,
		Message:     chatMessage.Message,
		MessageType: chatMessage.MessageType,
		Phase:       chatMessage.Phase,
		Timestamp:   chatMessage.CreatedAt,
	}

	// Broadcast to all players in the game
	m.BroadcastToGame(game, MessageTypeChatMessage, payload)

	return nil
}

// GetChatHistory retrieves chat messages for a game and phase
func (m *Manager) GetChatHistory(roomCode string, phase string, limit int) ([]ChatMessagePayload, error) {
	game := m.getGame(roomCode)
	if game == nil {
		return nil, fmt.Errorf("game not found")
	}

	if limit <= 0 || limit > 100 {
		limit = 50 // Default limit
	}

	// Get messages from database
	messages, err := m.getChatMessages(game.ID, phase, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat messages: %w", err)
	}

	// Convert to payload format
	payloads := make([]ChatMessagePayload, 0, len(messages))
	for _, msg := range messages {
		// Get player name
		playerName := "Unknown"
		if player, exists := game.Players[msg.PlayerID]; exists {
			playerName = player.Name
		} else {
			// Fallback: get from database
			var dbPlayer models.Player
			if err := database.GetDB().First(&dbPlayer, "id = ?", msg.PlayerID).Error; err == nil {
				playerName = dbPlayer.Name
			}
		}

		payloads = append(payloads, ChatMessagePayload{
			ID:          msg.ID,
			PlayerID:    msg.PlayerID,
			PlayerName:  playerName,
			Message:     msg.Message,
			MessageType: msg.MessageType,
			Phase:       msg.Phase,
			Timestamp:   msg.CreatedAt,
		})
	}

	return payloads, nil
}

// SendSystemMessage sends a system message (e.g., "Player joined", "Round started")
func (m *Manager) SendSystemMessage(roomCode string, message string) error {
	game := m.getGame(roomCode)
	if game == nil {
		return fmt.Errorf("game not found")
	}

	// Determine current phase
	currentPhase := "lobby"
	if game.Status == models.GameStatusInProgress && game.CurrentRound != nil {
		currentPhase = string(game.CurrentRound.Status)
	}

	// Create system message with a system player ID (using nil UUID)
	systemPlayerID := uuid.Nil
	chatMessage := models.ChatMessage{
		ID:          uuid.New(),
		GameID:      game.ID,
		PlayerID:    systemPlayerID,
		Message:     message,
		MessageType: "system",
		Phase:       currentPhase,
		IsVisible:   true,
		CreatedAt:   time.Now(),
	}

	// Persist to database
	if err := m.persistChatMessage(&chatMessage); err != nil {
		logger.Error("Failed to persist system message", "error", err)
		// Continue anyway - system messages are not critical
	}

	// Create payload
	payload := ChatMessagePayload{
		ID:          chatMessage.ID,
		PlayerID:    systemPlayerID,
		PlayerName:  "System",
		Message:     chatMessage.Message,
		MessageType: chatMessage.MessageType,
		Phase:       chatMessage.Phase,
		Timestamp:   chatMessage.CreatedAt,
	}

	// Broadcast to all players in the game
	m.BroadcastToGame(game, MessageTypeChatMessage, payload)

	return nil
}

// Database persistence methods

func (m *Manager) persistChatMessage(chatMessage *models.ChatMessage) error {
	db := database.GetDB()
	return db.Create(chatMessage).Error
}

func (m *Manager) getChatMessages(gameID uuid.UUID, phase string, limit int) ([]models.ChatMessage, error) {
	db := database.GetDB()
	var messages []models.ChatMessage

	query := db.Where("game_id = ? AND is_visible = ?", gameID, true)

	if phase != "" && phase != "all" {
		query = query.Where("phase = ?", phase)
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// Reverse to get chronological order
	for i := 0; i < len(messages)/2; i++ {
		j := len(messages) - 1 - i
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// startCleanupService runs a background goroutine to clean up inactive games
func (m *Manager) startCleanupService() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()

	logger.Info("Game cleanup service started",
		"check_interval", m.cleanupInterval,
		"empty_room_timeout", "10m",
		"occupied_room_timeout", "30m")

	for {
		select {
		case <-ticker.C:
			m.cleanupInactiveGames()
		case <-m.stopCleanup:
			logger.Info("Game cleanup service stopped")
			return
		}
	}
}

// cleanupInactiveGames removes games that have been inactive for too long
func (m *Manager) cleanupInactiveGames() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var toRemove []string
	var emptyRooms []string
	var occupiedRooms []string

	for roomCode, game := range m.games {
		game.mu.RLock()

		// Count active/connected players
		activePlayerCount := 0
		for _, player := range game.Players {
			if player.IsConnected {
				activePlayerCount++
			}
		}

		emptyRoomTimeout := 10 * time.Minute    // Empty rooms: 10 minutes
		occupiedRoomTimeout := 30 * time.Minute // Rooms with players: 30 minutes

		var shouldRemove bool
		var reason string

		if activePlayerCount == 0 {
			// Empty room - use shorter timeout
			shouldRemove = time.Since(game.LastActivity) > emptyRoomTimeout
			if shouldRemove {
				emptyRooms = append(emptyRooms, roomCode)
				reason = "Game closed - empty room (10 minutes)"
			}
		} else {
			// Room has players - use longer timeout
			shouldRemove = time.Since(game.LastActivity) > occupiedRoomTimeout
			if shouldRemove {
				occupiedRooms = append(occupiedRooms, roomCode)
				reason = "Game closed due to inactivity (30 minutes)"
			}
		}

		if shouldRemove {
			toRemove = append(toRemove, roomCode)
			// Store reason for later use
			game.mu.RUnlock()

			// Notify all connected players that the game is being closed
			m.broadcastGameClosure(game, reason)

			// Mark game as abandoned in database
			m.markGameAsAbandoned(game)
		} else {
			game.mu.RUnlock()
		}
	}

	if len(toRemove) > 0 {
		logger.Info("Cleaning up inactive games",
			"total_count", len(toRemove),
			"empty_rooms", len(emptyRooms),
			"occupied_rooms", len(occupiedRooms),
			"empty_room_codes", emptyRooms,
			"occupied_room_codes", occupiedRooms)

		// Remove from memory
		for _, roomCode := range toRemove {
			delete(m.games, roomCode)
		}
	}
}

// broadcastGameClosure notifies all players that the game is being closed
func (m *Manager) broadcastGameClosure(game *GameState, reason string) {
	game.mu.RLock()
	defer game.mu.RUnlock()

	message := GameMessage{
		Type: MessageTypeError,
		Payload: ErrorPayload{
			Message: reason,
		},
	}

	for _, player := range game.Players {
		if player.Connection != nil && player.IsConnected {
			if err := player.Connection.WriteJSON(message); err != nil {
				logger.Error("Failed to notify player of game closure", "error", err, "player_id", player.ID)
			}
		}
	}
}

// markGameAsAbandoned updates the game status in the database
func (m *Manager) markGameAsAbandoned(game *GameState) {
	db := database.GetDB()

	err := db.Model(&models.Game{}).
		Where("room_code = ?", game.RoomCode).
		Updates(map[string]interface{}{
			"status":     models.GameStatusAbandoned,
			"updated_at": time.Now(),
		}).Error

	if err != nil {
		logger.Error("Failed to mark game as abandoned", "error", err, "room_code", game.RoomCode)
	}
}

// StopCleanupService stops the background cleanup service
func (m *Manager) StopCleanupService() {
	if m.stopCleanup != nil {
		close(m.stopCleanup)
	}
}
