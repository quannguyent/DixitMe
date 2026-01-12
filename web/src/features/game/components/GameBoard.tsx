import React, { useState, useEffect, useRef } from 'react';
import { useGameStore } from '../stores/gameStore';
import { PlayerHand } from './PlayerHand';
import { VotingPhase } from './VotingPhase';
import { MainLayout } from '../../../layouts/MainLayout';
import styles from './GameBoard.module.css';

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
      <MainLayout>
        <div className={styles.container}>
          <div className={styles.loading}>
            <div className={styles.loadingContent}>
              <div className={styles.spinner}></div>
              <p>Loading game...</p>
            </div>
          </div>
        </div>
      </MainLayout>
    );
  }

  const players = Object.values(gameState.players);
  const onlineCount = players.filter((player) => player.is_connected).length;
  const phaseTag = gameState.phase.replace(/_/g, ' ');
  const isStoryteller = !!gameState.current_round && currentPlayer.id === gameState.current_round.storyteller_id;

  return (
    <MainLayout>
      <div className={styles.container}>
        <div className={styles.gameBoard}>
          <div className={styles.gameHeader}>
            <div className={styles.gameInfo}>
              <div className={styles.roomPill}>Room {gameState.room_code}</div>
              <h1>Table</h1>
              <div className={styles.connectionStatus}>
                <span className={`${styles.statusIndicator} ${isConnected ? styles.connected : styles.disconnected}`}>
                  {isConnected ? '🟢 Connected' : '🔴 Disconnected'}
                </span>
              </div>
            </div>

            <div className={styles.roundInfo}>
              <div className={styles.roundNumber}>Round {gameState.round_number} of {gameState.max_rounds}</div>
              {gameState.current_round && (
                <div className={styles.storytellerInfo}>
                  Storyteller: <strong>{getStorytellerName()}</strong>
                </div>
              )}
              <div className={styles.roleBadge}>{isStoryteller ? 'You are the storyteller' : 'You are a player'}</div>
            </div>

            <button onClick={handleLeaveGame} className={styles.leaveBtn}>
              Leave Game
            </button>
          </div>

          <div className={styles.phaseIndicator}>
            <h2>{getGamePhase()}</h2>
            <div className={styles.phaseMeta}>
              <span className={styles.phasePill}>{phaseTag}</span>
            </div>
            {gameState.current_round?.clue && (
              <div className={styles.clueDisplay}>
                <strong>Clue:</strong> "{gameState.current_round.clue}"
              </div>
            )}
          </div>

          <div className={styles.playersSection}>
            <div className={styles.playersHeader}>
              <h3>Players</h3>
              <div className={styles.playersCount}>{onlineCount} online</div>
            </div>
            <div className={styles.playersGrid}>
              {players.map((player) => (
                <div
                  key={player.id}
                  className={`${styles.playerCard} ${player.id === currentPlayer.id ? styles.currentPlayer : ''} ${player.id === gameState.current_round?.storyteller_id ? styles.storyteller : ''
                    }`}
                >
                  <div className={styles.playerName}>{player.name}</div>
                  <div className={styles.playerScore}>Score: {player.score}</div>
                  <div className={styles.playerStatus}>
                    {player.is_connected ? '🟢' : '🔴'}
                  </div>
                  {gameState.current_round?.submissions[player.id] && (
                    <div className={styles.submissionIndicator}>📤 Submitted</div>
                  )}
                  {gameState.current_round?.votes[player.id] && (
                    <div className={styles.voteIndicator}>🗳️ Voted</div>
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
            <div className={styles.clueFormSection}>
              <div className={styles.clueForm}>
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
                <div className={styles.clueActions}>
                  <span className={styles.selectedIndicator}>
                    {selectedCard ? `Card ${selectedCard} selected` : 'Select a card first'}
                  </span>
                  <button
                    onClick={handleSubmitClue}
                    disabled={!selectedCard || !clueText.trim()}
                    className={styles.submitClueBtn}
                  >
                    Submit Clue
                  </button>
                </div>
              </div>
            </div>
          )}

          <div className={styles.chatPanel}>
            <div className={styles.chatHeader}>Chat</div>
            <div className={styles.chatMessages}>
              {chatMessages.length === 0 && (
                <div className={styles.chatEmpty}>No messages yet.</div>
              )}
              {chatMessages.map((msg) => (
                <div
                  key={msg.id}
                  className={`${styles.chatMessage} ${msg.player_name === 'System' ? styles.system : ''}`}
                >
                  <span className={styles.chatName}>{msg.player_name}:</span>
                  <span className={styles.chatText}>{msg.message}</span>
                </div>
              ))}
              <div ref={chatEndRef} />
            </div>
            <div className={styles.chatInput}>
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
      </div>
    </MainLayout>
  );
};
