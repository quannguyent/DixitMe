import React, { useState } from 'react';
import { Card } from './Card';
import { RevealedCard } from '../types';
import styles from './VotingPhase.module.css';

interface VotingPhaseProps {
  revealedCards: RevealedCard[];
  isStoryteller: boolean;
  hasVoted: boolean;
  onVote: (cardId: number) => void;
}

export const VotingPhase: React.FC<VotingPhaseProps> = ({
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
    <div className={styles.container}>
      <div className={styles.header}>
        <h3>
          {isStoryteller
            ? 'Waiting for players to vote...'
            : hasVoted
              ? 'You have voted! Waiting for other players...'
              : 'Which card belongs to the storyteller?'
          }
        </h3>
        {!isStoryteller && !hasVoted && (
          <p className={styles.instructions}>
            Click a card to select it, then click again or use the Vote button to confirm
          </p>
        )}
      </div>

      <div className={styles.revealedCards}>
        <div className={styles.cardsGrid}>
          {revealedCards.map((revealedCard, index) => (
            <div key={`${revealedCard.card_id}-${index}`} className={styles.cardWrapper}>
              <Card
                id={revealedCard.card_id}
                isSelected={selectedCard === revealedCard.card_id}
                isClickable={!isStoryteller && !hasVoted}
                onClick={handleCardClick}
                size="large"
                className={styles.revealedCard}
              />
              <div className={styles.cardNumber}>Card {String.fromCharCode(65 + index)}</div>
              {revealedCard.vote_count > 0 && (
                <div className={styles.voteCount}>
                  {revealedCard.vote_count} vote{revealedCard.vote_count !== 1 ? 's' : ''}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {!isStoryteller && !hasVoted && selectedCard && (
        <div className={styles.voteSection}>
          <button onClick={handleVote} className={styles.voteBtn}>
            Vote for Card {String.fromCharCode(65 + revealedCards.findIndex(c => c.card_id === selectedCard))}
          </button>
        </div>
      )}
    </div>
  );
};
