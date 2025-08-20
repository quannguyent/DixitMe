// Package models provides all database entity models for the DixitMe application.
// Models are organized into separate files by domain for better maintainability:
//
//   - user.go: User authentication and session management
//   - player.go: Player entities and game participation
//   - game.go: Game sessions and game history
//   - round.go: Game rounds, submissions, and votes
//   - card.go: Card entities and tag system
//   - chat.go: Chat messages and communication
//
// All models use GORM for ORM mapping and UUID for primary keys where applicable.
package models
