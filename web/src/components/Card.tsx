import React from 'react';

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
        return 'card-small';
      case 'large':
        return 'card-large';
      default:
        return 'card-medium';
    }
  };

  return (
    <div
      className={`card ${getSizeClass()} ${isSelected ? 'selected' : ''} ${
        isClickable ? 'clickable' : ''
      } ${className}`}
      onClick={handleClick}
    >
      <div className="card-inner">
        {showCardBack ? (
          <div className="card-back">
            <div className="card-pattern">
              <div className="pattern-circle"></div>
              <div className="pattern-text">Dixit</div>
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

      <style jsx>{`
        .card {
          border-radius: 12px;
          overflow: hidden;
          box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
          transition: all 0.3s ease;
          background: white;
          position: relative;
        }

        .card-small {
          width: 80px;
          height: 120px;
        }

        .card-medium {
          width: 120px;
          height: 180px;
        }

        .card-large {
          width: 160px;
          height: 240px;
        }

        .card.clickable {
          cursor: pointer;
        }

        .card.clickable:hover {
          transform: translateY(-5px);
          box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
        }

        .card.selected {
          border: 3px solid #667eea;
          box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.3);
        }

        .card-inner {
          width: 100%;
          height: 100%;
          position: relative;
        }

        .card img {
          width: 100%;
          height: 100%;
          object-fit: cover;
        }

        .card-back {
          width: 100%;
          height: 100%;
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          display: flex;
          align-items: center;
          justify-content: center;
          position: relative;
          overflow: hidden;
        }

        .card-pattern {
          text-align: center;
          color: white;
          position: relative;
          z-index: 2;
        }

        .pattern-circle {
          width: 60px;
          height: 60px;
          border: 3px solid rgba(255, 255, 255, 0.3);
          border-radius: 50%;
          margin: 0 auto 8px;
          position: relative;
        }

        .pattern-circle::before {
          content: '';
          position: absolute;
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
          width: 30px;
          height: 30px;
          border: 2px solid rgba(255, 255, 255, 0.5);
          border-radius: 50%;
        }

        .pattern-text {
          font-weight: bold;
          font-size: 14px;
          letter-spacing: 1px;
        }

        .card-back::before {
          content: '';
          position: absolute;
          top: -50%;
          left: -50%;
          width: 200%;
          height: 200%;
          background: repeating-linear-gradient(
            45deg,
            transparent,
            transparent 10px,
            rgba(255, 255, 255, 0.05) 10px,
            rgba(255, 255, 255, 0.05) 20px
          );
          animation: pattern-move 20s linear infinite;
        }

        @keyframes pattern-move {
          0% {
            transform: translateX(-50px) translateY(-50px);
          }
          100% {
            transform: translateX(0px) translateY(0px);
          }
        }

        @media (max-width: 768px) {
          .card-small {
            width: 60px;
            height: 90px;
          }

          .card-medium {
            width: 100px;
            height: 150px;
          }

          .card-large {
            width: 140px;
            height: 210px;
          }

          .pattern-text {
            font-size: 12px;
          }

          .pattern-circle {
            width: 50px;
            height: 50px;
          }

          .pattern-circle::before {
            width: 25px;
            height: 25px;
          }
        }
      `}</style>
    </div>
  );
};

export default Card;
