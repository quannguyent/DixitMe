# DixitMe - Online Dixit Card Game

A full-stack implementation of the popular Dixit card game with real-time multiplayer support.

## 🎮 What is DixitMe?

DixitMe brings the beloved board game Dixit to the web with:
- **Real-time multiplayer** for 3-6 players
- **Complete Dixit gameplay** with storytelling, voting, and scoring
- **AI bot players** with multiple difficulty levels
- **Guest & registered play** options
- **Modern web interface** built with React and Go

## 🚀 Quick Start

### Prerequisites
- Go 1.21+ 
- Node.js 18+
- Docker & Docker Compose

### Setup & Run
   ```bash
# 1. Clone and setup environment
   git clone <repository-url>
   cd DixitMe
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

**Access the game**: http://localhost:3000

## 🎯 How to Play

1. **Create or join** a game with a 4-character room code
2. **Wait for 3-6 players** to join
3. **Play rounds**:
   - Storyteller gives a clue and picks a card
   - Other players submit cards that fit the clue
   - Everyone votes for the storyteller's card
   - Points awarded based on correct guesses
4. **First to 30 points wins!**

## 🛠️ Tech Stack

**Backend**: Go, PostgreSQL, Redis, WebSocket  
**Frontend**: React, TypeScript, CSS Modules  
**Deployment**: Docker, MinIO  

## 📁 Project Structure

```
DixitMe/
├── cmd/                     # Application entry points
├── internal/                # Private application code
│   ├── services/game/       # Core game mechanics
│   ├── transport/           # HTTP & WebSocket handlers
│   ├── models/              # Database models
│   ├── utils/               # Common utility functions & input validation
│   └── ...                  # Other backend packages
├── web/                     # React frontend
│   ├── src/components/      # UI components
│   ├── src/store/           # State management
│   └── ...                  # Other frontend files
├── assets/                  # Game cards & data
├── configs/                 # Configuration files
└── deployments/             # Docker configs
```

## 🔧 Development

**For detailed development information**, see:
- **[ONBOARD_GUIDE.md](./ONBOARD_GUIDE.md)** - Complete project walkthrough, setup, and architecture
- **[PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md)** - Detailed file structure [[memory:6899868]]

**For beginners**, start with:
- **[Go Beginner's Guide](./guides/golang-beginners.md)** - Learn Go programming fundamentals
- **[React Beginner's Guide](./guides/react-beginners.md)** - Learn React & TypeScript basics

### Quick Commands
```bash
# Backend
go run cmd/server/main.go          # Start server
go test ./...                      # Run tests
go run cmd/seed/main.go            # Seed database

# Frontend  
cd web && npm start                # Development server
cd web && npm run build            # Production build

# Database
docker exec -it dixitme-postgres psql -U postgres -d dixitme
```

### API Documentation
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Generate docs**: `./scripts/generate-swagger.sh`

## 🚢 Deployment

### Docker
```bash
# Development
docker-compose -f deployments/docker/docker-compose.dev.yml up

# Production
docker-compose -f deployments/docker/docker-compose.yml up
```

### Environment Variables
Key variables in `.env`:
- `DATABASE_URL` - PostgreSQL connection
- `REDIS_URL` - Redis connection
- `JWT_SECRET` - Authentication secret
- `PORT` - Server port (default: 8080)

## 🎯 Game Features

- ✅ **Complete Dixit rules** with authentic scoring
- ✅ **Real-time gameplay** via WebSocket
- ✅ **Bot AI players** (Easy/Medium/Hard)
- ✅ **Multiple auth types** (Guest/Password/Google SSO)
- ✅ **Chat system** with phase restrictions
- ✅ **Game history** and player statistics
- ✅ **Mobile-responsive** design

## 📝 License

This project is for educational purposes. Dixit is a trademark of Libellud.