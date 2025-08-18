# DixitMe - Online Dixit Card Game

A full-stack implementation of the popular Dixit card game with real-time multiplayer support.

## Features

- **Real-time multiplayer gameplay** with WebSocket connections
- **Complete Dixit game flow**: storytelling, card submission, voting, and scoring
- **Modern web interface** built with React and TypeScript
- **Scalable backend** with Go, PostgreSQL, and Redis
- **Beautiful card animations** and responsive design
- **Game lobby system** with room codes
- **Player statistics** and game history
- **Real-time chat** in lobby and voting phases with system notifications
- **Flexible authentication** supporting registered users, Google SSO, and guest access

## Tech Stack

### Backend (Go)
- **Framework**: Gin web framework
- **Database**: PostgreSQL with GORM
- **Cache**: Redis for state synchronization  
- **Storage**: MinIO object storage for card images
- **WebSockets**: Gorilla WebSocket for real-time communication
- **API Documentation**: Swagger/OpenAPI 2.0 with interactive UI
- **Logging**: Structured logging with slog (JSON/text formats)
- **AI System**: Heuristic bot players with weighted random selection
- **Asset Management**: Automated database seeding with 84+ cards
- **Chat System**: Real-time messaging with phase-based restrictions and system notifications
- **Authentication**: JWT-based auth with Google OAuth2, password, and guest support
- **Architecture**: Monolithic with clean separation of concerns

### Frontend (React)
- **Framework**: React 18 with TypeScript
- **State Management**: Zustand for game state
- **Styling**: CSS-in-JS with responsive design
- **WebSocket Client**: Native WebSocket API with reconnection logic

## Game Rules

### Setup
- 3-6 players per game
- Each player gets 6 cards from the Dixit deck
- Game lasts for 2 rounds per player (storyteller rotates)

### Round Flow
1. **Storytelling**: Storyteller picks a card and gives a clue
2. **Submission**: Other players submit cards that fit the clue
3. **Voting**: Players vote for the storyteller's card among shuffled submissions
4. **Scoring**: Points awarded based on voting results

