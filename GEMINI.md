# DixitMe Context Guide

## Project Overview
**DixitMe** is a full-stack implementation of the Dixit card game, featuring real-time multiplayer gameplay, bots, chat, and card management. The project is structured as a monorepo containing a Go backend and a React frontend.

## Technology Stack

### Backend (Go)
*   **Language:** Go 1.23+
*   **Framework:** Gin (`github.com/gin-gonic/gin`)
*   **Database:** PostgreSQL (via GORM)
*   **Caching/PubSub:** Redis (via `go-redis`)
*   **Storage:** MinIO/S3 (optional, for card images)
*   **WebSockets:** Gorilla WebSocket (`github.com/gorilla/websocket`) for real-time game state and chat.
*   **Authentication:** JWT-based auth with support for Email/Password, Google SSO, and Guest sessions.
*   **Documentation:** Swagger/OpenAPI (`swaggo/swag`).

### Frontend (React)
*   **Framework:** React (Create React App)
*   **Language:** TypeScript
*   **State Management:** Custom stores (likely Zustand or Context API based on `useGameStore` pattern).
*   **Styling:** CSS Modules / Standard CSS.

### Infrastructure
*   **Docker:** `docker-compose.yml` for running the full stack, `docker-compose.dev.yml` for backing services (Postgres, Redis) only.

## Architecture
The backend follows a **Clean/Hexagonal Architecture** pattern within the `internal/` directory:

*   **`cmd/`**: Application entry points (`server`, `seed`).
*   **`internal/app/`**: Composition root, wiring services together.
*   **`internal/game/`**: Core game domain.
    *   **`core/`**: Pure game logic, state machine, and rules (e.g., `room.go`, `match.go`).
    *   **`domain/`**: Type definitions and entities.
    *   **`ws/`**: WebSocket adapter/hub.
    *   **`bot/`**: AI bot logic.
*   **`internal/data/`**: Data access layer (Repositories, Models, Seeding).
*   **`internal/auth/`**: Authentication service and handlers.
*   **`web/`**: The React frontend application.

## Key Development Commands

### Setup
1.  **Environment:** Copy the example config:
    ```bash
    cp config.env.example .env
    ```
2.  **Dependencies:**
    ```bash
    go mod tidy
    cd web && npm install
    ```

### Running the Application
*   **Backing Services (DB/Redis):**
    ```bash
    docker compose -f docker-compose.dev.yml up -d
    ```
*   **Backend:**
    ```bash
    go run cmd/server/main.go
    ```
    *server runs on port 8080 by default.*
*   **Frontend:**
    ```bash
    cd web && npm start
    ```
    *frontend runs on port 3000 and proxies API requests to 8080.*

### Seeding Data
The application requires initial data (cards/tags) to function correctly.
```bash
go run cmd/seed/main.go          # Seed everything
go run cmd/seed/main.go -force   # Force reseed
```

### Testing
*   **Backend:** `go test ./...`
*   **Frontend:** `cd web && npm test`

## Important Considerations
*   **Game State:** The game state is persisted as JSONB snapshots in Postgres for reliability and recovery. Redis is used for caching and potentially for pub/sub (though specific usage should be verified).
*   **WebSockets:** The game relies heavily on WebSockets. The main loop handles events like `submit_clue`, `submit_card`, `submit_vote` and broadcasts state updates.
*   **Assets:** Card images are stored in `assets/cards/` and served statically or via MinIO.
