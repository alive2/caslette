package game

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestBaseGameEngine(t *testing.T) {
	t.Run("NewBaseGameEngine", func(t *testing.T) {
		engine := NewBaseGameEngine("test-game")
		if engine.gameID != "test-game" {
			t.Errorf("Expected game ID 'test-game', got %s", engine.gameID)
		}
		if engine.GetState() != GameStateWaiting {
			t.Errorf("Expected initial state %v, got %v", GameStateWaiting, engine.GetState())
		}
	})

	t.Run("AddPlayer", func(t *testing.T) {
		engine := NewBaseGameEngine("test-game")
		player := &Player{
			ID:   "player1",
			Name: "Player 1",
		}

		err := engine.AddPlayer(player)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(engine.GetPlayers()) != 1 {
			t.Errorf("Expected 1 player, got %d", len(engine.GetPlayers()))
		}

		// Try to add same player again
		err = engine.AddPlayer(player)
		if err == nil {
			t.Error("Expected error when adding duplicate player")
		}
	})

	t.Run("AddPlayerWhenNotWaiting", func(t *testing.T) {
		engine := NewBaseGameEngine("test-game")
		engine.SetState(GameStateInProgress)

		player := &Player{
			ID:   "player1",
			Name: "Player 1",
		}

		err := engine.AddPlayer(player)
		if err == nil {
			t.Error("Expected error when adding player to non-waiting game")
		}
	})

	t.Run("RemovePlayer", func(t *testing.T) {
		engine := NewBaseGameEngine("test-game")
		player := &Player{
			ID:   "player1",
			Name: "Player 1",
		}

		engine.AddPlayer(player)
		err := engine.RemovePlayer("player1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if len(engine.GetPlayers()) != 0 {
			t.Errorf("Expected 0 players, got %d", len(engine.GetPlayers()))
		}

		// Try to remove non-existent player
		err = engine.RemovePlayer("nonexistent")
		if err == nil {
			t.Error("Expected error when removing non-existent player")
		}
	})

	t.Run("GetPlayer", func(t *testing.T) {
		engine := NewBaseGameEngine("test-game")
		player := &Player{
			ID:   "player1",
			Name: "Player 1",
		}

		engine.AddPlayer(player)

		retrieved, err := engine.GetPlayer("player1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if retrieved.ID != "player1" {
			t.Errorf("Expected player ID 'player1', got %s", retrieved.ID)
		}

		_, err = engine.GetPlayer("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent player")
		}
	})

	t.Run("StateTransitions", func(t *testing.T) {
		engine := NewBaseGameEngine("test-game")

		// Add a player so we can start
		player := &Player{ID: "player1", Name: "Player 1"}
		engine.AddPlayer(player)

		// Test start
		err := engine.Start()
		if err != nil {
			t.Errorf("Unexpected error starting game: %v", err)
		}
		if engine.GetState() != GameStateInProgress {
			t.Errorf("Expected state %v after start, got %v", GameStateInProgress, engine.GetState())
		}

		// Test pause
		err = engine.Pause()
		if err != nil {
			t.Errorf("Unexpected error pausing game: %v", err)
		}
		if engine.GetState() != GameStatePaused {
			t.Errorf("Expected state %v after pause, got %v", GameStatePaused, engine.GetState())
		}

		// Test resume
		err = engine.Resume()
		if err != nil {
			t.Errorf("Unexpected error resuming game: %v", err)
		}
		if engine.GetState() != GameStateInProgress {
			t.Errorf("Expected state %v after resume, got %v", GameStateInProgress, engine.GetState())
		}

		// Test end
		err = engine.End()
		if err != nil {
			t.Errorf("Unexpected error ending game: %v", err)
		}
		if engine.GetState() != GameStateFinished {
			t.Errorf("Expected state %v after end, got %v", GameStateFinished, engine.GetState())
		}
	})

	t.Run("StartWithoutPlayers", func(t *testing.T) {
		engine := NewBaseGameEngine("test-game")
		err := engine.Start()
		if err == nil {
			t.Error("Expected error when starting game without players")
		}
	})

	t.Run("EventSubscription", func(t *testing.T) {
		engine := NewBaseGameEngine("test-game")

		var receivedEvent *GameEvent
		engine.SubscribeToEvents(func(event *GameEvent) {
			receivedEvent = event
		})

		// Add player to trigger event
		player := &Player{ID: "player1", Name: "Player 1"}
		engine.AddPlayer(player)

		// Give a moment for the goroutine to execute
		time.Sleep(10 * time.Millisecond)

		if receivedEvent == nil {
			t.Error("Expected to receive an event")
		}
		if receivedEvent.Type != "player_joined" {
			t.Errorf("Expected event type 'player_joined', got %s", receivedEvent.Type)
		}
	})

	t.Run("NextTurn", func(t *testing.T) {
		engine := NewBaseGameEngine("test-game")

		// Add multiple players
		for i := 1; i <= 3; i++ {
			player := &Player{
				ID:   string(rune('0' + i)),
				Name: "Player " + string(rune('0'+i)),
			}
			engine.AddPlayer(player)
		}

		initialPlayer := engine.GetCurrentPlayerID()

		err := engine.NextTurn()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		newPlayer := engine.GetCurrentPlayerID()
		if newPlayer == initialPlayer {
			t.Error("Expected different player after NextTurn")
		}
	})

	t.Run("Reset", func(t *testing.T) {
		engine := NewBaseGameEngine("test-game")
		player := &Player{ID: "player1", Name: "Player 1"}
		engine.AddPlayer(player)
		engine.Start()

		err := engine.Reset()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if engine.GetState() != GameStateWaiting {
			t.Errorf("Expected state %v after reset, got %v", GameStateWaiting, engine.GetState())
		}
		if len(engine.GetEvents()) != 0 {
			t.Errorf("Expected 0 events after reset, got %d", len(engine.GetEvents()))
		}
	})
}

