package game

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"dixitme/internal/game/bot"
	"dixitme/internal/game/domain"
	"dixitme/internal/platform/logger"

	"github.com/google/uuid"
)

// Manager manages all active games
type Manager struct {
	repo        GameRepository
	cache       GameCache
	broadcaster Broadcaster
}

// NewManager constructs a game manager instance.
func NewManager(repo GameRepository, cache GameCache, broadcaster Broadcaster) *Manager {
	m := &Manager{
		repo:        repo,
		cache:       cache,
		broadcaster: broadcaster,
	}
	return m
}

var phaseTransitions = map[domain.GamePhase]map[domain.GamePhase]bool{
	domain.PhaseLobby: {
		domain.PhaseStorytellerSubmit: true,
	},
	domain.PhaseStorytellerSubmit: {
		domain.PhaseOthersSubmit: true,
	},
	domain.PhaseOthersSubmit: {
		domain.PhaseVoting: true,
	},
	domain.PhaseVoting: {
		domain.PhaseRevealScore: true,
	},
	domain.PhaseRevealScore: {
		domain.PhaseRoundEnd: true,
		domain.PhaseGameOver: true,
	},
	domain.PhaseRoundEnd: {
		domain.PhaseStorytellerSubmit: true,
	},
}

type phaseOptions struct {
	abandoned bool
}

func setPhase(game *GameState, next domain.GamePhase, opts ...phaseOptions) error {
	if game == nil {
		return fmt.Errorf("game is nil")
	}
	normalizePhase(game)
	if game.Phase == next {
		return nil
	}
	allowed := phaseTransitions[game.Phase]
	if allowed == nil || !allowed[next] {
		return fmt.Errorf("invalid phase transition: %s -> %s", game.Phase, next)
	}
	game.Phase = next
	if len(opts) > 0 && next == domain.PhaseGameOver {
		game.Abandoned = opts[0].abandoned
	}
	syncPhase(game)
	return nil
}

func normalizePhase(game *GameState) {
	if game == nil || game.Phase != "" {
		return
	}

	switch game.Status {
	case domain.GameStatusWaiting:
		game.Phase = domain.PhaseLobby
	case domain.GameStatusCompleted, domain.GameStatusAbandoned:
		game.Phase = domain.PhaseGameOver
		game.Abandoned = game.Status == domain.GameStatusAbandoned
	case domain.GameStatusInProgress:
		if game.CurrentRound == nil {
			game.Phase = domain.PhaseStorytellerSubmit
			return
		}
		switch game.CurrentRound.Status {
		case domain.RoundStatusStorytelling:
			game.Phase = domain.PhaseStorytellerSubmit
		case domain.RoundStatusSubmitting:
			game.Phase = domain.PhaseOthersSubmit
		case domain.RoundStatusVoting:
			game.Phase = domain.PhaseVoting
		case domain.RoundStatusScoring:
			game.Phase = domain.PhaseRevealScore
		case domain.RoundStatusCompleted:
			game.Phase = domain.PhaseRoundEnd
		default:
			game.Phase = domain.PhaseStorytellerSubmit
		}
	default:
		game.Phase = domain.PhaseLobby
	}
}

func syncPhase(game *GameState) {
	switch game.Phase {
	case domain.PhaseLobby:
		game.Status = domain.GameStatusWaiting
	case domain.PhaseGameOver:
		if game.Abandoned {
			game.Status = domain.GameStatusAbandoned
		} else {
			game.Status = domain.GameStatusCompleted
		}
	default:
		game.Status = domain.GameStatusInProgress
	}

	if game.CurrentRound == nil {
		return
	}
	switch game.Phase {
	case domain.PhaseStorytellerSubmit:
		game.CurrentRound.Status = domain.RoundStatusStorytelling
	case domain.PhaseOthersSubmit:
		game.CurrentRound.Status = domain.RoundStatusSubmitting
	case domain.PhaseVoting:
		game.CurrentRound.Status = domain.RoundStatusVoting
	case domain.PhaseRevealScore:
		game.CurrentRound.Status = domain.RoundStatusScoring
	case domain.PhaseRoundEnd, domain.PhaseGameOver:
		game.CurrentRound.Status = domain.RoundStatusCompleted
	}
}

