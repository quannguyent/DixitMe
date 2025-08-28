# DixitMe - Project Onboarding Guide

Welcome to DixitMe! This guide will help you understand the project structure, set up your development environment, and explore the features we've built. This is your roadmap to getting productive quickly.

> 🔰 **New to Go or React?** Start with our beginner guides:
> - [Go Beginner's Guide](./guides/golang-beginners.md) - Learn Go programming fundamentals
> - [React Beginner's Guide](./guides/react-beginners.md) - Learn React & TypeScript basics

## 🎯 What is DixitMe?

DixitMe is a digital implementation of the popular board game "Dixit" - a creative storytelling and guessing game for 3-6 players.

## 📚 Quick Start

### 1. Game Overview 🎮

**Dixit Basics:**
- **Players**: 3-6 players per game
- **Cards**: 84 beautifully illustrated cards with abstract imagery  
- **Goal**: Score 30 points through creative storytelling and guessing

**How a Round Works:**
1. **Storytelling**: Current storyteller picks a card and gives a clue
2. **Submission**: Other players submit cards that match the clue
3. **Voting**: Players vote for which card they think belongs to the storyteller
4. **Scoring**: Points awarded based on correct guesses and deception

**Victory**: First to 30 points wins, or highest score when deck runs out

### 2. Project Architecture Flow 🏗️

**High-Level Architecture:**
```
Frontend (React)  ←→  Backend (Go)  ←→  Database (PostgreSQL)
     ↑                    ↑                    ↑
  User Interface      Game Logic           Data Storage
     ↓                    ↓                    ↓
- Game screens      - Room management     - Game history
- Real-time UI      - Round logic        - Player data  
- WebSocket conn    - Bot AI             - Chat logs
```

**Data Flow Overview:**
1. **User Action** → Frontend captures user input (join game, submit card, etc.)
2. **WebSocket Message** → Frontend sends structured message to backend
3. **Business Logic** → Backend processes game rules and validates actions
4. **Database Update** → Backend persists game state and history
5. **Broadcast** → Backend sends updates to all connected players
6. **UI Update** → Frontend receives updates and refreshes game state

**Key Components:**
- **Frontend**: React app with real-time WebSocket communication
- **Backend**: Go server handling game logic, WebSocket connections, and API
- **Database**: PostgreSQL for persistence, Redis for sessions
- **Assets**: Card images and game data
- **Bot AI**: Intelligent computer players for better gameplay

### 3. Development Environment Setup 🛠️

**Prerequisites:**
- Go 1.21+
- Node.js 18+  
- Docker & Docker Compose

**Quick Setup:**
```bash
# 1. Environment setup
cp configs/config.env.example .env

# 2. Start dependencies  
docker-compose -f deployments/docker/docker-compose.dev.yml up -d

# 3. Install dependencies
go mod download
cd web && npm install && cd ..

# 4. Start backend
go run cmd/server/main.go

# 5. Start frontend (new terminal)
cd web && npm start
```

**Access Points:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080  
- Swagger Docs: http://localhost:8080/swagger/

### 4. Project Structure Overview 🏗️

```
DixitMe/
├── cmd/                     # Application entry points
│   ├── server/             # Main server application
│   └── seed/               # Database seeding utility
├── internal/                # Private application code
│   ├── app/                # Application initialization
│   ├── config/             # Configuration management
│   ├── models/             # Database models
│   ├── services/           # Business logic
│   │   ├── auth/          # Authentication & authorization
│   │   ├── bot/           # AI bot logic
│   │   └── game/          # Core game mechanics
│   ├── transport/          # HTTP & WebSocket handlers
│   └── storage/            # File storage (MinIO)
├── web/                     # React frontend application
│   ├── src/
│   │   ├── components/    # React components
│   │   ├── store/         # State management
│   │   └── types/         # TypeScript definitions
│   └── public/            # Static assets
├── assets/                  # Game assets (card images)
├── configs/                 # Configuration files
├── deployments/            # Docker & deployment configs
└── docs/                   # API documentation
```

> 📋 **For detailed structure**, see [PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md)

## 🎮 Core Features & Game Flow

**Complete Game Journey:**
```
1. Lobby → 2. Game Start → 3. Round Cycle → 4. Game End
     ↓             ↓              ↓              ↓
- Create/Join  - Deal cards   - Storytelling   - Winner
- Wait for     - Select       - Submission     - Scores
  players        storyteller   - Voting        - History
                              - Scoring
```

### Detailed Game Flow Walkthrough

Let's trace through how a complete game works, from creation to completion:

#### 1. Game Creation Flow 🎯

**Step-by-Step Process:**

**Frontend (User Creates Game):**
1. User enters their name on landing page (`GameLanding.tsx`)
2. Clicks "Create Game" → generates 4-character room code (e.g., "XYZW")
3. WebSocket connection established with authentication token
4. `createGame()` function sends message: `{type: 'create_game', payload: {room_code, player_name}}`

**Backend (Game Manager Processing):**
1. WebSocket handler (`websocket/handlers.go`) receives message
2. Calls `game.Manager.CreateGame()` in business logic layer
3. **Validation**: Checks room code format, ensures uniqueness
4. **Game State Creation**: 
   - Creates new `GameState` object with unique ID
   - Initializes empty player map, shuffled deck of 84 cards
   - Sets status to "waiting"
   - Adds creator as first player
5. **Persistence**: Saves game to PostgreSQL database
6. **Response**: Sends `game_created` message back to frontend

**What Happens:**
- Game appears in creator's browser instantly
- Room code displayed for sharing with friends
- Creator can invite others or add bots
- Game state stored both in-memory (for speed) and database (for persistence)

#### 2. Joining Existing Games 🚪

**Join Flow Process:**

**Frontend (User Joins):**
1. User enters room code and name
2. Validation: Checks code is 4 characters
3. Sends `{type: 'join_game', payload: {room_code, player_name}}`

**Backend Processing:**
1. `Manager.JoinGame()` receives request
2. **Validation**:
   - Game exists and is joinable
   - Room not full (max 6 players)
   - Player name not already taken
3. **Player Addition**:
   - Creates new `Player` object
   - Adds to game's player map
   - Updates game state
4. **Broadcasting**: Sends `player_joined` to ALL players in room
5. **Database**: Updates player records

**Real-time Updates:**
- All existing players see new player immediately
- Lobby updates with current player count
- Ready status indicators appear
- Chat becomes available

#### 3. Game Start & Card Dealing Flow 🎴

**When Creator Starts Game:**

**Frontend Trigger:**
1. Creator clicks "Start Game" button in lobby
2. Validates minimum 3 players present
3. Sends `{type: 'start_game'}` message

**Backend Game Initialization:**
1. `Manager.StartGame()` processes request
2. **Business Rules Check**:
   - Minimum 3 players validation
   - Game status must be "waiting"
   - Only creator can start
3. **Card Dealing Process**:
   - Deals 6 cards to each player from shuffled deck
   - Removes dealt cards from main deck
   - Updates each player's hand in game state
4. **Round Setup**:
   - Selects first storyteller (usually creator)
   - Creates first `Round` object with unique ID
   - Sets round status to "storytelling"
   - Assigns storyteller for this round
5. **State Updates**:
   - Game status → "in_progress"
   - Round number → 1
   - Current phase → "storytelling"
6. **Broadcasting**: Sends `game_started` to all players with:
   - Updated game state
   - Each player's hand (private)
   - Current storyteller assignment

**Frontend Response:**
- UI transitions from lobby to game board
- Players see their 6 cards
- Game phase indicator shows "Storytelling Phase"
- Only storyteller sees clue input interface

#### 4. Round Management: The Core Game Loop 🔄

**Phase 1: Storytelling Phase**

**Storyteller's Turn:**
1. Storyteller selects one card from their hand
2. Types a creative clue (word or phrase)
3. Clicks "Submit Clue" 
4. Frontend sends: `{type: 'submit_clue', payload: {clue, card_id}}`

**Backend Processing:**
1. `Manager.SubmitClue()` handles request
2. **Validation**:
   - Only current storyteller can submit
   - Game must be in storytelling phase
   - Card must be in storyteller's hand
3. **State Updates**:
   - Stores clue text in round object
   - Records storyteller's card ID
   - Removes card from storyteller's hand
   - Changes round status to "submission"
4. **Broadcasting**: Sends `clue_submitted` with clue text to all players

**UI Transition:**
- All players see the clue
- Storyteller's interface shows "Waiting for others..."
- Other players see card selection interface

**Phase 2: Card Submission Phase**

**Each Player's Submission:**
1. Players read the clue and examine their cards
2. Select one card that fits the clue
3. Click selected card → sends `{type: 'submit_card', payload: {card_id}}`

**Backend Processing per Submission:**
1. `Manager.SubmitCard()` processes each submission
2. **Validation**:
   - Player cannot be the storyteller
   - Player hasn't already submitted
   - Card exists in player's hand
3. **Storage**:
   - Creates `CardSubmission` object
   - Links player ID to card ID
   - Removes card from player's hand
4. **Progress Tracking**:
   - Counts total submissions
   - When all non-storytellers have submitted → auto-advance
5. **Phase Transition**: Calls `startVotingPhase()` when ready

**Voting Phase Setup:**
1. **Card Collection**:
   - Gathers storyteller's card + all submitted cards
   - Creates `RevealedCard` objects with card IDs
2. **Shuffling**: Randomizes card order so storyteller's isn't obvious
3. **State Update**: Round status → "voting"
4. **Broadcasting**: Sends `voting_started` with shuffled card array

**Phase 3: Voting Phase**

**Player Voting:**
1. All players (except storyteller) see shuffled cards
2. Each player votes for storyteller's card
3. Frontend sends: `{type: 'submit_vote', payload: {card_id}}`

**Backend Vote Processing:**
1. `Manager.SubmitVote()` handles each vote
2. **Validation**:
   - Storyteller cannot vote
   - Player hasn't voted yet
   - Valid card selection
3. **Vote Storage**:
   - Creates `Vote` object linking player to card
   - Tracks vote in round data
4. **Completion Check**: When all votes received → `calculateScores()`

**Phase 4: Scoring & Results**

**Score Calculation Process:**
1. **Vote Analysis**:
   - Counts votes for storyteller's actual card
   - Identifies who voted correctly
   - Counts votes for each submitted card
2. **Dixit Scoring Rules**:
   - If 0 or ALL players guessed correctly: Everyone except storyteller gets 2 points
   - Otherwise: Storyteller + correct guessers get 3 points
   - Bonus: +1 point for each vote your submitted card received
3. **State Updates**:
   - Updates each player's total score
   - Records round results
   - Checks for game end conditions (30 points or deck empty)

**End Round Processing:**
- Saves round results to database
- If game continues: starts new round with next storyteller
- If game ends: transitions to final results screen

### 5. Authentication & Session Management 🔐

**Guest Authentication Flow:**
1. User enters name on landing page
2. Frontend calls `/auth/guest` endpoint
3. Backend `AuthService.CreateGuestSession()`:
   - Creates session with NULL user_id
   - Generates JWT token with guest info
   - Sets 24-hour expiration
   - Returns token to frontend
4. Token stored in localStorage for persistence
5. All WebSocket connections include this token

**Password Authentication Flow:**
1. User submits email/username + password
2. Backend `AuthService.LoginWithPassword()`:
   - Looks up user in database
   - Verifies password with bcrypt
   - Creates session linked to user account
   - Generates JWT with user info
3. Session tracking includes IP, user agent for security

### 6. AFK Detection & Player Management 🕐

The system includes sophisticated AFK (Away From Keyboard) detection and automatic player replacement to ensure games continue smoothly when players disconnect or leave.

**AFK Detection Logic:**

Players are considered AFK in these situations:
1. **Disconnected Players**: Lost WebSocket connection for >3 minutes
2. **Players Who Left**: Manually left during an active game
3. **Inactive Players**: No WebSocket activity for >3 minutes

**Automatic Bot Replacement Flow:**

```
Player Disconnects/Leaves → Mark as AFK → Wait 3 mins → Replace with Bot
                                ↓
                          Update LastActivity
                                ↓
                        Cleanup Service Detects
                                ↓
                         Automatic Replacement
```

**Backend AFK Processing:**

1. **Activity Tracking**:
   - Every WebSocket message updates player's last activity timestamp
   - System tracks both connection status and manual leave actions
   - Players marked as AFK if disconnected for more than timeout duration OR if they left the game manually

2. **Automatic Replacement Logic**:
   - When a player is detected as AFK, system creates a replacement bot
   - Bot inherits the original player's score, position, and cards
   - Game state is preserved seamlessly - other players see no disruption
   - All players are notified of the replacement via broadcast message

3. **Periodic Monitoring Service**:
   - Background service runs every 2 minutes checking all active games
   - First checks if ALL human players in a game are AFK
   - If all are AFK: ends the game immediately
   - If only some are AFK: replaces individual AFK players with bots
   - Abandoned games are cleaned up faster (2 minutes vs 30 minutes)

**Game Ending Due to All AFK:**

When all human players go AFK:
1. **Detection**: System checks if every human player is either disconnected or has left
2. **Game End**: Game status automatically changed to "abandoned" 
3. **Cleanup**: Abandoned games are removed after 2 minutes (vs 30 minutes for normal games)
4. **Notification**: All players receive "Game ended: All players went AFK" message

**Leave Game Behavior:**

- **Waiting Games**: Players are completely removed
- **Active Games**: Players marked as inactive (AFK) but stay in game for potential bot replacement

**Key System Components:**
- **AFK Scanner** - Periodically scans all games for inactive players
- **All-AFK Detector** - Identifies when all human players have gone AFK
- **Game Ender** - Gracefully terminates abandoned games
- **Activity Tracker** - Monitors WebSocket activity to detect disconnections

This system ensures games never stall due to disconnections while maintaining fair gameplay.

## 🔧 Key Functions Overview

### Core Backend Functions

**Game Management (`internal/services/game/`)**
- `CreateGame()` - Sets up new game room with shuffled deck and initial player
- `JoinGame()` - Adds players to existing games with validation
- `StartGame()` - Deals cards and begins first round
- `GetGame()` - Retrieves active game state from memory
- `RemovePlayer()` - Handles player leaving (removes from waiting games, marks AFK in active games)
- `LeaveGame()` - Player-initiated leave (calls RemovePlayer)

**Round Management**
- `SubmitClue()` - Handles storyteller's card and clue submission
- `SubmitCard()` - Processes player card submissions for clues
- `SubmitVote()` - Records votes and triggers scoring when complete
- `calculateScores()` - Applies Dixit scoring rules and updates player totals

**AFK Detection & Bot Replacement**
- **Player Replacement** - Replaces AFK/disconnected players with bots (preserves score/cards)
- **AFK Scanning** - Periodically scans games for inactive players and replaces them
- **All-AFK Detection** - Identifies when all human players are AFK and ends the game
- **Game Termination** - Gracefully ends games when all players go AFK
- **Activity Monitoring** - Tracks player WebSocket activity to detect disconnections

**WebSocket Communication (`internal/transport/websocket/`)**
- `handleMessage()` - Routes incoming messages to appropriate handlers
- `BroadcastToGame()` - Sends real-time updates to all players in a room

**Database Operations (`internal/services/game/persistence.go`)**
- `PersistGame()` - Saves game state to PostgreSQL
- `PersistRound()` - Stores round data, submissions, and votes
- `LoadGameState()` - Reconstructs game from database for recovery

### Frontend Functions

**State Management (`web/src/store/gameStore.ts`)**
- `useGameStore` - Central Zustand store for all game state
- `connect()` - Establishes WebSocket connection with authentication
- `sendMessage()` - Formats and sends actions to backend
- `handleMessage()` - Processes incoming updates and updates UI

**UI Components (`web/src/components/`)**
- `GameLanding` - Handles game creation and joining
- `GameBoard` - Main game interface that switches between phases
- Phase-specific components for storytelling, submission, voting

### Bot AI Functions

**Decision Making (`internal/services/bot/ai.go`)**
- `SelectCardAsStoryteller()` - AI picks card and generates clue based on difficulty
- `SelectCardForClue()` - AI analyzes clue and chooses best matching card
- `calculateRelevanceScore()` - Scores how well cards match given clues

## 🔗 How Functions Work Together

### Simple Function Flow Examples

**Creating a Game:**
```
User clicks "Create" → Frontend sends WebSocket message → Backend creates GameState → 
Saves to database → Broadcasts to frontend → UI updates with new game
```

**Player Joins:**
```
User enters room code → Frontend validates → Backend checks game exists → 
Adds player to game → Broadcasts to all players → All UIs update with new player
```

**Submitting a Clue:**
```
Storyteller picks card + clue → Frontend sends submission → Backend validates → 
Updates round state → Broadcasts clue → All players see submission phase
```

### Data Storage Strategy
- **In-Memory**: Active game states for fast real-time access
- **Database**: Persistent storage for game history and recovery
- **WebSocket**: Real-time synchronization between all players

## 🛠️ Technical Architecture

### Backend Technologies

**Go (Golang):**
- RESTful API with Gin framework
- WebSocket connections for real-time features
- GORM for database operations
- Structured logging with custom logger
- JWT authentication and session management

**Database:**
- **PostgreSQL**: Primary data storage (games, users, rounds, chat)
- **Redis**: Session storage and caching
- **MinIO**: Object storage for future file uploads

### Frontend Technologies

**React with TypeScript:**
- Create React App for quick development [[memory:6475929]]
- Component-based architecture with CSS modules
- Zustand for state management (lightweight alternative to Redux)
- Real-time WebSocket integration
- Responsive design for multiple screen sizes

## 🚀 Development Workflow

### Development Best Practices

**1. Start with Setup:**
- Use the setup instructions above to get your environment running
- Familiarize yourself with the project structure
- Test that you can create a game and see real-time updates

**2. Understand the Flow:**
- Create a game → Join with multiple browser tabs → Play through a round
- Watch the network tab to see WebSocket messages
- Observe how state changes propagate between players

**3. Code Organization:**
- **Backend**: Start with `internal/services/game/` for game logic
- **Frontend**: Explore `web/src/components/` for UI components  
- **Models**: Check `internal/models/` for database structures
- **WebSocket**: Review `internal/transport/websocket/` for real-time logic

## 📋 Available Scripts & Commands

```bash
# Run server in development mode
go run cmd/server/main.go

# Run tests
go test ./...

# Seed database with cards
go run cmd/seed/main.go

# Generate Swagger docs  
./scripts/generate-swagger.sh
```

### Frontend Development  
```bash
cd web

# Start development server
npm start

# Build for production
npm run build

# Run tests
npm test
```

### Database Operations
```bash
# Start PostgreSQL and Redis
docker-compose -f deployments/docker/docker-compose.dev.yml up -d

# Connect to database
docker exec -it dixitme-postgres psql -U postgres -d dixitme

# View logs
docker-compose -f deployments/docker/docker-compose.dev.yml logs -f
```

## 🎯 Getting Productive Quickly

### Common Development Tasks

**1. Add a New Game Feature:**
- Define the feature in `internal/models/` (if new data needed)
- Implement business logic in `internal/services/game/`
- Add WebSocket handlers in `internal/transport/websocket/`
- Create frontend components in `web/src/components/`
- Update state management in `web/src/store/`

**2. Debug Issues:**
- Check browser console for frontend errors
- View backend logs for server-side issues
- Use browser Network tab for WebSocket message inspection
- Add strategic console.log/logger statements

**3. Test New Features:**
- Open multiple browser tabs to simulate multiple players
- Use incognito mode for different user sessions
- Test bot integration by adding AI players
- Verify database persistence by checking data after game completion

## 🔧 Key Configuration Files

- **Backend Config**: `configs/config.env.example` → copy to `.env`
- **Database Setup**: `deployments/docker/docker-compose.dev.yml`
- **Frontend Config**: `web/package.json` and `web/tsconfig.json`
- **Server Config**: `configs/server.yaml`

## 🎮 Feature Roadmap & Extensibility

**Current Features:**
- ✅ Core Dixit gameplay (all phases implemented)
- ✅ Real-time multiplayer via WebSocket
- ✅ Bot AI with difficulty levels
- ✅ Guest and password authentication
- ✅ Chat system with phase restrictions
- ✅ Persistent game history

**Future Enhancements:**
- 🔄 Google SSO integration
- 🔄 Custom card packs
- 🔄 Spectator mode
- 🔄 Voice chat integration

**Easy First Contributions:**
- Add more bot personality types
- Implement player ready status in lobby
- Add game replay functionality  
- Create admin dashboard for game monitoring
- Improve UI/UX with animations and transitions

## 🚀 Performance & Scaling Considerations

**Current Architecture Supports:**
- ~100 concurrent games (in-memory game state)
- Real-time updates via efficient WebSocket broadcasting
- Database optimization with GORM relationship loading
- Redis session management for scalability

**For Higher Scale:**
- Move game state to Redis for multi-server deployment
- Implement horizontal scaling with load balancers
- Add CDN for static assets (card images)
- Consider microservices architecture for specialized features

Happy coding! 🎮✨