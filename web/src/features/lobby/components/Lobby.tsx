import React, { useState, useEffect, useRef } from 'react';
import { useGameStore } from '../../game/stores/gameStore';
import { useAuthStore } from '../../auth/stores/authStore';
import { UserInfo } from '../../auth';
import styles from './Lobby.module.css';
import homeStyles from '../../../pages/HomePage.module.css';

export const Lobby: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'join' | 'create'>('join');
  const [roomCode, setRoomCode] = useState('');
  const [playerName, setPlayerName] = useState('');
  const [chatText, setChatText] = useState('');
  const autoJoinAttemptedRef = useRef(false);
  const [isAutoJoining, setIsAutoJoining] = useState(false);

  const { user, logout } = useAuthStore();

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

  const handleLogout = async () => {
    if (window.confirm('Are you sure you want to logout?')) {
      await logout();
    }
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
      <div className={homeStyles.container}>
        <div className={homeStyles.header}>
          <div className={homeStyles.logo}>
            <h1>DixitMe</h1>
            <p>Online Dixit Card Game</p>
          </div>

          <div className={homeStyles.userSection}>
            {user ? (
              <div className={homeStyles.userInfo}>
                <div className={homeStyles.userDetails}>
                  <span className={homeStyles.userName}>{user.name}</span>
                  <span className={homeStyles.userType}>
                    {user.auth_type === 'guest' ? '👤 Guest' : '🔐 Member'}
                  </span>
                </div>
                <div className={homeStyles.userActions}>
                  <button className={homeStyles.historyBtn} title="View game history">
                    📊 History
                  </button>
                  <button onClick={handleLogout} className={homeStyles.logoutBtn}>
                    ↗️ Logout
                  </button>
                </div>
              </div>
            ) : (
              <div className={homeStyles.userInfo}>
                <div className={homeStyles.userDetails}>
                  <span className={homeStyles.userName}>Guest</span>
                  <span className={homeStyles.userType}>👤 Guest</span>
                </div>
              </div>
            )}
          </div>
        </div>

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
                      {isLoading ? 'Starting...' : 'Start Game'}
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
    );
  }

  return (
    <div className={styles.lobbyContainer}>
      <div className={styles.lobbyForm}>
        <UserInfo />
        <div className={styles.formHeader}>
          <h1>DixitMe</h1>
          <p>Online Dixit Card Game</p>
          <div className={styles.connectionStatus}>
            <span
              className={`${styles.statusIndicator} ${
                isConnected ? styles.connected : styles.disconnected
              }`}
            >
              {isConnected ? '🟢 Connected' : '🔴 Connecting...'}
            </span>
            {isAutoJoining && (
              <span className={styles.statusText}>Rejoining room...</span>
            )}
          </div>
        </div>

        <div className={styles.tabs}>
          <button
            className={`${styles.tab} ${activeTab === 'join' ? styles.active : ''}`}
            onClick={() => setActiveTab('join')}
          >
            Join Game
          </button>
          <button
            className={`${styles.tab} ${activeTab === 'create' ? styles.active : ''}`}
            onClick={() => setActiveTab('create')}
          >
            Create Game
          </button>
        </div>

        {activeTab === 'join' && (
          <form onSubmit={handleJoinGame} className={styles.gameForm}>
            <div className={styles.formGroup}>
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
            <div className={styles.formGroup}>
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
              className={styles.submitBtn}
              disabled={isLoading || !isConnected}
            >
              {isLoading ? 'Joining...' : 'Join Game'}
            </button>
          </form>
        )}

        {activeTab === 'create' && (
          <form onSubmit={handleCreateGame} className={styles.gameForm}>
            <div className={styles.formGroup}>
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
            <div className={styles.formGroup}>
              <label htmlFor="create-room">Room Code</label>
              <div className={styles.roomCodeInput}>
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
                  className={styles.generateBtn}
                >
                  Generate
                </button>
              </div>
            </div>
            <button
              type="submit"
              className={styles.submitBtn}
              disabled={isLoading || !isConnected}
            >
              {isLoading ? 'Creating...' : 'Create Game'}
            </button>
          </form>
        )}

        {error && (
          <div className={styles.errorMessage}>
            {error}
          </div>
        )}
      </div>
    </div>
  );
};
