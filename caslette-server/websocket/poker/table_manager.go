package poker

import (
	"caslette-server/models"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// TableManager handles table-related operations
type TableManager struct {
	db          *gorm.DB
	gameManager *GameManager

	// Broadcast callbacks - will be set by router
	broadcastTableUpdate     func(tableID uint, msg PokerMessage)
	broadcastTableListUpdate func()
}

func NewTableManager(db *gorm.DB, gameManager *GameManager) *TableManager {
	return &TableManager{
		db:          db,
		gameManager: gameManager,
	}
}

func (tm *TableManager) HandleCreateTable(client Client, msg PokerMessage) {
	var req CreateTableRequest
	data, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(data, &req); err != nil {
		client.SendError("Invalid create table request")
		return
	}

	// Validate request
	if req.Name == "" || req.MinBuyIn <= 0 || req.MaxBuyIn <= req.MinBuyIn ||
		req.SmallBlind <= 0 || req.BigBlind <= req.SmallBlind {
		client.SendError("Invalid table parameters")
		return
	}

	// Set defaults
	if req.GameType == "" {
		req.GameType = "texas_holdem"
	}
	if req.MaxPlayers == 0 {
		req.MaxPlayers = 9
	}
	if req.RakePercent == 0 {
		req.RakePercent = 0.05
	}

	// Create table in database
	table := models.PokerTable{
		Name:        req.Name,
		GameType:    req.GameType,
		MaxPlayers:  req.MaxPlayers,
		MinBuyIn:    req.MinBuyIn,
		MaxBuyIn:    req.MaxBuyIn,
		SmallBlind:  req.SmallBlind,
		BigBlind:    req.BigBlind,
		RakePercent: req.RakePercent,
		MaxRake:     req.MaxRake,
		Status:      "waiting",
		CreatedBy:   client.GetUserID(),
		IsPrivate:   req.IsPrivate,
		Password:    req.Password,
	}

	if err := tm.db.Create(&table).Error; err != nil {
		client.SendError("Failed to create table")
		return
	}

	// Initialize table in game manager
	tm.gameManager.InitializeTable(table.ID)

	client.SendSuccess(MsgCreateTable, map[string]interface{}{
		"table_id": table.ID,
		"message":  "Table created successfully",
	})

	// Broadcast table list update
	tm.BroadcastTableListUpdate()
}

func (tm *TableManager) HandleListTables(client Client, msg PokerMessage) {
	var tables []models.PokerTable
	query := tm.db.Preload("Creator").Preload("Players.User")

	// Only show public tables unless user owns private tables
	query = query.Where("is_private = false OR created_by = ?", client.GetUserID())

	if err := query.Find(&tables).Error; err != nil {
		client.SendError("Failed to fetch tables")
		return
	}

	// Build response
	var tableResponses []TableListResponse
	for _, table := range tables {
		response := tm.buildTableResponse(&table)
		tableResponses = append(tableResponses, response)
	}

	client.SendSuccess(MsgListTables, map[string]interface{}{
		"tables": tableResponses,
	})
}

func (tm *TableManager) HandleJoinTable(client Client, msg PokerMessage) {
	var req JoinTableRequest
	data, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(data, &req); err != nil {
		client.SendError("Invalid join table request")
		return
	}

	// Get table lock
	tableLock := tm.gameManager.GetTableLock(req.TableID)
	tableLock.Lock()
	defer tableLock.Unlock()

	// Start transaction
	tx := tm.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get table
	var table models.PokerTable
	if err := tx.First(&table, req.TableID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			client.SendError("Table not found")
		} else {
			client.SendError("Failed to fetch table")
		}
		return
	}

	// Check password for private tables
	if table.IsPrivate && table.Password != req.Password {
		tx.Rollback()
		client.SendError("Invalid password for private table")
		return
	}

	// Check if user is already at the table
	var existingPlayer models.TablePlayer
	if err := tx.Where("table_id = ? AND user_id = ?", table.ID, client.GetUserID()).First(&existingPlayer).Error; err == nil {
		tx.Rollback()
		client.SendError("You are already seated at this table")
		return
	}

	// Validate buy-in limits
	if req.BuyInAmount < table.MinBuyIn || req.BuyInAmount > table.MaxBuyIn {
		tx.Rollback()
		client.SendError("Buy-in amount must be within table limits")
		return
	}

	// TODO: Add diamond balance validation and transaction processing

	// Check if table has available seats
	var playerCount int64
	tx.Model(&models.TablePlayer{}).Where("table_id = ?", table.ID).Count(&playerCount)
	if int(playerCount) >= table.MaxPlayers {
		tx.Rollback()
		client.SendError("Table is full")
		return
	}

	// Find available seat
	seatNumber := tm.findAvailableSeat(tx, table.ID, table.MaxPlayers)

	// Create table player
	player := models.TablePlayer{
		TableID:    table.ID,
		UserID:     client.GetUserID(),
		SeatNumber: seatNumber,
		ChipCount:  req.BuyInAmount,
		Status:     "sitting",
		JoinedAt:   time.Now(),
	}

	if err := tx.Create(&player).Error; err != nil {
		tx.Rollback()
		client.SendError("Failed to join table")
		return
	}

	if err := tx.Commit().Error; err != nil {
		client.SendError("Failed to complete join table operation")
		return
	}

	client.SendSuccess(MsgJoinTable, map[string]interface{}{
		"table_id":    table.ID,
		"seat_number": seatNumber,
		"chip_count":  req.BuyInAmount,
		"message":     "Successfully joined table",
	})

	// Broadcast table update
	tm.BroadcastTableUpdate(req.TableID)

	// Start game if enough players
	if int(playerCount)+1 >= 2 && table.Status == "waiting" {
		tm.gameManager.CheckAndStartGame(req.TableID)
	}
}

