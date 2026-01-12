import React, { useState, useEffect, useRef } from 'react';
import { useGameStore } from '../features/game/stores/gameStore';
import { Lobby } from '../features/lobby/components/Lobby';
import { MainLayout } from '../layouts/MainLayout';
import { GameBoard } from '../features/game';
import { useAuthStore } from '../features/auth/stores/authStore';
import styles from './HomePage.module.css';

export const HomePage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'join' | 'create'>('join');
  const [showAuth, setShowAuth] = useState(false);
  const [authTab, setAuthTab] = useState<'login' | 'register'>('login');
  const [gameForm, setGameForm] = useState({
    playerName: '',
    roomCode: '',
  });
  const [authForm, setAuthForm] = useState({
    name: '',
    email: '',
    password: '',
    confirmPassword: '',
  });
  const [isAutoJoining, setIsAutoJoining] = useState(false);

  const {
    gameState,
    isConnected,
    isLoading: gameLoading,
    error: gameError,
    connect,
    createGame,
    joinGame,
    setError: setGameError,
  } = useGameStore();

  const autoJoinAttemptedRef = useRef(false);

  const {
    user,
    isLoading: authLoading,
    error: authError,
    login,
    register,
    loginAsGuest,
    clearError,
  } = useAuthStore();

  const shouldShowLobby = !!gameState && gameState.status === 'waiting';
  const shouldShowGame = !!gameState && gameState.status !== 'waiting';

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

  // Auto-connect to WebSocket
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
    const roomCode = params.get('room_code');
    const playerName = gameForm.playerName.trim();
    if (roomCode && playerName && isConnected && !gameState) {
      autoJoinAttemptedRef.current = true;
      setIsAutoJoining(true);
      joinGame(roomCode.toUpperCase(), playerName);
    }
  }, [gameForm.playerName, isConnected, joinGame, gameState]);

  useEffect(() => {
    if (gameState) {
      setIsAutoJoining(false);
      return;
    }
    if (gameError && autoJoinAttemptedRef.current) {
      setIsAutoJoining(false);
      clearRoomCodeFromUrl();
    }
  }, [gameState, gameError]);

  // Auto-fill player name if user is authenticated
  useEffect(() => {
    if (user && !gameForm.playerName) {
      setGameForm(prev => ({ ...prev, playerName: user.name }));
    }
  }, [user, gameForm.playerName]);

  // Try to restore guest name from localStorage if no user
  useEffect(() => {
    if (!user) {
      const savedGuestName = localStorage.getItem('dixitme-guest-name');
      if (savedGuestName && !gameForm.playerName) {
        setGameForm(prev => ({ ...prev, playerName: savedGuestName }));
      }
    }
  }, [user, gameForm.playerName]);

  const handleGameFormChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setGameForm({
      ...gameForm,
      [e.target.name]: e.target.value,
    });
  };

  const handleAuthFormChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setAuthForm({
      ...authForm,
      [e.target.name]: e.target.value,
    });
  };

  const generateRoomCode = () => {
    const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
    let result = '';
    for (let i = 0; i < 6; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    const timestamp = Date.now().toString().slice(-3);
    result = result.substring(0, 3) + timestamp;
    setGameForm(prev => ({ ...prev, roomCode: result }));
  };

  const handleCreateGame = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!gameForm.playerName.trim() || !gameForm.roomCode.trim()) {
      setGameError('Please fill in all fields');
      return;
    }
    if (!isConnected) {
      setGameError('Not connected to server');
      return;
    }

    // Auto-login as guest if not authenticated
    if (!user) {
      try {
        localStorage.setItem('dixitme-guest-name', gameForm.playerName.trim());
        await loginAsGuest(gameForm.playerName.trim());
      } catch (error) {
        console.error('Guest login failed:', error);
        return;
      }
    }

    createGame(gameForm.roomCode.toUpperCase(), gameForm.playerName.trim());
  };

  const handleJoinGame = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!gameForm.playerName.trim() || !gameForm.roomCode.trim()) {
      setGameError('Please fill in all fields');
      return;
    }
    if (!isConnected) {
      setGameError('Not connected to server');
      return;
    }

    // Auto-login as guest if not authenticated
    if (!user) {
      try {
        localStorage.setItem('dixitme-guest-name', gameForm.playerName.trim());
        await loginAsGuest(gameForm.playerName.trim());
      } catch (error) {
        console.error('Guest login failed:', error);
        return;
      }
    }

    joinGame(gameForm.roomCode.toUpperCase(), gameForm.playerName.trim());
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await login(authForm.email, authForm.password);
      setShowAuth(false);
      setAuthForm({ name: '', email: '', password: '', confirmPassword: '' });
    } catch (error) {
      // Error handled by store
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    if (authForm.password !== authForm.confirmPassword) {
      return;
    }
    try {
      await register(authForm.name, authForm.email, authForm.password);
      setShowAuth(false);
      setAuthForm({ name: '', email: '', password: '', confirmPassword: '' });
    } catch (error) {
      // Error handled by store
    }
  };



  // Clear errors after 5 seconds
  useEffect(() => {
    if (gameError) {
      const timer = setTimeout(() => setGameError(null), 5000);
      return () => clearTimeout(timer);
    }
  }, [gameError, setGameError]);

  useEffect(() => {
    if (authError) {
      const timer = setTimeout(() => clearError(), 5000);
      return () => clearTimeout(timer);
    }
  }, [authError, clearError]);

  if (shouldShowLobby) {
    return <Lobby />;
  }

  if (shouldShowGame) {
    return <GameBoard />;
  }

  return (
    <MainLayout
      headerActions={
        !user && (
          <button
            onClick={() => setShowAuth(true)}
            className={styles.signInBtn}
          >
            🔑 Sign In
          </button>
        )
      }
    >
      {/* Main game section */}
      <div className={styles.gameSection}>
        <div className={styles.connectionStatus}>
          <span className={`${styles.statusIndicator} ${isConnected ? styles.connected : styles.disconnected}`}>
            {isConnected ? '🟢 Connected' : '🔴 Connecting...'}
          </span>
          {isAutoJoining && (
            <span className={styles.statusText}>Rejoining room...</span>
          )}
        </div>

        <div className={styles.gameTabs}>
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
                name="playerName"
                type="text"
                value={gameForm.playerName}
                onChange={handleGameFormChange}
                placeholder="Enter your name"
                maxLength={20}
                required
              />
            </div>
            <div className={styles.formGroup}>
              <label htmlFor="join-room">Room Code</label>
              <input
                id="join-room"
                name="roomCode"
                type="text"
                value={gameForm.roomCode}
                onChange={handleGameFormChange}
                placeholder="Enter room code (e.g., ABC123)"
                maxLength={6}
                style={{ textTransform: 'uppercase' }}
                required
              />
            </div>
            <button
              type="submit"
              className={styles.gameBtn}
              disabled={gameLoading || !isConnected}
            >
              {gameLoading ? 'Joining...' : 'Join Game'}
            </button>
          </form>
        )}

        {activeTab === 'create' && (
          <form onSubmit={handleCreateGame} className={styles.gameForm}>
            <div className={styles.formGroup}>
              <label htmlFor="create-name">Your Name</label>
              <input
                id="create-name"
                name="playerName"
                type="text"
                value={gameForm.playerName}
                onChange={handleGameFormChange}
                placeholder="Enter your name"
                maxLength={20}
                required
              />
            </div>
            <div className={styles.formGroup}>
              <label htmlFor="create-room">Room Code</label>
              <div className={styles.roomCodeGroup}>
                <input
                  id="create-room"
                  name="roomCode"
                  type="text"
                  value={gameForm.roomCode}
                  onChange={handleGameFormChange}
                  placeholder="Enter or generate room code"
                  maxLength={6}
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
              className={styles.gameBtn}
              disabled={gameLoading || !isConnected}
            >
              {gameLoading ? 'Creating...' : 'Create Game'}
            </button>
          </form>
        )}

        {gameError && (
          <div className={styles.errorMessage}>
            {gameError}
          </div>
        )}
      </div>

      {/* Optional sign-in overlay */}
      {showAuth && (
        <div className={styles.authOverlay}>
          <div className={styles.authModal}>
            <div className={styles.authHeader}>
              <h3>Sign In to DixitMe</h3>
              <button
                onClick={() => setShowAuth(false)}
                className={styles.closeBtn}
              >
                ×
              </button>
            </div>

            <div className={styles.authTabs}>
              <button
                className={`${styles.authTab} ${authTab === 'login' ? styles.active : ''}`}
                onClick={() => setAuthTab('login')}
              >
                Login
              </button>
              <button
                className={`${styles.authTab} ${authTab === 'register' ? styles.active : ''}`}
                onClick={() => setAuthTab('register')}
              >
                Register
              </button>
            </div>

            {authTab === 'login' && (
              <form onSubmit={handleLogin} className={styles.authForm}>
                <div className={styles.formGroup}>
                  <input
                    name="email"
                    type="email"
                    value={authForm.email}
                    onChange={handleAuthFormChange}
                    placeholder="Email"
                    required
                  />
                </div>
                <div className={styles.formGroup}>
                  <input
                    name="password"
                    type="password"
                    value={authForm.password}
                    onChange={handleAuthFormChange}
                    placeholder="Password"
                    required
                  />
                </div>
                <button
                  type="submit"
                  className={styles.authBtn}
                  disabled={authLoading}
                >
                  {authLoading ? 'Logging in...' : 'Login'}
                </button>
              </form>
            )}

            {authTab === 'register' && (
              <form onSubmit={handleRegister} className={styles.authForm}>
                <div className={styles.formGroup}>
                  <input
                    name="name"
                    type="text"
                    value={authForm.name}
                    onChange={handleAuthFormChange}
                    placeholder="Name"
                    required
                  />
                </div>
                <div className={styles.formGroup}>
                  <input
                    name="email"
                    type="email"
                    value={authForm.email}
                    onChange={handleAuthFormChange}
                    placeholder="Email"
                    required
                  />
                </div>
                <div className={styles.formGroup}>
                  <input
                    name="password"
                    type="password"
                    value={authForm.password}
                    onChange={handleAuthFormChange}
                    placeholder="Password"
                    minLength={6}
                    required
                  />
                </div>
                <div className={styles.formGroup}>
                  <input
                    name="confirmPassword"
                    type="password"
                    value={authForm.confirmPassword}
                    onChange={handleAuthFormChange}
                    placeholder="Confirm Password"
                    minLength={6}
                    required
                  />
                </div>
                <button
                  type="submit"
                  className={styles.authBtn}
                  disabled={authLoading}
                >
                  {authLoading ? 'Creating account...' : 'Create Account'}
                </button>
              </form>
            )}

            {authError && (
              <div className={styles.errorMessage}>
                {authError}
              </div>
            )}

            <div className={styles.authFooter}>
              <p>Sign in to save your game history and progress!</p>
            </div>
          </div>
        </div>
      )}
    </MainLayout>
  );
};