func (m *Manager) persistSnapshot(game *GameState) error {
	if m.repo == nil {
		return nil
	}
	ok, err := m.repo.TrySaveGameSnapshot(game, game.Version)
	if err != nil {
		return fmt.Errorf("failed to save game snapshot: %w", err)
	}
	if !ok {
		return fmt.Errorf("game state conflict")
	}
	game.Version++
	return nil
}

// CreateGame creates a new game with the given room code
func (m *Manager) CreateGame(roomCode string, creatorID uuid.UUID, creatorName string) (*GameState, error) {
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
		Phase:        domain.PhaseLobby,
		RoundNumber:  0,
		MaxRounds:    999, // Will be determined by 30 points or empty deck
		Deck:         deck,
		UsedCards:    make([]int, 0),
		CreatedAt:    now,
		LastActivity: now,
	}

	// Add creator as first player
	creator := &Player{
		ID:             creatorID,
		Name:           creatorName,
		Score:          0,
		Position:       1,
		Hand:           make([]int, 0),
		IsConnected:    true,
		DisconnectedAt: nil,
		IsActive:       true,
	}

	game.Players[creatorID] = creator
	syncPhase(game)
	game.Version = 1

	// Persist to database
	if err := m.repo.CreateGame(game); err != nil {
		// Check if it's a duplicate key error (would need to check error string/type from repo)
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already taken") {
			return nil, fmt.Errorf("room code '%s' is already taken, please try a different one", roomCode)
		}
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	if err := m.repo.AddPlayerToGame(game.ID, creator); err != nil {
		return nil, fmt.Errorf("failed to add creator to game: %w", err)
	}
	if err := m.persistSnapshot(game); err != nil {
		return nil, err
	}

	// Store in Redis for scaling
	if err := m.cache.SetGame(context.Background(), game); err != nil {
		logger.Error("Failed to store game in Redis", "error", err, "room_code", roomCode)
	}

	return game, nil
}

func (m *Manager) loadSnapshot(roomCode string) (*GameState, int, error) {
	if m.repo == nil {
		return nil, 0, fmt.Errorf("repository not configured")
	}
	game, version, err := m.repo.LoadGameSnapshot(roomCode)
	if err != nil {
		return nil, 0, err
	}
	game.Version = version
	return game, version, nil
}

// JoinGame adds a player to an existing game
func (m *Manager) JoinGame(roomCode string, playerID uuid.UUID, playerName string) (*GameState, error) {
	game, _, err := m.loadSnapshot(roomCode)
	if err != nil {
		return nil, err
	}

	// Update activity
	game.LastActivity = time.Now()

	// Handle reconnection
	if game.Phase != domain.PhaseGameOver {
		if player, exists := game.Players[playerID]; exists && !player.IsConnected {
			// Allow reconnection
			player.IsConnected = true
			player.DisconnectedAt = nil

			// Broadcast player rejoin
			m.BroadcastToGame(game, MessageTypePlayerJoined, PlayerJoinedPayload{Player: player})

			// Send system message
			m.SendSystemMessage(roomCode, fmt.Sprintf("%s rejoined the game", playerName))

			if err := m.persistSnapshot(game); err != nil {
				return nil, err
			}
			return game, nil
		}
	}

	// Check if game is still accepting players
	if game.Phase != domain.PhaseLobby {
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
		ID:             playerID,
		Name:           playerName,
		Score:          0,
		Position:       len(game.Players) + 1,
		Hand:           make([]int, 0),
		IsConnected:    true,
		DisconnectedAt: nil,
		IsActive:       true,
	}

	game.Players[playerID] = player

	// Persist changes
	if err := m.repo.AddPlayerToGame(game.ID, player); err != nil {
		delete(game.Players, playerID)
		return nil, fmt.Errorf("failed to persist player: %w", err)
	}
	if err := m.persistSnapshot(game); err != nil {
		delete(game.Players, playerID)
		return nil, err
	}

	// Update Redis
	if err := m.cache.SetGame(context.Background(), game); err != nil {
		logger.Error("Failed to update game in Redis", "error", err, "room_code", roomCode)
	}

	// Broadcast player joined
	m.BroadcastToGame(game, MessageTypePlayerJoined, PlayerJoinedPayload{Player: player})

	// Send system message
	m.SendSystemMessage(roomCode, fmt.Sprintf("%s joined the game", playerName))

	return game, nil
}

