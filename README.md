# DixitMe - Online Dixit Card Game

A full-stack implementation of Dixit with real-time multiplayer, bots, chat, and card/tag management. Go backend + React frontend.

## Prerequisites
- Go 1.23+ (toolchain 1.24.x recommended)
- Node.js 18+ and npm 9+
- Docker & Docker Compose (optional, for local Postgres/Redis)
- Optional: Swagger generator
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```

## Feature Highlights
- Real-time gameplay over WebSocket: lobby, storytelling, submissions, voting, scoring; win at 30 points or empty deck.
- Authentication: email/password, Google SSO, guest sessions; JWT cookies/headers; refresh/logout/status.
- Bots: heuristic card/vote selection using weighted tags; add bots to games; bot stats.
- Cards & Tags: 84 seeded cards with semantic tags; MinIO/S3 image hosting; card/tag CRUD for authenticated users.
- Chat: phase-aware chat (lobby/voting) with history, stats, system messages, and moderation visibility flag.
- Admin/ops: seed tags/cards, DB stats, cleanup old games.
- Docs: Swagger/OpenAPI for REST; typed WebSocket message types documented below.

## Configuration
Copy the example env and adjust as needed:

```bash
cp config.env.example .env
```

Key settings:
- `DATABASE_URL` (e.g., `postgres://dixitme:dixitme_password@localhost/dixitme?sslmode=disable`)
- `REDIS_URL` (e.g., `redis://localhost:6379`)
- `PORT` (default 8080), `GIN_MODE` (debug|release)
- `JWT_SECRET` (set a strong value in prod)
- `ENABLE_SSO` (true|false) and Google OAuth creds if using SSO
- `MINIO_*` (optional; dev can rely on filesystem fallback)

## Quick Start
Backend
```bash
cp config.env.example .env
go mod tidy
go run cmd/server/main.go
```
- Requires PostgreSQL + Redis reachable via `DATABASE_URL` / `REDIS_URL` in `.env`.
- Seeds tags/cards on startup (see `internal/data/seeder`).
- Swagger UI: `http://localhost:8080/swagger/index.html`

Frontend
```bash
cd web
npm install
npm start
```
- CRA dev server proxies API/WS to `http://localhost:8080` (see `web/package.json`).

## Start Postgres & Redis (Dev)
Bring up backing services with Docker (optional, recommended for local dev):

```bash
docker compose -f docker-compose.dev.yml up -d
# Stop:
docker compose -f docker-compose.dev.yml down
``` 

## Build frontend and serve via Go server
`internal/app/server.go` serves `./web/build` at `/` and `/static`.

```bash
cd web && npm run build && cd ..
# Then start the backend (serves ./web/build at /)
go run cmd/server/main.go
```

## Run everything with Docker (optional)
Use the provided `docker-compose.yml` to run the app + Postgres + Redis together.

```bash
docker compose up -d --build
# App: http://localhost:8080
# Swagger: http://localhost:8080/swagger/index.html
```

## Seeder CLI
```bash
go run cmd/seed/main.go          # Seed everything
go run cmd/seed/main.go -tags    # Tags only
go run cmd/seed/main.go -cards   # Cards only (requires tags)
go run cmd/seed/main.go -force   # Force reseed
```

## Testing
```bash
# Backend
go test ./...

# Frontend (jest via CRA)
cd web && npm test
```

## Project Structure
```
cmd/                # Entrypoints (server, seed)
internal/
  app/              # Composition root wiring data + feature services
  auth/             # Auth service + JWT + HTTP middleware/handlers
  data/             # Infra: DB/Redis/MinIO store, repos, seeding, models
  game/             # Game domain (core/bot/domain) + HTTP + WS hub
  config/           # Config loading
  platform/logger/  # Slog-based logging
docs/               # Swagger artifacts + detailed docs
web/                # React frontend
assets/cards/       # Card images
scripts/            # Utility scripts
```

