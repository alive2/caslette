package poker

import (
	"caslette-server/models"
	"encoding/json"
	"log"

	"gorm.io/gorm"
)

// PokerRouter coordinates all poker operations
type PokerRouter struct {
	db            *gorm.DB
	tableManager  *TableManager
	gameManager   *GameManager
	clientManager *ClientManager
}

func NewPokerRouter(db *gorm.DB) *PokerRouter {
	gameManager := NewGameManager(db)
	tableManager := NewTableManager(db, gameManager)
	clientManager := NewClientManager()

	router := &PokerRouter{
		db:            db,
		tableManager:  tableManager,
		gameManager:   gameManager,
		clientManager: clientManager,
	}

	// Set up broadcast callbacks
	router.ConnectTableManagerBroadcasts()

	return router
}

// HandleMessage routes incoming poker messages to appropriate handlers
func (pr *PokerRouter) HandleMessage(client Client, message []byte) {
	var msg PokerMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		client.SendError("Invalid message format")
		return
	}

	log.Printf("Handling poker message: %s from user %d", msg.Type, client.GetUserID())

	switch msg.Type {
	// Table management messages
	case MsgCreateTable:
		pr.tableManager.HandleCreateTable(client, msg)
	case MsgListTables:
		pr.tableManager.HandleListTables(client, msg)
	case MsgJoinTable:
		pr.tableManager.HandleJoinTable(client, msg)
	case MsgLeaveTable:
		pr.tableManager.HandleLeaveTable(client, msg)

	// Game action messages
	case MsgGameAction:
		pr.gameManager.HandleGameAction(client, msg)
	case MsgGetGameState:
		pr.handleGetGameState(client, msg)

	// Spectator messages
	case MsgSpectateTable:
		pr.handleSpectateTable(client, msg)
	case MsgStopSpectating:
		pr.handleStopSpectating(client, msg)

	// Admin messages (future implementation)
	case MsgPauseTable:
		pr.handlePauseTable(client, msg)
	case MsgResumeTable:
		pr.handleResumeTable(client, msg)

	default:
		client.SendError("Unknown message type: " + msg.Type)
	}
}

// AddClient registers a new client connection
func (pr *PokerRouter) AddClient(userID uint, client Client) {
	pr.clientManager.AddClient(userID, client)
	log.Printf("Poker client added for user %d", userID)
}

// RemoveClient unregisters a client connection
func (pr *PokerRouter) RemoveClient(userID uint) {
	// Check if user is at any tables and handle cleanup
	pr.handleClientDisconnect(userID)
	pr.clientManager.RemoveClient(userID)
	log.Printf("Poker client removed for user %d", userID)
}

// BroadcastToTable sends a message to all clients at a specific table
func (pr *PokerRouter) BroadcastToTable(tableID uint, msg PokerMessage) {
	// Get all players at the table
	var players []models.TablePlayer
	if err := pr.db.Where("table_id = ?", tableID).Find(&players).Error; err != nil {
		log.Printf("Error fetching players for broadcast: %v", err)
		return
	}

	// Send message to each player's client
	for _, player := range players {
		if client, exists := pr.clientManager.GetClient(player.UserID); exists {
			if client.IsConnected() {
				client.SendMessage(msg)
			}
		}
	}
}

// BroadcastTableListUpdate notifies all clients about table list changes
func (pr *PokerRouter) BroadcastTableListUpdate() {
	// Get all tables
	var tables []models.PokerTable
	if err := pr.db.Preload("Creator").Preload("Players.User").Find(&tables).Error; err != nil {
		log.Printf("Error fetching tables for broadcast: %v", err)
		return
	}

	// Build table responses
	var tableResponses []TableListResponse
	for _, table := range tables {
		response := pr.tableManager.buildTableResponse(&table)
		tableResponses = append(tableResponses, response)
	}

	// Create broadcast message
	msg := PokerMessage{
		Type: MsgTableListUpdate,
		Data: map[string]interface{}{
			"tables": tableResponses,
		},
	}

	// Send to all connected clients
	clients := pr.clientManager.GetAllClients()
	for _, client := range clients {
		client.SendMessage(msg)
	}
}

