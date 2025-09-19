package game

import (
	"fmt"
	"time"
)

// GameEngineFactory creates game engines
type GameEngineFactory interface {
	CreateEngine(gameType GameType, settings TableSettings) (GameEngine, error)
}

// TableJoinMode defines how a player joins the table
type TableJoinMode string

const (
	JoinModePlayer   TableJoinMode = "player"
	JoinModeObserver TableJoinMode = "observer"
)

// TableError represents table operation errors
type TableError struct {
	Code    string
	Message string
}

func (e *TableError) Error() string {
	return fmt.Sprintf("table error [%s]: %s", e.Code, e.Message)
}

// Common table error codes
var (
	ErrTableNotFound        = &TableError{"TABLE_NOT_FOUND", "Table not found"}
	ErrTableFull            = &TableError{"TABLE_FULL", "Table is full"}
	ErrPlayerNotAtTable     = &TableError{"PLAYER_NOT_AT_TABLE", "Player is not at this table"}
	ErrPlayerAlreadyAtTable = &TableError{"PLAYER_ALREADY_AT_TABLE", "Player is already at this table"}
	ErrTableNotJoinable     = &TableError{"TABLE_NOT_JOINABLE", "Table is not in a joinable state"}
	ErrObserversNotAllowed  = &TableError{"OBSERVERS_NOT_ALLOWED", "Observers are not allowed at this table"}
	ErrInvalidPosition      = &TableError{"INVALID_POSITION", "Invalid table position"}
	ErrGameInProgress       = &TableError{"GAME_IN_PROGRESS", "Cannot perform action while game is in progress"}
	ErrInvalidPassword      = &TableError{"INVALID_PASSWORD", "Invalid table password"}
	ErrNotTableCreator      = &TableError{"NOT_TABLE_CREATOR", "Only table creator can perform this action"}
)

// TableJoinRequest represents a request to join a table
type TableJoinRequest struct {
	TableID  string        `json:"table_id"`
	PlayerID string        `json:"player_id"`
	Username string        `json:"username"`
	Mode     TableJoinMode `json:"mode"`               // player or observer
	Position int           `json:"position,omitempty"` // specific position (optional)
	Password string        `json:"password,omitempty"` // for private tables
}

// TableLeaveRequest represents a request to leave a table
type TableLeaveRequest struct {
	TableID  string `json:"table_id"`
	PlayerID string `json:"player_id"`
}

// TableCreateRequest represents a request to create a table
type TableCreateRequest struct {
	Name        string        `json:"name"`
	GameType    GameType      `json:"game_type"`
	CreatedBy   string        `json:"created_by"`
	Username    string        `json:"username"`
	Settings    TableSettings `json:"settings"`
	Description string        `json:"description,omitempty"`
	Tags        []string      `json:"tags,omitempty"`
}

// UserLimitState tracks rate limiting state for a user
type UserLimitState struct {
	// Table creation limits
	CreatedTables  []string    // List of table IDs created by user
	CreateAttempts []time.Time // Timestamps of recent create attempts

	// Join attempt limits
	JoinAttempts []time.Time // Timestamps of recent join attempts

	// Current state
	ObservedTables []string // Tables currently being observed
	ActiveTables   []string // Tables currently playing in

	LastActivity time.Time // Last activity timestamp
}

// Compatibility constructors for existing test files
func NewTableManager(factory GameEngineFactory) *ActorTableManager {
	return NewActorTableManager(factory)
}

func NewRateLimiter() *ActorRateLimiter {
	return NewActorRateLimiter()
}
