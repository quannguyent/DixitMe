import React, { useState, useEffect, useRef } from 'react';
import { useGameStore } from '../../game/stores/gameStore';
import { useAuthStore } from '../../auth/stores/authStore';
import { MainLayout } from '../../../layouts/MainLayout';
import styles from './Lobby.module.css';

export const Lobby: React.FC = () => {
  const [playerName, setPlayerName] = useState('');
  const [chatText, setChatText] = useState('');
  const autoJoinAttemptedRef = useRef(false);

  const { user } = useAuthStore();

  const {
    gameState,
    isConnected,
    isLoading,
    error,
    connect,
    joinGame,
    addBot,
    startGame,
    sendChat,
    chatMessages,
    leaveGame,
  } = useGameStore();

  const clearRoomCodeFromUrl = () => {
    try {
      const url = new URL(window.location.href);
      if (!url.searchParams.has('room_code')) {
        return;
      }
      url.searchParams.delete('room_code');
      window.history.replaceState({}, '', url.toString());
    } catch (error) {
      console.warn('Failed to clear room code from URL:', error);
    }
  };

  // Auto-fill player name from authenticated user
  useEffect(() => {
    if (user && !playerName) {
      setPlayerName(user.name);
    }
  }, [user, playerName]);

  useEffect(() => {
    if (!isConnected) {
      connect();
    }
  }, [isConnected, connect]);

  // Auto-join from URL if room_code is present.
  useEffect(() => {
    if (autoJoinAttemptedRef.current) {
      return;
    }
    const params = new URLSearchParams(window.location.search);
    const urlRoomCode = params.get('room_code');
    const trimmedName = playerName.trim();
    if (urlRoomCode && trimmedName && isConnected && !gameState) {
      autoJoinAttemptedRef.current = true;
      joinGame(urlRoomCode.toUpperCase(), trimmedName);
    }
  }, [playerName, isConnected, joinGame, gameState]);

  useEffect(() => {
    if (gameState) {
      return;
    }
    if (error && autoJoinAttemptedRef.current) {
      clearRoomCodeFromUrl();
    }
  }, [gameState, error]);

  const handleStartGame = () => {
    if (gameState) {
      startGame(gameState.room_code);
    }
  };

  const handleAddBot = () => {
    if (gameState) {
      addBot(gameState.room_code);
    }
  };

  const handleLeaveLobby = () => {
    if (!gameState) return;
    if (window.confirm('Leave the lobby?')) {
      leaveGame(gameState.room_code);
    }
  };

  const handleSendChat = () => {
    if (!gameState) return;
    if (!chatText.trim()) return;
    sendChat(gameState.room_code, chatText.trim());
    setChatText('');
  };

  if (gameState) {
    return (
      <MainLayout>
        <div className={styles.lobbyContainer}>
          <div className={styles.lobbyMain}>
            <div className={styles.gameLobby}>
              <div className={styles.roomBar}>
                <div className={styles.roomCode}>Room Code: <strong>{gameState.room_code}</strong></div>
                <button
                  onClick={handleLeaveLobby}
                  className={styles.leaveBtn}
                  disabled={isLoading}
                >
                  Leave Lobby
                </button>
              </div>

              <div className={styles.playersSection}>
                <h3>Players ({Object.keys(gameState.players).length}/6)</h3>
                <div className={styles.playersList}>
                  {Object.values(gameState.players).map((player) => (
                    <div key={player.id} className={styles.playerItem}>
                      <span className={styles.playerName}>{player.name}</span>
                      <span className={styles.playerStatus}>
                        {player.is_connected ? '🟢' : '🔴'}
                      </span>
                    </div>
                  ))}
                </div>
              </div>

              <div className={styles.chatPanel}>
                <div className={styles.chatHeader}>Chat</div>
                <div className={styles.chatMessages}>
                  {chatMessages.length === 0 && (
                    <div className={styles.chatEmpty}>No messages yet.</div>
                  )}
                  {chatMessages.map((msg) => (
                    <div key={msg.id} className={styles.chatMessage}>
                      <span className={styles.chatName}>{msg.player_name}:</span>
                      <span className={styles.chatText}>{msg.message}</span>
                    </div>
                  ))}
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

              <div className={styles.lobbyActions}>
                {gameState.status === 'waiting' && (
                  <>
                    {Object.keys(gameState.players).length < 6 && (
                      <button
                        onClick={handleAddBot}
                        className={styles.addBotBtn}
                        disabled={isLoading}
                      >
                        {isLoading ? 'Adding...' : 'Add Bot'}
                      </button>
                    )}
                    {Object.keys(gameState.players).length >= 3 && (
                      <button
                        onClick={handleStartGame}
                        className={styles.startGameBtn}
                        disabled={isLoading}
                      >
                        {isLoading ? 'Start Game' : 'Start Game'}
                      </button>
                    )}
                  </>
                )}
              </div>

              {error && (
                <div className={styles.errorMessage}>
                  {error}
                </div>
              )}
            </div>
          </div>
        </div>
      </MainLayout>
    );
  }

  // Fallback
  return (
    <MainLayout>
      <div className={styles.lobbyContainer}>
        <div className={`${styles.lobbyMain} glass-panel`} style={{ flexDirection: 'column', alignItems: 'center' }}>
          <h2>Lobby</h2>
          <p>Please join a game from the Home Page.</p>
        </div>
      </div>
    </MainLayout>
  );
};