func (tm *TableManager) HandleLeaveTable(client Client, msg PokerMessage) {
	var req LeaveTableRequest
	data, _ := json.Marshal(msg.Data)
	if err := json.Unmarshal(data, &req); err != nil {
		client.SendError("Invalid leave table request")
		return
	}

	tableLock := tm.gameManager.GetTableLock(req.TableID)
	tableLock.Lock()
	defer tableLock.Unlock()

	// Start transaction
	tx := tm.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find player at table
	var player models.TablePlayer
	if err := tx.Where("table_id = ? AND user_id = ?", req.TableID, client.GetUserID()).First(&player).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			client.SendError("You are not seated at this table")
		} else {
			client.SendError("Failed to find player")
		}
		return
	}

	// TODO: Process cash-out and diamond transactions

	// Remove player from table (soft delete)
	if err := tx.Delete(&player).Error; err != nil {
		tx.Rollback()
		client.SendError("Failed to leave table")
		return
	}

	if err := tx.Commit().Error; err != nil {
		client.SendError("Failed to complete leave table operation")
		return
	}

	client.SendSuccess(MsgLeaveTable, map[string]interface{}{
		"message":        "Successfully left table",
		"chips_returned": player.ChipCount,
	})

	tm.BroadcastTableUpdate(req.TableID)
}

// Helper methods

func (tm *TableManager) buildTableResponse(table *models.PokerTable) TableListResponse {
	var playerCount int64
	tm.db.Model(&models.TablePlayer{}).Where("table_id = ?", table.ID).Count(&playerCount)

	// Get occupied seats
	var occupiedSeats []int
	tm.db.Model(&models.TablePlayer{}).Where("table_id = ?", table.ID).Pluck("seat_number", &occupiedSeats)

	// Calculate available seats
	var availableSeats []int
	for i := 1; i <= table.MaxPlayers; i++ {
		occupied := false
		for _, seat := range occupiedSeats {
			if seat == i {
				occupied = true
				break
			}
		}
		if !occupied {
			availableSeats = append(availableSeats, i)
		}
	}

	return TableListResponse{
		ID:             table.ID,
		Name:           table.Name,
		GameType:       table.GameType,
		MaxPlayers:     table.MaxPlayers,
		MinBuyIn:       table.MinBuyIn,
		MaxBuyIn:       table.MaxBuyIn,
		SmallBlind:     table.SmallBlind,
		BigBlind:       table.BigBlind,
		Status:         table.Status,
		IsPrivate:      table.IsPrivate,
		CreatedBy:      table.Creator.Username,
		PlayerCount:    playerCount,
		AvailableSeats: availableSeats,
	}
}

func (tm *TableManager) findAvailableSeat(tx *gorm.DB, tableID uint, maxPlayers int) int {
	var occupiedSeats []int
	tx.Model(&models.TablePlayer{}).Where("table_id = ?", tableID).Pluck("seat_number", &occupiedSeats)

	for i := 1; i <= maxPlayers; i++ {
		occupied := false
		for _, seat := range occupiedSeats {
			if seat == i {
				occupied = true
				break
			}
		}
		if !occupied {
			return i
		}
	}
	return 1 // Fallback
}

func (tm *TableManager) BroadcastTableUpdate(tableID uint) {
	// Implementation will be handled by the main manager
	if tm.broadcastTableUpdate != nil {
		msg := PokerMessage{
			Type: MsgTableUpdate,
			Data: map[string]interface{}{
				"table_id": tableID,
			},
		}
		tm.broadcastTableUpdate(tableID, msg)
	}
}

func (tm *TableManager) BroadcastTableListUpdate() {
	// Implementation will be handled by the main manager
	if tm.broadcastTableListUpdate != nil {
		tm.broadcastTableListUpdate()
	}
}
