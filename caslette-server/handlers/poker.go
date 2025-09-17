package handlers

import (
	"caslette-server/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PokerHandler struct {
	db *gorm.DB
}

type CreateTableRequest struct {
	Name        string  `json:"name" binding:"required"`
	GameType    string  `json:"game_type"`
	MaxPlayers  int     `json:"max_players"`
	MinBuyIn    int64   `json:"min_buy_in" binding:"required,min=1"`
	MaxBuyIn    int64   `json:"max_buy_in" binding:"required,min=1"`
	SmallBlind  int64   `json:"small_blind" binding:"required,min=1"`
	BigBlind    int64   `json:"big_blind" binding:"required,min=1"`
	RakePercent float64 `json:"rake_percent"`
	MaxRake     int64   `json:"max_rake" binding:"required,min=1"`
	IsPrivate   bool    `json:"is_private"`
	Password    string  `json:"password"`
}

type JoinTableRequest struct {
	BuyInAmount int64  `json:"buy_in_amount" binding:"required,min=1"`
	Password    string `json:"password"`
}

type TableResponse struct {
	*models.PokerTable
	PlayerCount    int                   `json:"player_count"`
	AvailableSeats []int                 `json:"available_seats"`
	Players        []TablePlayerResponse `json:"players"`
}

type TablePlayerResponse struct {
	*models.TablePlayer
	Username string `json:"username"`
}

func NewPokerHandler(db *gorm.DB) *PokerHandler {
	return &PokerHandler{db: db}
}

// CreateTable creates a new poker table
func (h *PokerHandler) CreateTable(c *gin.Context) {
	var req CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Validate business rules
	if req.MaxBuyIn <= req.MinBuyIn {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Max buy-in must be greater than min buy-in"})
		return
	}

	if req.BigBlind <= req.SmallBlind {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Big blind must be greater than small blind"})
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
		req.RakePercent = 0.05 // 5% default rake
	}

	// Create table
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
		CreatedBy:   userID.(uint),
		IsPrivate:   req.IsPrivate,
		Password:    req.Password,
	}

	if err := h.db.Create(&table).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create table"})
		return
	}

	// Load the table with creator info
	h.db.Preload("Creator").First(&table, table.ID)

	response := h.buildTableResponse(&table)
	c.JSON(http.StatusCreated, response)
}

// ListTables returns all available poker tables
func (h *PokerHandler) ListTables(c *gin.Context) {
	var tables []models.PokerTable

	query := h.db.Preload("Creator").Preload("Players.User")

	// Filter by game type if specified
	if gameType := c.Query("game_type"); gameType != "" {
		query = query.Where("game_type = ?", gameType)
	}

	// Filter by status if specified
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Only show public tables unless user owns private tables
	userID, exists := c.Get("user_id")
	if exists {
		query = query.Where("is_private = false OR created_by = ?", userID.(uint))
	} else {
		query = query.Where("is_private = false")
	}

	if err := query.Find(&tables).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tables"})
		return
	}

	var responses []TableResponse
	for _, table := range tables {
		responses = append(responses, h.buildTableResponse(&table))
	}

	c.JSON(http.StatusOK, gin.H{"tables": responses})
}

// GetTable returns a specific poker table
func (h *PokerHandler) GetTable(c *gin.Context) {
	tableIDStr := c.Param("id")
	tableID, err := strconv.ParseUint(tableIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	var table models.PokerTable
	if err := h.db.Preload("Creator").Preload("Players.User").First(&table, uint(tableID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch table"})
		}
		return
	}

	// Check if user can access private table
	if table.IsPrivate {
		userID, exists := c.Get("user_id")
		if !exists || table.CreatedBy != userID.(uint) {
			// Check if user is already a player at the table
			if exists {
				var playerCount int64
				h.db.Model(&models.TablePlayer{}).Where("table_id = ? AND user_id = ?", table.ID, userID.(uint)).Count(&playerCount)
				if playerCount == 0 {
					c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to private table"})
					return
				}
			} else {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to private table"})
				return
			}
		}
	}

	response := h.buildTableResponse(&table)
	c.JSON(http.StatusOK, response)
}

// JoinTable allows a player to join a poker table
func (h *PokerHandler) JoinTable(c *gin.Context) {
	tableIDStr := c.Param("id")
	tableID, err := strconv.ParseUint(tableIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	var req JoinTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get table
	var table models.PokerTable
	if err := tx.First(&table, uint(tableID)).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch table"})
		}
		return
	}

	// Check password for private tables
	if table.IsPrivate && table.Password != req.Password {
		tx.Rollback()
		c.JSON(http.StatusForbidden, gin.H{"error": "Invalid password for private table"})
		return
	}

	// Check if user is already at the table
	var existingPlayer models.TablePlayer
	if err := tx.Where("table_id = ? AND user_id = ?", table.ID, userID.(uint)).First(&existingPlayer).Error; err == nil {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "You are already seated at this table"})
		return
	}

	// Check buy-in limits
	if req.BuyInAmount < table.MinBuyIn || req.BuyInAmount > table.MaxBuyIn {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Buy-in amount must be between %d and %d diamonds", table.MinBuyIn, table.MaxBuyIn),
		})
		return
	}

	// Check if user has enough diamonds
	var user models.User
	if err := tx.Preload("Diamonds").First(&user, userID.(uint)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data"})
		return
	}

	// Calculate current balance
	var currentBalance int64
	for _, diamond := range user.Diamonds {
		currentBalance += diamond.Amount
	}

	if currentBalance < req.BuyInAmount {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient diamond balance"})
		return
	}

	// Check if table has available seats
	var playerCount int64
	tx.Model(&models.TablePlayer{}).Where("table_id = ?", table.ID).Count(&playerCount)
	if int(playerCount) >= table.MaxPlayers {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "Table is full"})
		return
	}

	// Find available seat
	var occupiedSeats []int
	tx.Model(&models.TablePlayer{}).Where("table_id = ?", table.ID).Pluck("seat_number", &occupiedSeats)

	seatNumber := 1
	for i := 1; i <= table.MaxPlayers; i++ {
		occupied := false
		for _, seat := range occupiedSeats {
			if seat == i {
				occupied = true
				break
			}
		}
		if !occupied {
			seatNumber = i
			break
		}
	}

	// Create transaction for buy-in
	transaction := models.Transaction{
		UserID:      userID.(uint),
		TableID:     &table.ID,
		Amount:      -req.BuyInAmount, // Negative because it's leaving the user's balance
		Type:        "buy_in",
		Description: fmt.Sprintf("Buy-in to table: %s", table.Name),
		Status:      "completed",
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process buy-in transaction"})
		return
	}

	// Create diamond transaction record
	newBalance := currentBalance - req.BuyInAmount
	diamond := models.Diamond{
		UserID:        userID.(uint),
		Amount:        -req.BuyInAmount,
		Balance:       newBalance,
		TransactionID: transaction.TransactionID,
		Type:          "buy_in",
		Description:   fmt.Sprintf("Buy-in to poker table: %s", table.Name),
	}

	if err := tx.Create(&diamond).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update diamond balance"})
		return
	}

	// Create table player
	player := models.TablePlayer{
		TableID:    table.ID,
		UserID:     userID.(uint),
		SeatNumber: seatNumber,
		ChipCount:  req.BuyInAmount,
		Status:     "sitting",
	}

	if err := tx.Create(&player).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join table"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete join table operation"})
		return
	}

	// Load player with user info
	h.db.Preload("User").First(&player, player.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully joined table",
		"player":  h.buildTablePlayerResponse(&player),
	})
}

