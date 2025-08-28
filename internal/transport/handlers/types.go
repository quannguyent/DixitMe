package handlers

import (
	"dixitme/internal/models"
	"time"
)

// Player related types
type CreatePlayerRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreatePlayerResponse struct {
	Player *models.Player `json:"player"`
}

// Game related types
type GetGameRequest struct {
	RoomCode string `uri:"room_code" binding:"required"`
}

type GetGameResponse struct {
	Game   *models.Game `json:"game"`
	IsLive bool         `json:"is_live"`
}

type GetGamesResponse struct {
	Games []models.Game `json:"games"`
}

type AddBotRequest struct {
	RoomCode string `json:"room_code" binding:"required"`
	BotLevel string `json:"bot_level"` // easy, medium, hard
}

type RemovePlayerRequest struct {
	RoomCode string `json:"room_code" binding:"required"`
	PlayerID string `json:"player_id" binding:"required"`
}

type LeaveGameRequest struct {
	RoomCode string `json:"room_code" binding:"required"`
}

type DeleteGameRequest struct {
	RoomCode string `json:"room_code"`
}

// Card related types
type UploadCardImageResponse struct {
	Message  string `json:"message"`
	ImageURL string `json:"image_url"`
}

type CreateCardRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Extension   string `json:"extension"`
	TagIDs      []int  `json:"tag_ids"`
}

type CreateCardResponse struct {
	Card *models.Card `json:"card"`
}

type TagResponse struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	Category string  `json:"category"`
	Color    string  `json:"color"`
	Weight   float64 `json:"weight"`
}

type CardWithTagsResponse struct {
	ID          int           `json:"id"`
	ImageURL    string        `json:"image_url"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Extension   string        `json:"extension"`
	IsActive    bool          `json:"is_active"`
	Tags        []TagResponse `json:"tags"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type PaginationResponse struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
	Pages int64 `json:"pages"`
}

type CardsListResponse struct {
	Cards      []CardWithTagsResponse `json:"cards"`
	Pagination PaginationResponse     `json:"pagination"`
}

type ListCardsResponse struct {
	Cards []models.Card `json:"cards"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}

// Tag related types
type CreateTagRequest struct {
	Name        string  `json:"name" binding:"required"`
	Slug        string  `json:"slug"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
	Weight      float64 `json:"weight"`
	Category    string  `json:"category"`
}

type CreateTagResponse struct {
	Tag *models.Tag `json:"tag"`
}

type ListTagsResponse struct {
	Tags []models.Tag `json:"tags"`
}

type TagsListResponse struct {
	Tags       []models.Tag       `json:"tags"`
	Pagination PaginationResponse `json:"pagination"`
}

// Chat related types
type SendChatMessageRequest struct {
	GameID      string `json:"game_id" binding:"required"`
	Message     string `json:"message" binding:"required"`
	MessageType string `json:"message_type"` // chat, system, emote
}

type GetChatHistoryRequest struct {
	GameID string `form:"game_id" binding:"required"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

type ChatHistoryResponse struct {
	Messages []models.ChatMessage `json:"messages"`
	Total    int64                `json:"total"`
}

type ChatStatsResponse struct {
	TotalMessages    int64 `json:"total_messages"`
	ActiveGames      int64 `json:"active_games"`
	MessagesToday    int64 `json:"messages_today"`
	MessagesThisWeek int64 `json:"messages_this_week"`
}

// Bot related types
type BotStatsResponse struct {
	TotalBots      int64                  `json:"total_bots"`
	ActiveBots     int64                  `json:"active_bots"`
	BotsByLevel    map[string]int64       `json:"bots_by_level"`
	BotPerformance map[string]interface{} `json:"bot_performance"`
}

// Admin related types
type DatabaseStatsResponse struct {
	Stats interface{} `json:"stats"`
}

// Player stats types
type PlayerStatsResponse struct {
	PlayerID           string  `json:"player_id"`
	TotalGames         int64   `json:"total_games"`
	GamesWon           int64   `json:"games_won"`
	WinRate            float64 `json:"win_rate"`
	AverageScore       float64 `json:"average_score"`
	TotalScore         int64   `json:"total_score"`
	FavoriteRole       string  `json:"favorite_role"`
	GamesAsStoryteller int64   `json:"games_as_storyteller"`
}

type GameHistoryResponse struct {
	Games []models.GameHistory `json:"games"`
	Total int64                `json:"total"`
	Page  int                  `json:"page"`
	Limit int                  `json:"limit"`
}
