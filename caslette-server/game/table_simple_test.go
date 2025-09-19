package game

import (
	"context"
	"testing"
	"time"
)

// TestTableBasics tests basic table functionality without websockets
func TestTableBasics(t *testing.T) {
	settings := DefaultTableSettings()
	settings.SmallBlind = 10
	settings.BigBlind = 20
	settings.BuyIn = 1000

	table := NewGameTable("test123", "Test Table", GameTypeTexasHoldem, "user1", settings)

	if table.ID != "test123" {
		t.Errorf("Expected ID 'test123', got '%s'", table.ID)
	}

	if table.Name != "Test Table" {
		t.Errorf("Expected name 'Test Table', got '%s'", table.Name)
	}

	if table.GameType != GameTypeTexasHoldem {
		t.Errorf("Expected game type '%s', got '%s'", GameTypeTexasHoldem, table.GameType)
	}

	if table.Status != TableStatusWaiting {
		t.Errorf("Expected status '%s', got '%s'", TableStatusWaiting, table.Status)
	}

	if table.CreatedBy != "user1" {
		t.Errorf("Expected created by 'user1', got '%s'", table.CreatedBy)
	}

	if table.MaxPlayers != 8 {
		t.Errorf("Expected max players 8, got %d", table.MaxPlayers)
	}

	if table.GetPlayerCount() != 0 {
		t.Errorf("Expected 0 players, got %d", table.GetPlayerCount())
	}
}

func TestTablePlayerSlots(t *testing.T) {
	table := NewGameTable("test", "Test", GameTypeTexasHoldem, "creator", DefaultTableSettings())

	// Test initial state
	if table.IsPlayerAtTable("user1") {
		t.Error("User should not be at table initially")
	}

	if !table.CanJoinAsPlayer("user1") {
		t.Error("User should be able to join as player initially")
	}

	// Test position management
	availableSlots := table.GetAvailableSlots()
	if len(availableSlots) != 8 {
		t.Errorf("Expected 8 available slots, got %d", len(availableSlots))
	}

	// Test manual slot assignment
	table.PlayerSlots[0] = PlayerSlot{
		Position: 0,
		PlayerID: "user1",
		Username: "User1",
		IsReady:  false,
		JoinedAt: time.Now(),
	}

	if !table.IsPlayerAtTable("user1") {
		t.Error("User should be at table after assignment")
	}

	if table.GetPlayerCount() != 1 {
		t.Errorf("Expected 1 player, got %d", table.GetPlayerCount())
	}

	position := table.GetPlayerPosition("user1")
	if position != 0 {
		t.Errorf("Expected position 0, got %d", position)
	}
}

func TestTableObservers(t *testing.T) {
	settings := DefaultTableSettings()
	settings.ObserversAllowed = true // Enable observers
	table := NewGameTable("test", "Test", GameTypeTexasHoldem, "creator", settings)

	// Test observer management
	if !table.CanJoinAsObserver("user2") {
		t.Error("User should be able to join as observer")
	}

	table.Observers = append(table.Observers, TableObserver{
		PlayerID: "user2",
		Username: "User2",
		JoinedAt: time.Now(),
	})

	if !table.IsObserver("user2") {
		t.Error("User should be observer after adding")
	}

	if table.GetObserverCount() != 1 {
		t.Errorf("Expected 1 observer, got %d", table.GetObserverCount())
	}

	if table.GetTotalCount() != 1 {
		t.Errorf("Expected total count 1, got %d", table.GetTotalCount())
	}
}

