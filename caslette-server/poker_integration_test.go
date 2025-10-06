package main

import (
	"caslette-server/auth"
	"caslette-server/config"
	"caslette-server/game"
	"caslette-server/websocket_v2"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiplayerPokerComplete is a comprehensive integration test for the multiplayer poker system
func TestMultiplayerPokerComplete(t *testing.T) {
	// Setup test server
	server, wsURL := setupTestServer(t)
	defer server.Close()

	// Create test users
	player1 := &TestUser{UserID: "user1", Username: "alice", Token: "token1"}
	player2 := &TestUser{UserID: "user2", Username: "bob", Token: "token2"}
	observer := &TestUser{UserID: "user3", Username: "charlie", Token: "token3"}

	// Connect users via WebSocket
	conn1 := connectUser(t, wsURL, player1)
	defer conn1.Close()

	conn2 := connectUser(t, wsURL, player2)
	defer conn2.Close()

	conn3 := connectUser(t, wsURL, observer)
	defer conn3.Close()

	t.Run("Complete Poker Game Flow", func(t *testing.T) {
		// 1. Create a table
		tableID := createTable(t, conn1, "Texas Hold'em Test", game.GameTypeTexasHoldem)

		// 2. Second player joins table
		joinTable(t, conn2, tableID, game.JoinModePlayer)

		// 3. Observer joins table room
		joinTableRoom(t, conn3, tableID)

		// 4. Players set ready
		setReady(t, conn1, tableID, true)
		setReady(t, conn2, tableID, true)

		// 5. Start game
		startGame(t, conn1, tableID)

		// 6. Play a round of poker
		playPokerRound(t, conn1, conn2, conn3, tableID)

		// 7. Get game statistics
		getGameStats(t, conn1, tableID)

		// 8. Close table
		closeTable(t, conn1, tableID)
	})
}

func TestPokerActionValidation(t *testing.T) {
	server, wsURL := setupTestServer(t)
	defer server.Close()

	player1 := &TestUser{UserID: "user1", Username: "alice", Token: "token1"}
	player2 := &TestUser{UserID: "user2", Username: "bob", Token: "token2"}

	conn1 := connectUser(t, wsURL, player1)
	defer conn1.Close()

	conn2 := connectUser(t, wsURL, player2)
	defer conn2.Close()

	t.Run("Invalid Actions", func(t *testing.T) {
		// Create and setup table
		tableID := createTable(t, conn1, "Action Test", game.GameTypeTexasHoldem)
		joinTable(t, conn2, tableID, game.JoinModePlayer)
		setReady(t, conn1, tableID, true)
		setReady(t, conn2, tableID, true)
		startGame(t, conn1, tableID)

		// Test invalid action - wrong player's turn
		testInvalidAction(t, conn2, tableID, "fold", 0, "not player's turn")

		// Test invalid action - negative bet amount
		testInvalidAction(t, conn1, tableID, "raise", -100, "invalid amount")

		// Test invalid action - insufficient chips
		testInvalidAction(t, conn1, tableID, "raise", 999999, "insufficient chips")
	})
}

func TestTableManagement(t *testing.T) {
	server, wsURL := setupTestServer(t)
	defer server.Close()

	player1 := &TestUser{UserID: "user1", Username: "alice", Token: "token1"}
	conn1 := connectUser(t, wsURL, player1)
	defer conn1.Close()

	t.Run("Table CRUD Operations", func(t *testing.T) {
		// Create table
		tableID := createTable(t, conn1, "CRUD Test", game.GameTypeTexasHoldem)

		// Get table info
		tableInfo := getTable(t, conn1, tableID)
		assert.Equal(t, "CRUD Test", tableInfo["name"])
		assert.Equal(t, string(game.GameTypeTexasHoldem), tableInfo["game_type"])

		// List tables
		tables := listTables(t, conn1)
		assert.GreaterOrEqual(t, len(tables), 1)

		// Close table
		closeTable(t, conn1, tableID)

		// Verify table is closed
		time.Sleep(100 * time.Millisecond)
		tableInfo = getTable(t, conn1, tableID)
		assert.Equal(t, string(game.TableStatusClosed), tableInfo["status"])
	})
}

func TestObserverFunctionality(t *testing.T) {
	server, wsURL := setupTestServer(t)
	defer server.Close()

	player1 := &TestUser{UserID: "user1", Username: "alice", Token: "token1"}
	player2 := &TestUser{UserID: "user2", Username: "bob", Token: "token2"}
	observer := &TestUser{UserID: "user3", Username: "charlie", Token: "token3"}

	conn1 := connectUser(t, wsURL, player1)
	defer conn1.Close()

	conn2 := connectUser(t, wsURL, player2)
	defer conn2.Close()

	conn3 := connectUser(t, wsURL, observer)
	defer conn3.Close()

	t.Run("Observer Access", func(t *testing.T) {
		// Create table
		tableID := createTable(t, conn1, "Observer Test", game.GameTypeTexasHoldem)
		joinTable(t, conn2, tableID, game.JoinModePlayer)

		// Observer joins as observer
		joinTable(t, conn3, tableID, game.JoinModeObserver)

		// Observer can get game state
		gameState := getGameState(t, conn3, tableID)
		assert.NotNil(t, gameState)

		// Observer can get hand history
		history := getHandHistory(t, conn3, tableID, 10)
		assert.NotNil(t, history)

		// Observer cannot perform game actions
		testUnauthorizedAction(t, conn3, tableID, "fold")
	})
}

// Helper functions for test scenarios

type TestUser struct {
	UserID   string
	Username string
	Token    string
}

func setupTestServer(t *testing.T) (*httptest.Server, string) {
	// Create minimal config for testing
	cfg := &config.Config{
		JWTSecret: "test-secret",
	}

	// Initialize auth service
	authService := auth.NewAuthService(cfg.JWTSecret)

	// Initialize WebSocket server
	wsServer := websocket_v2.NewServer(authService)

	// Setup poker system (reuse our main setup function)
	setupPokerSystemForTest(wsServer)

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wsServer.ServeHTTP(w, r)
	}))

	// Convert HTTP URL to WebSocket URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	return server, wsURL
}

