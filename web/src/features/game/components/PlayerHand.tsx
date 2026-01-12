import React from 'react';
import { Card } from './Card';
import styles from './PlayerHand.module.css';

interface PlayerHandProps {
  cards: number[];
  selectedCard: number | null;
  onCardSelect: (cardId: number | null) => void;
  canSelect: boolean;
  canSubmit: boolean;
  onSubmit: (cardId: number) => void;
}

export const PlayerHand: React.FC<PlayerHandProps> = ({
  cards,
  selectedCard,
  onCardSelect,
  canSelect,
  canSubmit,
  onSubmit,
}) => {
  const [isModalOpen, setIsModalOpen] = React.useState(false);
  const [modalCardId, setModalCardId] = React.useState<number | null>(null);

  const handleCardClick = (cardId: number) => {
    if (!canSelect) return;
    setModalCardId(cardId);
    setIsModalOpen(true);
  };

  const handleSubmit = () => {
    if (selectedCard && canSubmit) {
      onSubmit(selectedCard);
    }
  };

  const handleModalConfirm = () => {
    if (!modalCardId) return;
    if (canSubmit) {
      onSubmit(modalCardId);
    } else {
      onCardSelect(modalCardId);
    }
    setIsModalOpen(false);
    setModalCardId(null);
  };

  const handleModalCancel = () => {
    setIsModalOpen(false);
    setModalCardId(null);
  };

  if (cards.length === 0) {
    return (
      <div className={styles.container}>
        <div className={styles.header}>
          <h3>Your Hand</h3>
        </div>
        <div className={styles.emptyHand}>
          <p>No cards in hand</p>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h3>Your Hand ({cards.length} cards)</h3>
        {canSelect && (
          <div className={styles.instructions}>
            {canSubmit ? (
              selectedCard ? (
                <span>Review another card or use Submit to confirm</span>
              ) : (
                <span>Select a card to inspect and confirm</span>
              )
            ) : (
              <span>Select a card to inspect for your clue</span>
            )}
          </div>
        )}
      </div>

      <div className={styles.cardsContainer}>
        <div className={styles.cardsScroll}>
          {cards.map((cardId) => (
            <div key={cardId} className={styles.cardWrapper}>
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
        <div className={styles.submitSection}>
          <button onClick={handleSubmit} className={styles.submitBtn}>
            Submit Card {selectedCard}
          </button>
        </div>
      )}

      {isModalOpen && modalCardId && (
        <div className={styles.modalBackdrop} onClick={handleModalCancel}>
          <div className={styles.modal} onClick={(e: React.MouseEvent) => e.stopPropagation()}>
            <div className={styles.modalImage}>
              <img src={`/cards/${modalCardId}.jpg`} alt={`Card ${modalCardId}`} />
            </div>
            <div className={styles.modalActions}>
              <button className={styles.modalCancel} onClick={handleModalCancel}>
                Cancel
              </button>
              <button className={styles.modalConfirm} onClick={handleModalConfirm}>
                {canSubmit ? 'Confirm & Submit' : 'Confirm'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
