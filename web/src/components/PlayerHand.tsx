import React from 'react';
import Card from './Card';

interface PlayerHandProps {
  cards: number[];
  selectedCard: number | null;
  onCardSelect: (cardId: number) => void;
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

        <style jsx>{`
          .player-hand {
            background: rgba(255, 255, 255, 0.95);
            border-radius: 12px 12px 0 0;
            padding: 1.5rem;
            position: fixed;
            bottom: 0;
            left: 0;
            right: 0;
            box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.1);
          }

          .hand-header {
            text-align: center;
            margin-bottom: 1rem;
          }

          .hand-header h3 {
            margin: 0;
            color: #333;
          }

          .empty-hand {
            text-align: center;
            color: #666;
            padding: 2rem;
          }
        `}</style>
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

      <style jsx>{`
        .player-hand {
          background: rgba(255, 255, 255, 0.95);
          border-radius: 12px 12px 0 0;
          padding: 1.5rem 1rem 1rem;
          position: fixed;
          bottom: 0;
          left: 0;
          right: 0;
          box-shadow: 0 -4px 12px rgba(0, 0, 0, 0.1);
          backdrop-filter: blur(10px);
          max-height: 50vh;
          display: flex;
          flex-direction: column;
        }

        .hand-header {
          text-align: center;
          margin-bottom: 1rem;
          flex-shrink: 0;
        }

        .hand-header h3 {
          margin: 0 0 0.5rem 0;
          color: #333;
        }

        .hand-instructions {
          color: #667eea;
          font-size: 0.9rem;
        }

        .cards-container {
          flex: 1;
          overflow: hidden;
          margin-bottom: 1rem;
        }

        .cards-scroll {
          display: flex;
          gap: 0.75rem;
          overflow-x: auto;
          overflow-y: hidden;
          padding: 0.5rem 0;
          scroll-behavior: smooth;
        }

        .cards-scroll::-webkit-scrollbar {
          height: 8px;
        }

        .cards-scroll::-webkit-scrollbar-track {
          background: #f1f1f1;
          border-radius: 4px;
        }

        .cards-scroll::-webkit-scrollbar-thumb {
          background: #c1c1c1;
          border-radius: 4px;
        }

        .cards-scroll::-webkit-scrollbar-thumb:hover {
          background: #a1a1a1;
        }

        .card-wrapper {
          flex-shrink: 0;
          position: relative;
        }

        .submit-section {
          text-align: center;
          flex-shrink: 0;
        }

        .submit-card-btn {
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

        .submit-card-btn:hover {
          background: #16a34a;
        }

        @media (max-width: 768px) {
          .player-hand {
            padding: 1rem 0.5rem 0.5rem;
            max-height: 40vh;
          }

          .cards-scroll {
            gap: 0.5rem;
          }

          .hand-header h3 {
            font-size: 1rem;
          }

          .hand-instructions {
            font-size: 0.8rem;
          }

          .submit-card-btn {
            padding: 0.5rem 1.5rem;
            font-size: 0.9rem;
          }
        }

        @media (max-width: 480px) {
          .player-hand {
            max-height: 35vh;
          }
        }
      `}</style>
    </div>
  );
};

export default PlayerHand;
