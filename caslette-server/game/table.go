package game

import (
	"time"
)

// TableStatus represents the current state of a game table
type TableStatus string

const (
	TableStatusWaiting  TableStatus = "waiting"  // Waiting for players
	TableStatusActive   TableStatus = "active"   // Game in progress
	TableStatusPaused   TableStatus = "paused"   // Game paused
	TableStatusFinished TableStatus = "finished" // Game completed
	TableStatusClosed   TableStatus = "closed"   // Table closed
)

// GameType represents the type of game being played
type GameType string

const (
	GameTypeTexasHoldem GameType = "texas_holdem"
	// Add more game types as they're implemented
)

// TableSettings contains configurable settings for a table
type TableSettings struct {
	// Game-specific settings
	SmallBlind     int  `json:"small_blind"`
	BigBlind       int  `json:"big_blind"`
	BuyIn          int  `json:"buy_in"`
	MaxBuyIn       int  `json:"max_buy_in"`
	AutoStart      bool `json:"auto_start"`      // Auto start when enough players join
	TimeLimit      int  `json:"time_limit"`      // Turn time limit in seconds
	TournamentMode bool `json:"tournament_mode"` // Tournament vs cash game

	// Table behavior
	ObserversAllowed bool   `json:"observers_allowed"`  // Allow spectators
	Private          bool   `json:"private"`            // Requires invitation
	Password         string `json:"password,omitempty"` // Password protection
}

// PlayerSlot represents a player's position at the table
type PlayerSlot struct {
	Position int       `json:"position"`
	PlayerID string    `json:"player_id,omitempty"`
	Username string    `json:"username,omitempty"`
	IsReady  bool      `json:"is_ready"`
	JoinedAt time.Time `json:"joined_at,omitempty"`
}

// TableObserver represents an observer watching the table
type TableObserver struct {
	PlayerID string    `json:"player_id"`
	Username string    `json:"username"`
	JoinedAt time.Time `json:"joined_at"`
}

// GameTable represents a game table where players can join and play
type GameTable struct {
	// Basic info
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	GameType  GameType    `json:"game_type"`
	Status    TableStatus `json:"status"`
	CreatedBy string      `json:"created_by"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`

	// Player management
	MaxPlayers  int             `json:"max_players"`
	MinPlayers  int             `json:"min_players"`
	PlayerSlots []PlayerSlot    `json:"player_slots"`
	Observers   []TableObserver `json:"observers"` // Observers watching the game

	// Game state
	GameEngine GameEngine    `json:"-"` // Don't serialize the engine
	Settings   TableSettings `json:"settings"`
	RoomID     string        `json:"room_id"` // Associated websocket room

	// Metadata
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// NewGameTable creates a new game table
func NewGameTable(id, name string, gameType GameType, createdBy string, settings TableSettings) *GameTable {
	now := time.Now()

	// Determine max players based on game type
	maxPlayers := 8 // Default for Texas Hold'em
	minPlayers := 2

	switch gameType {
	case GameTypeTexasHoldem:
		maxPlayers = 8
		minPlayers = 2
	}

	// Initialize player slots
	playerSlots := make([]PlayerSlot, maxPlayers)
	for i := range playerSlots {
		playerSlots[i] = PlayerSlot{
			Position: i,
		}
	}

	return &GameTable{
		ID:          id,
		Name:        name,
		GameType:    gameType,
		Status:      TableStatusWaiting,
		CreatedBy:   createdBy,
		CreatedAt:   now,
		UpdatedAt:   now,
		MaxPlayers:  maxPlayers,
		MinPlayers:  minPlayers,
		PlayerSlots: playerSlots,
		Observers:   make([]TableObserver, 0),
		Settings:    settings,
		RoomID:      "table_" + id, // Default room naming
	}
}

// GetPlayerCount returns the number of active players
func (t *GameTable) GetPlayerCount() int {
	count := 0
	for _, slot := range t.PlayerSlots {
		if slot.PlayerID != "" {
			count++
		}
	}
	return count
}

// GetObserverCount returns the number of observers
func (t *GameTable) GetObserverCount() int {
	return len(t.Observers)
}

// GetTotalCount returns total number of people at the table (players + observers)
func (t *GameTable) GetTotalCount() int {
	return t.GetPlayerCount() + t.GetObserverCount()
}

// IsPlayerAtTable checks if a player is sitting at the table
func (t *GameTable) IsPlayerAtTable(playerID string) bool {
	for _, slot := range t.PlayerSlots {
		if slot.PlayerID == playerID {
			return true
		}
	}
	return false
}

// IsObserver checks if a player is observing the table
func (t *GameTable) IsObserver(playerID string) bool {
	for _, observer := range t.Observers {
		if observer.PlayerID == playerID {
			return true
		}
	}
	return false
}

// GetPlayerPosition returns the position of a player at the table (-1 if not found)
func (t *GameTable) GetPlayerPosition(playerID string) int {
	for _, slot := range t.PlayerSlots {
		if slot.PlayerID == playerID {
			return slot.Position
		}
	}
	return -1
}

// CanJoinAsPlayer checks if a player can join as a player
func (t *GameTable) CanJoinAsPlayer(playerID string) bool {
	// Check if table is in a joinable state
	if t.Status != TableStatusWaiting && t.Status != TableStatusPaused {
		return false
	}

	// Check if player is already at table
	if t.IsPlayerAtTable(playerID) {
		return false
	}

	// Check if there's an available slot
	return t.GetPlayerCount() < t.MaxPlayers
}

// CanJoinAsObserver checks if a player can join as an observer
func (t *GameTable) CanJoinAsObserver(playerID string) bool {
	// Check if observers are allowed
	if !t.Settings.ObserversAllowed {
		return false
	}

	// Check if table is closed
	if t.Status == TableStatusClosed {
		return false
	}

	// Check if already observing or playing
	return !t.IsPlayerAtTable(playerID) && !t.IsObserver(playerID)
}

// GetAvailableSlots returns positions of available player slots
func (t *GameTable) GetAvailableSlots() []int {
	var available []int
	for _, slot := range t.PlayerSlots {
		if slot.PlayerID == "" {
			available = append(available, slot.Position)
		}
	}
	return available
}

// GetTableInfo returns public information about the table
func (t *GameTable) GetTableInfo() map[string]interface{} {
	return map[string]interface{}{
		"id":             t.ID,
		"name":           t.Name,
		"game_type":      t.GameType,
		"status":         t.Status,
		"created_by":     t.CreatedBy,
		"created_at":     t.CreatedAt,
		"updated_at":     t.UpdatedAt,
		"max_players":    t.MaxPlayers,
		"min_players":    t.MinPlayers,
		"player_count":   t.GetPlayerCount(),
		"observer_count": len(t.Observers),
		"settings":       t.Settings,
		"description":    t.Description,
		"tags":           t.Tags,
		"room_id":        t.RoomID,
	}
}

// GetDetailedInfo returns detailed information including player slots (for players/observers)
func (t *GameTable) GetDetailedInfo() map[string]interface{} {
	info := t.GetTableInfo()

	// Add detailed player information
	info["player_slots"] = t.PlayerSlots
	info["observers"] = t.Observers

	return info
}

// Touch updates the UpdatedAt timestamp
func (t *GameTable) Touch() {
	t.UpdatedAt = time.Now()
}