// Handler implementations for specific message types

func (pr *PokerRouter) handleGetGameState(client Client, msg PokerMessage) {
	var req GetGameStateRequest
	data, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(data, &req); err != nil {
		client.SendError("Invalid get game state request")
		return
	}

	gameState, err := pr.gameManager.GetGameState(req.TableID, client.GetUserID())
	if err != nil {
		client.SendError("Failed to get game state: " + err.Error())
		return
	}

	client.SendSuccess(MsgGetGameState, gameState)
}

func (pr *PokerRouter) handleSpectateTable(client Client, msg PokerMessage) {
	var req SpectateTableRequest
	data, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(data, &req); err != nil {
		client.SendError("Invalid spectate request")
		return
	}

	// TODO: Implement spectator functionality
	// For now, just send game state
	gameState, err := pr.gameManager.GetGameState(req.TableID, 0) // 0 = no hole cards shown
	if err != nil {
		client.SendError("Failed to spectate table: " + err.Error())
		return
	}

	client.SendSuccess(MsgSpectateTable, gameState)
}

func (pr *PokerRouter) handleStopSpectating(client Client, msg PokerMessage) {
	// TODO: Implement stop spectating functionality
	client.SendSuccess(MsgStopSpectating, map[string]interface{}{
		"message": "Stopped spectating",
	})
}

func (pr *PokerRouter) handlePauseTable(client Client, msg PokerMessage) {
	// TODO: Implement table pause functionality (admin only)
	client.SendError("Pause table not implemented")
}

func (pr *PokerRouter) handleResumeTable(client Client, msg PokerMessage) {
	// TODO: Implement table resume functionality (admin only)
	client.SendError("Resume table not implemented")
}

func (pr *PokerRouter) handleClientDisconnect(userID uint) {
	// Find all tables where user is playing
	var players []models.TablePlayer
	if err := pr.db.Where("user_id = ?", userID).Find(&players).Error; err != nil {
		log.Printf("Error finding player tables during disconnect: %v", err)
		return
	}

	// Handle disconnect for each table
	for _, player := range players {
		pr.handlePlayerDisconnect(player.TableID, userID)
	}
}

func (pr *PokerRouter) handlePlayerDisconnect(tableID, userID uint) {
	// Get table lock
	tableLock := pr.gameManager.GetTableLock(tableID)
	tableLock.Lock()
	defer tableLock.Unlock()

	// TODO: Implement disconnect handling:
	// 1. Mark player as sitting out if in game
	// 2. Set timeout for automatic fold/leave
	// 3. Broadcast player status change
	// 4. Handle chip preservation

	log.Printf("Player %d disconnected from table %d", userID, tableID)

	// For now, just broadcast table update
	pr.BroadcastToTable(tableID, PokerMessage{
		Type: MsgPlayerDisconnected,
		Data: map[string]interface{}{
			"table_id": tableID,
			"user_id":  userID,
		},
	})
}

// Helper methods for table manager integration

func (pr *PokerRouter) ConnectTableManagerBroadcasts() {
	// Set up broadcast callbacks for table manager
	pr.tableManager.broadcastTableUpdate = func(tableID uint, msg PokerMessage) {
		pr.BroadcastToTable(tableID, msg)
	}
	pr.tableManager.broadcastTableListUpdate = func() {
		pr.BroadcastTableListUpdate()
	}
}

// GetStats returns poker system statistics
func (pr *PokerRouter) GetStats() map[string]interface{} {
	clients := pr.clientManager.GetAllClients()

	var activeTables int64
	pr.db.Model(&models.PokerTable{}).Where("status = ?", "playing").Count(&activeTables)

	var totalPlayers int64
	pr.db.Model(&models.TablePlayer{}).Count(&totalPlayers)

	return map[string]interface{}{
		"connected_clients": len(clients),
		"active_tables":     activeTables,
		"total_players":     totalPlayers,
	}
}
