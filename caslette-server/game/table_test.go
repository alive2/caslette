package game

import (
	"context"
	"testing"
	"time"
)

// Mock implementations for testing

// MockGameEngineFactory implements GameEngineFactory for testing
type MockGameEngineFactory struct{}

func (f *MockGameEngineFactory) CreateEngine(gameType GameType, settings TableSettings) (GameEngine, error) {
	return NewMockGameEngine("test"), nil
}

// MockWebSocketHub implements WebSocketHub for testing
type MockWebSocketHub struct {
	broadcastCalls []BroadcastCall
}

type BroadcastCall struct {
	RoomID  string
	Message interface{}
}

func (h *MockWebSocketHub) BroadcastToRoom(roomID string, msg interface{}) error {
	h.broadcastCalls = append(h.broadcastCalls, BroadcastCall{
		RoomID:  roomID,
		Message: msg,
	})
	return nil
}

func (h *MockWebSocketHub) GetRoomUsers(roomID string) []map[string]interface{} {
	return []map[string]interface{}{}
}

// MockWebSocketConnection implements WebSocketConnection for testing
type MockWebSocketConnection struct {
	userID   string
	username string
	messages []interface{}
	rooms    []string
}

func (c *MockWebSocketConnection) GetUserID() string {
	return c.userID
}

func (c *MockWebSocketConnection) GetUsername() string {
	return c.username
}

func (c *MockWebSocketConnection) SendMessage(msg interface{}) error {
	c.messages = append(c.messages, msg)
	return nil
}

func (c *MockWebSocketConnection) JoinRoom(roomID string) error {
	c.rooms = append(c.rooms, roomID)
	return nil
}

func (c *MockWebSocketConnection) LeaveRoom(roomID string) error {
	for i, room := range c.rooms {
		if room == roomID {
			c.rooms = append(c.rooms[:i], c.rooms[i+1:]...)
			break
		}
	}
	return nil
}

func NewMockConnection(userID, username string) *MockWebSocketConnection {
	return &MockWebSocketConnection{
		userID:   userID,
		username: username,
		messages: make([]interface{}, 0),
		rooms:    make([]string, 0),
	}
}

// Test GameTable

