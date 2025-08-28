# DixitMe - Complete Developer Guide

Welcome! This comprehensive guide will help you understand **both the codebase architecture and business flow** of DixitMe. Whether you're new to Go/React or an experienced developer, this guide will get you up to speed systematically.

## 🎯 Learning Strategy

Instead of learning technologies in isolation, we'll explore **the actual codebase and business logic together**. This approach gives you practical context while building your understanding of both the technical implementation and game mechanics.

## 📚 Phase 1: Foundation & Setup (30-45 minutes)

### 1. Understand the Business: What is Dixit? 🎮

Before diving into code, understand what we're building:

**Game Overview:**
- **Players**: 3-6 players per game
- **Cards**: 84 beautifully illustrated cards with abstract imagery
- **Goal**: Score 30 points through creative storytelling and guessing

**Round Flow (This drives our entire codebase):**
1. **Storytelling**: Storyteller picks a card and gives a clue
2. **Submission**: Other players submit cards that fit the clue  
3. **Voting**: All players (except storyteller) vote for the storyteller's card
4. **Scoring**: Points awarded based on voting results

**Victory Conditions:**
- First player to 30 points OR deck runs empty
- Storyteller role rotates each round

> 💡 **Why This Matters**: Every component, API endpoint, and database table exists to support this game flow!

### 2. Development Environment Setup 🛠️

```bash
# 1. Prerequisites check
go version    # Should be 1.21+
node --version # Should be 18+
docker --version

# 2. Clone and setup
cd DixitMe
cp configs/config.env.example .env

# 3. Start dependencies
docker-compose -f deployments/docker/docker-compose.dev.yml up -d postgres redis

# 4. Install dependencies
go mod download
cd web && npm install && cd ..

# 5. Test startup (we fixed the database issues!)
go run cmd/server/main.go

# 6. Start frontend (in another terminal)
cd web && npm start
```

### 3. Project Structure Overview 🏗️

```
DixitMe/ - Your game development workspace
├── cmd/                     # 🚀 Application entry points
├── internal/                # 🔒 Private game logic
│   ├── services/game/       # 🎮 Core game mechanics (THE HEART)
│   ├── transport/           # 🌐 API & WebSocket handlers
│   ├── models/              # 📊 Database structures
│   └── [other packages]     # 🔧 Supporting infrastructure
├── web/                     # ⚛️ React frontend
└── assets/                  # 🎨 Game cards & data
```

> 📋 **For complete structure details**, see [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md)

## 📚 Phase 2: Business Flow Deep Dive (60-90 minutes)

Let's trace the **complete game flow** through our codebase:

### 4. Game Creation Flow 🎯

**Business Need**: Players need to create and join game rooms

#### Frontend: User Action
**File**: `web/src/components/GameLanding.tsx`
```typescript
const handleCreateGame = async () => {
    // 1. Generate unique room code
    const roomCode = generateRoomCode(); // e.g., "ABCD"
    
    // 2. Connect to backend via WebSocket
    connect(userToken);
    
    // 3. Send create game request
    createGame(roomCode, playerName);
};
```

#### WebSocket Message Flow
**File**: `web/src/store/gameStore.ts`
```typescript
const createGame = (roomCode: string, playerName: string) => {
    // Send structured message to backend
    websocket.send(JSON.stringify({
        type: 'create_game',
        payload: { room_code: roomCode, player_name: playerName }
    }));
};
```

#### Backend: WebSocket Handler
**File**: `internal/transport/websocket/handlers.go`
```go
func handleMessage(conn *websocket.Conn, playerID uuid.UUID, msg ConnectionMessage) error {
    switch msg.Type {
    case "create_game":
        // Parse the incoming message
        var payload CreateGamePayload
        json.Unmarshal(msg.Payload, &payload)
        
        // Call business logic layer
        manager := game.GetManager()
        gameState, err := manager.CreateGame(payload.RoomCode, playerID, payload.PlayerName)
        
        // Send success response back to frontend
        return conn.WriteJSON(ConnectionMessage{
            Type: "game_created",
            Payload: GameCreatedPayload{Game: gameState},
        })
    }
}
```

