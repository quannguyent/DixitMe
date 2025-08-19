import React, { useEffect } from 'react';
import { useGameStore } from './store/gameStore';
import { useAuthStore } from './store/authStore';
import GameLanding from './components/GameLanding';
import Lobby from './components/Lobby';
import GameBoard from './components/GameBoard';

function App() {
  const { gameState, setCards } = useGameStore();

  useEffect(() => {
    // Load cards from API
    fetch('/api/v1/cards')
      .then(response => response.json())
      .then(data => {
        if (data.cards) {
          setCards(data.cards);
        }
      })
      .catch(error => {
        console.error('Failed to load cards:', error);
        // Generate fallback cards
        const fallbackCards = Array.from({ length: 84 }, (_, i) => ({
          id: i + 1,
          url: `/cards/${i + 1}.jpg`
        }));
        setCards(fallbackCards);
      });
  }, [setCards]);

  return (
    <div className="App">
      {gameState && gameState.status !== 'waiting' ? (
        <GameBoard />
      ) : gameState && gameState.status === 'waiting' ? (
        <Lobby />
      ) : (
        <GameLanding />
      )}
    </div>
  );
}

export default App;
