# React & TypeScript Beginner's Guide for DixitMe

Welcome to React! This guide will help you understand React, TypeScript, and how they work together in the DixitMe project.

## 🚀 What is React?

React is a JavaScript library for building user interfaces, especially web applications. It's designed around:
- **Components**: Reusable UI pieces
- **State Management**: Handling data that changes
- **Event Handling**: Responding to user actions
- **Virtual DOM**: Efficient UI updates

## 📚 React Basics

### 1. Components & JSX

```tsx
// Basic component
function Welcome() {
    return <h1>Welcome to DixitMe!</h1>;
}

// Component with props (parameters)
interface PlayerProps {
    name: string;
    score: number;
}

function Player({ name, score }: PlayerProps) {
    return (
        <div className="player">
            <h3>{name}</h3>
            <p>Score: {score}</p>
        </div>
    );
}

// Using components
function App() {
    return (
        <div>
            <Welcome />
            <Player name="Alice" score={15} />
            <Player name="Bob" score={12} />
        </div>
    );
}
```

**Key Concepts:**
- **JSX**: HTML-like syntax in JavaScript
- **Props**: Data passed to components
- **Curly braces `{}`**: Insert JavaScript in JSX

### 2. State (Data that Changes)

```tsx
import { useState } from 'react';

function GameLobby() {
    // useState hook - manages changing data
    const [playerName, setPlayerName] = useState('');
    const [roomCode, setRoomCode] = useState('');
    const [isLoading, setIsLoading] = useState(false);

    const handleJoinGame = () => {
        setIsLoading(true);
        // Join game logic...
        console.log(`${playerName} joining room ${roomCode}`);
    };

    return (
        <div>
            <input 
                value={playerName}
                onChange={(e) => setPlayerName(e.target.value)}
                placeholder="Your name"
            />
            <input 
                value={roomCode}
                onChange={(e) => setRoomCode(e.target.value)}
                placeholder="Room code"
            />
            <button 
                onClick={handleJoinGame}
                disabled={isLoading}
            >
                {isLoading ? 'Joining...' : 'Join Game'}
            </button>
        </div>
    );
}
```

### 3. Event Handling

```tsx
function Card({ cardId, onCardClick }: CardProps) {
    const handleClick = () => {
        console.log(`Card ${cardId} clicked!`);
        onCardClick(cardId);
    };

    const handleHover = () => {
        console.log(`Hovering over card ${cardId}`);
    };

    return (
        <div 
            className="card"
            onClick={handleClick}
            onMouseEnter={handleHover}
        >
            <img src={`/cards/${cardId}.jpg`} alt={`Card ${cardId}`} />
        </div>
    );
}
```

### 4. useEffect (Side Effects)

```tsx
import { useState, useEffect } from 'react';

function GameBoard() {
    const [gameState, setGameState] = useState(null);

    // Run when component mounts
    useEffect(() => {
        console.log('Component mounted!');
        
        // Cleanup when component unmounts
        return () => {
            console.log('Component unmounting!');
        };
    }, []); // Empty array = run once

    // Run when gameState changes
    useEffect(() => {
        if (gameState) {
            console.log('Game state updated:', gameState);
        }
    }, [gameState]); // Runs when gameState changes

    return <div>Game Board Content</div>;
}
```

### 5. Conditional Rendering

```tsx
function GamePhase({ phase, isStoryteller }: GamePhaseProps) {
    // Show different UI based on conditions
    if (phase === 'lobby') {
        return <LobbyView />;
    }

    if (phase === 'storytelling' && isStoryteller) {
        return <StorytellingView />;
    }

    if (phase === 'voting') {
        return <VotingView />;
    }

    // Default case
    return <div>Unknown phase: {phase}</div>;
}
```

### 6. Lists & Keys

```tsx
interface Player {
    id: string;
    name: string;
    score: number;
}

function PlayerList({ players }: { players: Player[] }) {
    return (
        <div className="player-list">
            {players.map(player => (
                <div key={player.id} className="player-item">
                    <span>{player.name}</span>
                    <span>{player.score} points</span>
                </div>
            ))}
        </div>
    );
}
```

## 🎯 TypeScript Basics

TypeScript adds type safety to JavaScript:

### 1. Basic Types

```tsx
// Basic types
const playerName: string = "Alice";
const score: number = 25;
const isActive: boolean = true;

// Arrays
const cardIds: number[] = [1, 2, 3, 4, 5];
const players: string[] = ["Alice", "Bob", "Charlie"];

// Objects
const player: {
    id: string;
    name: string;
    score: number;
} = {
    id: "123",
    name: "Alice", 
    score: 15
};
```

### 2. Interfaces

