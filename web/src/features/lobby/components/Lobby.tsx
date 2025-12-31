import React, { useState, useEffect, useRef } from 'react';
import { useGameStore } from '../../game/stores/gameStore';
import { useAuthStore } from '../../auth/stores/authStore';
import { UserInfo } from '../../auth';
import styles from './Lobby.module.css';

export const Lobby: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'join' | 'create'>('join');
  const [roomCode, setRoomCode] = useState('');
  const [playerName, setPlayerName] = useState('');
  const [chatText, setChatText] = useState('');
  const autoJoinAttemptedRef = useRef(false);
  const [isAutoJoining, setIsAutoJoining] = useState(false);

  const { user } = useAuthStore();

  const {
    gameState,
    isConnected,
    isLoading,
    error,
    connect,
    createGame,
    joinGame,
    addBot,
    startGame,
    setError,
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
      setIsAutoJoining(true);
      joinGame(urlRoomCode.toUpperCase(), trimmedName);
    }
  }, [playerName, isConnected, joinGame, gameState]);

  useEffect(() => {
    if (gameState) {
      setIsAutoJoining(false);
      return;
    }
    if (error && autoJoinAttemptedRef.current) {
      setIsAutoJoining(false);
      clearRoomCodeFromUrl();
    }
  }, [gameState, error]);

  const handleCreateGame = (e: React.FormEvent) => {
    e.preventDefault();
    if (!playerName.trim() || !roomCode.trim()) {
      setError('Please fill in all fields');
      return;
    }
    if (!isConnected) {
      setError('Not connected to server');
      return;
    }
    createGame(roomCode.toUpperCase(), playerName.trim());
  };

  const handleJoinGame = (e: React.FormEvent) => {
    e.preventDefault();
    if (!playerName.trim() || !roomCode.trim()) {
      setError('Please fill in all fields');
      return;
    }
    if (!isConnected) {
      setError('Not connected to server');
      return;
    }
    joinGame(roomCode.toUpperCase(), playerName.trim());
  };

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

  const generateRoomCode = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    let result = '';
    for (let i = 0; i < 6; i++) { // Increased to 6 characters for more uniqueness
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    // Add timestamp suffix to make it more unique
    const timestamp = Date.now().toString().slice(-3);
    result = result.substring(0, 3) + timestamp;
    setRoomCode(result);
  };

  if (gameState) {
    return (
      <div className="lobby-container">
        <div className="game-lobby">
          <UserInfo />
          <div className="lobby-header">
            <h2>Game Lobby</h2>
            <div className="room-code">Room Code: <strong>{gameState.room_code}</strong></div>
            <div className="connection-status">
              <span className={`status-indicator ${isConnected ? 'connected' : 'disconnected'}`}>
                {isConnected ? '🟢 Connected' : '🔴 Disconnected'}
              </span>
              {isAutoJoining && (
                <span className={styles.statusText}>Rejoining room...</span>
              )}
            </div>
            <button
              onClick={handleLeaveLobby}
              className={styles.leaveBtn}
              disabled={isLoading}
            >
              Leave Lobby
            </button>
          </div>

          <div className="players-section">
            <h3>Players ({Object.keys(gameState.players).length}/6)</h3>
            <div className="players-list">
              {Object.values(gameState.players).map((player) => (
                <div key={player.id} className="player-item">
                  <span className="player-name">{player.name}</span>
                  <span className={`player-status ${player.is_connected ? 'online' : 'offline'}`}>
                    {player.is_connected ? '🟢' : '🔴'}
                  </span>
                </div>
              ))}
            </div>
          </div>

          <div className="chat-panel">
            <div className="chat-header">Chat</div>
            <div className="chat-messages">
              {chatMessages.length === 0 && (
                <div className="chat-empty">No messages yet.</div>
              )}
              {chatMessages.map((msg) => (
                <div key={msg.id} className="chat-message">
                  <span className="chat-name">{msg.player_name}:</span>
                  <span className="chat-text">{msg.message}</span>
                </div>
              ))}
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

          <div className="game-info">
            <p>Game Status: <strong>{gameState.status}</strong></p>
            {gameState.status === 'waiting' && (
              <p>Waiting for players to join... (Minimum 3 players required)</p>
            )}
          </div>

          <div className="lobby-actions">
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
                    {isLoading ? 'Starting...' : 'Start Game'}
                  </button>
                )}
              </>
            )}
          </div>

          {error && (
            <div className="error-message">
              {error}
            </div>
          )}
        </div>
      </div>
    );
  }

  return (
    <div className="lobby-container">
      <div className="lobby-form">
        <UserInfo />
        <div className="form-header">
          <h1>DixitMe</h1>
          <p>Online Dixit Card Game</p>
          <div className="connection-status">
            <span className={`status-indicator ${isConnected ? 'connected' : 'disconnected'}`}>
              {isConnected ? '🟢 Connected' : '🔴 Connecting...'}
            </span>
            {isAutoJoining && (
              <span className={styles.statusText}>Rejoining room...</span>
            )}
          </div>
        </div>

        <div className="tabs">
          <button
            className={`tab ${activeTab === 'join' ? 'active' : ''}`}
            onClick={() => setActiveTab('join')}
          >
            Join Game
          </button>
          <button
            className={`tab ${activeTab === 'create' ? 'active' : ''}`}
            onClick={() => setActiveTab('create')}
          >
            Create Game
          </button>
        </div>

        {activeTab === 'join' && (
          <form onSubmit={handleJoinGame} className="game-form">
            <div className="form-group">
              <label htmlFor="join-name">Your Name</label>
              <input
                id="join-name"
                type="text"
                value={playerName}
                onChange={(e) => setPlayerName(e.target.value)}
                placeholder="Enter your name"
                maxLength={20}
                required
              />
            </div>
            <div className="form-group">
              <label htmlFor="join-room">Room Code</label>
              <input
                id="join-room"
                type="text"
                value={roomCode}
                onChange={(e) => setRoomCode(e.target.value.toUpperCase())}
                placeholder="Enter room code"
                maxLength={4}
                style={{ textTransform: 'uppercase' }}
                required
              />
            </div>
            <button
              type="submit"
              className="submit-btn"
              disabled={isLoading || !isConnected}
            >
              {isLoading ? 'Joining...' : 'Join Game'}
            </button>
          </form>
        )}

        {activeTab === 'create' && (
          <form onSubmit={handleCreateGame} className="game-form">
            <div className="form-group">
              <label htmlFor="create-name">Your Name</label>
              <input
                id="create-name"
                type="text"
                value={playerName}
                onChange={(e) => setPlayerName(e.target.value)}
                placeholder="Enter your name"
                maxLength={20}
                required
              />
            </div>
            <div className="form-group">
              <label htmlFor="create-room">Room Code</label>
              <div className="room-code-input">
                <input
                  id="create-room"
                  type="text"
                  value={roomCode}
                  onChange={(e) => setRoomCode(e.target.value.toUpperCase())}
                  placeholder="Enter room code"
                  maxLength={4}
                  style={{ textTransform: 'uppercase' }}
                  required
                />
                <button
                  type="button"
                  onClick={generateRoomCode}
                  className="generate-btn"
                >
                  Generate
                </button>
              </div>
            </div>
            <button
              type="submit"
              className="submit-btn"
              disabled={isLoading || !isConnected}
            >
              {isLoading ? 'Creating...' : 'Create Game'}
            </button>
          </form>
        )}

        {error && (
          <div className="error-message">
            {error}
          </div>
        )}
      </div>
    </div>
  );
};
