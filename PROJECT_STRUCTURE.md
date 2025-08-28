# 📁 Detailed Project Structure

This document provides a comprehensive overview of the DixitMe project structure, following the [Go standard project layout](https://github.com/golang-standards/project-layout) with a clean three-layer architecture.

## 🏗️ Complete Directory Structure

```
DixitMe/
├── cmd/                     # 🚀 Application entry points
│   ├── server/main.go       #     → Main server application 
│   └── seed/main.go         #     → Database seeding CLI tool

├── internal/                # 🔒 Private application code (cannot be imported externally)
│   ├── app/                 # 🎯 Application initialization & dependency injection
│   │   └── app.go           #     → App struct, NewApp(), Run(), graceful shutdown
│   ├── transport/           # 🌐 Transport layer (HTTP handlers, WebSocket, routing)
│   │   ├── handlers/        #     → HTTP API endpoints (domain-separated)
│   │   │   ├── common.go    #         ┣ Health checks & CORS middleware
│   │   │   ├── player.go    #         ┣ Player management & statistics
│   │   │   ├── game.go      #         ┣ Game management & bot operations
│   │   │   ├── card.go      #         ┣ Card management & image uploads
│   │   │   ├── tag.go       #         ┣ Tag management for categorization
│   │   │   ├── chat.go      #         ┣ Chat messages & communication
│   │   │   ├── admin.go     #         ┣ Administrative operations
│   │   │   ├── dependencies.go #      ┣ Dependency injection container
│   │   │   └── types.go     #         ┗ Request/response type definitions
│   │   ├── router/          #     → HTTP routing & middleware setup
│   │   │   └── router.go    #         ┗ Organized route definitions
│   │   └── websocket/       #     → WebSocket communication (real-time)
│   │       ├── auth.go      #         ┣ Authentication & token extraction
│   │       ├── connection.go#         ┣ WebSocket connection management
│   │       ├── handlers.go  #         ┣ Message routing & game actions
│   │       └── types.go     #         ┗ WebSocket message type definitions
│   ├── services/            # 💼 Business logic layer (core application logic)
│   │   ├── auth/            #     → Authentication & JWT services
│   │   │   ├── handlers.go  #         ┣ Auth HTTP endpoints
│   │   │   ├── jwt.go       #         ┣ JWT token management
│   │   │   ├── middleware.go#         ┣ Auth middleware
│   │   │   └── service.go   #         ┗ Authentication business logic
│   │   ├── game/            #     → 🎮 Core game logic & management
│   │   │   ├── manager.go   #         ┣ Main game manager with dependency injection
│   │   │   ├── game_actions.go #      ┣ Game lifecycle (create, join, start)
│   │   │   ├── round_actions.go #     ┣ Round management & gameplay
│   │   │   ├── bot_actions.go #       ┣ Bot AI integration
│   │   │   ├── chat_actions.go #      ┣ Real-time chat system
│   │   │   ├── persistence.go #       ┣ Database operations with context & transactions
│   │   │   ├── cleanup.go   #         ┣ Inactive game cleanup service
│   │   │   ├── broadcasting.go #      ┣ WebSocket message broadcasting
│   │   │   ├── game_types.go #        ┣ Core game domain entities (GameState, Player, Round)
│   │   │   ├── websocket_types.go #   ┣ WebSocket communication types & payloads
│   │   │   └── chat_types.go #        ┗ Chat-specific types & message structures
│   │   └── bot/             #     → AI bot players with heuristic algorithms
│   │       └── ai.go        #         ┗ Difficulty-based clue generation
│   ├── models/              # 📊 Database layer (data persistence)
│   │   ├── user.go          #     → User accounts & authentication
│   │   ├── player.go        #     → Player entities & game participation
│   │   ├── game.go          #     → Game sessions & game history
│   │   ├── round.go         #     → Game rounds, submissions & votes
│   │   ├── card.go          #     → Card entities & tag system
│   │   ├── chat.go          #     → Chat messages & communication
│   │   └── models.go        #     → Package documentation
│   ├── database/            #     → Database connection & migrations
│   │   └── database.go      #         ┗ PostgreSQL setup with GORM
│   ├── config/              #     → Configuration management
│   │   └── config.go        #         ┗ Environment variable loading
│   ├── logger/              #     → Structured logging
│   │   └── logger.go        #         ┗ slog configuration (JSON/text)
│   ├── redis/               #     → Redis cache integration
│   │   └── redis.go         #         ┗ Redis client setup
│   ├── storage/             #     → File storage (MinIO object storage)
│   │   └── minio.go         #         ┗ MinIO client & file operations
│   ├── utils/               #     → Common utility functions & input validation
│   │   └── strings.go       #         ┗ String manipulation, generation & validation
│   └── seeder/              #     → Database seeding logic
│       └── cards.go         #         ┗ Seed cards, tags, and default data
├── configs/                 # ⚙️ Static configuration files
│   ├── server.yaml          #     → Server & application settings
│   ├── config.env.example   #     → Environment template
│   └── config.env.development #   → Development environment
├── api/                     # 📋 API documentation & specifications
│   └── v1/                  #     → API version 1
│       ├── swagger.json     #         ┣ OpenAPI specification (JSON)
│       └── swagger.yaml     #         ┗ OpenAPI specification (YAML)
├── deployments/             # 🚀 Deployment configurations
│   └── docker/              #     → Docker deployment files
│       ├── Dockerfile       #         ┣ Multi-stage container build
│       ├── docker-compose.yml #       ┣ Production compose file
│       └── docker-compose.dev.yml #   ┗ Development compose file
├── web/                     # ⚛️ React frontend (TypeScript + CSS Modules)
│   ├── public/              #     → Static public assets
│   │   └── index.html       #         ┗ HTML template with root div
│   ├── src/                 #     → Source code
│   │   ├── components/      #         → React UI components
│   │   │   ├── Auth.tsx     #             ┣ Authentication modal (login/register/guest)
│   │   │   ├── Auth.module.css #          ┣ Auth component styles
│   │   │   ├── Card.tsx     #             ┣ Individual card display with animations
│   │   │   ├── Card.module.css #          ┣ Card component styles
│   │   │   ├── Chat.tsx     #             ┣ Real-time chat with emoji picker
│   │   │   ├── Chat.module.css #          ┣ Chat component styles
│   │   │   ├── GameBoard.tsx #            ┣ Main game interface during play
│   │   │   ├── GameLanding.tsx #          ┣ Primary landing page (join/create)
│   │   │   ├── GameLanding.module.css #   ┣ Landing page styles
│   │   │   ├── GamePhaseIndicator.tsx #   ┣ Visual game phase tracker
│   │   │   ├── GamePhaseIndicator.module.css # ┣ Phase indicator styles
│   │   │   ├── Lobby.tsx    #             ┣ Game lobby for waiting players
│   │   │   ├── Lobby.module.css #         ┣ Lobby component styles
│   │   │   ├── PlayerHand.tsx #           ┣ Player's card hand interface
│   │   │   ├── UserInfo.tsx #             ┣ User profile & guest upgrade
│   │   │   ├── UserInfo.module.css #      ┣ User info styles
│   │   │   └── VotingPhase.tsx #          ┗ Voting interface for cards
│   │   ├── store/           #         → State management (Zustand)
│   │   │   ├── authStore.ts #             ┣ Authentication state & actions
│   │   │   └── gameStore.ts #             ┗ Game state, WebSocket & actions
│   │   ├── types/           #         → TypeScript definitions
│   │   │   └── game.ts      #             ┗ Game interfaces & message types
│   │   ├── App.tsx          #         → Main app with routing logic
│   │   ├── index.tsx        #         → React entry point
│   │   ├── index.css        #         → Global styles
│   │   └── react-app-env.d.ts #       → React TypeScript environment types
│   ├── package.json         #     → NPM dependencies and scripts
│   ├── package-lock.json    #     → Dependency lock file
│   └── tsconfig.json        #     → TypeScript configuration
├── assets/                  # 🎨 Static game assets
│   ├── cards/               #     → Card image files (84 Dixit cards)
│   └── tags.json            #     → Card categorization tags for bot AI
├── scripts/                 # 🔧 Utility scripts
│   ├── generate-swagger.sh  #     → API documentation generation
│   └── dev-setup.sh         #     → Development environment setup
├── go.mod / go.sum          # 📦 Go dependency management
├── README.md                # 📖 Project documentation
├── PROJECT_STRUCTURE.md     # 📋 This detailed structure document
└── DEVELOPER_GUIDE.md       # 🎓 Comprehensive developer learning guide
```

## 🏗️ Architecture Overview

### Three-Layer Architecture

The project follows a clean three-layer architecture pattern:

#### 🌐 Transport Layer (`/internal/transport/`)
- **HTTP Handlers**: REST API endpoints organized by domain
- **WebSocket Communication**: Real-time game communication
- **Routing & Middleware**: Request routing and middleware setup
- **Dependency Injection**: Handler dependencies management

#### 💼 Business Layer (`/internal/services/`)
- **Game Logic**: Core game mechanics and state management
- **Authentication**: User auth, JWT, Google SSO, guest sessions
- **Bot AI**: Heuristic algorithms for AI players
- **Interface-Driven Design**: Clean contracts between layers

#### 📊 Data Layer (`/internal/models/`, `/internal/database/`)
- **Database Models**: GORM entities for PostgreSQL
- **Migrations**: Automatic schema management
- **Persistence**: Context-aware database operations with transactions

### 🎯 Key Architectural Features

#### **Dependency Injection**
- Clean separation of concerns with injected dependencies
- Interface-based design for testability
- Centralized dependency management in `/internal/app/`

#### **Type Organization**
- **game_types.go**: Core game domain entities (GameState, Player, Round)
- **websocket_types.go**: WebSocket communication types and payloads
- **chat_types.go**: Chat-specific types and message structures

#### **Context & Transactions**
- All persistence methods accept `context.Context` for cancellation/timeouts
- Multi-step database operations wrapped in transactions
- Structured logging with success/failure tracking

#### **Redis Integration**
- Comprehensive caching strategy with graceful degradation
- Full game state serialization for performance
- Documented partial caching approach with structured logging

## 📁 Directory Guidelines

### Adding New Features

#### **Reusable Utilities**
- Place in `/pkg/` (can be imported by other projects)
- Examples: validation functions, string utilities

#### **Business Logic**
- Place in `/internal/services/`
- Follow interface-driven design patterns
- Implement dependency injection

#### **HTTP Endpoints**
- Place in `/internal/transport/handlers/`
- Organize by domain (game, auth, chat, etc.)
- Add Swagger documentation

#### **Database Models**
- Place in `/internal/models/`
- Use GORM conventions
- Include appropriate indexes and relationships

#### **Configuration**
- Static configs in `/configs/`
- Code-based config in `/internal/config/`

### File Organization Principles

1. **Domain Separation**: Group related functionality together
2. **Layer Isolation**: Keep transport, business, and data layers separate
3. **Interface Contracts**: Define clear interfaces between components
4. **Go Standards**: Follow official Go project layout conventions
5. **Testability**: Structure code for easy unit testing

## 🧪 Testing Strategy

### Directory Structure for Tests
```
internal/
├── services/
│   ├── game/
│   │   ├── manager_test.go
│   │   ├── game_actions_test.go
│   │   └── testutils/
│   └── auth/
│       └── service_test.go
└── models/
    └── user_test.go
```

### Testing Guidelines
- Place test files alongside source files
- Use `testutils/` for shared test utilities
- Mock external dependencies using interfaces
- Test business logic thoroughly in the service layer

## 📦 Deployment Structure

### Docker Configuration
- **Multi-stage builds** for optimized production images
- **Development compose** with hot reloading
- **Production compose** with health checks and restart policies

### Environment Management
- **Development**: `configs/config.env.development`
- **Production**: Environment variables via orchestration
- **Examples**: `configs/config.env.example` for reference

This structure ensures maintainability, scalability, and follows Go best practices while providing clear separation of concerns and easy navigation for developers.
