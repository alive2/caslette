package main

import (
	"caslette-server/game"
	"caslette-server/websocket_v2"
	"context"
	"fmt"
	"log"
)

// Example of how to integrate the table system with your application

func setupTableSystem() {
	// Create websocket hub (assuming you have one)
	hub := websocket_v2.NewHub()
	
	// Create table integration
	tableIntegration := game.NewTableGameIntegration(hub)
	
	// Get the table manager for direct operations
	tableManager := tableIntegration.GetTableManager()
	
	// Register table message handlers with the hub
	tableHandlers := tableIntegration.GetMessageHandlers()
	for messageType, handler := range tableHandlers {
		hub.RegisterMessageHandler(messageType, func(ctx context.Context, conn *websocket_v2.Connection, msg *websocket_v2.Message) *websocket_v2.Message {
			// Adapter to convert between websocket types
			tableConn := &WebSocketConnectionAdapter{conn: conn}
			tableMsg := &game.WebSocketMessage{
				Type:      msg.Type,
				RequestID: msg.RequestID,
				Data:      msg.Data,
			}
			
			response := handler(ctx, tableConn, tableMsg)
			if response == nil {
				return nil
			}
			
			return &websocket_v2.Message{
				Type:      response.Type,
				RequestID: response.RequestID,
				Success:   response.Success,
				Error:     response.Error,
				Data:      response.Data,
			}
		})
	}
	
	// Example: Create a table programmatically
	ctx := context.Background()
	
	createRequest := &game.TableCreateRequest{
		Name:        "Texas Hold'em - High Stakes",
		GameType:    game.GameTypeTexasHoldem,
		CreatedBy:   "admin",
		Username:    "Admin",
		Settings:    game.DefaultTableSettings(),
		Description: "High stakes Texas Hold'em table",
		Tags:        []string{"high-stakes", "texas-holdem"},
	}
	
	table, err := tableManager.CreateTable(ctx, createRequest)
	if err != nil {
		log.Printf("Failed to create table: %v", err)
		return
	}
	
	log.Printf("Created table: %s (ID: %s)", table.Name, table.ID)
	
	// Example: List all tables
	tables := tableManager.ListTables(map[string]interface{}{})
	log.Printf("Total tables: %d", len(tables))
	
	for _, t := range tables {
		info := t.GetTableInfo()
		log.Printf("- %s: %d/%d players, Status: %s", 
			info["name"], info["player_count"], info["max_players"], info["status"])
	}
	
	// Example: Get statistics
	stats := tableManager.GetStats()
	log.Printf("Table statistics: %+v", stats)
}

// WebSocketConnectionAdapter adapts websocket_v2.Connection to game.WebSocketConnection
type WebSocketConnectionAdapter struct {
	conn *websocket_v2.Connection
}

func (w *WebSocketConnectionAdapter) GetUserID() string {
	return w.conn.UserID
}

func (w *WebSocketConnectionAdapter) GetUsername() string {
	return w.conn.Username
}

func (w *WebSocketConnectionAdapter) SendMessage(msg interface{}) error {
	return w.conn.SendMessage(msg)
}

func (w *WebSocketConnectionAdapter) JoinRoom(roomID string) error {
	// Implementation depends on your websocket system
	return nil
}

func (w *WebSocketConnectionAdapter) LeaveRoom(roomID string) error {
	// Implementation depends on your websocket system
	return nil
}

// Example usage in your main application
func main() {
	fmt.Println("Setting up table system...")
	setupTableSystem()
	fmt.Println("Table system setup complete!")
}

/*
Usage Examples:

1. Creating a table via WebSocket:
{
  "type": "table_create",
  "request_id": "req123",
  "data": {
    "name": "My Poker Table",
    "game_type": "texas_holdem",
    "description": "Casual poker game",
    "settings": {
      "small_blind": 10,
      "big_blind": 20,
      "buy_in": 1000,
      "observers_allowed": true
    }
  }
}

2. Joining a table:
{
  "type": "table_join",
  "request_id": "req124",
  "data": {
    "table_id": "abc123",
    "mode": "player"
  }
}

3. Listing tables:
{
  "type": "table_list",
  "request_id": "req125",
  "data": {
    "game_type": "texas_holdem",
    "has_space": true
  }
}

4. Getting table info:
{
  "type": "table_get",
  "request_id": "req126",
  "data": {
    "table_id": "abc123"
  }
}

5. Leaving a table:
{
  "type": "table_leave",
  "request_id": "req127",
  "data": {
    "table_id": "abc123"
  }
}

API Integration:
- POST   /api/tables           - Create table
- GET    /api/tables           - List tables (with filters)
- GET    /api/tables/:id       - Get table details
- POST   /api/tables/:id/join  - Join table
- POST   /api/tables/:id/leave - Leave table
- DELETE /api/tables/:id       - Close table (creator only)
- GET    /api/tables/stats     - Get statistics
- GET    /api/tables/my        - Get user's tables

Database Models:
- game_tables        - Table metadata
- table_players      - Player assignments
- table_observers    - Observer assignments  
- game_sessions      - Game history

Features:
✅ Table creation and management
✅ Player and observer management
✅ Position-based seating
✅ Real-time WebSocket updates
✅ RESTful API endpoints
✅ Configurable game settings
✅ Private/password protected tables
✅ Auto-start functionality
✅ Database persistence
✅ Comprehensive filtering
✅ Statistics and monitoring
✅ Game engine integration
✅ Security and validation
*/