func setupPokerSystemForTest(wsServer *websocket_v2.Server) {
	// Create WebSocket hub adapter
	hubAdapter := &TestWebSocketHubAdapter{server: wsServer}

	// Create table integration
	tableIntegration := game.NewTableGameIntegration(hubAdapter)

	// Register all table message handlers
	tableHandlers := tableIntegration.GetMessageHandlers()
	for messageType, handler := range tableHandlers {
		registerTableHandler(wsServer, messageType, handler)
	}

	// Register poker action handlers
	registerPokerActionHandlers(wsServer, tableIntegration.GetTableManager())
}

type TestWebSocketHubAdapter struct {
	server *websocket_v2.Server
}

func (w *TestWebSocketHubAdapter) BroadcastToRoom(roomID string, msg interface{}) error {
	return w.server.BroadcastToRoom(roomID, msg)
}

func (w *TestWebSocketHubAdapter) GetRoomUsers(roomID string) []map[string]interface{} {
	// Simplified for testing
	return []map[string]interface{}{}
}

func connectUser(t *testing.T, wsURL string, user *TestUser) *websocket.Conn {
	// Create custom header with auth token for testing
	header := http.Header{}
	header.Set("Authorization", "Bearer "+user.Token)

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	require.NoError(t, err)

	// Send auth message (simulate the auth process)
	authMsg := map[string]interface{}{
		"type":       "auth",
		"request_id": generateRequestID(),
		"data": map[string]interface{}{
			"token": user.Token,
		},
	}

	err = conn.WriteJSON(authMsg)
	require.NoError(t, err)

	// Read auth response
	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	assert.True(t, response["success"].(bool))

	return conn
}