### Scoring Rules
- If all or no players guess correctly: Storyteller gets 0 points, others get 2
- Otherwise: Storyteller + correct guessers get 3 points
- Players get 1 additional point for each vote their card receives (except storyteller's card)

## Installation & Setup

### Prerequisites
- Go 1.21+ 
- Node.js 16+
- PostgreSQL 13+
- Redis 6+

### Backend Setup

1. **Clone and navigate to project:**
   ```bash
   git clone <repository-url>
   cd DixitMe
   ```

2. **Install Go dependencies:**
   ```bash
   go mod tidy
   ```

3. **Set up environment variables:**
   ```bash
   cp config.env.example .env
   # Edit .env with your database and Redis URLs
   ```

4. **Create PostgreSQL database:**
   ```sql
   CREATE DATABASE dixitme;
   ```

5. **Generate Swagger documentation (optional):**
   ```bash
   ./scripts/generate-swagger.sh
   ```

6. **Run the backend:**
   ```bash
   go run cmd/server/main.go
   ```

   The server will start on `http://localhost:8080`
   - **API**: `http://localhost:8080/api/v1/`
   - **Swagger UI**: `http://localhost:8080/swagger/index.html`

### Frontend Setup

1. **Navigate to frontend directory:**
   ```bash
   cd web
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Start development server:**
   ```bash
   npm start
   ```

   The frontend will start on `http://localhost:3000`

### Production Build

1. **Build frontend:**
   ```bash
   cd web
   npm run build
   ```

2. **Build backend:**
   ```bash
   go build -o dixitme cmd/server/main.go
   ```

3. **Run production server:**
   ```bash
   ./dixitme
   ```

## API Endpoints

### 📊 Interactive API Documentation
- **Swagger UI**: `GET /swagger/index.html` - Interactive API documentation
- **OpenAPI JSON**: `GET /swagger/doc.json` - Machine-readable API specification
- **OpenAPI YAML**: Access the YAML specification in `docs/swagger.yaml`

### REST API
- `GET /health` - Health check
- `POST /api/v1/players` - Create player
- `GET /api/v1/players/:id` - Get player info
- `GET /api/v1/player/:player_id/stats` - Get player statistics
- `GET /api/v1/player/:player_id/history` - Get player game history
- `GET /api/v1/games` - List games (with pagination and filtering)
- `GET /api/v1/games/:room_code` - Get game info
- `GET /api/v1/cards` - Get card list

### WebSocket
- `GET /ws` - WebSocket connection for real-time game updates

> **💡 Tip**: Visit `http://localhost:8080/swagger/index.html` after starting the server to explore the API interactively!

## WebSocket Messages

### Client → Server
- `create_game` - Create new game room
- `join_game` - Join existing game
- `start_game` - Start game (when enough players)
- `submit_clue` - Storyteller submits clue + card
- `submit_card` - Player submits card for round
- `submit_vote` - Player votes for storyteller's card
- `leave_game` - Leave current game

### Server → Client
- `connection_established` - Connection confirmation
- `game_state` - Full game state update
- `player_joined/left` - Player join/leave notifications
- `game_started` - Game start notification
- `round_started` - New round notification
- `clue_submitted` - Clue announcement
- `voting_started` - Voting phase with revealed cards
- `round_completed` - Round results with scores
- `game_completed` - Final game results

## File Structure

```
DixitMe/
├── cmd/server/main.go          # Main application entry
├── internal/
│   ├── config/                 # Configuration management
│   ├── database/              # Database setup and migrations
│   ├── redis/                 # Redis client setup
│   ├── models/                # Database models
│   ├── game/                  # Game logic and state management
│   ├── websocket/             # WebSocket handlers
│   └── handlers/              # HTTP API handlers
├── web/                       # React frontend
│   ├── src/
│   │   ├── components/        # React components
│   │   ├── store/            # Zustand store
│   │   └── types/            # TypeScript types
│   └── public/               # Static assets
├── assets/cards/             # Card images
└── go.mod                    # Go module definition
```

## Development

### API Documentation
- **Regenerate Swagger docs** after making API changes:
  ```bash
  ./scripts/generate-swagger.sh
  ```
- **Swagger annotations**: Add/update `@Summary`, `@Description`, etc. in handler functions
- **Custom types**: Document request/response structures with proper JSON tags

### Adding New Cards
1. Add card images to `assets/cards/` (numbered 1.jpg, 2.jpg, etc.)
2. Update card count in `handlers/handlers.go` if needed
3. Restart server to serve new assets

### Database Migrations
The application uses GORM AutoMigrate for database schema management. Schema changes are applied automatically on startup.

### Testing WebSocket Connection
You can test the WebSocket connection using browser developer tools:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (event) => console.log(JSON.parse(event.data));
ws.send(JSON.stringify({
  type: 'create_game',
  payload: { room_code: 'TEST', player_name: 'TestPlayer' }
}));
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is for educational purposes. Dixit is a trademark of Libellud.

## Deployment

### Docker (Optional)
You can containerize the application:

```dockerfile
# Backend Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o dixitme cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/dixitme .
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/web/build ./web/build
CMD ["./dixitme"]
```

### Environment Variables
- `DATABASE_URL` - PostgreSQL connection string
- `REDIS_URL` - Redis connection string  
- `PORT` - Server port (default: 8080)
- `GIN_MODE` - Gin mode (debug/release)
- `LOG_LEVEL` - Logging level (debug/info/warn/error, default: info)
- `LOG_FORMAT` - Log output format (text/json, default: text)

## Card Asset Management

### Default Card Database

The system includes a comprehensive card database with 84 pre-designed cards and a tagging system:

#### Card Categories
- **Nature**: Forest, Ocean, Mountain, Sky, Desert, Garden, Storm (20 cards)
- **Fantasy**: Magic, Dragons, Wizards, Fairies, Castles, Quests (20 cards)  
- **Emotions/Activities**: Happy, Sad, Love, Fear, Dance, Music, Art (20 cards)
- **Objects**: Keys, Mirrors, Clocks, Books, Crowns, Instruments (20 cards)
- **Abstract**: Dreams, Memory, Balance, Transformation (4 cards)

#### Tag System
- **50+ semantic tags** organized by category (emotion, nature, fantasy, activity, object, abstract, time)
- **Weighted scoring** for bot AI decision making
- **Color-coded** tags for UI organization
- **Many-to-many** relationships between cards and tags

### Database Seeding

#### Automatic Seeding
Cards and tags are automatically seeded when the server starts:

```bash
go run cmd/server/main.go
# Automatically seeds 50+ tags and 84 cards on first run
```

#### Manual Seeding CLI
Use the dedicated seeding command for more control:

```bash
# Seed everything (tags + cards)
go run cmd/seed/main.go

# Seed only tags
go run cmd/seed/main.go -tags

# Seed only cards (requires existing tags)
go run cmd/seed/main.go -cards

# Force complete reseed (deletes existing data)
go run cmd/seed/main.go -force

# Show help
go run cmd/seed/main.go -help
```

#### API Seeding Endpoints
Manage seeding through the REST API:

```bash
# Complete database seeding
curl -X POST http://localhost:8080/api/v1/admin/seed

# Seed only tags
curl -X POST http://localhost:8080/api/v1/admin/seed/tags

# Seed only cards
curl -X POST http://localhost:8080/api/v1/admin/seed/cards

# Get database statistics
curl http://localhost:8080/api/v1/admin/stats
```

### Adding Custom Cards

#### Via API
```bash
# Create a new tag
curl -X POST http://localhost:8080/api/v1/tags \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Steampunk",
    "category": "fantasy",
    "description": "Victorian-era technology and aesthetics",
    "color": "#8B4513",
    "weight": 1.2
  }'

# Create a card with tags
curl -X POST http://localhost:8080/api/v1/cards \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Clockwork Dragon",
    "description": "Mechanical dragon with brass gears",
    "extension": ".jpg",
    "tag_ids": [1, 5, 12]
  }'

# Upload card image
curl -X POST http://localhost:8080/api/v1/cards/85/image \
  -F "image=@/path/to/clockwork-dragon.jpg"
```

#### Card Data Structure
Each card in the asset list contains:

```go
type CardData struct {
    ID          int      `json:"id"`          // Unique card ID
    Title       string   `json:"title"`       // Display name
    Description string   `json:"description"` // Detailed description
    Extension   string   `json:"extension"`   // Image file extension
    Tags        []string `json:"tags"`        // Associated tag slugs
}
```

## Real-Time Chat System

### Chat Functionality

The game includes a comprehensive real-time chat system that allows players to communicate during specific game phases:

**Allowed Phases**:
- **Lobby**: Players can chat freely while waiting for the game to start
- **Voting Phase**: Strategic discussions during card voting

**Restricted Phases**: 
- Chat is disabled during storytelling and card submission to prevent cheating

### Chat Features

**Message Types**:
- **Regular Chat**: Standard text messages between players
- **Emotes**: Special emoji-style messages for quick reactions  
- **System Messages**: Automated notifications for game events (player joins/leaves, game state changes)

**Real-time Delivery**:
- Instant message broadcasting via WebSockets
- Message persistence in PostgreSQL database
- Chat history retrieval for late-joining players

**Moderation**:
- Message length limits (500 characters max)
- Content filtering support (extensible for profanity filters)
- Message visibility controls for moderation

### Chat API Endpoints

#### REST API
```bash
# Send chat message
curl -X POST http://localhost:8080/api/v1/chat/send \
  -H "Content-Type: application/json" \
  -d '{
    "room_code": "GAME123",
    "player_id": "123e4567-e89b-12d3-a456-426614174000",
    "message": "Good luck everyone!",
    "message_type": "chat"
  }'

# Get chat history
curl "http://localhost:8080/api/v1/chat/history?room_code=GAME123&phase=lobby&limit=50"

# Get chat statistics
curl "http://localhost:8080/api/v1/chat/stats?room_code=GAME123"
```

#### WebSocket API
```javascript
// Send chat message via WebSocket
websocket.send(JSON.stringify({
  type: "send_chat",
  payload: {
    room_code: "GAME123",
    message: "Great card choice!",
    message_type: "chat"
  }
}));

// Request chat history
websocket.send(JSON.stringify({
  type: "get_chat_history", 
  payload: {
    room_code: "GAME123",
    phase: "voting",
    limit: 20
  }
}));

// Receive chat messages
websocket.onmessage = (event) => {
  const message = JSON.parse(event.data);
  if (message.type === "chat_message") {
    // Display chat message
    console.log(`${message.payload.player_name}: ${message.payload.message}`);
  }
};
```

### System Messages

Automatic system notifications are sent for:
- **Player Events**: "Alice joined the game", "Bob left the game"
- **Game Events**: "Game started! Let the storytelling begin!"
- **Bot Events**: "Bot Charlie (medium difficulty) joined the game"

### Chat Data Structure

```go
type ChatMessage struct {
    ID          uuid.UUID `json:"id"`
    GameID      uuid.UUID `json:"game_id"`
    PlayerID    uuid.UUID `json:"player_id"`
    Message     string    `json:"message"`
    MessageType string    `json:"message_type"` // chat, system, emote
    Phase       string    `json:"phase"`        // lobby, voting, etc.
    IsVisible   bool      `json:"is_visible"`   // For moderation
    CreatedAt   time.Time `json:"created_at"`
}
```

### Usage Examples

**Player Chat Flow**:
1. Player joins lobby → System message: "Player joined"
2. Players chat in lobby → Real-time message broadcasting
3. Game starts → Chat disabled during storytelling/submission
4. Voting phase begins → Chat re-enabled for discussions
5. Player leaves → System message: "Player left"

**Integration with Game Flow**:
- Chat permissions automatically adjust based on game phase
- System messages provide context for game state changes
- Bot players don't participate in chat (system messages only)
- Chat history persists across browser refreshes

## Authentication System

### Player Tracking & Sessions

The game supports three types of player authentication, providing flexibility for different user preferences:

**1. Registered Users (Username/Password)**
- Full account creation with email and password
- Persistent player statistics and game history
- Cross-device session management
- Personalized profile with display name and avatar

**2. Google OAuth2 SSO**  
- One-click sign-in with Google account
- Automatic account creation on first login
- Secure authentication without password management
- Profile information synced from Google

**3. Guest Access**
- Play immediately without registration
- Temporary session-based identification
- Basic functionality with limited persistence
- Can upgrade to registered account later

### Session Management

**JWT Token-Based Authentication**:
- Secure stateless authentication using JWT tokens
- 24-hour token expiration with refresh capability
- Session tracking in PostgreSQL database
- Support for multiple concurrent sessions

**Flexible Token Delivery**:
- HTTP Authorization header (`Bearer <token>`)
- HTTP-only cookies (recommended for web)
- WebSocket query parameters for real-time connections

### Authentication Flow

#### User Registration
```bash
# Register new account
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "player@example.com",
    "username": "player123", 
    "display_name": "Cool Player",
    "password": "securepassword123"
  }'
```

#### Password Login
```bash
# Login with email/username
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email_or_username": "player@example.com",
    "password": "securepassword123"
  }'
```

#### Google OAuth Login
```bash
# Login with Google OAuth token
curl -X POST http://localhost:8080/api/v1/auth/google \
  -H "Content-Type: application/json" \
  -d '{
    "access_token": "google_oauth_access_token"
  }'
```

#### Guest Session
```bash
# Create guest session
curl -X POST http://localhost:8080/api/v1/auth/guest \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Guest Player"
  }'
```

### WebSocket Authentication

The WebSocket connection supports multiple authentication methods:

```javascript
// Authenticated connection with token in URL
const ws = new WebSocket('ws://localhost:8080/ws?token=jwt_token_here');

// Guest connection with player ID
const ws = new WebSocket('ws://localhost:8080/ws?player_id=guest_uuid');

// Cookie-based authentication (automatic)
const ws = new WebSocket('ws://localhost:8080/ws');
```

### API Security Levels

**Public Endpoints** (no authentication required):
- Card browsing and details
- Tag listing
- Bot statistics
- Health checks

**Session Required** (guest or registered):
- Game creation and joining
- Chat messaging
- Player statistics

**Authentication Required** (registered users only):
- Card creation and image uploads
- Tag management
- Account management

**Admin Only** (authenticated + admin privileges):
- Database seeding
- System administration

### Environment Configuration

Add authentication settings to your `.env` file:

```bash
# Authentication configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
GOOGLE_CLIENT_ID=your-google-oauth-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-oauth-client-secret
```

### Security Features

- **Password Hashing**: bcrypt with configurable cost
- **Token Security**: HMAC-signed JWT with expiration
- **Session Management**: Database-tracked with cleanup
- **CORS Protection**: Configurable origins
- **Rate Limiting**: Ready for implementation
- **XSS Protection**: HTTP-only cookies option

## Troubleshooting

### Common Issues

1. **WebSocket connection fails**
   - Check if backend is running on correct port
   - Verify firewall settings
   - Check browser console for CORS errors

2. **Database connection error**
   - Ensure PostgreSQL is running
   - Verify connection string in .env
   - Check database exists and user has permissions

3. **Redis connection error**
   - Ensure Redis server is running
   - Verify Redis URL in .env
   - Check Redis auth if required

4. **Cards not displaying**
   - Ensure card images are in `assets/cards/`
   - Check file permissions
   - Verify static file serving is enabled

### Performance Optimization

- Use Redis for session storage in production
- Enable PostgreSQL connection pooling
- Add CDN for card images
- Implement WebSocket connection limits
- Add rate limiting for API endpoints

## Future Enhancements

- [ ] Spectator mode
- [ ] Private rooms with passwords
- [ ] Multiple card decks/expansions
- [ ] Tournament mode
- [ ] Mobile app
- [ ] Voice chat integration
- [ ] Replay system
- [ ] Advanced statistics
- [ ] Custom card uploads
- [ ] AI players