// LeaveTable allows a player to leave a poker table
func (h *PokerHandler) LeaveTable(c *gin.Context) {
	tableIDStr := c.Param("id")
	tableID, err := strconv.ParseUint(tableIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find player at table
	var player models.TablePlayer
	if err := tx.Where("table_id = ? AND user_id = ?", uint(tableID), userID.(uint)).First(&player).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "You are not seated at this table"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find player"})
		}
		return
	}

	// Get table info
	var table models.PokerTable
	if err := tx.First(&table, uint(tableID)).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch table"})
		return
	}

	// TODO: Check if player is in an active hand - for now we'll allow leaving

	// Cash out the player's chips
	if player.ChipCount > 0 {
		// Create transaction for cash out
		transaction := models.Transaction{
			UserID:      userID.(uint),
			TableID:     &table.ID,
			Amount:      player.ChipCount, // Positive because it's returning to user's balance
			Type:        "cash_out",
			Description: fmt.Sprintf("Cash-out from table: %s", table.Name),
			Status:      "completed",
		}

		if err := tx.Create(&transaction).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process cash-out transaction"})
			return
		}

		// Calculate new balance
		var currentBalance int64
		var diamonds []models.Diamond
		tx.Where("user_id = ?", userID.(uint)).Find(&diamonds)
		for _, diamond := range diamonds {
			currentBalance += diamond.Amount
		}
		newBalance := currentBalance + player.ChipCount

		// Create diamond transaction record
		diamond := models.Diamond{
			UserID:        userID.(uint),
			Amount:        player.ChipCount,
			Balance:       newBalance,
			TransactionID: transaction.TransactionID,
			Type:          "cash_out",
			Description:   fmt.Sprintf("Cash-out from poker table: %s", table.Name),
		}

		if err := tx.Create(&diamond).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update diamond balance"})
			return
		}
	}

	// Remove player from table (soft delete)
	if err := tx.Delete(&player).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave table"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete leave table operation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Successfully left table",
		"chips_returned": player.ChipCount,
	})
}

// Helper functions

func (h *PokerHandler) buildTableResponse(table *models.PokerTable) TableResponse {
	var playerCount int64
	h.db.Model(&models.TablePlayer{}).Where("table_id = ?", table.ID).Count(&playerCount)

	// Get occupied seats
	var occupiedSeats []int
	h.db.Model(&models.TablePlayer{}).Where("table_id = ?", table.ID).Pluck("seat_number", &occupiedSeats)

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

	// Build player responses
	var players []TablePlayerResponse
	for _, player := range table.Players {
		players = append(players, h.buildTablePlayerResponse(&player))
	}

	// Hide password from response
	tableCopy := *table
	tableCopy.Password = ""

	return TableResponse{
		PokerTable:     &tableCopy,
		PlayerCount:    int(playerCount),
		AvailableSeats: availableSeats,
		Players:        players,
	}
}

func (h *PokerHandler) buildTablePlayerResponse(player *models.TablePlayer) TablePlayerResponse {
	username := ""
	if player.User.Username != "" {
		username = player.User.Username
	}

	return TablePlayerResponse{
		TablePlayer: player,
		Username:    username,
	}
}