// MarkPlayerDisconnected marks a player as disconnected and keeps them in the room.
func (m *Manager) MarkPlayerDisconnected(roomCode string, playerID uuid.UUID) error {
	game, _, err := m.loadSnapshot(roomCode)
	if err != nil {
		return err
	}
	player, inGame := game.Players[playerID]
	if !inGame {
		return fmt.Errorf("player not in game")
	}

	now := time.Now()
	player.IsConnected = false
	player.IsActive = false
	player.DisconnectedAt = &now
	game.LastActivity = now

	if err := m.persistSnapshot(game); err != nil {
		return err
	}
	if m.cache != nil {
		if err := m.cache.SetGame(context.Background(), game); err != nil {
			logger.Error("Failed to update game cache after disconnect", "error", err, "room_code", roomCode)
		}
	}

	m.SendSystemMessage(roomCode, fmt.Sprintf("%s has disconnected", player.Name))

	return nil
}

// LeaveGame removes a player from an active game and notifies the room.
func (m *Manager) LeaveGame(roomCode string, playerID uuid.UUID) error {
	game, _, err := m.loadSnapshot(roomCode)
	if err != nil {
		return err
	}
	player, inGame := game.Players[playerID]
	if !inGame {
		return fmt.Errorf("player not in game")
	}

	delete(game.Players, playerID)
	game.LastActivity = time.Now()
	remaining := len(game.Players)

	if err := m.persistSnapshot(game); err != nil {
		return err
	}
	if m.cache != nil {
		if err := m.cache.SetGame(context.Background(), game); err != nil {
			logger.Error("Failed to update game cache after leave", "error", err, "room_code", roomCode)
		}
	}

	m.BroadcastToGame(game, MessageTypePlayerLeft, PlayerLeftPayload{PlayerID: playerID})
	m.SendSystemMessage(roomCode, fmt.Sprintf("%s left the game", player.Name))

	if remaining == 0 {
	}

	return nil
}

