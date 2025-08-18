package game

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"dixitme/internal/database"
	"dixitme/internal/models"
	"dixitme/internal/redis"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Manager manages all active games
type Manager struct {
	games map[string]*GameState
	mu    sync.RWMutex
}

var gameManager *Manager

// GetManager returns the singleton game manager
func GetManager() *Manager {
	if gameManager == nil {
		gameManager = &Manager{
			games: make(map[string]*GameState),
		}
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
	game := &GameState{
		ID:          gameID,
		RoomCode:    roomCode,
		Players:     make(map[uuid.UUID]*Player),
		Status:      models.GameStatusWaiting,
		RoundNumber: 0,
		MaxRounds:   6, // Will be adjusted based on player count
		CreatedAt:   time.Now(),
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
		return nil, fmt.Errorf("failed to persist game: %w", err)
	}

	// Store in Redis for scaling
	if err := m.storeGameInRedis(game); err != nil {
		log.Printf("Failed to store game in Redis: %v", err)
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

	// Adjust max rounds based on player count (each player gets 2 rounds as storyteller)
	game.MaxRounds = len(game.Players) * 2

	// Persist changes
	if err := m.persistGamePlayer(game.ID, player); err != nil {
		delete(game.Players, playerID)
		return nil, fmt.Errorf("failed to persist player: %w", err)
	}

	// Update Redis
	if err := m.storeGameInRedis(game); err != nil {
		log.Printf("Failed to update game in Redis: %v", err)
	}

	// Broadcast player joined
	m.BroadcastToGame(game, MessageTypePlayerJoined, PlayerJoinedPayload{Player: player})

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
	game.MaxRounds = len(game.Players) * 2

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

	// Remove card from storyteller's hand
	for i, handCard := range player.Hand {
		if handCard == cardID {
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
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

	// Remove card from player's hand
	for i, handCard := range player.Hand {
		if handCard == cardID {
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
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
	// Simple card dealing - in a real implementation, you'd have a proper deck
	cardID := 1
	for _, player := range game.Players {
		player.Hand = make([]int, 6) // Each player gets 6 cards
		for i := 0; i < 6; i++ {
			player.Hand[i] = cardID
			cardID++
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

	// Check if game is complete
	if game.RoundNumber >= game.MaxRounds {
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
	highestScore := -1
	finalScores := make(map[uuid.UUID]int)

	for playerID, player := range game.Players {
		finalScores[playerID] = player.Score
		if player.Score > highestScore {
			highestScore = player.Score
			winnerID = playerID
		}
	}

	// Persist game completion
	if err := m.persistGameCompletion(game.ID, winnerID); err != nil {
		log.Printf("Failed to persist game completion: %v", err)
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
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	for _, player := range game.Players {
		if player.Connection != nil && player.IsConnected {
			if err := player.Connection.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
				log.Printf("Failed to send message to player %s: %v", player.ID, err)
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
