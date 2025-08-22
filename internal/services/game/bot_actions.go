package game

import (
	"math/rand"
	"time"

	"dixitme/internal/logger"
	"dixitme/internal/services/bot"

	"github.com/google/uuid"
)

// BotService defines bot-related operations
type BotService interface {
	AddBot(roomCode string, botLevel string) (*GameState, error)
	ProcessBotActions(gameState *GameState)
}

// ProcessBotActions handles bot actions based on game phase
func (m *Manager) ProcessBotActions(game *GameState) {
	if game.CurrentRound == nil {
		return
	}

	switch game.CurrentRound.Status {
	case "storytelling":
		m.processBotStorytelling(game)
	case "submitting":
		m.processBotSubmissions(game)
	case "voting":
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
