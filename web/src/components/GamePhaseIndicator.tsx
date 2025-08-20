import React from 'react';
import styles from './GamePhaseIndicator.module.css';

interface Phase {
  id: string;
  name: string;
  icon: string;
  description: string;
}

interface GamePhaseIndicatorProps {
  currentPhase: string;
  roundNumber: number;
  isStoryteller: boolean;
}

const GamePhaseIndicator: React.FC<GamePhaseIndicatorProps> = ({ 
  currentPhase, 
  roundNumber, 
  isStoryteller 
}) => {
  const phases: Phase[] = [
    {
      id: 'storytelling',
      name: 'Storytelling',
      icon: '📚',
      description: isStoryteller ? 'Choose a card and give a clue' : 'Wait for the storyteller\'s clue'
    },
    {
      id: 'submitting',
      name: 'Submission',
      icon: '🎨',
      description: isStoryteller ? 'Wait for others to submit cards' : 'Submit a card that matches the clue'
    },
    {
      id: 'voting',
      name: 'Voting',
      icon: '🗳️',
      description: isStoryteller ? 'Watch others vote for your card' : 'Vote for the storyteller\'s card'
    },
    {
      id: 'scoring',
      name: 'Results',
      icon: '🏆',
      description: 'View round results and scores'
    }
  ];

  const getCurrentPhaseIndex = () => {
    return phases.findIndex(phase => phase.id === currentPhase);
  };

  const getCurrentPhase = () => {
    return phases.find(phase => phase.id === currentPhase);
  };

  const currentPhaseIndex = getCurrentPhaseIndex();
  const currentPhaseData = getCurrentPhase();

  const getPhaseStatus = (index: number) => {
    if (index < currentPhaseIndex) return 'completed';
    if (index === currentPhaseIndex) return 'active';
    return 'pending';
  };

  const getProgressPercentage = () => {
    if (currentPhaseIndex === -1) return 0;
    return ((currentPhaseIndex + 1) / phases.length) * 100;
  };

  return (
    <div className={styles.phaseIndicator}>
      <div className={styles.header}>
        <div className={styles.roundInfo}>
          <span className={styles.roundLabel}>Round</span>
          <span className={styles.roundNumber}>{roundNumber}</span>
        </div>
        
        <div className={styles.roleIndicator}>
          <span className={`${styles.role} ${isStoryteller ? styles.storyteller : styles.player}`}>
            {isStoryteller ? '📖 Storyteller' : '🎭 Player'}
          </span>
        </div>
      </div>

      <div className={styles.currentPhase}>
        <div className={styles.phaseIcon}>
          {currentPhaseData?.icon || '⏳'}
        </div>
        <div className={styles.phaseInfo}>
          <div className={styles.phaseName}>
            {currentPhaseData?.name || 'Loading...'}
          </div>
          <div className={styles.phaseDescription}>
            {currentPhaseData?.description || 'Preparing game...'}
          </div>
        </div>
      </div>

      <div className={styles.progressBar}>
        <div 
          className={styles.progressFill}
          style={{ width: `${getProgressPercentage()}%` }}
        />
      </div>

      <div className={styles.phasesList}>
        {phases.map((phase, index) => {
          const status = getPhaseStatus(index);
          return (
            <div
              key={phase.id}
              className={`${styles.phaseStep} ${styles[status]}`}
            >
              <div className={styles.stepIcon}>
                {status === 'completed' ? '✅' : 
                 status === 'active' ? phase.icon : 
                 '⭕'}
              </div>
              <div className={styles.stepName}>{phase.name}</div>
              {index < phases.length - 1 && (
                <div className={`${styles.connector} ${
                  status === 'completed' ? styles.connectorCompleted : ''
                }`} />
              )}
            </div>
          );
        })}
      </div>

      {/* Quick tips based on phase */}
      <div className={styles.tips}>
        {currentPhase === 'storytelling' && isStoryteller && (
          <div className={styles.tip}>
            <span className={styles.tipIcon}>💡</span>
            <span>Give a clue that's not too obvious, not too vague!</span>
          </div>
        )}
        
        {currentPhase === 'submitting' && !isStoryteller && (
          <div className={styles.tip}>
            <span className={styles.tipIcon}>🎯</span>
            <span>Choose a card that fits the clue but isn't too obvious.</span>
          </div>
        )}
        
        {currentPhase === 'voting' && !isStoryteller && (
          <div className={styles.tip}>
            <span className={styles.tipIcon}>🤔</span>
            <span>Try to guess which card belongs to the storyteller!</span>
          </div>
        )}
        
        {currentPhase === 'scoring' && (
          <div className={styles.tip}>
            <span className={styles.tipIcon}> ️⭐</span>
            <span>Points awarded! First to 30 points wins the game.</span>
          </div>
        )}
      </div>
    </div>
  );
};

export default GamePhaseIndicator;
