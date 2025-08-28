import React, { useState } from 'react';
import { useGameStore } from '../store/gameStore';
import { useAuthStore } from '../store/authStore';
import styles from './GameControls.module.css';

interface GameControlsProps {
  roomCode: string;
  isLobbyManager?: boolean;
  onGameDeleted?: () => void;
  onPlayerLeft?: () => void;
}

export const GameControls: React.FC<GameControlsProps> = ({
  roomCode,
  isLobbyManager = false,
  onGameDeleted,
  onPlayerLeft
}) => {
  const { deleteGame, leaveGame, removePlayer, isLoading, error } = useGameStore();
  const [showConfirmDialog, setShowConfirmDialog] = useState<'delete' | 'leave' | null>(null);
  const [isActioning, setIsActioning] = useState(false);

  const handleDeleteGame = async () => {
    if (!isLobbyManager) return;
    
    setIsActioning(true);
    try {
      await deleteGame(roomCode);
      setShowConfirmDialog(null);
      onGameDeleted?.();
      // Optionally redirect to home
      window.location.href = '/';
    } catch (error) {
      console.error('Failed to delete game:', error);
    } finally {
      setIsActioning(false);
    }
  };

  const handleLeaveGame = async () => {
    setIsActioning(true);
    try {
      await leaveGame(roomCode);
      setShowConfirmDialog(null);
      onPlayerLeft?.();
      // Optionally redirect to home
      window.location.href = '/';
    } catch (error) {
      console.error('Failed to leave game:', error);
    } finally {
      setIsActioning(false);
    }
  };



  const ConfirmDialog = ({ type }: { type: 'delete' | 'leave' }) => (
    <div className={styles.overlay}>
      <div className={styles.dialog}>
        <h3>
          {type === 'delete' ? 'Delete Game?' : 'Leave Game?'}
        </h3>
        <p>
          {type === 'delete' 
            ? 'This will permanently delete the game room and remove all players. This action cannot be undone.'
            : 'Are you sure you want to leave this game room?'
          }
        </p>
        <div className={styles.dialogButtons}>
          <button
            className={styles.cancelButton}
            onClick={() => setShowConfirmDialog(null)}
            disabled={isActioning}
          >
            Cancel
          </button>
          <button
            className={`${styles.confirmButton} ${type === 'delete' ? styles.deleteButton : styles.leaveButton}`}
            onClick={type === 'delete' ? handleDeleteGame : handleLeaveGame}
            disabled={isActioning}
          >
            {isActioning ? 'Processing...' : (type === 'delete' ? 'Delete Game' : 'Leave Game')}
          </button>
        </div>
      </div>
    </div>
  );

  return (
    <div className={styles.gameControls}>
      <div className={styles.controlsHeader}>
        <h4>Game Controls</h4>
        <span className={styles.roomCode}>Room: {roomCode}</span>
      </div>

      {error && (
        <div className={styles.error}>
          {error}
        </div>
      )}

      <div className={styles.buttonGroup}>
        {isLobbyManager && (
          <button
            className={`${styles.controlButton} ${styles.deleteButton}`}
            onClick={() => setShowConfirmDialog('delete')}
            disabled={isLoading}
            title="Delete this game room (Manager only)"
          >
            🗑️ Delete Room
          </button>
        )}

        <button
          className={`${styles.controlButton} ${styles.leaveButton}`}
          onClick={() => setShowConfirmDialog('leave')}
          disabled={isLoading}
          title="Leave this game room"
        >
          🚪 Leave Game
        </button>
      </div>

      <div className={styles.info}>
        {isLobbyManager ? (
          <p className={styles.managerInfo}>
            ✨ You are the lobby manager and can delete the room or remove players
          </p>
        ) : (
          <p className={styles.playerInfo}>
            👤 You can leave the game at any time
          </p>
        )}
      </div>

      {showConfirmDialog && <ConfirmDialog type={showConfirmDialog} />}
    </div>
  );
};

export default GameControls;
