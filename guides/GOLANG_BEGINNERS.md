# Go (Golang) Beginner's Guide for DixitMe

Welcome to Go! This guide will help you understand Go programming language fundamentals and how they're applied in the DixitMe project.

## 🚀 What is Go?

Go is a programming language developed by Google that's designed for:
- **Simplicity**: Easy to learn and read
- **Performance**: Fast compilation and execution
- **Concurrency**: Built-in support for concurrent programming
- **Reliability**: Strong type system and memory safety

## 📚 Go Basics

### 1. Hello World & Project Structure

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

**Key Concepts:**
- `package main` - Entry point package
- `import` - Include other packages
- `func main()` - Program starts here
- Go files end with `.go`

### 2. Variables & Types

```go
// Variable declarations
var name string = "DixitMe"
var playerCount int = 6
var isActive bool = true

// Short declaration (type inferred)
roomCode := "ABCD"
score := 30

// Constants
const MaxPlayers = 6
const GameName = "Dixit"
```

**Common Types:**
- `string` - Text
- `int` - Integers
- `bool` - true/false
- `float64` - Decimal numbers

### 3. Functions

```go
// Basic function
func createRoomCode() string {
    return "ABCD"
}

// Function with parameters
func addPlayer(name string, id int) bool {
    // Add player logic
    return true
}

// Multiple return values (common in Go)
func validatePlayer(name string) (bool, error) {
    if name == "" {
        return false, errors.New("name cannot be empty")
    }
    return true, nil
}
```

### 4. Structs (Like Objects)

```go
// Define a struct
type Player struct {
    ID    string
    Name  string
    Score int
    Hand  []int  // slice of card IDs
}

// Create and use struct
func main() {
    player := Player{
        ID:    "123",
        Name:  "Alice",
        Score: 0,
        Hand:  []int{1, 2, 3, 4, 5, 6},
    }
    
    fmt.Println(player.Name) // Access field
    player.Score = 10        // Modify field
}
```

### 5. Slices & Maps

```go
// Slices (dynamic arrays)
cards := []int{1, 2, 3, 4, 5}
cards = append(cards, 6)        // Add element
fmt.Println(len(cards))         // Length: 6

// Maps (key-value pairs)
players := make(map[string]*Player)
players["alice"] = &Player{Name: "Alice"}
players["bob"] = &Player{Name: "Bob"}

// Check if key exists
if player, exists := players["alice"]; exists {
    fmt.Println("Found:", player.Name)
}
```

### 6. Error Handling

```go
func joinGame(roomCode string) error {
    if roomCode == "" {
        return errors.New("room code required")
    }
    
    // Game logic here...
    
    return nil // No error
}

// Using the function
func main() {
    err := joinGame("ABCD")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("Successfully joined!")
}
```

### 7. Interfaces

```go
// Define interface
type GameService interface {
    CreateGame(roomCode string) error
    JoinGame(roomCode string, playerName string) error
}

// Implement interface
type Manager struct {
    games map[string]*Game
}

func (m *Manager) CreateGame(roomCode string) error {
    // Implementation
    return nil
}

func (m *Manager) JoinGame(roomCode string, playerName string) error {
    // Implementation
    return nil
}
```

### 8. Goroutines (Concurrency)

```go
// Run function concurrently
go processPlayerAction(playerID)

// Channel communication
messages := make(chan string)

// Send to channel (in goroutine)
go func() {
    messages <- "Hello from goroutine!"
}()

// Receive from channel
msg := <-messages
fmt.Println(msg)
```

## 🎮 Go in DixitMe Project

### Project Structure Patterns

**1. Package Organization:**
```
internal/
├── models/          # Data structures
├── services/        # Business logic
├── transport/       # HTTP/WebSocket handling
└── database/        # Data persistence
```

**2. Dependency Injection:**
```go
type Manager struct {
    db     *gorm.DB        // Database connection
    logger *slog.Logger    // Logging
    games  map[string]*GameState  // In-memory storage
}

func NewManager(db *gorm.DB, logger *slog.Logger) *Manager {
    return &Manager{
        db:     db,
        logger: logger,
        games:  make(map[string]*GameState),
    }
}
```