#### Business Logic: Game Creation
**File**: `internal/services/game/game_actions.go`
```go
func (m *Manager) CreateGame(roomCode string, creatorID uuid.UUID, creatorName string) (*GameState, error) {
    // 1. Business validation
    if roomCode == "" || len(roomCode) != 4 {
        return nil, errors.New("room code must be 4 characters")
    }
    
    // 2. Create in-memory game state (for real-time play)
    game := &GameState{
        ID:       uuid.New(),
        RoomCode: roomCode,
        Players:  make(map[uuid.UUID]*Player),
        Status:   GameStatusWaiting,  // Business state
        Deck:     shuffleDeck(),      // Prepare 84 cards
        CreatedAt: time.Now(),
    }
    
    // 3. Add creator as first player
    creator := &Player{
        ID:   creatorID,
        Name: creatorName,
        Hand: []int{}, // Empty until game starts
        Type: PlayerTypeHuman,
    }
    game.Players[creatorID] = creator
    
    // 4. Store in memory for real-time access
    m.games[roomCode] = game
    
    // 5. Persist to database for history/recovery
    err := m.PersistGame(context.Background(), game)
    
    return game, err
}
```

#### Database Persistence
**File**: `internal/services/game/persistence.go`
```go
func (m *Manager) PersistGame(ctx context.Context, gameState *GameState) error {
    // Convert in-memory state to database model
    dbGame := &models.Game{
        ID:       gameState.ID,
        RoomCode: gameState.RoomCode,
        Status:   gameState.Status,      // Enum: waiting, playing, completed
        CreatedAt: gameState.CreatedAt,
    }
    
    // Save to PostgreSQL with context and logging
    if err := m.db.WithContext(ctx).Create(dbGame).Error; err != nil {
        logger.GetLogger().Error("Failed to persist game", 
            "game_id", gameState.ID, "error", err)
        return fmt.Errorf("failed to persist game: %w", err)
    }
    
    return nil
}
```

**Key Learning**: Notice how we have **three representations** of the same game:
1. **Frontend State** (React/Zustand) - For UI updates
2. **In-Memory State** (Go structs) - For real-time gameplay 
3. **Database State** (PostgreSQL) - For persistence and history

### 5. Game Start & Card Dealing Flow 🎴

**Business Need**: When enough players join, start the game and deal cards

#### Business Logic: Game Initialization
**File**: `internal/services/game/game_actions.go`
```go
func (m *Manager) StartGame(roomCode string, playerID uuid.UUID) error {
    game := m.GetGame(roomCode)
    
    // Business rules validation
    if len(game.Players) < 3 {
        return errors.New("need at least 3 players")
    }
    
    if game.Status != GameStatusWaiting {
        return errors.New("game already started")
    }
    
    // Deal 6 cards to each player (core Dixit rule)
    for playerID, player := range game.Players {
        player.Hand = m.dealCards(game, 6)
    }
    
    // Start first round
    if err := m.startNewRound(game); err != nil {
        return fmt.Errorf("failed to start first round: %w", err)
    }
    
    // Update game state
    game.Status = GameStatusInProgress
    
    // Broadcast to all connected players
    m.BroadcastToGame(game, MessageTypeGameStarted, GameStartedPayload{
        GameState: game,
    })
    
    return nil
}
```

#### Card Dealing Algorithm
```go
func (m *Manager) dealCards(game *GameState, count int) []int {
    if len(game.Deck) < count {
        return nil // Not enough cards (business rule violation)
    }
    
    // Take cards from top of deck
    cards := game.Deck[:count]
    game.Deck = game.Deck[count:]
    
    // Move dealt cards to used pile
    game.UsedCards = append(game.UsedCards, cards...)
    
    return cards
}
```

### 6. Round Management: The Core Game Loop 🔄

**Business Need**: Manage the storytelling → submission → voting → scoring cycle