```tsx
// Define object shape
interface GameState {
    id: string;
    roomCode: string;
    players: Player[];
    currentPhase: 'lobby' | 'storytelling' | 'voting' | 'results';
    roundNumber: number;
}

interface Player {
    id: string;
    name: string;
    score: number;
    hand: number[];
}

// Use interface
function updateGameState(newState: GameState) {
    console.log(`Game ${newState.roomCode} updated`);
}
```

### 3. Props with TypeScript

```tsx
// Component props interface
interface CardProps {
    cardId: number;
    isSelected: boolean;
    onCardClick: (cardId: number) => void;
    className?: string; // Optional prop
}

function Card({ cardId, isSelected, onCardClick, className = '' }: CardProps) {
    return (
        <div 
            className={`card ${className} ${isSelected ? 'selected' : ''}`}
            onClick={() => onCardClick(cardId)}
        >
            <img src={`/cards/${cardId}.jpg`} alt={`Card ${cardId}`} />
        </div>
    );
}
```

### 4. State with TypeScript

```tsx
// Typed state
const [gameState, setGameState] = useState<GameState | null>(null);
const [selectedCard, setSelectedCard] = useState<number | null>(null);
const [players, setPlayers] = useState<Player[]>([]);

// Custom hooks with types
function useWebSocket(url: string) {
    const [socket, setSocket] = useState<WebSocket | null>(null);
    const [isConnected, setIsConnected] = useState<boolean>(false);
    
    useEffect(() => {
        const ws = new WebSocket(url);
        ws.onopen = () => setIsConnected(true);
        ws.onclose = () => setIsConnected(false);
        setSocket(ws);
        
        return () => ws.close();
    }, [url]);
    
    return { socket, isConnected };
}
```

## 🎮 React in DixitMe Project

### State Management with Zustand

DixitMe uses Zustand for simple state management:

```tsx
// Store definition
interface GameStore {
    // State
    gameState: GameState | null;
    currentPhase: GamePhase;
    isConnected: boolean;
    
    // Actions
    setGameState: (state: GameState) => void;
    setPhase: (phase: GamePhase) => void;
    connect: (token: string) => void;
}

const useGameStore = create<GameStore>((set, get) => ({
    // Initial state
    gameState: null,
    currentPhase: 'lobby',
    isConnected: false,
    
    // Actions
    setGameState: (state) => set({ gameState: state }),
    setPhase: (phase) => set({ currentPhase: phase }),
    connect: (token) => {
        // WebSocket connection logic
        set({ isConnected: true });
    },
}));

// Using the store
function GameBoard() {
    const { gameState, setGameState } = useGameStore();
    
    return (
        <div>
            {gameState ? (
                <div>Game: {gameState.roomCode}</div>
            ) : (
                <div>No game loaded</div>
            )}
        </div>
    );
}
```

### Component Structure in DixitMe

**1. Page Components:**
```tsx
// Main game component
function GameBoard() {
    const { gameState, currentPhase } = useGameStore();
    
    const renderPhaseContent = () => {
        switch (currentPhase) {
            case 'storytelling':
                return <StorytellingPhase />;
            case 'submission':
                return <SubmissionPhase />;
            case 'voting':
                return <VotingPhase />;
            default:
                return <Lobby />;
        }
    };
    
    return (
        <div className="game-board">
            <PlayerList />
            <GamePhaseIndicator phase={currentPhase} />
            {renderPhaseContent()}
            <Chat />
        </div>
    );
}
```

**2. Reusable Components:**
```tsx
// Card component
interface CardProps {
    cardId: number;
    isSelected?: boolean;
    onClick?: (cardId: number) => void;
    size?: 'small' | 'medium' | 'large';
}

function Card({ cardId, isSelected = false, onClick, size = 'medium' }: CardProps) {
    return (
        <div 
            className={`card card--${size} ${isSelected ? 'card--selected' : ''}`}
            onClick={() => onClick?.(cardId)}
        >
            <img 
                src={`/assets/cards/${cardId}.jpg`} 
                alt={`Card ${cardId}`}
                loading="lazy"
            />
        </div>
    );
}
```

### WebSocket Integration

```tsx
function useWebSocketConnection() {
    const [socket, setSocket] = useState<WebSocket | null>(null);
    const { setGameState, setPhase } = useGameStore();
    
    const connect = (token: string) => {
        const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);
        
        ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            
            switch (message.type) {
                case 'game_created':
                    setGameState(message.payload.game);
                    break;
                case 'round_started':
                    setPhase('storytelling');
                    break;
                case 'voting_started':
                    setPhase('voting');
                    break;
            }
        };
        
        setSocket(ws);
    };
    
    const sendMessage = (type: string, payload: any) => {
        if (socket) {
            socket.send(JSON.stringify({ type, payload }));
        }
    };
    
    return { connect, sendMessage };
}
```

### CSS Modules Styling