### Common Patterns in DixitMe

**1. Error Handling Pattern:**
```go
func (m *Manager) CreateGame(roomCode string, creatorID uuid.UUID) (*GameState, error) {
    // Validate input
    if roomCode == "" {
        return nil, errors.New("room code required")
    }
    
    // Business logic
    game := &GameState{
        ID:       uuid.New(),
        RoomCode: roomCode,
        Players:  make(map[uuid.UUID]*Player),
    }
    
    // Save to database
    if err := m.db.Create(game).Error; err != nil {
        return nil, fmt.Errorf("failed to save game: %w", err)
    }
    
    return game, nil
}
```

**2. HTTP Handler Pattern:**
```go
func (h *GameHandlers) CreateGame(c *gin.Context) {
    var req CreateGameRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    game, err := h.gameService.CreateGame(req.RoomCode, req.PlayerID)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"game": game})
}
```

**3. WebSocket Message Handling:**
```go
func handleMessage(conn *websocket.Conn, message []byte) error {
    var msg ConnectionMessage
    if err := json.Unmarshal(message, &msg); err != nil {
        return err
    }
    
    switch msg.Type {
    case "create_game":
        return handleCreateGame(conn, msg.Payload)
    case "join_game":
        return handleJoinGame(conn, msg.Payload)
    default:
        return errors.New("unknown message type")
    }
}
```

### Key Go Libraries Used

**1. Gin (Web Framework):**
```go
import "github.com/gin-gonic/gin"

router := gin.Default()
router.POST("/api/games", createGameHandler)
router.GET("/api/games/:id", getGameHandler)
```

**2. GORM (Database ORM):**
```go
import "gorm.io/gorm"

type Game struct {
    ID       uuid.UUID `gorm:"primaryKey"`
    RoomCode string    `gorm:"uniqueIndex"`
    Status   string
}

// Query database
var game Game
db.Where("room_code = ?", "ABCD").First(&game)
```

**3. UUID (Unique Identifiers):**
```go
import "github.com/google/uuid"

playerID := uuid.New()
```

**4. WebSocket (Real-time Communication):**
```go
import "github.com/gorilla/websocket"

upgrader := websocket.Upgrader{}
conn, err := upgrader.Upgrade(w, r, nil)
if err != nil {
    return err
}

// Send message
conn.WriteJSON(message)
```

## 🛠️ Development Tips

### 1. Code Organization
- **One package per directory**
- **Short, descriptive names**
- **Interfaces for testing**
- **Keep functions small**

### 2. Go Conventions
- **Public functions start with uppercase**: `CreateGame()`
- **Private functions start with lowercase**: `validateInput()`
- **Use `gofmt`** to format code automatically
- **Handle all errors explicitly**

### 3. Testing
```go
func TestCreateGame(t *testing.T) {
    manager := NewManager(db, logger)
    
    game, err := manager.CreateGame("TEST", uuid.New())
    
    assert.NoError(t, err)
    assert.Equal(t, "TEST", game.RoomCode)
}
```

### 4. Common Mistakes to Avoid
- **Ignoring errors**: Always check `if err != nil`
- **Not using pointers**: Use `*Player` for structs
- **Forgetting to close resources**: Use `defer` for cleanup
- **Not handling nil pointers**: Check before accessing

## 📖 Learning Resources

**Official Documentation:**
- [Go Tour](https://tour.golang.org/) - Interactive tutorial
- [Go by Example](https://gobyexample.com/) - Practical examples
- [Effective Go](https://golang.org/doc/effective_go.html) - Best practices

**Books:**
- "The Go Programming Language" by Donovan & Kernighan
- "Go in Action" by William Kennedy

**Practice:**
1. Complete the Go Tour
2. Build small CLI tools
3. Read DixitMe source code
4. Write tests for existing functions

## 🎯 Next Steps

1. **Set up Go environment** and run `go version`
2. **Complete Go Tour** (tour.golang.org)
3. **Read DixitMe code** starting with `internal/models/`
4. **Try modifying** simple functions
5. **Write tests** for new features

Remember: Go is designed to be simple. If you're writing complex code, there's probably a simpler way!
