import React, { useState, useEffect } from 'react';
import { useGameStore } from '../store/gameStore';
import Card from './Card';
import PlayerHand from './PlayerHand';
import VotingPhase from './VotingPhase';

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

  const getGamePhase = () => {
    if (!gameState?.current_round) return 'Waiting for game to start...';
    
    const round = gameState.current_round;
    const isStoryteller = currentPlayer?.id === round.storyteller_id;
    
    switch (round.status) {
      case 'storytelling':
        return isStoryteller ? 'Choose a card and give a clue' : 'Waiting for storyteller...';
      case 'submitting':
        return isStoryteller ? 'Waiting for other players to submit cards...' : 'Choose a card that fits the clue';
      case 'voting':
        return isStoryteller ? 'Waiting for players to vote...' : 'Vote for the storyteller\'s card';
      case 'scoring':
        return 'Calculating scores...';
      case 'completed':
        return 'Round completed!';
      default:
        return 'Unknown phase';
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
          <div className="round-number">Round {gameState.round_number} of {gameState.max_rounds}</div>
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

      <div className="phase-indicator">
        <h2>{getGamePhase()}</h2>
        {gameState.current_round?.clue && (
          <div className="clue-display">
            <strong>Clue:</strong> "{gameState.current_round.clue}"
          </div>
        )}
      </div>

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

      <style jsx>{`
        .game-board {
          min-height: 100vh;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          padding: 1rem;
        }

        .game-board.loading {
          display: flex;
          justify-content: center;
          align-items: center;
        }

        .loading-content {
          text-align: center;
          color: white;
        }

        .spinner {
          width: 40px;
          height: 40px;
          border: 3px solid rgba(255, 255, 255, 0.3);
          border-top: 3px solid white;
          border-radius: 50%;
          animation: spin 1s linear infinite;
          margin: 0 auto 1rem;
        }

        @keyframes spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }

        .game-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          background: rgba(255, 255, 255, 0.1);
          backdrop-filter: blur(10px);
          border-radius: 12px;
          padding: 1rem;
          margin-bottom: 1rem;
          color: white;
        }

        .game-info h1 {
          margin: 0;
          font-size: 1.5rem;
        }

        .connection-status {
          font-size: 0.9rem;
          margin-top: 0.25rem;
        }

        .round-info {
          text-align: center;
        }

        .round-number {
          font-size: 1.1rem;
          font-weight: bold;
          margin-bottom: 0.25rem;
        }

        .storyteller-info {
          font-size: 0.9rem;
        }

        .leave-btn {
          background: rgba(220, 38, 38, 0.8);
          color: white;
          border: none;
          padding: 0.5rem 1rem;
          border-radius: 8px;
          cursor: pointer;
          transition: background-color 0.2s;
        }

        .leave-btn:hover {
          background: rgba(220, 38, 38, 1);
        }

        .phase-indicator {
          text-align: center;
          background: rgba(255, 255, 255, 0.95);
          border-radius: 12px;
          padding: 1.5rem;
          margin-bottom: 2rem;
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }

        .phase-indicator h2 {
          margin: 0 0 0.5rem 0;
          color: #333;
        }

        .clue-display {
          color: #667eea;
          font-size: 1.1rem;
        }

        .players-section {
          margin-bottom: 2rem;
        }

        .players-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
          gap: 1rem;
        }

        .player-card {
          background: rgba(255, 255, 255, 0.9);
          border-radius: 12px;
          padding: 1rem;
          text-align: center;
          transition: all 0.2s;
        }

        .player-card.current-player {
          border: 2px solid #22c55e;
          background: rgba(34, 197, 94, 0.1);
        }

        .player-card.storyteller {
          border: 2px solid #f59e0b;
          background: rgba(245, 158, 11, 0.1);
        }

        .player-name {
          font-weight: bold;
          margin-bottom: 0.5rem;
          color: #333;
        }

        .player-score {
          color: #666;
          margin-bottom: 0.5rem;
        }

        .submission-indicator,
        .vote-indicator {
          font-size: 0.8rem;
          color: #22c55e;
          margin-top: 0.25rem;
        }

        .clue-form-section {
          margin-bottom: 2rem;
        }

        .clue-form {
          background: rgba(255, 255, 255, 0.95);
          border-radius: 12px;
          padding: 1.5rem;
          max-width: 500px;
          margin: 0 auto;
        }

        .clue-form h3 {
          margin: 0 0 1rem 0;
          text-align: center;
          color: #333;
        }

        .clue-form input {
          width: 100%;
          padding: 0.75rem;
          border: 1px solid #d1d5db;
          border-radius: 8px;
          font-size: 1rem;
          margin-bottom: 1rem;
          box-sizing: border-box;
        }

        .clue-form input:focus {
          outline: none;
          border-color: #667eea;
          box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }

        .clue-actions {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .selected-indicator {
          color: #666;
          font-size: 0.9rem;
        }

        .submit-clue-btn {
          background: #667eea;
          color: white;
          border: none;
          padding: 0.75rem 1.5rem;
          border-radius: 8px;
          cursor: pointer;
          transition: background-color 0.2s;
        }

        .submit-clue-btn:hover:not(:disabled) {
          background: #5a67d8;
        }

        .submit-clue-btn:disabled {
          background: #9ca3af;
          cursor: not-allowed;
        }

        @media (max-width: 768px) {
          .game-header {
            flex-direction: column;
            gap: 1rem;
            text-align: center;
          }

          .players-grid {
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
          }

          .clue-actions {
            flex-direction: column;
            gap: 1rem;
            align-items: stretch;
          }
        }
      `}</style>
    </div>
  );
};

export default GameBoard;