## Game Rules (summary)
- 3–6 players, 6 cards each.
- Flow: storyteller gives clue + picks card → others submit → shuffle/reveal → vote → score.
- Scoring: if all/none guess storyteller → storyteller 0, others 2; otherwise storyteller + correct guessers 3; +1 per vote on your card (non-storyteller).
- Ends at 30 points or empty deck.

## Working Flow
- Startup: `NewServer` wires DB/Redis, creates a `Manager`, and connects it to the WebSocket hub.
- State model: FSM phases live on `GameState.Phase`: `LOBBY`, `STORYTELLER_SUBMIT`, `OTHERS_SUBMIT`, `VOTING`, `REVEAL_SCORE`, `ROUND_END`, `GAME_OVER`.
- Persistence: live game state is persisted as JSONB in `games.state_snapshot` after each state change, with `games.version` used for optimistic concurrency.
- History: normal relational tables (rounds, votes, game history, chat) remain for history/leaderboards and analytics.
- Connect: client opens `GET /ws`; server replies with `connection_established` and stores the socket.
- Create/Join: client sends `create_game` or `join_game`; server loads the latest snapshot from Postgres, applies the command, and saves with a version check.
- Gameplay: clients send `start_game`, `submit_clue`, `submit_card`, `submit_vote`, `leave_game`; server always loads the latest snapshot, applies, and saves with a version check, then broadcasts WS events.
- Chat: client sends `send_chat` over WS; server persists and broadcasts `chat_message`.

## Endpoint Design Notes
- WebSocket is used for real-time game actions and live chat: create/join/start/leave game, submit clue/card/vote, send chat.
- HTTP is used for authentication, CRUD/admin operations, and read-only queries like game lists, player stats, and chat history.
- Chat history is HTTP-only; WebSocket only handles live chat messages.

## REST API (key endpoints)
- Base path: `/api/v1` (Swagger: `/swagger/index.html`)
- Health: `GET /health`
- Auth: `POST /auth/register|login|google|guest|refresh|logout`, `GET /auth/me|validate|status`
- Players: `POST /players`, `GET /players/:id`, `GET /player/:player_id/stats|history`
- Games: `GET /games`, `GET /games/:room_code`, `POST /games/add-bot`
- Cards/Tags: `GET /cards|/cards/:card_id|/cards/legacy`, `POST /cards`, `POST /cards/:card_id/image`, `GET /tags`, `POST /tags`
- Bots: `GET /bots/stats`
- Admin: `POST /admin/seed|seed/tags|seed/cards|cleanup`, `GET /admin/stats`
- Chat: `GET /chat/history`, `GET /chat/stats`

## WebSocket
- Path: `GET /ws`
### Messages
Client → Server
- `create_game`, `join_game`, `start_game`, `submit_clue`, `submit_card`, `submit_vote`, `leave_game`, `send_chat`

Server → Client (message types)
- `connection_established`, `game_state`, `player_joined/left`, `game_started`, `round_started`, `clue_submitted`, `card_submitted`, `voting_started`, `vote_submitted`, `round_completed`, `game_completed`, `chat_message`, `error`

## Data & Assets
- Models live in `internal/data/models`.
- Infra wiring (DB/Redis/MinIO) in `internal/data/store.go`.
- Seeder: `internal/data/seeder` seeds tags/cards; assets under `assets/cards/`.
- MinIO/S3 config via `.env` (`MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, etc.).
- Note: MinIO/S3 is optional in development. If MinIO isn’t available, the seeder falls back to local paths, and the server serves card images from `./assets/cards` at `/cards`.

## Troubleshooting
- WebSocket fails: confirm backend is running on the expected port, check browser console for CORS issues, and verify you’re hitting `/ws`.
- DB/Redis errors: ensure services are running and `DATABASE_URL`/`REDIS_URL` are set correctly.
- Cards not showing: confirm images in `assets/cards/` and static serving; rerun seeding if needed.
- Auth issues: check `JWT_SECRET` and cookies/headers; `/auth/status` shows enabled methods.