func TestTableManagerBasic(t *testing.T) {
	factory := &MockGameEngineFactory{}
	manager := NewTableManager(factory)

	if manager.GetTableCount() != 0 {
		t.Errorf("Expected 0 tables initially, got %d", manager.GetTableCount())
	}

	ctx := context.Background()

	req := &TableCreateRequest{
		Name:        "Test Table",
		GameType:    GameTypeTexasHoldem,
		CreatedBy:   "user1",
		Username:    "User1",
		Settings:    DefaultTableSettings(),
		Description: "Test description",
		Tags:        []string{"casual", "beginner"},
	}

	table, err := manager.CreateTable(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	if table.Name != req.Name {
		t.Errorf("Expected name '%s', got '%s'", req.Name, table.Name)
	}

	if manager.GetTableCount() != 1 {
		t.Errorf("Expected 1 table after creation, got %d", manager.GetTableCount())
	}

	// Test retrieval
	retrievedTable, err := manager.GetTable(table.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve table: %v", err)
	}

	if retrievedTable.ID != table.ID {
		t.Errorf("Retrieved table ID mismatch")
	}
}

func TestTableManagerFilters(t *testing.T) {
	factory := &MockGameEngineFactory{}
	manager := NewTableManager(factory)
	ctx := context.Background()

	// Create different types of tables
	tables := []*TableCreateRequest{
		{
			Name:      "Texas Hold'em Table",
			GameType:  GameTypeTexasHoldem,
			CreatedBy: "user1",
			Username:  "User1",
			Settings: func() TableSettings {
				s := DefaultTableSettings()
				s.ObserversAllowed = true
				return s
			}(),
		},
		{
			Name:      "Private Table",
			GameType:  GameTypeTexasHoldem,
			CreatedBy: "user2",
			Username:  "User2",
			Settings: func() TableSettings {
				s := DefaultTableSettings()
				s.Private = true
				s.ObserversAllowed = false // Disable observers for this table
				return s
			}(),
		},
	}

	for _, req := range tables {
		_, err := manager.CreateTable(ctx, req)
		if err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	// Test filtering by game type
	holdemTables := manager.ListTables(map[string]interface{}{
		"game_type": string(GameTypeTexasHoldem),
	})

	if len(holdemTables) != 2 {
		t.Errorf("Expected 2 Hold'em tables, got %d", len(holdemTables))
	}

	// Test filtering by creator
	user1Tables := manager.ListTables(map[string]interface{}{
		"created_by": "user1",
	})

	if len(user1Tables) != 1 {
		t.Errorf("Expected 1 table by user1, got %d", len(user1Tables))
	}

	// Test filtering by observers allowed
	observerTables := manager.ListTables(map[string]interface{}{
		"observers_allowed": true,
	})

	if len(observerTables) != 1 {
		t.Errorf("Expected 1 table with observers allowed, got %d", len(observerTables))
	}
}

func TestTableIntegration(t *testing.T) {
	hub := &MockWebSocketHub{}
	integration := NewTableGameIntegration(hub)

	if integration == nil {
		t.Fatal("Failed to create table integration")
	}

	tableManager := integration.GetTableManager()
	if tableManager == nil {
		t.Fatal("Table manager should not be nil")
	}

	wsHandler := integration.GetWebSocketHandler()
	if wsHandler == nil {
		t.Fatal("WebSocket handler should not be nil")
	}

	// Test message handlers
	handlers := integration.GetMessageHandlers()
	expectedHandlers := []string{
		"table_create", "table_join", "table_leave", "table_list",
		"table_get", "table_close", "table_set_ready", "table_start_game", "table_get_stats",
	}

	for _, expectedHandler := range expectedHandlers {
		if _, exists := handlers[expectedHandler]; !exists {
			t.Errorf("Missing handler for message type: %s", expectedHandler)
		}
	}
}

func TestTableSettings(t *testing.T) {
	// Test default settings
	defaults := DefaultTableSettings()
	if defaults.SmallBlind != 10 {
		t.Errorf("Expected small blind 10, got %d", defaults.SmallBlind)
	}
	if defaults.BigBlind != 20 {
		t.Errorf("Expected big blind 20, got %d", defaults.BigBlind)
	}
	if !defaults.ObserversAllowed {
		t.Error("Expected observers allowed by default")
	}

	// Test quick game settings
	quick := QuickGameSettings()
	if quick.SmallBlind != 5 {
		t.Errorf("Expected small blind 5, got %d", quick.SmallBlind)
	}
	if !quick.AutoStart {
		t.Error("Expected auto start for quick games")
	}

	// Test tournament settings
	tournament := TournamentSettings()
	if !tournament.TournamentMode {
		t.Error("Expected tournament mode to be true")
	}
	if tournament.SmallBlind != 25 {
		t.Errorf("Expected small blind 25, got %d", tournament.SmallBlind)
	}

	// Test private settings
	private := PrivateTableSettings("password123")
	if !private.Private {
		t.Error("Expected private to be true")
	}
	if private.Password != "password123" {
		t.Errorf("Expected password 'password123', got '%s'", private.Password)
	}
}
