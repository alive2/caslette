package handlers

import (
	"caslette-server/game"
	"caslette-server/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TableHandler handles HTTP requests for table operations
type TableHandler struct {
	db           *gorm.DB
	tableManager *game.TableManager
}

// NewTableHandler creates a new table handler
func NewTableHandler(db *gorm.DB, tableManager *game.TableManager) *TableHandler {
	return &TableHandler{
		db:           db,
		tableManager: tableManager,
	}
}

// CreateTable handles POST /api/tables
func (h *TableHandler) CreateTable(c *gin.Context) {
	var req game.TableCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get user info from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Authentication required",
		})
		return
	}

	username, _ := c.Get("username")
	
	// Set creator info
	req.CreatedBy = userID.(string)
	req.Username = username.(string)

	// Create table
	table, err := h.tableManager.CreateTable(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    table.GetTableInfo(),
	})
}

// GetTables handles GET /api/tables
func (h *TableHandler) GetTables(c *gin.Context) {
	// Parse query parameters for filtering
	filters := make(map[string]interface{})
	
	if gameType := c.Query("game_type"); gameType != "" {
		filters["game_type"] = gameType
	}
	
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	
	if createdBy := c.Query("created_by"); createdBy != "" {
		filters["created_by"] = createdBy
	}
	
	if hasSpace := c.Query("has_space"); hasSpace == "true" {
		filters["has_space"] = true
	}
	
	if observersAllowed := c.Query("observers_allowed"); observersAllowed == "true" {
		filters["observers_allowed"] = true
	}

	// Get tables
	tables := h.tableManager.ListTables(filters)

	// Convert to public info
	var tableList []map[string]interface{}
	for _, table := range tables {
		tableList = append(tableList, table.GetTableInfo())
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tableList,
		"count":   len(tableList),
	})
}

// GetTable handles GET /api/tables/:id
func (h *TableHandler) GetTable(c *gin.Context) {
	tableID := c.Param("id")
	
	// Get table
	table, err := h.tableManager.GetTable(tableID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Table not found",
		})
		return
	}

	// Check if user is authenticated and at table for detailed info
	userID, authenticated := c.Get("user_id")
	var tableInfo map[string]interface{}
	
	if authenticated {
		playerID := userID.(string)
		if table.IsPlayerAtTable(playerID) || table.IsObserver(playerID) {
			tableInfo = table.GetDetailedInfo()
		} else {
			tableInfo = table.GetTableInfo()
		}
	} else {
		tableInfo = table.GetTableInfo()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tableInfo,
	})
}

// JoinTable handles POST /api/tables/:id/join
func (h *TableHandler) JoinTable(c *gin.Context) {
	tableID := c.Param("id")
	
	var req struct {
		Mode     string `json:"mode"`     // "player" or "observer"
		Position int    `json:"position"` // optional specific position
		Password string `json:"password"` // for private tables
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get user info from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Authentication required",
		})
		return
	}

	username, _ := c.Get("username")

	// Create join request
	joinReq := &game.TableJoinRequest{
		TableID:  tableID,
		PlayerID: userID.(string),
		Username: username.(string),
		Mode:     game.TableJoinMode(req.Mode),
		Position: req.Position,
		Password: req.Password,
	}

	// Default to player mode
	if joinReq.Mode == "" {
		joinReq.Mode = game.JoinModePlayer
	}

	// Join table
	if err := h.tableManager.JoinTable(c.Request.Context(), joinReq); err != nil {
		statusCode := http.StatusBadRequest
		if tableErr, ok := err.(*game.TableError); ok {
			switch tableErr.Code {
			case "TABLE_NOT_FOUND":
				statusCode = http.StatusNotFound
			case "TABLE_FULL", "POSITION_TAKEN":
				statusCode = http.StatusConflict
			case "INVALID_PASSWORD":
				statusCode = http.StatusUnauthorized
			}
		}
		
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get updated table info
	table, _ := h.tableManager.GetTable(tableID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"table": table.GetDetailedInfo(),
			"mode":  req.Mode,
		},
	})
}