```tsx
// Component file: Card.tsx
import styles from './Card.module.css';

function Card({ cardId, isSelected }: CardProps) {
    return (
        <div className={`${styles.card} ${isSelected ? styles.selected : ''}`}>
            <img 
                src={`/cards/${cardId}.jpg`}
                className={styles.cardImage}
                alt={`Card ${cardId}`}
            />
        </div>
    );
}
```

```css
/* Card.module.css */
.card {
    width: 120px;
    height: 180px;
    border-radius: 8px;
    cursor: pointer;
    transition: transform 0.2s ease;
}

.card:hover {
    transform: scale(1.05);
}

.selected {
    border: 3px solid #007bff;
    transform: scale(1.1);
}

.cardImage {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: inherit;
}
```

## 🛠️ Development Patterns

### 1. Custom Hooks

```tsx
// Custom hook for game actions
function useGameActions() {
    const { sendMessage } = useWebSocketConnection();
    
    const createGame = (roomCode: string, playerName: string) => {
        sendMessage('create_game', { room_code: roomCode, player_name: playerName });
    };
    
    const submitCard = (cardId: number) => {
        sendMessage('submit_card', { card_id: cardId });
    };
    
    return { createGame, submitCard };
}
```

### 2. Error Boundaries

```tsx
interface ErrorBoundaryState {
    hasError: boolean;
}

class ErrorBoundary extends React.Component<React.PropsWithChildren<{}>, ErrorBoundaryState> {
    constructor(props: React.PropsWithChildren<{}>) {
        super(props);
        this.state = { hasError: false };
    }
    
    static getDerivedStateFromError(): ErrorBoundaryState {
        return { hasError: true };
    }
    
    render() {
        if (this.state.hasError) {
            return <div>Something went wrong. Please refresh the page.</div>;
        }
        
        return this.props.children;
    }
}
```

### 3. Loading States

```tsx
function GameLobby() {
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    
    const joinGame = async (roomCode: string) => {
        setIsLoading(true);
        setError(null);
        
        try {
            await gameService.joinGame(roomCode);
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Unknown error');
        } finally {
            setIsLoading(false);
        }
    };
    
    if (isLoading) return <div>Joining game...</div>;
    if (error) return <div>Error: {error}</div>;
    
    return <div>Lobby content</div>;
}
```

## 🎯 Common Patterns in DixitMe

### 1. Phase-Based Rendering

```tsx
function GameContent() {
    const { currentPhase, gameState } = useGameStore();
    
    switch (currentPhase) {
        case 'lobby':
            return <LobbyPhase />;
        case 'storytelling':
            return <StorytellingPhase storyteller={gameState?.currentStoryteller} />;
        case 'submission':
            return <SubmissionPhase clue={gameState?.currentClue} />;
        case 'voting':
            return <VotingPhase cards={gameState?.revealedCards} />;
        case 'results':
            return <ResultsPhase scores={gameState?.roundScores} />;
        default:
            return <div>Unknown phase</div>;
    }
}
```

### 2. Real-time Updates

```tsx
function PlayerList() {
    const { gameState } = useGameStore();
    
    // Re-renders automatically when gameState changes
    return (
        <div className="player-list">
            {gameState?.players.map(player => (
                <div key={player.id} className="player">
                    <span>{player.name}</span>
                    <span>{player.score} pts</span>
                    {player.isStoryteller && <span>📖 Storyteller</span>}
                </div>
            ))}
        </div>
    );
}
```

## 📖 Learning Resources

**Official Documentation:**
- [React Documentation](https://react.dev/) - Official React docs
- [TypeScript Handbook](https://www.typescriptlang.org/docs/) - Complete TypeScript guide

**Interactive Learning:**
- [React Tutorial](https://react.dev/learn) - Step-by-step tutorial
- [TypeScript Playground](https://www.typescriptlang.org/play) - Try TypeScript online

**Practice Projects:**
1. Build a simple counter app
2. Create a todo list with React
3. Add TypeScript to existing JavaScript project
4. Build a simple card game interface

## 🔧 Development Tools

### 1. VS Code Extensions
- **ES7+ React/Redux/React-Native snippets** - Code snippets
- **TypeScript Hero** - Auto imports and organization
- **Prettier** - Code formatting
- **ESLint** - Code linting

### 2. Browser DevTools
- **React Developer Tools** - Inspect React components
- **TypeScript errors** - Check console for type errors

### 3. Common Commands
```bash
# Start development server
npm start

# Type check
npm run type-check

# Build for production
npm run build

# Run tests
npm test
```

## 🎯 Next Steps

1. **Learn JavaScript basics** if you haven't already
2. **Complete React tutorial** (react.dev/learn)
3. **Learn TypeScript fundamentals** (typescriptlang.org/docs)
4. **Examine DixitMe components** starting with simple ones
5. **Try modifying** existing components
6. **Build your own** simple components

Remember: React is all about breaking down complex UIs into simple, reusable components. Start small and build up!