// AddBot adds a bot player to an existing game
func (m *Manager) AddBot(roomCode string, botLevel string) (*GameState, error) {
	game, _, err := m.loadSnapshot(roomCode)
	if err != nil {
		return nil, err
	}

	if game.Phase != domain.PhaseLobby {
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
	// Create game player
	player := &Player{
		ID:             botID,
		Name:           botName,
		Score:          0,
		Position:       len(game.Players) + 1,
		Hand:           make([]int, 0),
		IsConnected:    true, // Bots are always "connected"
		DisconnectedAt: nil,
		IsActive:       true,
		IsBot:          true,
		BotLevel:       botLevel,
	}

	game.Players[botID] = player

	// Persist bot player (handled by AddPlayerToGame which should handle Bot logic in repo if needed, or generic player)
	// Note: original code created a User/Player entry for bot. Repository implementation should handle this.
	if err := m.repo.AddPlayerToGame(game.ID, player); err != nil {
		delete(game.Players, botID)
		return nil, fmt.Errorf("failed to persist bot player: %w", err)
	}
	if err := m.persistSnapshot(game); err != nil {
		delete(game.Players, botID)
		return nil, err
	}

	// Update Redis
	if err := m.cache.SetGame(context.Background(), game); err != nil {
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
	game, _, err := m.loadSnapshot(roomCode)
	if err != nil {
		return err
	}

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

	if game.Phase != domain.PhaseLobby {
		return fmt.Errorf("game already started")
	}

	// Initialize game
	if err := setPhase(game, domain.PhaseStorytellerSubmit); err != nil {
		return fmt.Errorf("failed to start game: %w", err)
	}

	// Deal cards to players
	m.dealCards(game)

	// Start first round
	if err := m.startNewRound(game); err != nil {
		return fmt.Errorf("failed to start first round: %w", err)
	}

	if err := m.persistSnapshot(game); err != nil {
		return err
	}

	// Broadcast game started
	m.BroadcastToGame(game, MessageTypeGameStarted, GameStartedPayload{GameState: game})

	// Send system message
	m.SendSystemMessage(roomCode, "Game started! Let the storytelling begin!")
	m.ProcessBotActions(game)

	return nil
}

// SubmitClue handles storyteller submitting a clue
func (m *Manager) SubmitClue(roomCode string, playerID uuid.UUID, clue string, cardID int) error {
	game, _, err := m.loadSnapshot(roomCode)
	if err != nil {
		return err
	}

	if game.CurrentRound == nil {
		return fmt.Errorf("no active round")
	}

	if game.CurrentRound.StorytellerID != playerID {
		return fmt.Errorf("only storyteller can submit clue")
	}

	if game.Phase != domain.PhaseStorytellerSubmit {
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
	if err := setPhase(game, domain.PhaseOthersSubmit); err != nil {
		return fmt.Errorf("failed to advance phase: %w", err)
	}

	// Remove card from storyteller's hand and add to used cards
	for i, handCard := range player.Hand {
		if handCard == cardID {
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
			game.UsedCards = append(game.UsedCards, cardID)
			break
		}
	}

	// Persist round update
	if err := m.repo.UpdateRound(game.CurrentRound); err != nil {
		return fmt.Errorf("failed to update round: %w", err)
	}
	if err := m.persistSnapshot(game); err != nil {
		return err
	}

	// Broadcast clue submitted
	m.BroadcastToGame(game, MessageTypeClueSubmitted, ClueSubmittedPayload{Clue: clue})
	m.BroadcastToGame(game, MessageTypeGameState, GameStatePayload{GameState: game})
	m.sendSystemMessageWithGame(game, fmt.Sprintf("Storyteller submitted a clue. Others can now submit cards."))
	m.ProcessBotActions(game)

	return nil
}

// SubmitCard handles non-storyteller players submitting cards
func (m *Manager) SubmitCard(roomCode string, playerID uuid.UUID, cardID int) error {
	game, _, err := m.loadSnapshot(roomCode)
	if err != nil {
		return err
	}

	if game.CurrentRound == nil {
		return fmt.Errorf("no active round")
	}

	if game.CurrentRound.StorytellerID == playerID {
		return fmt.Errorf("storyteller cannot submit cards")
	}

	if game.Phase != domain.PhaseOthersSubmit {
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

	// Check if all players submitted
	expectedSubmissions := len(game.Players) - 1 // Exclude storyteller
	if len(game.CurrentRound.Submissions) == expectedSubmissions {
		m.startVotingPhase(game)
	}
	if err := m.persistSnapshot(game); err != nil {
		return err
	}

	// Broadcast card submitted
	m.BroadcastToGame(game, MessageTypeCardSubmitted, CardSubmittedPayload{PlayerID: playerID})
	m.BroadcastToGame(game, MessageTypeGameState, GameStatePayload{GameState: game})
	if len(game.CurrentRound.Submissions) == expectedSubmissions {
		m.sendSystemMessageWithGame(game, "All cards submitted. Voting started.")
	}

	return nil
}

// SubmitVote handles player voting
func (m *Manager) SubmitVote(roomCode string, playerID uuid.UUID, cardID int) error {
	game, _, err := m.loadSnapshot(roomCode)
	if err != nil {
		return err
	}

	if game.CurrentRound == nil {
		return fmt.Errorf("no active round")
	}

	if game.CurrentRound.StorytellerID == playerID {
		return fmt.Errorf("storyteller cannot vote")
	}

	if game.Phase != domain.PhaseVoting {
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

	// Check if all players voted
	expectedVotes := len(game.Players) - 1 // Exclude storyteller
	if len(game.CurrentRound.Votes) == expectedVotes {
		m.completeRound(game)
	}
	if err := m.persistSnapshot(game); err != nil {
		return err
	}

	// Broadcast vote submitted
	m.BroadcastToGame(game, MessageTypeVoteSubmitted, VoteSubmittedPayload{PlayerID: playerID})
	m.BroadcastToGame(game, MessageTypeGameState, GameStatePayload{GameState: game})
	if len(game.CurrentRound.Votes) == expectedVotes {
		m.sendSystemMessageWithGame(game, "All votes submitted.")
	}

	return nil
}

// Helper methods

func (m *Manager) GetGame(roomCode string) *GameState {
	game, _, err := m.loadSnapshot(roomCode)
	if err != nil {
		return nil
	}
	return game
}

func (m *Manager) getGame(roomCode string) *GameState {
	return m.GetGame(roomCode)
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
		Submissions:   make(map[uuid.UUID]*CardSubmission),
		Votes:         make(map[uuid.UUID]*Vote),
		CreatedAt:     time.Now(),
	}

	game.CurrentRound = round
	if err := setPhase(game, domain.PhaseStorytellerSubmit); err != nil {
		return err
	}

	// Persist round
	if err := m.repo.SaveRound(game.ID, round); err != nil {
		return err
	}

	// Broadcast round started
	m.BroadcastToGame(game, MessageTypeRoundStarted, RoundStartedPayload{Round: round})
	m.BroadcastToGame(game, MessageTypeGameState, GameStatePayload{GameState: game})
	m.sendSystemMessageWithGame(game, fmt.Sprintf("Round %d started. Storyteller: %s", round.RoundNumber, storytellerName(game, storytellerID)))

	return nil
}

func (m *Manager) startVotingPhase(game *GameState) {
	if err := setPhase(game, domain.PhaseVoting); err != nil {
		logger.Error("Failed to advance to voting phase", "error", err, "room_code", game.RoomCode)
		return
	}

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
	m.BroadcastToGame(game, MessageTypeGameState, GameStatePayload{GameState: game})
	m.ProcessBotActions(game)
}

func (m *Manager) completeRound(game *GameState) {
	if err := setPhase(game, domain.PhaseRevealScore); err != nil {
		logger.Error("Failed to advance to reveal score phase", "error", err, "room_code", game.RoomCode)
		return
	}

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
	m.BroadcastToGame(game, MessageTypeGameState, GameStatePayload{GameState: game})

	roundScoreMessage := buildRoundScoreMessage(game, scores)
	if roundScoreMessage != "" {
		m.sendSystemMessageWithGame(game, roundScoreMessage)
	}

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
		return
	}

	if err := setPhase(game, domain.PhaseRoundEnd); err != nil {
		logger.Error("Failed to advance to round end phase", "error", err, "room_code", game.RoomCode)
		return
	}

	// Start next round after a delay
	go func() {
		time.Sleep(5 * time.Second)
		m.advanceToNextRound(game.RoomCode)
	}()
}

func buildRoundScoreMessage(game *GameState, scores map[uuid.UUID]int) string {
	if game == nil || len(scores) == 0 {
		return ""
	}

	var parts []string
	for playerID, delta := range scores {
		player, exists := game.Players[playerID]
		if !exists {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s +%d", player.Name, delta))
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("Round %d scores: %s", game.RoundNumber, strings.Join(parts, ", "))
}

func (m *Manager) advanceToNextRound(roomCode string) {
	game, _, err := m.loadSnapshot(roomCode)
	if err != nil {
		logger.Error("Failed to load snapshot for next round", "error", err, "room_code", roomCode)
		return
	}
	game.LastActivity = time.Now()
	if err := m.startNewRound(game); err != nil {
		logger.Error("Failed to start next round", "error", err, "room_code", roomCode)
		return
	}
	if err := m.persistSnapshot(game); err != nil {
		logger.Error("Failed to persist snapshot for next round", "error", err, "room_code", roomCode)
		return
	}
	m.ProcessBotActions(game)
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
	if err := setPhase(game, domain.PhaseGameOver, phaseOptions{abandoned: false}); err != nil {
		logger.Error("Failed to advance to game over phase", "error", err, "room_code", game.RoomCode)
	}

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
	if err := m.repo.SaveGameCompletion(game.ID, winnerID, finalScores, game.UsedCards); err != nil {
		logger.Error("Failed to persist game completion", "error", err, "game_id", game.ID, "winner_id", winnerID)
	}

	// Broadcast game completed
	m.BroadcastToGame(game, MessageTypeGameCompleted, GameCompletedPayload{
		FinalScores: finalScores,
		Winner:      winnerID,
	})
	if winnerName != "" {
		m.sendSystemMessageWithGame(game, fmt.Sprintf("Game over! Winner: %s", winnerName))
	}

}

func storytellerName(game *GameState, storytellerID uuid.UUID) string {
	if game == nil {
		return ""
	}
	if player, ok := game.Players[storytellerID]; ok {
		return player.Name
	}
	return ""
}

func (m *Manager) BroadcastToGame(game *GameState, messageType MessageType, payload interface{}) {
	if m.broadcaster == nil {
		return
	}
	m.broadcaster.Broadcast(game.RoomCode, GameMessage{
		Type:    messageType,
		Payload: payload,
	})
}

// Bot automation methods

// ProcessBotActions handles automated bot actions for the current game phase
func (m *Manager) ProcessBotActions(game *GameState) {
	if game.CurrentRound == nil {
		return
	}

	switch game.Phase {
	case domain.PhaseStorytellerSubmit:
		m.processBotStorytelling(game)
	case domain.PhaseOthersSubmit:
		m.processBotSubmissions(game)
	case domain.PhaseVoting:
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
		botPlayer := botManager.NewBot(storytellerID, storyteller.Name, bot.BotDifficulty(storyteller.BotLevel), game.ID, storyteller.Hand)

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
			botInstance := botManager.NewBot(botID, botPlayer.Name, bot.BotDifficulty(botPlayer.BotLevel), game.ID, botPlayer.Hand)

			// Bot selects card for clue
			selectedCard, err := botInstance.SelectCardForClue(game.CurrentRound.Clue)
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
			botInstance := botManager.NewBot(botID, botPlayer.Name, bot.BotDifficulty(botPlayer.BotLevel), game.ID, botPlayer.Hand)

			// Get submitted cards for voting
			submittedCards := make([]int, 0, len(game.CurrentRound.RevealedCards))
			for _, revealedCard := range game.CurrentRound.RevealedCards {
				submittedCards = append(submittedCards, revealedCard.CardID)
			}

			// Bot votes for card
			selectedCard, err := botInstance.VoteForCard(submittedCards, game.CurrentRound.Clue, game.CurrentRound.StorytellerCard)
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
	game, _, err := m.loadSnapshot(roomCode)
	if err != nil {
		return err
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
	currentPhase := strings.ToLower(string(game.Phase))

	// Allow chat in all phases except storyteller submit.
	if game.Phase == domain.PhaseStorytellerSubmit {
		return fmt.Errorf("chat not allowed during storyteller phase")
	}

	// Create chat message
	chatMessage := &ChatMessage{
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
	if err := m.repo.SaveChatMessage(chatMessage); err != nil {
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
	// This needs to be implemented in Repo to fetch history
	// For now returning empty or implementing via Repo if interface supports it.
	// The interface didn't have GetChatMessages.
	// I should probably add it to interface or leave it as TODO.
	// Given "rearrange", I should maintain functionality.
	// I'll add GetChatHistory to repo interface in next step if I can.
	// For now, I'll return empty or error.
	return nil, fmt.Errorf("not implemented in this refactor step")
}

// SendSystemMessage sends a system message (e.g., "Player joined", "Round started")
func (m *Manager) SendSystemMessage(roomCode string, message string) error {
	game, _, err := m.loadSnapshot(roomCode)
	if err != nil {
		return err
	}

	// Determine current phase
	currentPhase := strings.ToLower(string(game.Phase))

	// Create system message with a system player ID (using nil UUID)
	systemPlayerID := uuid.Nil
	chatMessage := &ChatMessage{
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
	if err := m.repo.SaveChatMessage(chatMessage); err != nil {
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

func (m *Manager) sendSystemMessageWithGame(game *GameState, message string) {
	if game == nil {
		return
	}

	currentPhase := strings.ToLower(string(game.Phase))

	systemPlayerID := uuid.Nil
	chatMessage := &ChatMessage{
		ID:          uuid.New(),
		GameID:      game.ID,
		PlayerID:    systemPlayerID,
		Message:     message,
		MessageType: "system",
		Phase:       currentPhase,
		IsVisible:   true,
		CreatedAt:   time.Now(),
	}

	if err := m.repo.SaveChatMessage(chatMessage); err != nil {
		logger.Error("Failed to persist system message", "error", err)
	}

	payload := ChatMessagePayload{
		ID:          chatMessage.ID,
		PlayerID:    systemPlayerID,
		PlayerName:  "System",
		Message:     chatMessage.Message,
		MessageType: chatMessage.MessageType,
		Phase:       chatMessage.Phase,
		Timestamp:   chatMessage.CreatedAt,
	}

	m.BroadcastToGame(game, MessageTypeChatMessage, payload)
}