// LeaveTable handles POST /api/tables/:id/leave
func (h *TableHandler) LeaveTable(c *gin.Context) {
	tableID := c.Param("id")

	// Get user info from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Authentication required",
		})
		return
	}

	// Create leave request
	leaveReq := &game.TableLeaveRequest{
		TableID:  tableID,
		PlayerID: userID.(string),
	}

	// Leave table
	if err := h.tableManager.LeaveTable(c.Request.Context(), leaveReq); err != nil {
		statusCode := http.StatusBadRequest
		if tableErr, ok := err.(*game.TableError); ok {
			switch tableErr.Code {
			case "TABLE_NOT_FOUND":
				statusCode = http.StatusNotFound
			case "PLAYER_NOT_AT_TABLE":
				statusCode = http.StatusConflict
			}
		}
		
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"table_id": tableID,
		},
	})
}

// CloseTable handles DELETE /api/tables/:id
func (h *TableHandler) CloseTable(c *gin.Context) {
	tableID := c.Param("id")

	// Get user info from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Authentication required",
		})
		return
	}

	// Get table to check permissions
	table, err := h.tableManager.GetTable(tableID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Table not found",
		})
		return
	}

	// Check if user can close table (creator only)
	if table.CreatedBy != userID.(string) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Only table creator can close the table",
		})
		return
	}

	// Close table
	if err := h.tableManager.CloseTable(tableID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"table_id": tableID,
		},
	})
}

// GetTableStats handles GET /api/tables/stats
func (h *TableHandler) GetTableStats(c *gin.Context) {
	stats := h.tableManager.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetMyTables handles GET /api/tables/my
func (h *TableHandler) GetMyTables(c *gin.Context) {
	// Get user info from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Authentication required",
		})
		return
	}

	// Get tables created by user
	createdTables := h.tableManager.ListTables(map[string]interface{}{
		"created_by": userID.(string),
	})

	// Get all tables and filter where user is playing/observing
	allTables := h.tableManager.ListTables(map[string]interface{}{})
	var joinedTables []*game.GameTable
	
	playerIDStr := userID.(string)
	for _, table := range allTables {
		if table.IsPlayerAtTable(playerIDStr) || table.IsObserver(playerIDStr) {
			joinedTables = append(joinedTables, table)
		}
	}

	// Convert to public info
	var createdList []map[string]interface{}
	for _, table := range createdTables {
		createdList = append(createdList, table.GetTableInfo())
	}

	var joinedList []map[string]interface{}
	for _, table := range joinedTables {
		joinedList = append(joinedList, table.GetDetailedInfo())
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"created": createdList,
			"joined":  joinedList,
		},
	})
}

// Database operations for persistent storage

// SaveTableToDB saves a table to the database
func (h *TableHandler) SaveTableToDB(table *game.GameTable) error {
	// Convert game.GameTable to models.GameTable
	dbTable := &models.GameTable{
		ID:               table.ID,
		Name:             table.Name,
		GameType:         string(table.GameType),
		Status:           string(table.Status),
		MaxPlayers:       table.MaxPlayers,
		MinPlayers:       table.MinPlayers,
		Description:      table.Description,
		RoomID:           table.RoomID,
		CurrentPlayers:   table.GetPlayerCount(),
		CurrentObservers: table.GetObserverCount(),
		CreatedAt:        table.CreatedAt,
		UpdatedAt:        table.UpdatedAt,
	}

	// Convert CreatedBy string to uint
	if createdByID, err := strconv.ParseUint(table.CreatedBy, 10, 32); err == nil {
		dbTable.CreatedBy = uint(createdByID)
	}

	// Serialize settings and tags as JSON
	// TODO: Implement JSON serialization for settings and tags

	return h.db.Save(dbTable).Error
}

// LoadTableFromDB loads a table from the database
func (h *TableHandler) LoadTableFromDB(tableID string) (*models.GameTable, error) {
	var dbTable models.GameTable
	err := h.db.Preload("Creator").Preload("TablePlayers.User").Preload("TableObservers.User").First(&dbTable, "id = ?", tableID).Error
	return &dbTable, err
}