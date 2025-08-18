import React, { useState } from 'react';
import Card from './Card';
import { RevealedCard } from '../types/game';

interface VotingPhaseProps {
  revealedCards: RevealedCard[];
  isStoryteller: boolean;
  hasVoted: boolean;
  onVote: (cardId: number) => void;
}

const VotingPhase: React.FC<VotingPhaseProps> = ({
  revealedCards,
  isStoryteller,
  hasVoted,
  onVote,
}) => {
  const [selectedCard, setSelectedCard] = useState<number | null>(null);

  const handleCardClick = (cardId: number) => {
    if (isStoryteller || hasVoted) return;
    
    if (selectedCard === cardId) {
      // Double-click to vote
      onVote(cardId);
    } else {
      // Single-click to select
      setSelectedCard(cardId);
    }
  };

  const handleVote = () => {
    if (selectedCard) {
      onVote(selectedCard);
    }
  };

  return (
    <div className="voting-phase">
      <div className="voting-header">
        <h3>
          {isStoryteller 
            ? 'Waiting for players to vote...' 
            : hasVoted 
              ? 'You have voted! Waiting for other players...' 
              : 'Which card belongs to the storyteller?'
          }
        </h3>
        {!isStoryteller && !hasVoted && (
          <p className="voting-instructions">
            Click a card to select it, then click again or use the Vote button to confirm
          </p>
        )}
      </div>

      <div className="revealed-cards">
        <div className="cards-grid">
          {revealedCards.map((revealedCard, index) => (
            <div key={`${revealedCard.card_id}-${index}`} className="revealed-card-wrapper">
              <Card
                id={revealedCard.card_id}
                isSelected={selectedCard === revealedCard.card_id}
                isClickable={!isStoryteller && !hasVoted}
                onClick={handleCardClick}
                size="large"
                className="revealed-card"
              />
              <div className="card-number">Card {String.fromCharCode(65 + index)}</div>
              {revealedCard.vote_count > 0 && (
                <div className="vote-count">
                  {revealedCard.vote_count} vote{revealedCard.vote_count !== 1 ? 's' : ''}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {!isStoryteller && !hasVoted && selectedCard && (
        <div className="vote-section">
          <button onClick={handleVote} className="vote-btn">
            Vote for Card {String.fromCharCode(65 + revealedCards.findIndex(c => c.card_id === selectedCard))}
          </button>
        </div>
      )}

      <style jsx>{`
        .voting-phase {
          background: rgba(255, 255, 255, 0.95);
          border-radius: 12px;
          padding: 2rem;
          margin-bottom: 2rem;
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
        }

        .voting-header {
          text-align: center;
          margin-bottom: 2rem;
        }

        .voting-header h3 {
          margin: 0 0 0.5rem 0;
          color: #333;
          font-size: 1.5rem;
        }

        .voting-instructions {
          color: #667eea;
          margin: 0;
          font-size: 1rem;
        }

        .revealed-cards {
          margin-bottom: 2rem;
        }

        .cards-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
          gap: 1.5rem;
          justify-items: center;
        }

        .revealed-card-wrapper {
          text-align: center;
          position: relative;
        }

        .card-number {
          margin-top: 0.5rem;
          font-weight: bold;
          color: #333;
          font-size: 1.1rem;
        }

        .vote-count {
          position: absolute;
          top: -10px;
          right: -10px;
          background: #ef4444;
          color: white;
          border-radius: 50%;
          width: 30px;
          height: 30px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 0.8rem;
          font-weight: bold;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
        }

        .vote-section {
          text-align: center;
        }

        .vote-btn {
          background: #667eea;
          color: white;
          border: none;
          padding: 1rem 2rem;
          border-radius: 8px;
          font-size: 1.1rem;
          font-weight: 500;
          cursor: pointer;
          transition: background-color 0.2s;
          box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
        }

        .vote-btn:hover {
          background: #5a67d8;
          transform: translateY(-2px);
          box-shadow: 0 6px 16px rgba(102, 126, 234, 0.4);
        }

        @media (max-width: 768px) {
          .voting-phase {
            padding: 1.5rem 1rem;
          }

          .voting-header h3 {
            font-size: 1.3rem;
          }

          .voting-instructions {
            font-size: 0.9rem;
          }

          .cards-grid {
            grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
            gap: 1rem;
          }

          .vote-btn {
            padding: 0.75rem 1.5rem;
            font-size: 1rem;
          }
        }

        @media (max-width: 480px) {
          .cards-grid {
            grid-template-columns: repeat(2, 1fr);
            gap: 0.75rem;
          }

          .vote-btn {
            width: 100%;
            max-width: 280px;
          }
        }
      `}</style>
    </div>
  );
};

export default VotingPhase;
