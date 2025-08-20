// Package websocket provides WebSocket communication for real-time game interactions.
// Components are organized into separate files for better maintainability:
//
//   - connection.go: WebSocket connection management and upgrade logic
//   - handlers.go: Message routing and game action handlers
//   - auth.go: Authentication and token extraction
//   - types.go: Message type definitions and payload structures
//
// This package handles real-time communication between clients and the game server,
// supporting both authenticated users and guest players.
package websocket
