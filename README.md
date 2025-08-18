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

## Tech Stack

### Backend (Go)
- **Framework**: Gin web framework
- **Database**: PostgreSQL with GORM
- **Cache**: Redis for state synchronization
- **WebSockets**: Gorilla WebSocket for real-time communication
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

5. **Run the backend:**
   ```bash
   go run cmd/server/main.go
   ```

   The server will start on `http://localhost:8080`

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

### REST API
- `GET /health` - Health check
- `POST /api/v1/players` - Create player
- `GET /api/v1/players/:id` - Get player info
- `GET /api/v1/games` - List games
- `GET /api/v1/games/:room_code` - Get game info
- `GET /api/v1/cards` - Get card list

### WebSocket
- `GET /ws` - WebSocket connection for real-time game updates

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
