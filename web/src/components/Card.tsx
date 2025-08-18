import React from 'react';
import styles from './Card.module.css';

interface CardProps {
  id: number;
  isSelected?: boolean;
  isClickable?: boolean;
  showCardBack?: boolean;
  size?: 'small' | 'medium' | 'large';
  onClick?: (cardId: number) => void;
  className?: string;
}

const Card: React.FC<CardProps> = ({
  id,
  isSelected = false,
  isClickable = false,
  showCardBack = false,
  size = 'medium',
  onClick,
  className = '',
}) => {
  const handleClick = () => {
    if (isClickable && onClick) {
      onClick(id);
    }
  };

  const getSizeClass = () => {
    switch (size) {
      case 'small':
        return styles.cardSmall;
      case 'large':
        return styles.cardLarge;
      default:
        return styles.cardMedium;
    }
  };

  return (
    <div
      className={`${styles.card} ${getSizeClass()} ${isSelected ? styles.selected : ''} ${
        isClickable ? styles.clickable : ''
      } ${className}`}
      onClick={handleClick}
    >
      <div className={styles.cardInner}>
        {showCardBack ? (
          <div className={styles.cardBack}>
            <div className={styles.cardPattern}>
              <div className={styles.patternCircle}></div>
              <div className={styles.patternText}>Dixit</div>
            </div>
          </div>
        ) : (
          <img
            src={`/cards/${id}.jpg`}
            alt={`Card ${id}`}
            onError={(e) => {
              // Fallback to placeholder if card image doesn't exist
              e.currentTarget.src = `data:image/svg+xml,${encodeURIComponent(`
                <svg width="200" height="300" xmlns="http://www.w3.org/2000/svg">
                  <rect width="200" height="300" fill="#f0f0f0" stroke="#ccc" stroke-width="2"/>
                  <text x="100" y="150" text-anchor="middle" font-family="Arial" font-size="16" fill="#666">Card ${id}</text>
                </svg>
              `)}`;
            }}
          />
        )}
      </div>
    </div>
  );
};

export default Card;