#### Round Creation
**File**: `internal/services/game/round_actions.go`
```go
func (m *Manager) startNewRound(game *GameState) error {
    // Business logic: rotate storyteller
    storytellerID := m.getNextStoryteller(game)
    
    round := &Round{
        ID:            uuid.New(),
        RoundNumber:   game.RoundNumber + 1,
        StorytellerID: storytellerID,
        Status:        RoundStatusStorytelling,  // Initial phase
        Submissions:   make(map[uuid.UUID]*CardSubmission),
        Votes:         make(map[uuid.UUID]*Vote),
        CreatedAt:     time.Now(),
    }
    
    game.CurrentRound = round
    game.RoundNumber++
    
    // Persist round to database
    if err := m.PersistRound(context.Background(), game.ID, round); err != nil {
        return err
    }
    
    // Notify all players of new round
    m.BroadcastToGame(game, MessageTypeRoundStarted, RoundStartedPayload{
        Round: round,
    })
    
    return nil
}
```

#### Storytelling Phase
```go
func (m *Manager) SubmitClue(roomCode string, playerID uuid.UUID, clue string, cardID int) error {
    game := m.GetGame(roomCode)
    
    // Business validation
    if game.CurrentRound.StorytellerID != playerID {
        return errors.New("only storyteller can submit clue")
    }
    
    if game.CurrentRound.Status != RoundStatusStorytelling {
        return errors.New("not in storytelling phase")
    }
    
    // Validate card is in player's hand
    storyteller := game.Players[playerID]
    if !m.playerHasCard(storyteller, cardID) {
        return errors.New("card not in hand")
    }
    
    // Store clue and card
    game.CurrentRound.Clue = clue
    game.CurrentRound.StorytellerCard = cardID
    
    // Remove card from storyteller's hand
    m.removeCardFromHand(storyteller, cardID)
    
    // Advance to submission phase
    game.CurrentRound.Status = RoundStatusSubmission
    
    // Broadcast clue to all players
    m.BroadcastToGame(game, MessageTypeClueSubmitted, ClueSubmittedPayload{
        Clue: clue,
    })
    
    return nil
}
```

#### Card Submission Phase
```go
func (m *Manager) SubmitCard(roomCode string, playerID uuid.UUID, cardID int) error {
    game := m.GetGame(roomCode)
    
    // Business rules
    if playerID == game.CurrentRound.StorytellerID {
        return errors.New("storyteller cannot submit cards")
    }
    
    if _, exists := game.CurrentRound.Submissions[playerID]; exists {
        return errors.New("player already submitted")
    }
    
    // Store submission
    game.CurrentRound.Submissions[playerID] = &CardSubmission{
        PlayerID: playerID,
        CardID:   cardID,
    }
    
    // Check if all players submitted
    expectedSubmissions := len(game.Players) - 1 // Exclude storyteller
    if len(game.CurrentRound.Submissions) == expectedSubmissions {
        m.startVotingPhase(game)
    }
    
    return nil
}
```

#### Voting Phase & Scoring
```go
func (m *Manager) startVotingPhase(game *GameState) {
    // Collect all submitted cards + storyteller's card
    allCards := []RevealedCard{
        {CardID: game.CurrentRound.StorytellerCard, PlayerID: game.CurrentRound.StorytellerID},
    }
    
    for _, submission := range game.CurrentRound.Submissions {
        allCards = append(allCards, RevealedCard{
            CardID:   submission.CardID,
            PlayerID: submission.PlayerID,
        })
    }
    
    // Shuffle cards so storyteller's card isn't obvious
    rand.Shuffle(len(allCards), func(i, j int) {
        allCards[i], allCards[j] = allCards[j], allCards[i]
    })
    
    game.CurrentRound.RevealedCards = allCards
    game.CurrentRound.Status = RoundStatusVoting
    
    // Broadcast voting interface
    m.BroadcastToGame(game, MessageTypeVotingStarted, VotingStartedPayload{
        RevealedCards: allCards,
    })
}
```

### 7. Scoring System: Dixit Rules Implementation 🏆

