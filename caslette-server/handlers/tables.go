package handlers

import (
	"caslette-server/game"
	"caslette-server/models"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SecureTableHandler handles HTTP requests for table operations with security enhancements
type SecureTableHandler struct {
	db           *gorm.DB
	tableManager *game.ActorTableManager
	validator    *SecurityValidator
}

// SecureTableCreateRequest with additional validation
type SecureTableCreateRequest struct {
	Name       string `json:"name" binding:"required"`
	GameType   string `json:"game_type" binding:"required"`
	MinPlayers int    `json:"min_players" binding:"required,min=2,max=10"`
	MaxPlayers int    `json:"max_players" binding:"required,min=2,max=10"`
	BuyIn      int64  `json:"buy_in" binding:"required,min=1"`
	IsPrivate  bool   `json:"is_private"`
	Password   string `json:"password"`
}

// SecureTableResponse with sanitized data
type SecureTableResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	GameType    string `json:"game_type"`
	MinPlayers  int    `json:"min_players"`
	MaxPlayers  int    `json:"max_players"`
	BuyIn       int64  `json:"buy_in"`
	IsPrivate   bool   `json:"is_private"`
	CreatorID   uint   `json:"creator_id"`
	CreatorName string `json:"creator_name"`
	Status      string `json:"status"`
	PlayerCount int    `json:"player_count"`
	RequestID   string `json:"request_id"`
}

// NewSecureTableHandler creates a new secure table handler
func NewSecureTableHandler(db *gorm.DB, tableManager *game.ActorTableManager) *SecureTableHandler {
	return &SecureTableHandler{
		db:           db,
		tableManager: tableManager,
		validator:    NewSecurityValidator(),
	}
}

// CreateTable handles POST /api/tables with security validation
func (h *SecureTableHandler) CreateTable(c *gin.Context) {
	requestID, _ := c.Get("request_id")

	var req SecureTableCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Invalid request format",
			"request_id": requestID,
		})
		return
	}

	// Get user info from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error":      "Authentication required",
			"request_id": requestID,
		})
		return
	}

	username, _ := c.Get("username")

	// Validate and sanitize inputs
	tableName, err := h.validator.ValidateAndSanitizeString(req.Name, "name", 50)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return
	}

	gameType, err := h.validator.ValidateAndSanitizeString(req.GameType, "name", 20)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return
	}

	// Validate game type is allowed
	allowedGameTypes := []string{"texas_holdem", "omaha", "seven_card_stud"}
	isValidGameType := false
	for _, allowed := range allowedGameTypes {
		if gameType == allowed {
			isValidGameType = true
			break
		}
	}
	if !isValidGameType {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Invalid game type",
			"request_id": requestID,
		})
		return
	}

	// Validate player counts
	if req.MinPlayers > req.MaxPlayers {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Minimum players cannot exceed maximum players",
			"request_id": requestID,
		})
		return
	}

	// Validate buy-in amount (reasonable limits)
	if req.BuyIn < 1 || req.BuyIn > 1000000 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Buy-in must be between 1 and 1,000,000 diamonds",
			"request_id": requestID,
		})
		return
	}

	// Validate password if private table
	var tablePassword string
	if req.IsPrivate {
		if req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error":      "Private tables require a password",
				"request_id": requestID,
			})
			return
		}

		tablePassword, err = h.validator.ValidateAndSanitizeString(req.Password, "name", 100)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error":      "Invalid password format",
				"request_id": requestID,
			})
			return
		}
	}

	// Check user's diamond balance before allowing table creation
	var user models.User
	if err := h.db.Preload("Diamonds").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":    false,
			"error":      "User not found",
			"request_id": requestID,
		})
		return
	}

	// Get current diamond balance
	var currentBalance int64
	h.db.Model(&models.Diamond{}).Where("user_id = ?", userID).Order("created_at desc").Limit(1).Pluck("balance", &currentBalance)

	if currentBalance < req.BuyIn {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Insufficient diamond balance",
			"request_id": requestID,
		})
		return
	}

	// Create table request with validated data
	tableCreateReq := game.TableCreateRequest{
		Name:      tableName,
		GameType:  game.GameType(gameType),
		CreatedBy: fmt.Sprintf("%d", userID.(uint)),
		Username: func() string {
			if username != nil {
				return username.(string)
			}
			return ""
		}(),
		Settings: game.TableSettings{
			BuyIn:            int(req.BuyIn),
			Private:          req.IsPrivate,
			Password:         tablePassword,
			ObserversAllowed: true, // Default setting
		},
		Description: fmt.Sprintf("Table created by user %d", userID.(uint)),
	}

	// Create table through actor manager (thread-safe)
	table, err := h.tableManager.CreateTable(context.Background(), &tableCreateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error":      "Failed to create table",
			"request_id": requestID,
		})
		return
	}

	// Save table to database
	if err := h.SaveTableToDB(table); err != nil {
		// Cleanup: close table if DB save fails
		h.tableManager.CloseTable(table.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error":      "Failed to save table",
			"request_id": requestID,
		})
		return
	}

	// Return secure response
	response := SecureTableResponse{
		ID:         table.ID,
		Name:       table.Name,
		GameType:   string(table.GameType),
		MinPlayers: table.MinPlayers,
		MaxPlayers: table.MaxPlayers,
		BuyIn:      int64(table.Settings.BuyIn),
		IsPrivate:  table.Settings.Private,
		CreatorID: func() uint {
			if id, err := strconv.ParseUint(table.CreatedBy, 10, 32); err == nil {
				return uint(id)
			}
			return 0
		}(),
		CreatorName: tableCreateReq.Username,
		Status:      string(table.Status),
		PlayerCount: func() int {
			count := 0
			for _, slot := range table.PlayerSlots {
				if slot.PlayerID != "" {
					count++
				}
			}
			return count
		}(),
		RequestID: requestID.(string),
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"table":   response,
	})
}

