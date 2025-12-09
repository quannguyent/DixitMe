import React, { useEffect } from 'react';
import { Auth } from './features/auth';
import { useGameStore } from './features/game';

function App() {
  const { setCards } = useGameStore();

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
        <Auth />
      </div>

  );
}

export default App;