**File**: `internal/services/game/round_actions.go`
```go
func (m *Manager) calculateScores(game *GameState) map[uuid.UUID]int {
    scores := make(map[uuid.UUID]int)
    storytellerCard := game.CurrentRound.StorytellerCard
    storytellerID := game.CurrentRound.StorytellerID
    
    // Count votes for storyteller's card
    storytellerVotes := 0
    for _, vote := range game.CurrentRound.Votes {
        if vote.CardID == storytellerCard {
            storytellerVotes++
        }
    }
    
    totalVoters := len(game.Players) - 1 // Exclude storyteller
    
    // Dixit scoring rules implementation
    if storytellerVotes == 0 || storytellerVotes == totalVoters {
        // Perfect clue penalty: too obvious or too obscure
        // Everyone except storyteller gets 2 points
        for playerID := range game.Players {
            if playerID != storytellerID {
                scores[playerID] = 2
            }
        }
    } else {
        // Good clue: storyteller + correct guessers get 3 points
        scores[storytellerID] = 3
        
        for _, vote := range game.CurrentRound.Votes {
            if vote.CardID == storytellerCard {
                scores[vote.PlayerID] = 3
            }
        }
    }
    
    // Bonus points: +1 for each vote on your submitted card
    for _, vote := range game.CurrentRound.Votes {
        if vote.CardID != storytellerCard {
            // Find who submitted this card
            for _, submission := range game.CurrentRound.Submissions {
                if submission.CardID == vote.CardID {
                    scores[submission.PlayerID] += 1
                    break
                }
            }
        }
    }
    
    return scores
}
```

### 8. Game End Conditions 🏁

```go
func (m *Manager) checkGameEnd(game *GameState) (bool, uuid.UUID) {
    var winnerID uuid.UUID
    maxScore := 0
    
    // Check victory conditions
    for playerID, player := range game.Players {
        if player.Score >= 30 {  // Primary win condition
            if player.Score > maxScore {
                maxScore = player.Score
                winnerID = playerID
            }
        }
    }
    
    // Alternative win condition: deck empty
    if len(game.Deck) < len(game.Players)*6 { // Can't deal next round
        // Winner is highest score
        for playerID, player := range game.Players {
            if player.Score > maxScore {
                maxScore = player.Score
                winnerID = playerID
            }
        }
        return true, winnerID
    }
    
    return winnerID != uuid.Nil, winnerID
}
```

## 📚 Phase 3: Frontend-Backend Integration (45-60 minutes)

### 9. WebSocket Communication Patterns 🔄

**Understanding Real-time Game Updates**

#### Frontend WebSocket Store
**File**: `web/src/store/gameStore.ts`
```typescript
interface GameStore {
    // State
    gameState: GameState | null;
    currentPhase: 'lobby' | 'storytelling' | 'submission' | 'voting' | 'results';
    isConnected: boolean;
    
    // Actions
    connect: (token: string) => void;
    sendMessage: (type: string, payload: any) => void;
    
    // Game actions
    createGame: (roomCode: string, playerName: string) => void;
    submitClue: (clue: string, cardId: number) => void;
    submitCard: (cardId: number) => void;
    submitVote: (cardId: number) => void;
}

// WebSocket message handling
const handleMessage = (event: MessageEvent) => {
    const message = JSON.parse(event.data);
    
    switch (message.type) {
        case 'game_created':
            set({ gameState: message.payload.game });
            break;
            
        case 'round_started':
            set({ 
                currentPhase: 'storytelling',
                gameState: { ...gameState, current_round: message.payload.round }
            });
            break;
            
        case 'clue_submitted':
            set({ currentPhase: 'submission' });
            break;
            
        case 'voting_started':
            set({ 
                currentPhase: 'voting',
                revealedCards: message.payload.revealed_cards
            });
            break;
    }
};
```

#### Message Flow Diagram
```
Frontend                 Backend
   |                        |
   |--- create_game ------->|  (WebSocket)
   |                        |
   |<--- game_created ------|  (Response)
   |                        |
   |--- submit_clue ------->|  (Player action)
   |                        |
   |<--- clue_submitted ----|  (Broadcast to all)
   |                        |
   |--- submit_card ------->|  (Multiple players)
   |                        |
   |<--- voting_started ----|  (When all submitted)
```

### 10. Component-by-Component Breakdown 🎨