func TestGameEvent(t *testing.T) {
	t.Run("CreateEvent", func(t *testing.T) {
		event := &GameEvent{
			Type:     "test_event",
			PlayerID: "player1",
			Data: map[string]interface{}{
				"key": "value",
			},
			Timestamp: time.Now(),
		}

		if event.Type != "test_event" {
			t.Errorf("Expected type 'test_event', got %s", event.Type)
		}
		if event.PlayerID != "player1" {
			t.Errorf("Expected player ID 'player1', got %s", event.PlayerID)
		}
		if event.Data["key"] != "value" {
			t.Errorf("Expected data value 'value', got %v", event.Data["key"])
		}
	})
}

func TestGameAction(t *testing.T) {
	t.Run("CreateAction", func(t *testing.T) {
		action := &GameAction{
			Type:     "test_action",
			PlayerID: "player1",
			Data: map[string]interface{}{
				"parameter": "value",
			},
		}

		if action.Type != "test_action" {
			t.Errorf("Expected type 'test_action', got %s", action.Type)
		}
		if action.PlayerID != "player1" {
			t.Errorf("Expected player ID 'player1', got %s", action.PlayerID)
		}
		if action.Data["parameter"] != "value" {
			t.Errorf("Expected parameter value 'value', got %v", action.Data["parameter"])
		}
	})
}

func TestPlayer(t *testing.T) {
	t.Run("CreatePlayer", func(t *testing.T) {
		player := &Player{
			ID:       "player1",
			Name:     "Test Player",
			IsActive: true,
			Position: 0,
			Data: map[string]interface{}{
				"score": 100,
			},
		}

		if player.ID != "player1" {
			t.Errorf("Expected ID 'player1', got %s", player.ID)
		}
		if player.Name != "Test Player" {
			t.Errorf("Expected name 'Test Player', got %s", player.Name)
		}
		if !player.IsActive {
			t.Error("Expected player to be active")
		}
		if player.Position != 0 {
			t.Errorf("Expected position 0, got %d", player.Position)
		}
		if player.Data["score"] != 100 {
			t.Errorf("Expected score 100, got %v", player.Data["score"])
		}
	})
}