// GetTable handles GET /api/tables/:id with security validation
func (h *SecureTableHandler) GetTable(c *gin.Context) {
	requestID, _ := c.Get("request_id")

	// Validate table ID
	tableID, err := h.validator.ValidateIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Invalid table ID",
			"request_id": requestID,
		})
		return
	}

	tableIDStr := strconv.Itoa(int(tableID))

	// Get table from actor manager
	table, err := h.tableManager.GetTable(tableIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":    false,
			"error":      "Table not found",
			"request_id": requestID,
		})
		return
	}

	// Return secure response (no sensitive data like passwords)
	response := SecureTableResponse{
		ID:         table.ID,
		Name:       table.Name,
		GameType:   string(table.GameType),
		MinPlayers: table.MinPlayers,
		MaxPlayers: table.MaxPlayers,
		BuyIn:      int64(table.Settings.BuyIn),
		IsPrivate:  table.Settings.Private,
		CreatorID: func() uint {
			if id, err := strconv.ParseUint(table.CreatedBy, 10, 32); err == nil {
				return uint(id)
			}
			return 0
		}(),
		CreatorName: "", // Will be filled from username lookup
		Status:      string(table.Status),
		PlayerCount: func() int {
			count := 0
			for _, slot := range table.PlayerSlots {
				if slot.PlayerID != "" {
					count++
				}
			}
			return count
		}(),
		RequestID: requestID.(string),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"table":   response,
	})
}

// JoinTable handles POST /api/tables/:id/join with authorization and validation
func (h *SecureTableHandler) JoinTable(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error":      "Authentication required",
			"request_id": requestID,
		})
		return
	}

	// Validate table ID
	tableID, err := h.validator.ValidateIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Invalid table ID",
			"request_id": requestID,
		})
		return
	}

	tableIDStr := strconv.Itoa(int(tableID))

	// Parse password if provided
	var req struct {
		Password string `json:"password"`
	}
	c.ShouldBindJSON(&req)

	var password string
	if req.Password != "" {
		password, err = h.validator.ValidateAndSanitizeString(req.Password, "name", 100)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error":      "Invalid password format",
				"request_id": requestID,
			})
			return
		}
	}

	// Check user's diamond balance
	var user models.User
	if err := h.db.Preload("Diamonds").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":    false,
			"error":      "User not found",
			"request_id": requestID,
		})
		return
	}

	// Get current diamond balance
	var currentBalance int64
	h.db.Model(&models.Diamond{}).Where("user_id = ?", userID).Order("created_at desc").Limit(1).Pluck("balance", &currentBalance)

	// Get table to check buy-in requirement
	table, err := h.tableManager.GetTable(tableIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":    false,
			"error":      "Table not found",
			"request_id": requestID,
		})
		return
	}

	if currentBalance < int64(table.Settings.BuyIn) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Insufficient diamond balance",
			"request_id": requestID,
		})
		return
	}

	// Create join request
	username, _ := c.Get("username")
	joinReq := game.TableJoinRequest{
		TableID:  tableIDStr,
		PlayerID: fmt.Sprintf("%d", userID.(uint)),
		Username: username.(string),
		Mode:     game.JoinModePlayer, // Default to player mode
		Password: password,
	}

	// Join table through actor manager (thread-safe)
	if err := h.tableManager.JoinTable(c.Request.Context(), &joinReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      err.Error(),
			"request_id": requestID,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Successfully joined table",
		"request_id": requestID,
	})
}

// SaveTableToDB saves table to database with transaction safety
func (h *SecureTableHandler) SaveTableToDB(table *game.GameTable) error {
	tx := h.db.Begin()

	// Convert game table settings to JSON string
	settingsJSON, _ := json.Marshal(table.Settings)

	// Convert CreatedBy string to uint
	createdByID, _ := strconv.ParseUint(table.CreatedBy, 10, 32)

	gameTable := &models.GameTable{
		ID:          table.ID,
		Name:        table.Name,
		GameType:    string(table.GameType),
		Status:      string(table.Status),
		CreatedBy:   uint(createdByID),
		MaxPlayers:  table.MaxPlayers,
		MinPlayers:  table.MinPlayers,
		Description: table.Description,
		Settings:    string(settingsJSON),
		RoomID:      table.RoomID,
	}

	if err := tx.Create(gameTable).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