#### Game Landing Page
**File**: `web/src/components/GameLanding.tsx`
```typescript
function GameLanding() {
    const { createGame, joinGame } = useGameStore();
    const [roomCode, setRoomCode] = useState('');
    const [playerName, setPlayerName] = useState('');
    
    const handleCreateGame = () => {
        const code = generateRoomCode(); // Business logic
        createGame(code, playerName);    // WebSocket call
    };
    
    const handleJoinGame = () => {
        if (roomCode.length === 4) {     // Business validation
            joinGame(roomCode, playerName);
        }
    };
    
    return (
        <div className="game-landing">
            <h1>DixitMe</h1>
            <input 
                value={playerName} 
                onChange={(e) => setPlayerName(e.target.value)}
                placeholder="Your name"
            />
            
            <div className="game-actions">
                <button onClick={handleCreateGame}>
                    Create Game
                </button>
                
                <div className="join-section">
                    <input 
                        value={roomCode}
                        onChange={(e) => setRoomCode(e.target.value.toUpperCase())}
                        placeholder="ROOM CODE"
                        maxLength={4}
                    />
                    <button onClick={handleJoinGame}>
                        Join Game
                    </button>
                </div>
            </div>
        </div>
    );
}
```

#### Game Board (Main Game Interface)
**File**: `web/src/components/GameBoard.tsx`
```typescript
function GameBoard() {
    const { gameState, currentPhase, submitClue, submitCard, submitVote } = useGameStore();
    const [selectedCard, setSelectedCard] = useState<number | null>(null);
    const [clueText, setClueText] = useState('');
    
    // Render different UI based on game phase
    const renderPhaseContent = () => {
        switch (currentPhase) {
            case 'storytelling':
                return <StorytellingPhase 
                    onSubmitClue={(clue, cardId) => submitClue(clue, cardId)}
                />;
                
            case 'submission':
                return <SubmissionPhase 
                    clue={gameState?.current_round?.clue}
                    onSubmitCard={(cardId) => submitCard(cardId)}
                />;
                
            case 'voting':
                return <VotingPhase 
                    revealedCards={gameState?.current_round?.revealed_cards}
                    onSubmitVote={(cardId) => submitVote(cardId)}
                />;
                
            default:
                return <Lobby />;
        }
    };
    
    return (
        <div className="game-board">
            <GamePhaseIndicator phase={currentPhase} />
            <PlayerHand />
            {renderPhaseContent()}
            <Chat />
        </div>
    );
}
```

## 📚 Phase 4: Advanced Features & Architecture (45-60 minutes)

### 11. Bot AI System 🤖

**Business Need**: Fill games with AI players for better experience

#### Bot Decision Making
**File**: `internal/services/bot/ai.go`
```go
type BotPlayer struct {
    ID          uuid.UUID
    Level       string    // "easy", "medium", "hard"
    Hand        []int
    cardWeights map[int]float64
}

func (bot *BotPlayer) SelectCardAsStoryteller() (int, string, error) {
    // AI logic for storytelling
    selectedCard := bot.Hand[0] // Simplified selection
    
    // Generate clue based on card tags and difficulty
    tags := bot.getCardTags(selectedCard)
    clue := bot.generateClue(tags, bot.Level)
    
    return selectedCard, clue, nil
}

func (bot *BotPlayer) SelectCardForClue(clue string) (int, error) {
    scores := make(map[int]float64)
    
    // Analyze each card in hand against the clue
    for _, cardID := range bot.Hand {
        tags := bot.getCardTags(cardID)
        score := bot.calculateRelevanceScore(clue, tags)
        
        // Add difficulty-based randomness
        if bot.Level == "easy" {
            score += rand.Float64() * 0.5 // More random
        }
        
        scores[cardID] = score
    }
    
    // Select highest scoring card
    return bot.selectBestCard(scores), nil
}
```

#### Card Categorization System
**File**: `assets/tags.json`
```json
{
    "tags": [
        {"slug": "nature", "name": "Nature", "weight": 1.0},
        {"slug": "emotion", "name": "Emotion", "weight": 1.2},
        {"slug": "abstract", "name": "Abstract", "weight": 0.8},
        {"slug": "action", "name": "Action", "weight": 1.1}
    ],
    "card_associations": {
        "1": ["nature", "peace", "water"],
        "2": ["emotion", "joy", "celebration"],
        "3": ["mystery", "abstract", "darkness"]
    }
}
```

### 12. Authentication System 🔐

**Business Need**: Support registered users, guests, and Google SSO

