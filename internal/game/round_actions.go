package game

import (
	"fmt"
	"math/rand"
	"time"

	"dixitme/internal/logger"
	"dixitme/internal/models"

	"github.com/google/uuid"
)

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

// Card dealing and deck management

func (m *Manager) dealCards(game *GameState) {
	for _, player := range game.Players {
		for len(player.Hand) < 6 && len(game.Deck) > 0 {
			if !m.drawCard(game, player) {
				break
			}
		}
	}
	logger.Info("Cards dealt", "room_code", game.RoomCode, "remaining_deck", len(game.Deck))
}

func (m *Manager) drawCard(game *GameState, player *Player) bool {
	if len(game.Deck) == 0 {
		return false
	}

	// Draw from top of deck
	cardID := game.Deck[0]
	game.Deck = game.Deck[1:]
	player.Hand = append(player.Hand, cardID)
	return true
}

func (m *Manager) refillHands(game *GameState) {
	for _, player := range game.Players {
		for len(player.Hand) < 6 && len(game.Deck) > 0 {
			if !m.drawCard(game, player) {
				break
			}
		}
	}
}

// Round management

func (m *Manager) startNewRound(game *GameState) error {
	game.RoundNumber++

	// Choose storyteller (rotate through players)
	storytellerIndex := (game.RoundNumber - 1) % len(game.Players)
	var storytellerID uuid.UUID
	i := 0
	for playerID := range game.Players {
		if i == storytellerIndex {
			storytellerID = playerID
			break
		}
		i++
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
		return fmt.Errorf("failed to persist round: %w", err)
	}

	// Broadcast round started
	m.BroadcastToGame(game, MessageTypeRoundStarted, RoundStartedPayload{
		Round: round,
	})

	// Process bot storytelling if storyteller is a bot
	m.ProcessBotActions(game)

	logger.Info("New round started",
		"room_code", game.RoomCode,
		"round_number", game.RoundNumber,
		"storyteller_id", storytellerID)

	return nil
}

func (m *Manager) startVotingPhase(game *GameState) {
	round := game.CurrentRound
	round.Status = models.RoundStatusVoting

	// Create revealed cards (shuffle submissions + storyteller card)
	revealedCards := make([]RevealedCard, 0, len(round.Submissions)+1)

	// Add storyteller card
	revealedCards = append(revealedCards, RevealedCard{
		CardID:   round.StorytellerCard,
		PlayerID: round.StorytellerID,
	})

	// Add other submissions
	for _, submission := range round.Submissions {
		revealedCards = append(revealedCards, RevealedCard{
			CardID:   submission.CardID,
			PlayerID: submission.PlayerID,
		})
	}

	// Shuffle revealed cards
	for i := len(revealedCards) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		revealedCards[i], revealedCards[j] = revealedCards[j], revealedCards[i]
	}

	round.RevealedCards = revealedCards

	// Update round in database
	if err := m.updateRound(round); err != nil {
		logger.Error("Failed to update round for voting phase", "error", err)
	}

	// Broadcast voting started
	m.BroadcastToGame(game, MessageTypeVotingStarted, VotingStartedPayload{
		RevealedCards: revealedCards,
	})

	// Process bot voting
	m.ProcessBotActions(game)

	logger.Info("Voting phase started", "room_code", game.RoomCode, "round", game.RoundNumber)
}

func (m *Manager) completeRound(game *GameState) {
	round := game.CurrentRound
	round.Status = models.RoundStatusScoring

	// Calculate scores
	newScores := m.calculateScores(game)

	// Update round status
	if err := m.updateRound(round); err != nil {
		logger.Error("Failed to update round completion", "error", err)
	}

	// Broadcast round completed
	m.BroadcastToGame(game, MessageTypeRoundCompleted, RoundCompletedPayload{
		Scores:        newScores,
		RevealedCards: round.RevealedCards,
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
	round := game.CurrentRound
	storytellerID := round.StorytellerID

	// Count votes for storyteller's card
	storytellerVotes := 0
	totalVoters := len(round.Votes)

	for _, vote := range round.Votes {
		if vote.CardID == round.StorytellerCard {
			storytellerVotes++
		}
	}

	// Dixit scoring rules
	if storytellerVotes == 0 || storytellerVotes == totalVoters {
		// All or none guessed correctly: Storyteller gets 0, others get 2
		for playerID, player := range game.Players {
			if playerID != storytellerID {
				player.Score += 2
			}
		}
	} else {
		// Some guessed correctly: Storyteller + correct guessers get 3
		game.Players[storytellerID].Score += 3

		for _, vote := range round.Votes {
			if vote.CardID == round.StorytellerCard {
				game.Players[vote.PlayerID].Score += 3
			}
		}
	}

	// Count votes for each card and award points
	cardVotes := make(map[int]int)
	for _, vote := range round.Votes {
		cardVotes[vote.CardID]++
	}

	// Award points for votes received (except storyteller's card)
	for _, submission := range round.Submissions {
		if votes, exists := cardVotes[submission.CardID]; exists {
			game.Players[submission.PlayerID].Score += votes
		}
	}

	// Return current scores
	scores := make(map[uuid.UUID]int)
	for playerID, player := range game.Players {
		scores[playerID] = player.Score
	}

	logger.Info("Round scoring completed",
		"room_code", game.RoomCode,
		"round", game.RoundNumber,
		"storyteller_votes", storytellerVotes,
		"total_voters", totalVoters)

	return scores
}

func (m *Manager) completeGame(game *GameState) {
	game.Status = models.GameStatusCompleted

	// Find winner (highest score)
	var winnerID uuid.UUID
	var winnerName string
	var winnerScore int

	for playerID, player := range game.Players {
		if player.Score > winnerScore {
			winnerScore = player.Score
			winnerID = playerID
			winnerName = player.Name
		}
	}

	// Update game status in database
	if err := m.updateGameStatus(game.ID, models.GameStatusCompleted); err != nil {
		logger.Error("Failed to update game completion status", "error", err)
	}

	// Persist game completion
	if err := m.persistGameCompletion(game.ID, winnerID); err != nil {
		logger.Error("Failed to persist game completion", "error", err)
	}

	// Broadcast game completed
	finalScores := make(map[uuid.UUID]int)
	for playerID, player := range game.Players {
		finalScores[playerID] = player.Score
	}

	m.BroadcastToGame(game, MessageTypeGameCompleted, GameCompletedPayload{
		Winner:      winnerID,
		FinalScores: finalScores,
	})

	logger.Info("Game completed",
		"room_code", game.RoomCode,
		"winner_name", winnerName,
		"rounds_played", game.RoundNumber,
		"cards_remaining", len(game.Deck),
		"cards_used", len(game.UsedCards))
}
