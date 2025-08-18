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

        <style jsx>{`
          .lobby-container {
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            padding: 20px;
          }

          .game-lobby {
            background: white;
            border-radius: 12px;
            padding: 2rem;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
            min-width: 400px;
            max-width: 600px;
          }

          .lobby-header {
            text-align: center;
            margin-bottom: 2rem;
          }

          .lobby-header h2 {
            margin: 0 0 1rem 0;
            color: #333;
          }

          .room-code {
            font-size: 1.2rem;
            margin-bottom: 0.5rem;
          }

          .connection-status {
            font-size: 0.9rem;
          }

          .status-indicator.connected {
            color: #22c55e;
          }

          .status-indicator.disconnected {
            color: #ef4444;
          }

          .players-section {
            margin-bottom: 2rem;
          }

          .players-section h3 {
            margin: 0 0 1rem 0;
            color: #333;
          }

          .players-list {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
          }

          .player-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 0.75rem;
            background: #f8fafc;
            border-radius: 8px;
            border: 1px solid #e2e8f0;
          }

          .player-name {
            font-weight: 500;
          }

          .game-info {
            margin-bottom: 2rem;
            text-align: center;
          }

          .game-info p {
            margin: 0.5rem 0;
            color: #666;
          }

          .lobby-actions {
            display: flex;
            justify-content: center;
            gap: 1rem;
          }

          .start-game-btn {
            background: #22c55e;
            color: white;
            border: none;
            padding: 0.75rem 2rem;
            border-radius: 8px;
            font-size: 1rem;
            font-weight: 500;
            cursor: pointer;
            transition: background-color 0.2s;
          }

          .start-game-btn:hover:not(:disabled) {
            background: #16a34a;
          }

          .start-game-btn:disabled {
            background: #9ca3af;
            cursor: not-allowed;
          }

          .error-message {
            background: #fef2f2;
            color: #dc2626;
            padding: 0.75rem;
            border-radius: 8px;
            margin-top: 1rem;
            text-align: center;
            border: 1px solid #fecaca;
          }
        `}</style>
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

      <style jsx>{`
        .lobby-container {
          display: flex;
          justify-content: center;
          align-items: center;
          min-height: 100vh;
          padding: 20px;
        }

        .lobby-form {
          background: white;
          border-radius: 12px;
          padding: 2rem;
          box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
          min-width: 400px;
          max-width: 500px;
        }

        .form-header {
          text-align: center;
          margin-bottom: 2rem;
        }

        .form-header h1 {
          margin: 0 0 0.5rem 0;
          color: #333;
          font-size: 2.5rem;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
        }

        .form-header p {
          margin: 0 0 1rem 0;
          color: #666;
          font-size: 1.1rem;
        }

        .connection-status {
          font-size: 0.9rem;
        }

        .status-indicator.connected {
          color: #22c55e;
        }

        .status-indicator.disconnected {
          color: #ef4444;
        }

        .tabs {
          display: flex;
          border-bottom: 1px solid #e2e8f0;
          margin-bottom: 2rem;
        }

        .tab {
          flex: 1;
          padding: 0.75rem;
          background: none;
          border: none;
          cursor: pointer;
          font-size: 1rem;
          color: #666;
          transition: all 0.2s;
        }

        .tab.active {
          color: #667eea;
          border-bottom: 2px solid #667eea;
        }

        .tab:hover {
          color: #667eea;
        }

        .game-form {
          display: flex;
          flex-direction: column;
          gap: 1.5rem;
        }

        .form-group {
          display: flex;
          flex-direction: column;
        }

        .form-group label {
          margin-bottom: 0.5rem;
          font-weight: 500;
          color: #333;
        }

        .form-group input {
          padding: 0.75rem;
          border: 1px solid #d1d5db;
          border-radius: 8px;
          font-size: 1rem;
          transition: border-color 0.2s;
        }

        .form-group input:focus {
          outline: none;
          border-color: #667eea;
          box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }

        .room-code-input {
          display: flex;
          gap: 0.5rem;
        }

        .room-code-input input {
          flex: 1;
        }

        .generate-btn {
          background: #f3f4f6;
          color: #374151;
          border: 1px solid #d1d5db;
          padding: 0.75rem 1rem;
          border-radius: 8px;
          cursor: pointer;
          transition: background-color 0.2s;
        }

        .generate-btn:hover {
          background: #e5e7eb;
        }

        .submit-btn {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          color: white;
          border: none;
          padding: 0.75rem;
          border-radius: 8px;
          font-size: 1rem;
          font-weight: 500;
          cursor: pointer;
          transition: opacity 0.2s;
        }

        .submit-btn:hover:not(:disabled) {
          opacity: 0.9;
        }

        .submit-btn:disabled {
          opacity: 0.5;
          cursor: not-allowed;
        }

        .error-message {
          background: #fef2f2;
          color: #dc2626;
          padding: 0.75rem;
          border-radius: 8px;
          margin-top: 1rem;
          text-align: center;
          border: 1px solid #fecaca;
        }
      `}</style>
    </div>
  );
};

export default Lobby;
