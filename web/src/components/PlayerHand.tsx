import React from 'react';
import Card from './Card';

interface PlayerHandProps {
  cards: number[];
  selectedCard: number | null;
  onCardSelect: (cardId: number | null) => void;
  canSelect: boolean;
  canSubmit: boolean;
  onSubmit: (cardId: number) => void;
}

const PlayerHand: React.FC<PlayerHandProps> = ({
  cards,
  selectedCard,
  onCardSelect,
  canSelect,
  canSubmit,
  onSubmit,
}) => {
  const handleCardClick = (cardId: number) => {
    if (!canSelect) return;

    if (canSubmit && selectedCard === cardId) {
      // Double-click to submit
      onSubmit(cardId);
    } else {
      // Single-click to select
      onCardSelect(cardId === selectedCard ? null : cardId);
    }
  };

  const handleSubmit = () => {
    if (selectedCard && canSubmit) {
      onSubmit(selectedCard);
    }
  };

  if (cards.length === 0) {
    return (
      <div className="player-hand">
        <div className="hand-header">
          <h3>Your Hand</h3>
        </div>
        <div className="empty-hand">
          <p>No cards in hand</p>
        </div>
      </div>
    );
  }

  return (
    <div className="player-hand">
      <div className="hand-header">
        <h3>Your Hand ({cards.length} cards)</h3>
        {canSelect && (
          <div className="hand-instructions">
            {canSubmit ? (
              selectedCard ? (
                <span>Click selected card again or use Submit button to confirm</span>
              ) : (
                <span>Select a card to submit</span>
              )
            ) : (
              <span>Select a card for your clue</span>
            )}
          </div>
        )}
      </div>

      <div className="cards-container">
        <div className="cards-scroll">
          {cards.map((cardId) => (
            <div key={cardId} className="card-wrapper">
              <Card
                id={cardId}
                isSelected={selectedCard === cardId}
                isClickable={canSelect}
                onClick={handleCardClick}
                size="medium"
              />
            </div>
          ))}
        </div>
      </div>

      {canSubmit && selectedCard && (
        <div className="submit-section">
          <button onClick={handleSubmit} className="submit-card-btn">
            Submit Card {selectedCard}
          </button>
        </div>
      )}
    </div>
  );
};

export default PlayerHand;
