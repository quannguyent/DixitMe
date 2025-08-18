import React, { useState, useEffect } from 'react';
import { useGameStore } from '../store/gameStore';

const Lobby: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'join' | 'create'>('join');
  const [roomCode, setRoomCode] = useState('');
  const [playerName, setPlayerName] = useState('');

  const {
    gameState,
    isConnected,
    isLoading,
    error,
    connect,
    createGame,
    joinGame,
    startGame,
    setError,
  } = useGameStore();

  useEffect(() => {
    if (!isConnected) {
      connect();
    }
  }, [isConnected, connect]);

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

  const generateRoomCode = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    let result = '';
    for (let i = 0; i < 4; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    setRoomCode(result);
  };

  if (gameState) {
    return (
      <div className="lobby-container">
        <div className="game-lobby">
          <div className="lobby-header">
            <h2>Game Lobby</h2>
            <div className="room-code">Room Code: <strong>{gameState.room_code}</strong></div>
            <div className="connection-status">
              <span className={`status-indicator ${isConnected ? 'connected' : 'disconnected'}`}>
                {isConnected ? '🟢 Connected' : '🔴 Disconnected'}
              </span>
            </div>
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

          <div className="game-info">
            <p>Game Status: <strong>{gameState.status}</strong></p>
            {gameState.status === 'waiting' && (
              <p>Waiting for players to join... (Minimum 3 players required)</p>
            )}
          </div>

          <div className="lobby-actions">
            {gameState.status === 'waiting' && Object.keys(gameState.players).length >= 3 && (
              <button
                onClick={handleStartGame}
                className="start-game-btn"
                disabled={isLoading}
              >
                {isLoading ? 'Starting...' : 'Start Game'}
              </button>
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
        <div className="form-header">
          <h1>DixitMe</h1>
          <p>Online Dixit Card Game</p>
          <div className="connection-status">
            <span className={`status-indicator ${isConnected ? 'connected' : 'disconnected'}`}>
              {isConnected ? '🟢 Connected' : '🔴 Connecting...'}
            </span>
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

export default Lobby;