#### Authentication Flow
**File**: `internal/services/auth/service.go`
```go
// Three authentication types
func (a *AuthService) LoginWithPassword(emailOrUsername, password, ipAddress, userAgent string) (*models.User, *models.Session, string, error) {
    // 1. Find user by email or username
    var user models.User
    if err := a.db.Where("email = ? OR username = ?", emailOrUsername, emailOrUsername).First(&user).Error; err != nil {
        return nil, nil, "", errors.New("invalid credentials")
    }
    
    // 2. Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
        return nil, nil, "", errors.New("invalid credentials")
    }
    
    // 3. Create session
    session, token, err := a.createSession(&user, models.AuthTypePassword, ipAddress, userAgent)
    return &user, session, token, err
}

func (a *AuthService) CreateGuestSession(guestName, ipAddress, userAgent string) (*models.Session, string, error) {
    // Guest sessions don't need user accounts
    session := models.Session{
        ID:        uuid.New(),
        UserID:    nil,  // NULL for guests
        AuthType:  models.AuthTypeGuest,
        IPAddress: ipAddress,
        UserAgent: userAgent,
        ExpiresAt: time.Now().Add(24 * time.Hour),
        IsActive:  true,
    }
    
    // Generate JWT token with guest info
    token, err := a.jwtService.GenerateToken(nil, session, guestName, "")
    return &session, token, err
}
```

### 13. Chat System Implementation 💬

**Business Need**: Allow player communication during specific game phases

#### Real-time Chat
**File**: `internal/services/game/chat_actions.go`
```go
func (m *Manager) SendChatMessage(roomCode string, playerID uuid.UUID, message string, messageType string) error {
    game := m.GetGame(roomCode)
    
    // Business rules: when can players chat?
    if !m.canPlayerChat(game, playerID) {
        return errors.New("chat not allowed in current phase")
    }
    
    chatMessage := &models.ChatMessage{
        ID:          uuid.New(),
        GameID:      game.ID,
        PlayerID:    playerID,
        Message:     message,
        MessageType: messageType, // "chat", "system", "emote"
        Phase:       string(game.CurrentRound.Status),
        Timestamp:   time.Now(),
    }
    
    // Persist to database
    if err := m.PersistChatMessage(context.Background(), chatMessage); err != nil {
        return err
    }
    
    // Broadcast to all players in game
    m.BroadcastToGame(game, MessageTypeChatMessage, ChatMessagePayload{
        ID:          chatMessage.ID,
        PlayerID:    playerID,
        PlayerName:  game.Players[playerID].Name,
        Message:     message,
        MessageType: messageType,
        Phase:       chatMessage.Phase,
        Timestamp:   chatMessage.Timestamp,
    })
    
    return nil
}

func (m *Manager) canPlayerChat(game *GameState, playerID uuid.UUID) bool {
    // Business rule: no chat during voting to prevent influence
    if game.CurrentRound != nil && game.CurrentRound.Status == RoundStatusVoting {
        return false
    }
    return true
}
```

## 📚 Phase 5: Hands-On Exercises (60-90 minutes)

### 14. Exercise 1: Trace a Complete Feature 🔍

**Goal**: Follow "Submit Card" from button click to database

1. **Start**: Find the submit button in `web/src/components/SubmissionPhase.tsx`
2. **Frontend**: Trace through the click handler → Zustand store action
3. **WebSocket**: Find the message sending in `gameStore.ts`
4. **Backend**: Locate the handler in `websocket/handlers.go`
5. **Business Logic**: Follow to `round_actions.go` SubmitCard method
6. **Database**: See how it's persisted in `persistence.go`

**Questions to explore:**
- What validations happen at each layer?
- How are errors handled and communicated back?
- What happens if a player disconnects mid-submission?

### 15. Exercise 2: Add a New Feature 🛠️

**Goal**: Add a "ready" status for players in the lobby

#### Step 1: Database Model
```go
// Add to internal/models/player.go
type Player struct {
    // ... existing fields
    IsReady  bool `json:"is_ready" gorm:"default:false"`
}
```

#### Step 2: WebSocket Message
```go
// Add to internal/transport/websocket/types.go
const (
    ClientMessagePlayerReady = "player_ready"
    ServerMessagePlayerReady = "player_ready_updated"
)

type PlayerReadyPayload struct {
    PlayerID uuid.UUID `json:"player_id"`
    IsReady  bool      `json:"is_ready"`
}
```

