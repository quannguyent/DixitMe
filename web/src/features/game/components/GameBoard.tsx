import React, { useState, useEffect, useRef } from 'react';
import { useGameStore } from '../stores/gameStore';
import { PlayerHand } from './PlayerHand';
import { VotingPhase } from './VotingPhase';

export const GameBoard: React.FC = () => {
  const {
    gameState,
    currentPlayer,
    isConnected,
    submitClue,
    submitCard,
    leaveGame,
    sendChat,
    chatMessages,
  } = useGameStore();

  const [selectedCard, setSelectedCard] = useState<number | null>(null);
  const [clueText, setClueText] = useState('');
  const [showClueForm, setShowClueForm] = useState(false);
  const [chatText, setChatText] = useState('');
  const chatEndRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!gameState) return;
    const round = gameState.current_round;
    const isStoryteller = currentPlayer?.id === round?.storyteller_id;

    setShowClueForm(!!round && isStoryteller && gameState.phase === 'STORYTELLER_SUBMIT');

    if (gameState.phase !== 'STORYTELLER_SUBMIT') {
      setSelectedCard(null);
      setClueText('');
    }
  }, [gameState, currentPlayer]);

  useEffect(() => {
    if (!chatEndRef.current) return;
    chatEndRef.current.scrollIntoView({ behavior: 'smooth', block: 'end' });
  }, [chatMessages.length]);

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

  const handleSendChat = () => {
    if (!gameState) return;
    if (!chatText.trim()) return;
    sendChat(gameState.room_code, chatText.trim());
    setChatText('');
  };

  const getGamePhase = () => {
    if (!gameState?.current_round) return 'Waiting for game to start...';

    const round = gameState.current_round;
    const isStoryteller = currentPlayer?.id === round.storyteller_id;

    switch (gameState.phase) {
      case 'LOBBY':
        return 'Waiting for game to start...';
      case 'STORYTELLER_SUBMIT':
        return isStoryteller ? 'Choose a card and give a clue' : 'Waiting for storyteller...';
      case 'OTHERS_SUBMIT':
        return isStoryteller ? 'Waiting for other players to submit cards...' : 'Choose a card that fits the clue';
      case 'VOTING':
        return isStoryteller ? 'Waiting for players to vote...' : 'Vote for the storyteller\'s card';
      case 'REVEAL_SCORE':
        return 'Calculating scores...';
      case 'ROUND_END':
        return 'Round completed!';
      case 'GAME_OVER':
        return 'Game over!';
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

    return !isStoryteller && gameState.phase === 'OTHERS_SUBMIT' && !hasSubmitted;
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

  const players = Object.values(gameState.players);
  const onlineCount = players.filter((player) => player.is_connected).length;
  const phaseTag = gameState.phase.replace(/_/g, ' ');
  const isStoryteller = !!gameState.current_round && currentPlayer.id === gameState.current_round.storyteller_id;

  return (
    <div className="game-board">
      <div className="game-header">
        <div className="game-info">
          <div className="room-pill">Room {gameState.room_code}</div>
          <h1>Dixit Table</h1>
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
          <div className="role-badge">{isStoryteller ? 'You are the storyteller' : 'You are a player'}</div>
        </div>

        <button onClick={handleLeaveGame} className="leave-btn">
          Leave Game
        </button>
      </div>

      <div className="phase-indicator">
        <h2>{getGamePhase()}</h2>
        <div className="phase-meta">
          <span className="phase-pill">{phaseTag}</span>
        </div>
        {gameState.current_round?.clue && (
          <div className="clue-display">
            <strong>Clue:</strong> "{gameState.current_round.clue}"
          </div>
        )}
      </div>

      <div className="players-section">
        <div className="players-header">
          <h3>Players</h3>
          <div className="players-count">{onlineCount} online</div>
        </div>
        <div className="players-grid">
          {players.map((player) => (
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

      {gameState.phase === 'VOTING' && gameState.current_round?.revealed_cards && (
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

      <div className="chat-panel">
        <div className="chat-header">Chat</div>
        <div className="chat-messages">
          {chatMessages.length === 0 && (
            <div className="chat-empty">No messages yet.</div>
          )}
          {chatMessages.map((msg) => (
            <div
              key={msg.id}
              className={`chat-message ${msg.player_name === 'System' ? 'system' : ''}`}
            >
              <span className="chat-name">{msg.player_name}:</span>
              <span className="chat-text">{msg.message}</span>
            </div>
          ))}
          <div ref={chatEndRef} />
        </div>
        <div className="chat-input">
          <input
            type="text"
            value={chatText}
            onChange={(e) => setChatText(e.target.value)}
            placeholder="Type a message..."
            maxLength={200}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                handleSendChat();
              }
            }}
          />
          <button onClick={handleSendChat} disabled={!chatText.trim()}>
            Send
          </button>
        </div>
      </div>

      <PlayerHand
        cards={currentPlayer.hand}
        selectedCard={selectedCard}
        onCardSelect={setSelectedCard}
        canSelect={showClueForm || canSubmitCard()}
        canSubmit={canSubmitCard()}
        onSubmit={handleSubmitCard}
      />
    </div>
  );
};
