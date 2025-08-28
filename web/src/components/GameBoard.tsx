import React, { useState, useEffect } from 'react';
import { useGameStore } from '../store/gameStore';
import PlayerHand from './PlayerHand';
import VotingPhase from './VotingPhase';
import Chat from './Chat';
import GamePhaseIndicator from './GamePhaseIndicator';

const GameBoard: React.FC = () => {
  const {
    gameState,
    currentPlayer,
    isConnected,
    submitClue,
    submitCard,
    leaveGame,
  } = useGameStore();

  const [selectedCard, setSelectedCard] = useState<number | null>(null);
  const [clueText, setClueText] = useState('');
  const [showClueForm, setShowClueForm] = useState(false);
  const [isChatOpen, setIsChatOpen] = useState(false);

  useEffect(() => {
    if (gameState?.current_round) {
      const round = gameState.current_round;
      const isStoryteller = currentPlayer?.id === round.storyteller_id;
      
      setShowClueForm(isStoryteller && round.status === 'storytelling');
      
      if (round.status !== 'storytelling') {
        setSelectedCard(null);
        setClueText('');
      }
    }
  }, [gameState?.current_round, currentPlayer]);

  const handleSubmitClue = () => {
    if (!gameState || !selectedCard || !clueText.trim()) return;
    
    submitClue(gameState.room_code, clueText.trim(), selectedCard);
    setShowClueForm(false);
    setSelectedCard(null);
    setClueText('');
  };

  const handleSubmitCard = (cardId: number) => {
    if (!gameState) return;
    submitCard(gameState.room_code, cardId);
  };

  const handleLeaveGame = () => {
    if (gameState && window.confirm('Are you sure you want to leave the game?')) {
      leaveGame(gameState.room_code);
    }
  };



  const getStorytellerName = () => {
    if (!gameState?.current_round?.storyteller_id) return '';
    const storyteller = gameState.players[gameState.current_round.storyteller_id];
    return storyteller?.name || '';
  };

  const canSubmitCard = () => {
    if (!gameState?.current_round || !currentPlayer) return false;
    const round = gameState.current_round;
    const isStoryteller = currentPlayer.id === round.storyteller_id;
    const hasSubmitted = round.submissions[currentPlayer.id];
    
    return !isStoryteller && round.status === 'submitting' && !hasSubmitted;
  };

  if (!gameState || !currentPlayer) {
    return (
      <div className="game-board loading">
        <div className="loading-content">
          <div className="spinner"></div>
          <p>Loading game...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="game-board">
      <div className="game-header">
        <div className="game-info">
          <h1>Room: {gameState.room_code}</h1>
          <div className="connection-status">
            <span className={`status-indicator ${isConnected ? 'connected' : 'disconnected'}`}>
              {isConnected ? '🟢 Connected' : '🔴 Disconnected'}
            </span>
          </div>
        </div>
        
        <div className="round-info">
          <div className="round-number">Round {gameState.round_number}</div>
          <div className="game-rule">First to 30 points wins!</div>
          {gameState.current_round && (
            <div className="storyteller-info">
              Storyteller: <strong>{getStorytellerName()}</strong>
            </div>
          )}
        </div>

        <button onClick={handleLeaveGame} className="leave-btn">
          Leave Game
        </button>
      </div>

      {/* Game Phase Indicator */}
      {gameState.current_round && (
        <GamePhaseIndicator
          currentPhase={gameState.current_round.status}
          roundNumber={gameState.round_number}
          isStoryteller={currentPlayer.id === gameState.current_round.storyteller_id}
        />
      )}

      {gameState.current_round?.clue && (
        <div className="clue-display">
          <strong>Clue:</strong> "{gameState.current_round.clue}"
        </div>
      )}

      <div className="players-section">
        <div className="players-grid">
          {Object.values(gameState.players).map((player) => (
            <div
              key={player.id}
              className={`player-card ${player.id === currentPlayer.id ? 'current-player' : ''} ${
                player.id === gameState.current_round?.storyteller_id ? 'storyteller' : ''
              }`}
            >
              <div className="player-name">{player.name}</div>
              <div className="player-score">Score: {player.score}</div>
              <div className={`player-status ${player.is_connected ? 'online' : 'offline'}`}>
                {player.is_connected ? '🟢' : '🔴'}
              </div>
              {gameState.current_round?.submissions[player.id] && (
                <div className="submission-indicator">📤 Submitted</div>
              )}
              {gameState.current_round?.votes[player.id] && (
                <div className="vote-indicator">🗳️ Voted</div>
              )}
            </div>
          ))}
        </div>
      </div>

      {gameState.current_round?.status === 'voting' && gameState.current_round.revealed_cards && (
        <VotingPhase
          revealedCards={gameState.current_round.revealed_cards}
          isStoryteller={currentPlayer.id === gameState.current_round.storyteller_id}
          hasVoted={!!gameState.current_round.votes[currentPlayer.id]}
          onVote={(cardId) => {
            if (gameState) {
              useGameStore.getState().submitVote(gameState.room_code, cardId);
            }
          }}
        />
      )}

      {showClueForm && (
        <div className="clue-form-section">
          <div className="clue-form">
            <h3>Give your clue</h3>
            <input
              type="text"
              value={clueText}
              onChange={(e) => setClueText(e.target.value)}
              placeholder="Enter your clue..."
              maxLength={100}
              onKeyPress={(e) => {
                if (e.key === 'Enter' && selectedCard && clueText.trim()) {
                  handleSubmitClue();
                }
              }}
            />
            <div className="clue-actions">
              <span className="selected-indicator">
                {selectedCard ? `Card ${selectedCard} selected` : 'Select a card first'}
              </span>
              <button
                onClick={handleSubmitClue}
                disabled={!selectedCard || !clueText.trim()}
                className="submit-clue-btn"
              >
                Submit Clue
              </button>
            </div>
          </div>
        </div>
      )}

      <PlayerHand
        cards={currentPlayer.hand}
        selectedCard={selectedCard}
        onCardSelect={setSelectedCard}
        canSelect={showClueForm || canSubmitCard()}
        canSubmit={canSubmitCard()}
        onSubmit={handleSubmitCard}
      />

      {/* Chat Component */}
      <Chat 
        isOpen={isChatOpen} 
        onToggle={() => setIsChatOpen(!isChatOpen)} 
      />
    </div>
  );
};

export default GameBoard;