#### Step 3: Backend Handler
```go
// Add to internal/services/game/game_actions.go
func (m *Manager) SetPlayerReady(roomCode string, playerID uuid.UUID, isReady bool) error {
    game := m.GetGame(roomCode)
    
    if game.Status != GameStatusWaiting {
        return errors.New("can only set ready in lobby")
    }
    
    player := game.Players[playerID]
    player.IsReady = isReady
    
    // Broadcast to all players
    m.BroadcastToGame(game, ServerMessagePlayerReady, PlayerReadyPayload{
        PlayerID: playerID,
        IsReady:  isReady,
    })
    
    return nil
}
```

#### Step 4: Frontend Component
```typescript
// Add to web/src/components/Lobby.tsx
const ToggleReadyButton = () => {
    const { setPlayerReady, gameState } = useGameStore();
    const currentPlayer = getCurrentPlayer(gameState);
    
    return (
        <button 
            onClick={() => setPlayerReady(!currentPlayer.is_ready)}
            className={currentPlayer.is_ready ? 'ready' : 'not-ready'}
        >
            {currentPlayer.is_ready ? 'Ready!' : 'Not Ready'}
        </button>
    );
};
```

### 16. Exercise 3: Debug a Real Issue 🐛

**Scenario**: Players report that votes aren't being counted correctly

#### Investigation Steps:
1. **Reproduce**: Create a test game with 3 players
2. **Frontend Debugging**:
   ```typescript
   // Add to voting submission
   console.log('Submitting vote:', { cardId, playerId });
   ```
3. **Backend Logging**:
   ```go
   // Add to SubmitVote method
   logger.Info("Vote received", 
       "player", playerID, 
       "card", cardID, 
       "round", game.CurrentRound.ID)
   ```
4. **Database Verification**:
   ```sql
   -- Check vote records
   SELECT * FROM votes WHERE round_id = 'your-round-id';
   ```

## 💡 Technology Deep Dives

### Go Concepts in Action 🐹

#### 1. Interfaces & Dependency Injection
```go
// Define what we need, not how it's implemented
type GameService interface {
    CreateGame(roomCode string, creatorID uuid.UUID, creatorName string) (*GameState, error)
    JoinGame(roomCode string, playerID uuid.UUID, playerName string) (*GameState, error)
}

// Handler depends on interface, not concrete type
type GameHandlers struct {
    gameService GameService  // Can be real service or mock for testing
}
```

#### 2. Goroutines for Real-time Features
```go
func (m *Manager) ProcessBotActions(game *GameState) {
    // Run bot thinking in parallel
    go func() {
        for _, player := range game.Players {
            if player.IsBot {
                // Bot AI runs concurrently
                m.processBotTurn(player, game)
            }
        }
    }()
}
```

#### 3. Channels for Communication
```go
type Manager struct {
    games       map[string]*GameState
    gameUpdates chan GameUpdate  // For broadcasting
    stopCleanup chan bool        // For graceful shutdown
}
```

### React Patterns in Practice ⚛️

#### 1. State Management with Zustand
```typescript
// Simple, efficient state management
const useGameStore = create<GameStore>((set, get) => ({
    gameState: null,
    
    // Actions update state immutably
    updateGameState: (newState) => set({ gameState: newState }),
    
    // Complex logic in actions
    submitCard: (cardId) => {
        const { websocket, gameState } = get();
        websocket.send(JSON.stringify({
            type: 'submit_card',
            payload: { card_id: cardId }
        }));
    }
}));
```

#### 2. Component Composition
```typescript
// Each component has a single responsibility
function GameBoard() {
    return (
        <div className="game-board">
            <GamePhaseIndicator />  {/* Shows current phase */}
            <PlayerList />          {/* Shows all players */}
            <GameContent />         {/* Phase-specific content */}
            <PlayerHand />          {/* Player's cards */}
            <Chat />               {/* Communication */}
        </div>
    );
}
```

## 🚀 Advanced Topics

### Database Architecture 🗄️

#### GORM Relationships
```go
type Game struct {
    ID      uuid.UUID `gorm:"primaryKey"`
    Players []Player  `gorm:"foreignKey:GameID"`  // One-to-many
    Rounds  []Round   `gorm:"foreignKey:GameID"`  // One-to-many
}

type Player struct {
    GameID uuid.UUID `gorm:"index"`
    Game   Game      `gorm:"foreignKey:GameID"`  // Belongs-to
}
```

