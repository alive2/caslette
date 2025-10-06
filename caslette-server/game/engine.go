package game

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// GameState represents the current state of a game
type GameState string

const (
	GameStateWaiting    GameState = "waiting"    // Waiting for players
	GameStateStarting   GameState = "starting"   // Game is starting
	GameStateInProgress GameState = "inprogress" // Game is active
	GameStateFinished   GameState = "finished"   // Game has ended
	GameStatePaused     GameState = "paused"     // Game is paused
)

// Player represents a player in the game
type Player struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	IsActive bool                   `json:"isActive"`
	Position int                    `json:"position"`
	Data     map[string]interface{} `json:"data,omitempty"` // Game-specific player data
}

// GameEvent represents an event that occurs during the game
type GameEvent struct {
	Type      string                 `json:"type"`
	PlayerID  string                 `json:"playerId,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// GameAction represents an action a player can take
type GameAction struct {
	Type     string                 `json:"type"`
	PlayerID string                 `json:"playerId"`
	Data     map[string]interface{} `json:"data"`
}

// GameEngine is the abstract interface that all game engines must implement
type GameEngine interface {
	// Game lifecycle
	Initialize(config map[string]interface{}) error
	Start() error
	Pause() error
	Resume() error
	End() error
	Reset() error

	// Player management
	AddPlayer(player *Player) error
	RemovePlayer(playerID string) error
	GetPlayer(playerID string) (*Player, error)
	GetPlayers() []*Player
	GetActivePlayer() (*Player, error)

	// Game state
	GetState() GameState
	GetGameData() map[string]interface{}
	IsValidAction(action *GameAction) error

	// Action handling
	ProcessAction(ctx context.Context, action *GameAction) (*GameEvent, error)
	GetValidActions(playerID string) []string

	// Security methods for data filtering
	GetPublicGameState() map[string]interface{}
	GetPlayerState(playerID string) map[string]interface{}

	// Game flow
	NextTurn() error
	GetCurrentPlayerID() string
	IsGameOver() bool
	GetWinners() []*Player

	// Events
	GetEvents() []*GameEvent
	SubscribeToEvents(callback func(*GameEvent))

	// Additional methods for WebSocket integration
	GetGameState() map[string]interface{}
	GetHandHistory(limit int) []map[string]interface{}
	GetPlayerStats(playerID string) map[string]interface{}
}

// BaseGameEngine provides common functionality for all game engines
type BaseGameEngine struct {
	gameID      string
	state       GameState
	players     map[string]*Player
	gameData    map[string]interface{}
	events      []*GameEvent
	callbacks   []func(*GameEvent)
	currentTurn int
	config      map[string]interface{}
}

// NewBaseGameEngine creates a new base game engine
func NewBaseGameEngine(gameID string) *BaseGameEngine {
	return &BaseGameEngine{
		gameID:    gameID,
		state:     GameStateWaiting,
		players:   make(map[string]*Player),
		gameData:  make(map[string]interface{}),
		events:    make([]*GameEvent, 0),
		callbacks: make([]func(*GameEvent), 0),
		config:    make(map[string]interface{}),
	}
}

// GetState returns the current game state
func (b *BaseGameEngine) GetState() GameState {
	return b.state
}

// SetState sets the game state and emits an event
func (b *BaseGameEngine) SetState(state GameState) {
	oldState := b.state
	b.state = state
	b.emitEvent(&GameEvent{
		Type: "state_changed",
		Data: map[string]interface{}{
			"oldState": oldState,
			"newState": state,
		},
		Timestamp: time.Now(),
	})
}

// AddPlayer adds a player to the game
func (b *BaseGameEngine) AddPlayer(player *Player) error {
	if b.state != GameStateWaiting {
		return fmt.Errorf("cannot add player when game state is %s", b.state)
	}

	if _, exists := b.players[player.ID]; exists {
		return fmt.Errorf("player %s already exists", player.ID)
	}

	player.Position = len(b.players)
	player.IsActive = true
	b.players[player.ID] = player

	b.emitEvent(&GameEvent{
		Type:     "player_joined",
		PlayerID: player.ID,
		Data: map[string]interface{}{
			"player": player,
		},
		Timestamp: time.Now(),
	})

	return nil
}

// RemovePlayer removes a player from the game
func (b *BaseGameEngine) RemovePlayer(playerID string) error {
	player, exists := b.players[playerID]
	if !exists {
		return fmt.Errorf("player %s not found", playerID)
	}

	delete(b.players, playerID)

	b.emitEvent(&GameEvent{
		Type:     "player_left",
		PlayerID: playerID,
		Data: map[string]interface{}{
			"player": player,
		},
		Timestamp: time.Now(),
	})

	return nil
}

// GetPlayer returns a specific player
func (b *BaseGameEngine) GetPlayer(playerID string) (*Player, error) {
	player, exists := b.players[playerID]
	if !exists {
		return nil, fmt.Errorf("player %s not found", playerID)
	}
	return player, nil
}

// GetPlayers returns all players
func (b *BaseGameEngine) GetPlayers() []*Player {
	players := make([]*Player, 0, len(b.players))
	for _, player := range b.players {
		players = append(players, player)
	}
	return players
}

// GetActivePlayer returns the currently active player
func (b *BaseGameEngine) GetActivePlayer() (*Player, error) {
	activePlayers := b.getActivePlayers()

	if len(activePlayers) == 0 {
		return nil, fmt.Errorf("no active players")
	}

	// Return player whose turn it is based on current turn index
	if b.currentTurn >= 0 && b.currentTurn < len(activePlayers) {
		return activePlayers[b.currentTurn], nil
	}

	return activePlayers[0], nil
}

// getActivePlayers returns active players in a consistent order (sorted by Position)
func (b *BaseGameEngine) getActivePlayers() []*Player {
	activePlayers := make([]*Player, 0)
	for _, player := range b.players {
		if player.IsActive {
			activePlayers = append(activePlayers, player)
		}
	}

	// Sort by position to ensure consistent ordering
	sort.Slice(activePlayers, func(i, j int) bool {
		return activePlayers[i].Position < activePlayers[j].Position
	})

	return activePlayers
}

// GetGameData returns the current game data
func (b *BaseGameEngine) GetGameData() map[string]interface{} {
	return b.gameData
}

// SetGameData sets game data
func (b *BaseGameEngine) SetGameData(key string, value interface{}) {
	b.gameData[key] = value
}

// GetEvents returns all game events
func (b *BaseGameEngine) GetEvents() []*GameEvent {
	return b.events
}

// SubscribeToEvents subscribes to game events
func (b *BaseGameEngine) SubscribeToEvents(callback func(*GameEvent)) {
	b.callbacks = append(b.callbacks, callback)
}

// emitEvent emits an event to all subscribers
func (b *BaseGameEngine) emitEvent(event *GameEvent) {
	b.events = append(b.events, event)
	for _, callback := range b.callbacks {
		go callback(event)
	}
}

// NextTurn advances to the next player's turn
func (b *BaseGameEngine) NextTurn() error {
	activePlayers := b.getActivePlayers()

	if len(activePlayers) == 0 {
		return fmt.Errorf("no active players")
	}

	b.currentTurn = (b.currentTurn + 1) % len(activePlayers)

	currentPlayer := activePlayers[b.currentTurn]
	b.emitEvent(&GameEvent{
		Type:     "turn_changed",
		PlayerID: currentPlayer.ID,
		Data: map[string]interface{}{
			"currentPlayer": currentPlayer,
			"turnIndex":     b.currentTurn,
		},
		Timestamp: time.Now(),
	})

	return nil
}

// GetCurrentPlayerID returns the ID of the current player
func (b *BaseGameEngine) GetCurrentPlayerID() string {
	player, err := b.GetActivePlayer()
	if err != nil {
		return ""
	}
	return player.ID
}

// Initialize sets up the game with configuration
func (b *BaseGameEngine) Initialize(config map[string]interface{}) error {
	b.config = config
	b.SetState(GameStateWaiting)
	return nil
}

// Default implementations that can be overridden
func (b *BaseGameEngine) Start() error {
	if len(b.players) == 0 {
		return fmt.Errorf("cannot start game with no players")
	}
	b.SetState(GameStateInProgress)
	return nil
}

func (b *BaseGameEngine) Pause() error {
	if b.state != GameStateInProgress {
		return fmt.Errorf("can only pause a game in progress")
	}
	b.SetState(GameStatePaused)
	return nil
}

func (b *BaseGameEngine) Resume() error {
	if b.state != GameStatePaused {
		return fmt.Errorf("can only resume a paused game")
	}
	b.SetState(GameStateInProgress)
	return nil
}

func (b *BaseGameEngine) End() error {
	b.SetState(GameStateFinished)
	return nil
}

func (b *BaseGameEngine) Reset() error {
	b.state = GameStateWaiting
	b.events = make([]*GameEvent, 0)
	b.gameData = make(map[string]interface{})
	b.currentTurn = 0
	return nil
}

// These methods must be implemented by concrete game engines
func (b *BaseGameEngine) IsValidAction(action *GameAction) error {
	return fmt.Errorf("IsValidAction must be implemented by concrete game engine")
}

func (b *BaseGameEngine) ProcessAction(ctx context.Context, action *GameAction) (*GameEvent, error) {
	return nil, fmt.Errorf("ProcessAction must be implemented by concrete game engine")
}

func (b *BaseGameEngine) GetValidActions(playerID string) []string {
	return []string{}
}

func (b *BaseGameEngine) IsGameOver() bool {
	return b.state == GameStateFinished
}

func (b *BaseGameEngine) GetWinners() []*Player {
	return []*Player{}
}

// GetGameState returns the current game state for WebSocket clients
func (b *BaseGameEngine) GetGameState() map[string]interface{} {
	return map[string]interface{}{
		"game_id":      b.gameID,
		"state":        b.state,
		"current_turn": b.currentTurn,
		"players":      b.GetPlayers(),
		"events_count": len(b.events),
		"is_game_over": b.IsGameOver(),
	}
}

// GetHandHistory returns hand history (base implementation)
func (b *BaseGameEngine) GetHandHistory(limit int) []map[string]interface{} {
	// Base implementation returns game events as history
	history := make([]map[string]interface{}, 0)
	eventCount := len(b.events)
	start := 0
	if limit > 0 && eventCount > limit {
		start = eventCount - limit
	}

	for i := start; i < eventCount; i++ {
		event := b.events[i]
		history = append(history, map[string]interface{}{
			"type":      event.Type,
			"player_id": event.PlayerID,
			"data":      event.Data,
			"timestamp": event.Timestamp,
		})
	}

	return history
}

// GetPlayerStats returns player statistics (base implementation)
func (b *BaseGameEngine) GetPlayerStats(playerID string) map[string]interface{} {
	player, err := b.GetPlayer(playerID)
	if err != nil {
		return map[string]interface{}{
			"error": "Player not found",
		}
	}

	// Count player events
	eventCount := 0
	for _, event := range b.events {
		if event.PlayerID == playerID {
			eventCount++
		}
	}

	return map[string]interface{}{
		"player_id":   player.ID,
		"name":        player.Name,
		"is_active":   player.IsActive,
		"position":    player.Position,
		"event_count": eventCount,
	}
}
