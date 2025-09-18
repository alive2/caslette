package game

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// WebSocketConnection interface for websocket operations
type WebSocketConnection interface {
	GetUserID() string
	GetUsername() string
	SendMessage(msg interface{}) error
	JoinRoom(roomID string) error
	LeaveRoom(roomID string) error
}

// WebSocketHub interface for hub operations
type WebSocketHub interface {
	BroadcastToRoom(roomID string, msg interface{}) error
	GetRoomUsers(roomID string) []map[string]interface{}
}

// TableWebSocketHandler handles websocket messages for table operations
type TableWebSocketHandler struct {
	tableManager *TableManager
	hub          WebSocketHub
}

// NewTableWebSocketHandler creates a new table websocket handler
func NewTableWebSocketHandler(tableManager *TableManager, hub WebSocketHub) *TableWebSocketHandler {
	handler := &TableWebSocketHandler{
		tableManager: tableManager,
		hub:          hub,
	}
	
	// Register as webhook handler for table events
	tableManager.AddWebhookHandler(handler)
	
	return handler
}

// Message represents a websocket message
type WebSocketMessage struct {
	Type      string      `json:"type"`
	RequestID string      `json:"request_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Success   bool        `json:"success"`
	Error     string      `json:"error,omitempty"`
	Room      string      `json:"room,omitempty"`
}

// GetMessageHandlers returns all table-related message handlers
func (h *TableWebSocketHandler) GetMessageHandlers() map[string]func(ctx context.Context, conn WebSocketConnection, msg *WebSocketMessage) *WebSocketMessage {
	return map[string]func(ctx context.Context, conn WebSocketConnection, msg *WebSocketMessage) *WebSocketMessage{
		"table_create":         h.handleCreateTable,
		"table_join":          h.handleJoinTable,
		"table_leave":         h.handleLeaveTable,
		"table_list":          h.handleListTables,
		"table_get":           h.handleGetTable,
		"table_close":         h.handleCloseTable,
		"table_set_ready":     h.handleSetReady,
		"table_start_game":    h.handleStartGame,
		"table_get_stats":     h.handleGetStats,
	}
}

// handleCreateTable handles table creation requests
func (h *TableWebSocketHandler) handleCreateTable(ctx context.Context, conn WebSocketConnection, msg *WebSocketMessage) *WebSocketMessage {
	var req TableCreateRequest
	if err := h.parseMessageData(msg.Data, &req); err != nil {
		return h.errorResponse(msg.RequestID, "INVALID_DATA", "Invalid request data: "+err.Error())
	}
	
	// Set creator info from connection
	req.CreatedBy = conn.GetUserID()
	req.Username = conn.GetUsername()
	
	// Create table
	table, err := h.tableManager.CreateTable(ctx, &req)
	if err != nil {
		return h.errorResponse(msg.RequestID, "CREATE_FAILED", err.Error())
	}
	
	// Auto-join creator as player
	joinReq := &TableJoinRequest{
		TableID:  table.ID,
		PlayerID: conn.GetUserID(),
		Username: conn.GetUsername(),
		Mode:     JoinModePlayer,
	}
	
	if err := h.tableManager.JoinTable(ctx, joinReq); err != nil {
		log.Printf("Failed to auto-join creator to table: %v", err)
	}
	
	return h.successResponse(msg.RequestID, "table_created", table.GetDetailedInfo())
}

// handleJoinTable handles table join requests
func (h *TableWebSocketHandler) handleJoinTable(ctx context.Context, conn WebSocketConnection, msg *WebSocketMessage) *WebSocketMessage {
	var req TableJoinRequest
	if err := h.parseMessageData(msg.Data, &req); err != nil {
		return h.errorResponse(msg.RequestID, "INVALID_DATA", "Invalid request data: "+err.Error())
	}
	
	// Set player info from connection
	req.PlayerID = conn.GetUserID()
	req.Username = conn.GetUsername()
	
	// Default to player mode if not specified
	if req.Mode == "" {
		req.Mode = JoinModePlayer
	}
	
	// Join table
	if err := h.tableManager.JoinTable(ctx, &req); err != nil {
		return h.errorResponse(msg.RequestID, "JOIN_FAILED", err.Error())
	}
	
	// Get updated table info
	table, _ := h.tableManager.GetTable(req.TableID)
	
	return h.successResponse(msg.RequestID, "table_joined", map[string]interface{}{
		"table": table.GetDetailedInfo(),
		"mode":  req.Mode,
	})
}

// handleLeaveTable handles table leave requests
func (h *TableWebSocketHandler) handleLeaveTable(ctx context.Context, conn WebSocketConnection, msg *WebSocketMessage) *WebSocketMessage {
	var req TableLeaveRequest
	if err := h.parseMessageData(msg.Data, &req); err != nil {
		return h.errorResponse(msg.RequestID, "INVALID_DATA", "Invalid request data: "+err.Error())
	}
	
	// Set player info from connection
	req.PlayerID = conn.GetUserID()
	
	// Leave table
	if err := h.tableManager.LeaveTable(ctx, &req); err != nil {
		return h.errorResponse(msg.RequestID, "LEAVE_FAILED", err.Error())
	}
	
	return h.successResponse(msg.RequestID, "table_left", map[string]interface{}{
		"table_id": req.TableID,
	})
}

// handleListTables handles table listing requests
func (h *TableWebSocketHandler) handleListTables(ctx context.Context, conn WebSocketConnection, msg *WebSocketMessage) *WebSocketMessage {
	// Parse optional filters
	filters := make(map[string]interface{})
	if msg.Data != nil {
		if filterMap, ok := msg.Data.(map[string]interface{}); ok {
			filters = filterMap
		}
	}
	
	// Get tables
	tables := h.tableManager.ListTables(filters)
	
	// Convert to public info
	var tableList []map[string]interface{}
	for _, table := range tables {
		tableList = append(tableList, table.GetTableInfo())
	}
	
	return h.successResponse(msg.RequestID, "table_list", tableList)
}

// handleGetTable handles get table info requests
func (h *TableWebSocketHandler) handleGetTable(ctx context.Context, conn WebSocketConnection, msg *WebSocketMessage) *WebSocketMessage {
	var req struct {
		TableID string `json:"table_id"`
	}
	if err := h.parseMessageData(msg.Data, &req); err != nil {
		return h.errorResponse(msg.RequestID, "INVALID_DATA", "Invalid request data: "+err.Error())
	}
	
	// Get table
	table, err := h.tableManager.GetTable(req.TableID)
	if err != nil {
		return h.errorResponse(msg.RequestID, "TABLE_NOT_FOUND", err.Error())
	}
	
	// Return detailed info if player is at table, otherwise public info
	playerID := conn.GetUserID()
	var tableInfo map[string]interface{}
	if table.IsPlayerAtTable(playerID) || table.IsObserver(playerID) {
		tableInfo = table.GetDetailedInfo()
	} else {
		tableInfo = table.GetTableInfo()
	}
	
	return h.successResponse(msg.RequestID, "table_info", tableInfo)
}

// handleCloseTable handles table close requests
func (h *TableWebSocketHandler) handleCloseTable(ctx context.Context, conn WebSocketConnection, msg *WebSocketMessage) *WebSocketMessage {
	var req struct {
		TableID string `json:"table_id"`
	}
	if err := h.parseMessageData(msg.Data, &req); err != nil {
		return h.errorResponse(msg.RequestID, "INVALID_DATA", "Invalid request data: "+err.Error())
	}
	
	// Get table to check permissions
	table, err := h.tableManager.GetTable(req.TableID)
	if err != nil {
		return h.errorResponse(msg.RequestID, "TABLE_NOT_FOUND", err.Error())
	}
	
	// Check if user can close table (creator only)
	if table.CreatedBy != conn.GetUserID() {
		return h.errorResponse(msg.RequestID, "NOT_AUTHORIZED", "Only table creator can close the table")
	}
	
	// Close table
	if err := h.tableManager.CloseTable(req.TableID); err != nil {
		return h.errorResponse(msg.RequestID, "CLOSE_FAILED", err.Error())
	}
	
	return h.successResponse(msg.RequestID, "table_closed", map[string]interface{}{
		"table_id": req.TableID,
	})
}

// handleSetReady handles player ready state changes
func (h *TableWebSocketHandler) handleSetReady(ctx context.Context, conn WebSocketConnection, msg *WebSocketMessage) *WebSocketMessage {
	var req struct {
		TableID string `json:"table_id"`
		Ready   bool   `json:"ready"`
	}
	if err := h.parseMessageData(msg.Data, &req); err != nil {
		return h.errorResponse(msg.RequestID, "INVALID_DATA", "Invalid request data: "+err.Error())
	}
	
	// Get table
	table, err := h.tableManager.GetTable(req.TableID)
	if err != nil {
		return h.errorResponse(msg.RequestID, "TABLE_NOT_FOUND", err.Error())
	}
	
	// Check if player is at table
	playerID := conn.GetUserID()
	position := table.GetPlayerPosition(playerID)
	if position == -1 {
		return h.errorResponse(msg.RequestID, "NOT_AT_TABLE", "Player is not at this table")
	}
	
	// Update ready state
	table.mutex.Lock()
	table.PlayerSlots[position].IsReady = req.Ready
	table.Touch()
	table.mutex.Unlock()
	
	// Broadcast update to room
	h.broadcastTableUpdate(table, "player_ready_changed", map[string]interface{}{
		"player_id": playerID,
		"position":  position,
		"ready":     req.Ready,
	})
	
	return h.successResponse(msg.RequestID, "ready_updated", map[string]interface{}{
		"ready": req.Ready,
	})
}

// handleStartGame handles manual game start requests
func (h *TableWebSocketHandler) handleStartGame(ctx context.Context, conn WebSocketConnection, msg *WebSocketMessage) *WebSocketMessage {
	var req struct {
		TableID string `json:"table_id"`
	}
	if err := h.parseMessageData(msg.Data, &req); err != nil {
		return h.errorResponse(msg.RequestID, "INVALID_DATA", "Invalid request data: "+err.Error())
	}
	
	// Get table
	table, err := h.tableManager.GetTable(req.TableID)
	if err != nil {
		return h.errorResponse(msg.RequestID, "TABLE_NOT_FOUND", err.Error())
	}
	
	// Check permissions (creator or all players ready)
	playerID := conn.GetUserID()
	if table.CreatedBy != playerID {
		// Check if all players are ready
		allReady := true
		for _, slot := range table.PlayerSlots {
			if slot.PlayerID != "" && !slot.IsReady {
				allReady = false
				break
			}
		}
		if !allReady {
			return h.errorResponse(msg.RequestID, "NOT_READY", "All players must be ready or you must be the table creator")
		}
	}
	
	// Try to start game
	table.mutex.Lock()
	err = h.tableManager.tryStartGame(table)
	table.mutex.Unlock()
	
	if err != nil {
		return h.errorResponse(msg.RequestID, "START_FAILED", err.Error())
	}
	
	return h.successResponse(msg.RequestID, "game_started", map[string]interface{}{
		"table_id": req.TableID,
	})
}

// handleGetStats handles stats requests
func (h *TableWebSocketHandler) handleGetStats(ctx context.Context, conn WebSocketConnection, msg *WebSocketMessage) *WebSocketMessage {
	stats := h.tableManager.GetStats()
	return h.successResponse(msg.RequestID, "table_stats", stats)
}

// Webhook handler implementations (TableWebhookHandler interface)

// OnTableCreated broadcasts table creation event
func (h *TableWebSocketHandler) OnTableCreated(table *GameTable) {
	// Broadcast to global table list subscribers (if any)
	// For now, just log
	log.Printf("Table created: %s (%s)", table.Name, table.ID)
}

// OnTableClosed broadcasts table closure event
func (h *TableWebSocketHandler) OnTableClosed(table *GameTable) {
	h.broadcastTableUpdate(table, "table_closed", map[string]interface{}{
		"table_id": table.ID,
		"reason":   "closed",
	})
	log.Printf("Table closed: %s (%s)", table.Name, table.ID)
}

// OnPlayerJoined broadcasts player join event
func (h *TableWebSocketHandler) OnPlayerJoined(table *GameTable, playerID, username string, mode TableJoinMode) {
	h.broadcastTableUpdate(table, "player_joined", map[string]interface{}{
		"player_id": playerID,
		"username":  username,
		"mode":      mode,
		"table":     table.GetDetailedInfo(),
	})
}

// OnPlayerLeft broadcasts player leave event
func (h *TableWebSocketHandler) OnPlayerLeft(table *GameTable, playerID string, mode TableJoinMode) {
	h.broadcastTableUpdate(table, "player_left", map[string]interface{}{
		"player_id": playerID,
		"mode":      mode,
		"table":     table.GetDetailedInfo(),
	})
}

// OnGameStarted broadcasts game start event
func (h *TableWebSocketHandler) OnGameStarted(table *GameTable) {
	h.broadcastTableUpdate(table, "game_started", map[string]interface{}{
		"table_id": table.ID,
		"table":    table.GetDetailedInfo(),
	})
}

// OnGameFinished broadcasts game finish event
func (h *TableWebSocketHandler) OnGameFinished(table *GameTable) {
	h.broadcastTableUpdate(table, "game_finished", map[string]interface{}{
		"table_id": table.ID,
		"table":    table.GetDetailedInfo(),
	})
}

// Helper methods

// parseMessageData unmarshals message data into target struct
func (h *TableWebSocketHandler) parseMessageData(data interface{}, target interface{}) error {
	if data == nil {
		return nil
	}
	
	// Convert to JSON and back to properly handle type conversion
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(jsonData, target)
}

// successResponse creates a successful response message
func (h *TableWebSocketHandler) successResponse(requestID, msgType string, data interface{}) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      msgType,
		RequestID: requestID,
		Success:   true,
		Data:      data,
	}
}

// errorResponse creates an error response message
func (h *TableWebSocketHandler) errorResponse(requestID, code, message string) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      "error",
		RequestID: requestID,
		Success:   false,
		Error:     fmt.Sprintf("[%s] %s", code, message),
	}
}

// broadcastTableUpdate broadcasts an update to all users in the table room
func (h *TableWebSocketHandler) broadcastTableUpdate(table *GameTable, eventType string, data interface{}) {
	if h.hub != nil {
		msg := &WebSocketMessage{
			Type: eventType,
			Data: data,
			Room: table.RoomID,
		}
		
		if err := h.hub.BroadcastToRoom(table.RoomID, msg); err != nil {
			log.Printf("Failed to broadcast to room %s: %v", table.RoomID, err)
		}
	}
}