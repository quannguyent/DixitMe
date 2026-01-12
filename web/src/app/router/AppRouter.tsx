import React, { useEffect } from 'react';
import { Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import { HomePage } from '../../pages/HomePage';
import { GameBoard } from '../../features/game';
import { Auth } from '../../features/auth';
import { useGameStore } from '../../features/game/stores/gameStore';

export const AppRouter: React.FC = () => {
  const { gameState } = useGameStore();
  const navigate = useNavigate();

  // Handle programmatic navigation based on game state transitions
  useEffect(() => {
    if (gameState) {
      if (gameState.status === 'waiting') {
        // In lobby
        if (window.location.pathname !== '/') {
          navigate('/');
        }
      } else if (gameState.status === 'in_progress' || gameState.status === 'completed') {
        // In game
        const gamePath = `/game/${gameState.room_code}`;
        if (window.location.pathname !== gamePath) {
          navigate(gamePath);
        }
      }
    } else {
      // No active game, should be at home
      if (window.location.pathname.startsWith('/game/')) {
        navigate('/');
      }
    }
  }, [gameState, navigate]);

  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/login" element={<Auth />} />
      <Route path="/game/:roomCode" element={<GameBoard />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
};