func TestGameStateString(t *testing.T) {
	tests := []struct {
		state    GameState
		expected string
	}{
		{GameStateWaiting, "waiting"},
		{GameStateStarting, "starting"},
		{GameStateInProgress, "inprogress"},
		{GameStateFinished, "finished"},
		{GameStatePaused, "paused"},
	}

	for _, test := range tests {
		if string(test.state) != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, string(test.state))
		}
	}
}

// Mock implementation for testing abstract methods
type MockGameEngine struct {
	*BaseGameEngine
	validActions map[string][]string
	gameOver     bool
	winners      []*Player
}

func NewMockGameEngine(gameID string) *MockGameEngine {
	return &MockGameEngine{
		BaseGameEngine: NewBaseGameEngine(gameID),
		validActions:   make(map[string][]string),
		gameOver:       false,
		winners:        make([]*Player, 0),
	}
}

func (m *MockGameEngine) IsValidAction(action *GameAction) error {
	if action.Type == "invalid" {
		return fmt.Errorf("invalid action")
	}
	return nil
}

func (m *MockGameEngine) ProcessAction(ctx context.Context, action *GameAction) (*GameEvent, error) {
	if action.Type == "error" {
		return nil, fmt.Errorf("action processing error")
	}

	return &GameEvent{
		Type:      "action_processed",
		PlayerID:  action.PlayerID,
		Data:      action.Data,
		Timestamp: time.Now(),
	}, nil
}

func (m *MockGameEngine) GetValidActions(playerID string) []string {
	if actions, exists := m.validActions[playerID]; exists {
		return actions
	}
	return []string{"default_action"}
}

func (m *MockGameEngine) IsGameOver() bool {
	return m.gameOver
}

func (m *MockGameEngine) GetWinners() []*Player {
	return m.winners
}

// GetPublicGameState returns mock public game state
func (m *MockGameEngine) GetPublicGameState() map[string]interface{} {
	return map[string]interface{}{
		"game_id": m.gameID,
		"state":   m.state,
		"players": len(m.players),
	}
}

// GetPlayerState returns mock player state
func (m *MockGameEngine) GetPlayerState(playerID string) map[string]interface{} {
	player, err := m.GetPlayer(playerID)
	if err != nil || player == nil {
		return nil
	}
	
	return map[string]interface{}{
		"player_id": playerID,
		"position":  player.Position,
		"hand":      []string{}, // Mock empty hand
	}
}

func TestMockGameEngine(t *testing.T) {
	t.Run("ProcessValidAction", func(t *testing.T) {
		engine := NewMockGameEngine("test-game")

		action := &GameAction{
			Type:     "valid_action",
			PlayerID: "player1",
			Data:     map[string]interface{}{"test": "data"},
		}

		event, err := engine.ProcessAction(context.Background(), action)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if event.Type != "action_processed" {
			t.Errorf("Expected event type 'action_processed', got %s", event.Type)
		}
	})

	t.Run("ProcessInvalidAction", func(t *testing.T) {
		engine := NewMockGameEngine("test-game")

		action := &GameAction{
			Type:     "error",
			PlayerID: "player1",
			Data:     map[string]interface{}{},
		}

		_, err := engine.ProcessAction(context.Background(), action)
		if err == nil {
			t.Error("Expected error when processing error action")
		}
	})

	t.Run("ValidateAction", func(t *testing.T) {
		engine := NewMockGameEngine("test-game")

		validAction := &GameAction{Type: "valid"}
		invalidAction := &GameAction{Type: "invalid"}

		if err := engine.IsValidAction(validAction); err != nil {
			t.Errorf("Valid action should not return error: %v", err)
		}

		if err := engine.IsValidAction(invalidAction); err == nil {
			t.Error("Invalid action should return error")
		}
	})

	t.Run("GetValidActions", func(t *testing.T) {
		engine := NewMockGameEngine("test-game")

		// Test default actions
		actions := engine.GetValidActions("player1")
		if len(actions) != 1 || actions[0] != "default_action" {
			t.Errorf("Expected default action, got %v", actions)
		}

		// Test custom actions
		engine.validActions["player1"] = []string{"action1", "action2"}
		actions = engine.GetValidActions("player1")
		if len(actions) != 2 {
			t.Errorf("Expected 2 actions, got %d", len(actions))
		}
	})
}