#### Transaction Management
```go
func (m *Manager) CompleteRound(game *GameState) error {
    // Multiple operations must succeed together
    return m.db.Transaction(func(tx *gorm.DB) error {
        // 1. Update scores
        if err := m.updatePlayerScores(tx, game); err != nil {
            return err  // Rollback everything
        }
        
        // 2. Complete round
        if err := m.markRoundComplete(tx, game.CurrentRound); err != nil {
            return err  // Rollback everything
        }
        
        // 3. Check game end
        if winner := m.checkWinner(game); winner != nil {
            return m.endGame(tx, game, winner)
        }
        
        return nil  // Commit all changes
    })
}
```

### Performance Considerations ⚡

#### WebSocket Connection Management
```go
type ConnectionManager struct {
    connections map[uuid.UUID]*websocket.Conn
    broadcast   chan []byte
    register    chan *Connection
    unregister  chan *Connection
}

func (cm *ConnectionManager) Run() {
    for {
        select {
        case conn := <-cm.register:
            cm.connections[conn.PlayerID] = conn.WebSocket
            
        case conn := <-cm.unregister:
            delete(cm.connections, conn.PlayerID)
            
        case message := <-cm.broadcast:
            // Send to all connected players efficiently
            for playerID, conn := range cm.connections {
                if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
                    // Handle disconnection
                    delete(cm.connections, playerID)
                }
            }
        }
    }
}
```

## 🎯 Next Steps & Best Practices

### Development Workflow 🔄

1. **Start with Business Logic**: Understand what you're building
2. **Design the Data Flow**: Frontend → WebSocket → Business Logic → Database
3. **Implement in Layers**: Database models → Business logic → API → Frontend
4. **Test Incrementally**: Each layer should work before moving to the next
5. **Use the Debugger**: Set breakpoints in both Go and TypeScript

### Testing Strategy 🧪

```go
// Example unit test for game logic
func TestCreateGame(t *testing.T) {
    manager := NewManager()
    
    game, err := manager.CreateGame("TEST", uuid.New(), "Alice")
    
    assert.NoError(t, err)
    assert.Equal(t, "TEST", game.RoomCode)
    assert.Equal(t, GameStatusWaiting, game.Status)
    assert.Len(t, game.Players, 1)
}
```

### Common Patterns to Follow 📋

1. **Validate Early**: Check business rules before processing
2. **Handle Errors Gracefully**: Return meaningful error messages
3. **Log Important Events**: Use structured logging
4. **Keep Functions Small**: Single responsibility principle
5. **Use Interfaces**: Make code testable and flexible

## 🐛 Common Pitfalls & Solutions

1. **Go**: Forgetting to handle errors → Always check `if err != nil`
2. **React**: Forgetting useEffect dependencies → Use ESLint React hooks plugin
3. **WebSocket**: Not handling disconnections → Implement reconnection logic
4. **Database**: N+1 query problems → Use GORM preloading
5. **State Management**: Mutating state directly → Use immutable updates

## ❓ When You Get Stuck

1. **Read Error Messages Carefully** - Both Go and TypeScript give helpful errors
2. **Use Console/Logs** - Add logging to trace data flow
3. **Check the Network Tab** - See WebSocket messages in browser dev tools
4. **Isolate the Problem** - Test individual functions/components
5. **Ask Specific Questions** - "Why does this function return nil?" vs "It's broken"

## 🎉 Conclusion

You now understand both:
- **The Business**: How Dixit works as a game
- **The Code**: How our implementation supports the game flow
- **The Architecture**: How frontend, backend, and database work together
- **The Patterns**: Go and React best practices in a real project

**Ready to build amazing features!** 🚀

### Quick Reference
- **Game Flow**: Create → Join → Start → Round (Storytelling → Submission → Voting → Scoring) → Repeat → End
- **Data Flow**: Frontend State ↔ WebSocket ↔ Business Logic ↔ Database
- **Key Files**: 
  - `internal/services/game/` - All game logic
  - `web/src/store/gameStore.ts` - Frontend state
  - `internal/transport/websocket/` - Real-time communication

Happy coding! 🎮✨