func TestNewGameTable(t *testing.T) {
	settings := TableSettings{
		SmallBlind: 10,
		BigBlind:   20,
		BuyIn:      1000,
	}
	
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

func TestGameTablePlayerManagement(t *testing.T) {
	table := NewGameTable("test", "Test", GameTypeTexasHoldem, "creator", TableSettings{})
	
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
	
	if table.GetTotalCount() != 2 {
		t.Errorf("Expected total count 2, got %d", table.GetTotalCount())
	}
}

// Test TableManager

func TestTableManager(t *testing.T) {
	factory := &MockGameEngineFactory{}
	manager := NewTableManager(factory)
	
	if manager.GetTableCount() != 0 {
		t.Errorf("Expected 0 tables initially, got %d", manager.GetTableCount())
	}
}

func TestTableManagerCreateTable(t *testing.T) {
	factory := &MockGameEngineFactory{}
	manager := NewTableManager(factory)
	ctx := context.Background()
	
	req := &TableCreateRequest{
		Name:        "Test Table",
		GameType:    GameTypeTexasHoldem,
		CreatedBy:   "user1",
		Username:    "User1",
		Settings:    TableSettings{SmallBlind: 10, BigBlind: 20},
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
	
	if table.GameType != req.GameType {
		t.Errorf("Expected game type '%s', got '%s'", req.GameType, table.GameType)
	}
	
	if table.CreatedBy != req.CreatedBy {
		t.Errorf("Expected created by '%s', got '%s'", req.CreatedBy, table.CreatedBy)
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

func TestTableManagerJoinLeave(t *testing.T) {
	factory := &MockGameEngineFactory{}
	manager := NewTableManager(factory)
	ctx := context.Background()
	
	// Create table
	createReq := &TableCreateRequest{
		Name:     "Test Table",
		GameType: GameTypeTexasHoldem,
		CreatedBy: "creator",
		Username: "Creator",
		Settings: TableSettings{},
	}
	
	table, err := manager.CreateTable(ctx, createReq)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	
	// Test joining as player
	joinReq := &TableJoinRequest{
		TableID:  table.ID,
		PlayerID: "user1",
		Username: "User1",
		Mode:     JoinModePlayer,
	}
	
	err = manager.JoinTable(ctx, joinReq)
	if err != nil {
		t.Fatalf("Failed to join table: %v", err)
	}
	
	// Verify player joined
	updatedTable, _ := manager.GetTable(table.ID)
	if !updatedTable.IsPlayerAtTable("user1") {
		t.Error("Player should be at table after joining")
	}
	
	if updatedTable.GetPlayerCount() != 1 {
		t.Errorf("Expected 1 player, got %d", updatedTable.GetPlayerCount())
	}
	
	// Test joining as observer
	observerReq := &TableJoinRequest{
		TableID:  table.ID,
		PlayerID: "user2",
		Username: "User2",
		Mode:     JoinModeObserver,
	}
	
	err = manager.JoinTable(ctx, observerReq)
	if err != nil {
		t.Fatalf("Failed to join as observer: %v", err)
	}
	
	// Verify observer joined
	updatedTable, _ = manager.GetTable(table.ID)
	if !updatedTable.IsObserver("user2") {
		t.Error("User should be observer after joining")
	}
	
	// Test leaving
	leaveReq := &TableLeaveRequest{
		TableID:  table.ID,
		PlayerID: "user1",
	}
	
	err = manager.LeaveTable(ctx, leaveReq)
	if err != nil {
		t.Fatalf("Failed to leave table: %v", err)
	}
	
	// Verify player left
	updatedTable, _ = manager.GetTable(table.ID)
	if updatedTable.IsPlayerAtTable("user1") {
		t.Error("Player should not be at table after leaving")
	}
	
	if updatedTable.GetPlayerCount() != 0 {
		t.Errorf("Expected 0 players, got %d", updatedTable.GetPlayerCount())
	}
}

func TestTableManagerErrors(t *testing.T) {
	factory := &MockGameEngineFactory{}
	manager := NewTableManager(factory)
	ctx := context.Background()
	
	// Test joining non-existent table
	joinReq := &TableJoinRequest{
		TableID:  "nonexistent",
		PlayerID: "user1",
		Username: "User1",
		Mode:     JoinModePlayer,
	}
	
	err := manager.JoinTable(ctx, joinReq)
	if err != ErrTableNotFound {
		t.Errorf("Expected TABLE_NOT_FOUND error, got: %v", err)
	}
	
	// Test leaving non-existent table
	leaveReq := &TableLeaveRequest{
		TableID:  "nonexistent",
		PlayerID: "user1",
	}
	
	err = manager.LeaveTable(ctx, leaveReq)
	if err != ErrTableNotFound {
		t.Errorf("Expected TABLE_NOT_FOUND error, got: %v", err)
	}
}

func TestTableManagerFiltering(t *testing.T) {
	factory := &MockGameEngineFactory{}
	manager := NewTableManager(factory)
	ctx := context.Background()
	
	// Create different types of tables
	tables := []*TableCreateRequest{
		{
			Name:     "Texas Hold'em Table",
			GameType: GameTypeTexasHoldem,
			CreatedBy: "user1",
			Username: "User1",
			Settings: TableSettings{ObserversAllowed: true},
		},
		{
			Name:     "Private Table",
			GameType: GameTypeTexasHoldem,
			CreatedBy: "user2",
			Username: "User2",
			Settings: TableSettings{Private: true},
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

func TestTableManagerStats(t *testing.T) {
	factory := &MockGameEngineFactory{}
	manager := NewTableManager(factory)
	ctx := context.Background()
	
	// Create some tables and add players
	table1, _ := manager.CreateTable(ctx, &TableCreateRequest{
		Name:     "Table 1",
		GameType: GameTypeTexasHoldem,
		CreatedBy: "user1",
		Username: "User1",
		Settings: TableSettings{},
	})
	
	table2, _ := manager.CreateTable(ctx, &TableCreateRequest{
		Name:     "Table 2",
		GameType: GameTypeTexasHoldem,
		CreatedBy: "user2",
		Username: "User2",
		Settings: TableSettings{},
	})
	
	// Add some players
	manager.JoinTable(ctx, &TableJoinRequest{
		TableID:  table1.ID,
		PlayerID: "player1",
		Username: "Player1",
		Mode:     JoinModePlayer,
	})
	
	manager.JoinTable(ctx, &TableJoinRequest{
		TableID:  table1.ID,
		PlayerID: "observer1",
		Username: "Observer1",
		Mode:     JoinModeObserver,
	})
	
	manager.JoinTable(ctx, &TableJoinRequest{
		TableID:  table2.ID,
		PlayerID: "player2",
		Username: "Player2",
		Mode:     JoinModePlayer,
	})
	
	// Get stats
	stats := manager.GetStats()
	
	totalTables := stats["total_tables"].(int)
	totalPlayers := stats["total_players"].(int)
	totalObservers := stats["total_observers"].(int)
	
	if totalTables != 2 {
		t.Errorf("Expected 2 total tables, got %d", totalTables)
	}
	
	if totalPlayers != 2 {
		t.Errorf("Expected 2 total players, got %d", totalPlayers)
	}
	
	if totalObservers != 1 {
		t.Errorf("Expected 1 total observer, got %d", totalObservers)
	}
}

// Test WebSocket integration

func TestTableWebSocketHandler(t *testing.T) {
	factory := &MockGameEngineFactory{}
	manager := NewTableManager(factory)
	hub := &MockWebSocketHub{}
	
	handler := NewTableWebSocketHandler(manager, hub)
	
	if handler == nil {
		t.Fatal("Failed to create websocket handler")
	}
	
	// Test getting message handlers
	handlers := handler.GetMessageHandlers()
	
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

func TestTableWebSocketCreateTable(t *testing.T) {
	factory := &MockGameEngineFactory{}
	manager := NewTableManager(factory)
	hub := &MockWebSocketHub{}
	
	handler := NewTableWebSocketHandler(manager, hub)
	conn := NewMockConnection("user1", "User1")
	ctx := context.Background()
	
	handlers := handler.GetMessageHandlers()
	createHandler := handlers["table_create"]
	
	msg := &WebSocketMessage{
		Type:      "table_create",
		RequestID: "req123",
		Data: map[string]interface{}{
			"name":        "Test Table",
			"game_type":   "texas_holdem",
			"description": "Test description",
			"settings": map[string]interface{}{
				"small_blind": 10,
				"big_blind":   20,
			},
		},
	}
	
	response := createHandler(ctx, conn, msg)
	
	if response == nil {
		t.Fatal("Expected response from create handler")
	}
	
	if !response.Success {
		t.Errorf("Expected successful response, got error: %s", response.Error)
	}
	
	if response.Type != "table_created" {
		t.Errorf("Expected response type 'table_created', got '%s'", response.Type)
	}
	
	// Verify table was created
	if manager.GetTableCount() != 1 {
		t.Errorf("Expected 1 table after creation, got %d", manager.GetTableCount())
	}
}

func TestTableWebSocketJoinTable(t *testing.T) {
	factory := &MockGameEngineFactory{}
	manager := NewTableManager(factory)
	hub := &MockWebSocketHub{}
	
	handler := NewTableWebSocketHandler(manager, hub)
	ctx := context.Background()
	
	// Create a table first
	table, _ := manager.CreateTable(ctx, &TableCreateRequest{
		Name:     "Test Table",
		GameType: GameTypeTexasHoldem,
		CreatedBy: "creator",
		Username: "Creator",
		Settings: TableSettings{},
	})
	
	// Test joining
	conn := NewMockConnection("user1", "User1")
	handlers := handler.GetMessageHandlers()
	joinHandler := handlers["table_join"]
	
	msg := &WebSocketMessage{
		Type:      "table_join",
		RequestID: "req123",
		Data: map[string]interface{}{
			"table_id": table.ID,
			"mode":     "player",
		},
	}
	
	response := joinHandler(ctx, conn, msg)
	
	if response == nil {
		t.Fatal("Expected response from join handler")
	}
	
	if !response.Success {
		t.Errorf("Expected successful response, got error: %s", response.Error)
	}
	
	if response.Type != "table_joined" {
		t.Errorf("Expected response type 'table_joined', got '%s'", response.Type)
	}
	
	// Verify player joined
	updatedTable, _ := manager.GetTable(table.ID)
	if !updatedTable.IsPlayerAtTable("user1") {
		t.Error("Player should be at table after joining")
	}
}