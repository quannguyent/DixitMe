import React from 'react';
import { Card } from './Card';
import styles from './CardModal.module.css';

interface CardModalProps {
  cardId: number | null;
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  actionLabel?: string;
  onAction?: (cardId: number) => void;
  actionDisabled?: boolean;
}

export const CardModal: React.FC<CardModalProps> = ({
  cardId,
  isOpen,
  onClose,
  title,
  actionLabel,
  onAction,
  actionDisabled = false,
}) => {
  if (!isOpen || cardId === null) return null;

  const handleAction = () => {
    if (onAction) {
      onAction(cardId);
      onClose();
    }
  };

  const handleOverlayClick = (e: React.MouseEvent) => {
    if (e.target === e.currentTarget) {
      onClose();
    }
  };

  return (
    <div className={styles.overlay} onClick={handleOverlayClick}>
      <div className={styles.modal}>
        <button className={styles.closeBtn} onClick={onClose} aria-label="Close">
          &times;
        </button>
        
        {title && <h3 className={styles.title}>{title}</h3>}
        
        <div className={styles.cardContainer}>
          <Card id={cardId} size="large" className={styles.zoomedCard} />
        </div>
        
        <div className={styles.footer}>
          <button className={styles.cancelBtn} onClick={onClose}>
            Cancel
          </button>
          {actionLabel && onAction && (
            <button 
              className={styles.actionBtn} 
              onClick={handleAction}
              disabled={actionDisabled}
            >
              {actionLabel}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};