func createTable(t *testing.T, conn *websocket.Conn, name string, gameType game.GameType) string {
	req := map[string]interface{}{
		"type":       "table_create",
		"request_id": generateRequestID(),
		"data": map[string]interface{}{
			"name":        name,
			"game_type":   string(gameType),
			"description": "Test table",
			"settings": map[string]interface{}{
				"small_blind":       10,
				"big_blind":         20,
				"buy_in":            1000,
				"max_buy_in":        2000,
				"auto_start":        false,
				"time_limit":        30,
				"tournament_mode":   false,
				"observers_allowed": true,
				"private":           false,
			},
		},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	return data["id"].(string)
}

func joinTable(t *testing.T, conn *websocket.Conn, tableID string, mode game.TableJoinMode) {
	req := map[string]interface{}{
		"type":       "table_join",
		"request_id": generateRequestID(),
		"data": map[string]interface{}{
			"table_id": tableID,
			"mode":     string(mode),
		},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.True(t, response["success"].(bool))
}

func joinTableRoom(t *testing.T, conn *websocket.Conn, tableID string) {
	req := map[string]interface{}{
		"type":       "join_table_room",
		"request_id": generateRequestID(),
		"data": map[string]interface{}{
			"table_id": tableID,
		},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.True(t, response["success"].(bool))
}

func setReady(t *testing.T, conn *websocket.Conn, tableID string, ready bool) {
	req := map[string]interface{}{
		"type":       "table_set_ready",
		"request_id": generateRequestID(),
		"data": map[string]interface{}{
			"table_id": tableID,
			"ready":    ready,
		},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.True(t, response["success"].(bool))
}

func startGame(t *testing.T, conn *websocket.Conn, tableID string) {
	req := map[string]interface{}{
		"type":       "table_start_game",
		"request_id": generateRequestID(),
		"data": map[string]interface{}{
			"table_id": tableID,
		},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.True(t, response["success"].(bool))
}

func playPokerRound(t *testing.T, conn1, conn2, conn3 *websocket.Conn, tableID string) {
	// Get initial game state
	gameState := getGameState(t, conn1, tableID)
	assert.NotNil(t, gameState)

	// Player 1 action (assuming they're first to act)
	performPokerAction(t, conn1, tableID, "call", 20)

	// Player 2 action
	performPokerAction(t, conn2, tableID, "raise", 40)

	// Player 1 response
	performPokerAction(t, conn1, tableID, "call", 20)

	// Check that observer received game updates
	// Note: In a real test, we'd listen for broadcast messages
}

func performPokerAction(t *testing.T, conn *websocket.Conn, tableID, action string, amount int) {
	req := map[string]interface{}{
		"type":       "poker_action",
		"request_id": generateRequestID(),
		"data": map[string]interface{}{
			"table_id": tableID,
			"action":   action,
			"amount":   amount,
		},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.True(t, response["success"].(bool))
}

func getGameState(t *testing.T, conn *websocket.Conn, tableID string) map[string]interface{} {
	req := map[string]interface{}{
		"type":       "get_game_state",
		"request_id": generateRequestID(),
		"data": map[string]interface{}{
			"table_id": tableID,
		},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.True(t, response["success"].(bool))

	return response["data"].(map[string]interface{})
}

func getHandHistory(t *testing.T, conn *websocket.Conn, tableID string, limit int) []interface{} {
	req := map[string]interface{}{
		"type":       "get_hand_history",
		"request_id": generateRequestID(),
		"data": map[string]interface{}{
			"table_id": tableID,
			"limit":    limit,
		},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	return data["history"].([]interface{})
}

func getTable(t *testing.T, conn *websocket.Conn, tableID string) map[string]interface{} {
	req := map[string]interface{}{
		"type":       "table_get",
		"request_id": generateRequestID(),
		"data": map[string]interface{}{
			"table_id": tableID,
		},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.True(t, response["success"].(bool))

	return response["data"].(map[string]interface{})
}

func listTables(t *testing.T, conn *websocket.Conn) []interface{} {
	req := map[string]interface{}{
		"type":       "table_list",
		"request_id": generateRequestID(),
		"data":       map[string]interface{}{},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.True(t, response["success"].(bool))

	return response["data"].([]interface{})
}

func closeTable(t *testing.T, conn *websocket.Conn, tableID string) {
	req := map[string]interface{}{
		"type":       "table_close",
		"request_id": generateRequestID(),
		"data": map[string]interface{}{
			"table_id": tableID,
		},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.True(t, response["success"].(bool))
}

func getGameStats(t *testing.T, conn *websocket.Conn, tableID string) {
	req := map[string]interface{}{
		"type":       "table_get_stats",
		"request_id": generateRequestID(),
		"data":       map[string]interface{}{},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.True(t, response["success"].(bool))
}

func testInvalidAction(t *testing.T, conn *websocket.Conn, tableID, action string, amount int, expectedError string) {
	req := map[string]interface{}{
		"type":       "poker_action",
		"request_id": generateRequestID(),
		"data": map[string]interface{}{
			"table_id": tableID,
			"action":   action,
			"amount":   amount,
		},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.False(t, response["success"].(bool))
	assert.Contains(t, response["error"].(string), expectedError)
}

func testUnauthorizedAction(t *testing.T, conn *websocket.Conn, tableID, action string) {
	req := map[string]interface{}{
		"type":       "poker_action",
		"request_id": generateRequestID(),
		"data": map[string]interface{}{
			"table_id": tableID,
			"action":   action,
			"amount":   0,
		},
	}

	err := conn.WriteJSON(req)
	require.NoError(t, err)

	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	require.False(t, response["success"].(bool))
	assert.Contains(t, response["error"].(string), "not at table")
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// Run the test
func main() {
	// This would normally be run with `go test`
	log.Println("Multiplayer poker backend integration test suite ready")
	log.Println("Run with: go test -v ./poker_integration_test.go")
}
