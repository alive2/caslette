package game

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

// TableJoinMode defines how a player joins the table
type TableJoinMode string

const (
	JoinModePlayer   TableJoinMode = "player"
	JoinModeObserver TableJoinMode = "observer"
)

// TableError represents table operation errors
type TableError struct {
	Code    string
	Message string
}

func (e *TableError) Error() string {
	return fmt.Sprintf("table error [%s]: %s", e.Code, e.Message)
}

// Common table error codes
var (
	ErrTableNotFound        = &TableError{"TABLE_NOT_FOUND", "Table not found"}
	ErrTableFull            = &TableError{"TABLE_FULL", "Table is full"}
	ErrPlayerNotAtTable     = &TableError{"PLAYER_NOT_AT_TABLE", "Player is not at this table"}
	ErrPlayerAlreadyAtTable = &TableError{"PLAYER_ALREADY_AT_TABLE", "Player is already at this table"}
	ErrTableNotJoinable     = &TableError{"TABLE_NOT_JOINABLE", "Table is not in a joinable state"}
	ErrObserversNotAllowed  = &TableError{"OBSERVERS_NOT_ALLOWED", "Observers are not allowed at this table"}
	ErrInvalidPosition      = &TableError{"INVALID_POSITION", "Invalid table position"}
	ErrGameInProgress       = &TableError{"GAME_IN_PROGRESS", "Cannot perform action while game is in progress"}
	ErrInvalidPassword      = &TableError{"INVALID_PASSWORD", "Invalid table password"}
	ErrNotTableCreator      = &TableError{"NOT_TABLE_CREATOR", "Only table creator can perform this action"}
)

// TableJoinRequest represents a request to join a table
type TableJoinRequest struct {
	TableID  string        `json:"table_id"`
	PlayerID string        `json:"player_id"`
	Username string        `json:"username"`
	Mode     TableJoinMode `json:"mode"`               // player or observer
	Position int           `json:"position,omitempty"` // specific position (optional)
	Password string        `json:"password,omitempty"` // for private tables
}

// TableLeaveRequest represents a request to leave a table
type TableLeaveRequest struct {
	TableID  string `json:"table_id"`
	PlayerID string `json:"player_id"`
}

// TableCreateRequest represents a request to create a new table
type TableCreateRequest struct {
	Name        string        `json:"name"`
	GameType    GameType      `json:"game_type"`
	CreatedBy   string        `json:"created_by"`
	Username    string        `json:"username"`
	Settings    TableSettings `json:"settings"`
	Description string        `json:"description,omitempty"`
	Tags        []string      `json:"tags,omitempty"`
}

// TableManager manages all game tables
type TableManager struct {
	tables          map[string]*GameTable
	mutex           sync.RWMutex
	engineFactory   GameEngineFactory
	webhookHandlers []TableWebhookHandler
	validator       *TableValidator
	rateLimiter     *RateLimiter
	dataFilter      *DataFilter
	auditor         *SecurityAuditor
}

// TableWebhookHandler defines callbacks for table events
type TableWebhookHandler interface {
	OnTableCreated(table *GameTable)
	OnTableClosed(table *GameTable)
	OnPlayerJoined(table *GameTable, playerID, username string, mode TableJoinMode)
	OnPlayerLeft(table *GameTable, playerID string, mode TableJoinMode)
	OnGameStarted(table *GameTable)
	OnGameFinished(table *GameTable)
}

// GameEngineFactory creates game engines for specific game types
type GameEngineFactory interface {
	CreateEngine(gameType GameType, settings TableSettings) (GameEngine, error)
}

// NewTableManager creates a new table manager
func NewTableManager(engineFactory GameEngineFactory) *TableManager {
	return &TableManager{
		tables:          make(map[string]*GameTable),
		engineFactory:   engineFactory,
		webhookHandlers: make([]TableWebhookHandler, 0),
		validator:       NewTableValidator(),
		rateLimiter:     NewRateLimiter(),
		dataFilter:      NewDataFilter(),
		auditor:         NewSecurityAuditor(),
	}
}

// AddWebhookHandler adds a webhook handler for table events
func (tm *TableManager) AddWebhookHandler(handler TableWebhookHandler) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tm.webhookHandlers = append(tm.webhookHandlers, handler)
}

// generateTableID generates a unique table ID
func (tm *TableManager) generateTableID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateTable creates a new game table
func (tm *TableManager) CreateTable(ctx context.Context, req *TableCreateRequest) (*GameTable, error) {
	// Validate the request first
	if err := tm.validator.ValidateTableCreateRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check rate limits
	if err := tm.rateLimiter.CanCreateTable(req.CreatedBy); err != nil {
		return nil, err
	}

	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Sanitize inputs
	req.Name = tm.validator.SanitizeInput(req.Name)
	req.Description = tm.validator.SanitizeInput(req.Description)
	req.Settings.Password = tm.validator.SanitizeInput(req.Settings.Password)

	// Generate unique ID
	tableID := tm.generateTableID()

	// Create table
	table := NewGameTable(tableID, req.Name, req.GameType, req.CreatedBy, req.Settings)
	table.Description = req.Description
	table.Tags = req.Tags

	// Create game engine if factory is available
	if tm.engineFactory != nil {
		engine, err := tm.engineFactory.CreateEngine(req.GameType, req.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to create game engine: %w", err)
		}
		table.GameEngine = engine
	}

	// Store table
	tm.tables[tableID] = table

	// Record table creation for rate limiting
	tm.rateLimiter.RecordTableCreated(req.CreatedBy, tableID)

	// Notify webhook handlers
	for _, handler := range tm.webhookHandlers {
		go handler.OnTableCreated(table)
	}

	return table, nil
}

// GetTable retrieves a table by ID with security filtering
func (tm *TableManager) GetTable(tableID string) (*GameTable, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	table, exists := tm.tables[tableID]
	if !exists {
		return nil, ErrTableNotFound
	}

	return table, nil
}

// GetTableInfo retrieves filtered table information for a specific user
func (tm *TableManager) GetTableInfo(tableID string, requesterID string) (map[string]interface{}, error) {
	table, err := tm.GetTable(tableID)
	if err != nil {
		tm.auditor.LogAction(requesterID, tableID, "get_table_info", "failed", err.Error())
		return nil, err
	}

	// Validate access
	if err := tm.dataFilter.ValidateTableAccess(table, requesterID, "view"); err != nil {
		tm.auditor.LogAction(requesterID, tableID, "get_table_info", "access_denied", err.Error())
		return nil, err
	}

	tm.auditor.LogAction(requesterID, tableID, "get_table_info", "success", "")
	return tm.dataFilter.FilterTableInfo(table, requesterID, ""), nil
}

// ListTables returns all tables with optional filtering
func (tm *TableManager) ListTables(filters map[string]interface{}) []*GameTable {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var result []*GameTable

	for _, table := range tm.tables {
		if tm.matchesFilters(table, filters) {
			result = append(result, table)
		}
	}

	// Sort by created time (newest first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result
}

// matchesFilters checks if a table matches the given filters
func (tm *TableManager) matchesFilters(table *GameTable, filters map[string]interface{}) bool {
	if gameType, ok := filters["game_type"]; ok {
		if table.GameType != GameType(gameType.(string)) {
			return false
		}
	}

	if status, ok := filters["status"]; ok {
		if table.Status != TableStatus(status.(string)) {
			return false
		}
	}

	if createdBy, ok := filters["created_by"]; ok {
		if table.CreatedBy != createdBy.(string) {
			return false
		}
	}

	if hasSpace, ok := filters["has_space"]; ok {
		if hasSpace.(bool) && table.GetPlayerCount() >= table.MaxPlayers {
			return false
		}
	}

	if observersAllowed, ok := filters["observers_allowed"]; ok {
		if table.Settings.ObserversAllowed != observersAllowed.(bool) {
			return false
		}
	}

	return true
}

// JoinTable handles a player joining a table
func (tm *TableManager) JoinTable(ctx context.Context, req *TableJoinRequest) error {
	// Validate the request first
	if err := tm.validator.ValidateTableJoinRequest(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Check rate limits for joins
	if err := tm.rateLimiter.CanJoinTable(req.PlayerID, req.TableID); err != nil {
		return err
	}

	// For observers, check observer-specific limits
	if req.Mode == JoinModeObserver {
		if err := tm.rateLimiter.CanObserveTable(req.PlayerID, req.TableID); err != nil {
			return err
		}
	}

	table, err := tm.GetTable(req.TableID)
	if err != nil {
		return err
	}

	table.mutex.Lock()
	defer table.mutex.Unlock()

	// Check password for private tables
	if table.Settings.Private && table.Settings.Password != "" {
		if req.Password != table.Settings.Password {
			return ErrInvalidPassword
		}
	}

	switch req.Mode {
	case JoinModePlayer:
		return tm.joinAsPlayer(table, req)
	case JoinModeObserver:
		return tm.joinAsObserver(table, req)
	default:
		return &TableError{"INVALID_JOIN_MODE", "Invalid join mode"}
	}
}

// joinAsPlayer handles joining as a player
func (tm *TableManager) joinAsPlayer(table *GameTable, req *TableJoinRequest) error {
	// Check if can join as player
	if !table.CanJoinAsPlayer(req.PlayerID) {
		if table.IsPlayerAtTable(req.PlayerID) {
			return ErrPlayerAlreadyAtTable
		}
		if table.GetPlayerCount() >= table.MaxPlayers {
			return ErrTableFull
		}
		return ErrTableNotJoinable
	}

	// Find position
	position := req.Position
	if position == 0 { // Auto-assign position
		availableSlots := table.GetAvailableSlots()
		if len(availableSlots) == 0 {
			return ErrTableFull
		}
		position = availableSlots[0]
	} else {
		// Check if requested position is available
		if position < 0 || position >= len(table.PlayerSlots) {
			return ErrInvalidPosition
		}
		if table.PlayerSlots[position].PlayerID != "" {
			return &TableError{"POSITION_TAKEN", "Requested position is already taken"}
		}
	}

	// Add player to slot
	table.PlayerSlots[position] = PlayerSlot{
		Position: position,
		PlayerID: req.PlayerID,
		Username: req.Username,
		IsReady:  false,
		JoinedAt: time.Now(),
	}

	// Remove from observers if was observing
	tm.removeObserver(table, req.PlayerID)

	table.Touch()

	// Notify webhook handlers
	for _, handler := range tm.webhookHandlers {
		go handler.OnPlayerJoined(table, req.PlayerID, req.Username, JoinModePlayer)
	}

	// Check if game should auto-start
	if table.Settings.AutoStart && table.GetPlayerCount() >= table.MinPlayers {
		tm.tryStartGame(table)
	}

	return nil
}

// joinAsObserver handles joining as an observer
func (tm *TableManager) joinAsObserver(table *GameTable, req *TableJoinRequest) error {
	// Check if can join as observer
	if !table.CanJoinAsObserver(req.PlayerID) {
		if !table.Settings.ObserversAllowed {
			return ErrObserversNotAllowed
		}
		if table.IsPlayerAtTable(req.PlayerID) || table.IsObserver(req.PlayerID) {
			return ErrPlayerAlreadyAtTable
		}
		return ErrTableNotJoinable
	}

	// Add to observers
	observer := TableObserver{
		PlayerID: req.PlayerID,
		Username: req.Username,
		JoinedAt: time.Now(),
	}
	table.Observers = append(table.Observers, observer)
	table.Touch()

	// Record in rate limiter
	tm.rateLimiter.RecordObserverJoined(req.PlayerID, req.TableID)

	// Notify webhook handlers
	for _, handler := range tm.webhookHandlers {
		go handler.OnPlayerJoined(table, req.PlayerID, req.Username, JoinModeObserver)
	}

	return nil
}

// LeaveTable handles a player leaving a table
func (tm *TableManager) LeaveTable(ctx context.Context, req *TableLeaveRequest) error {
	table, err := tm.GetTable(req.TableID)
	if err != nil {
		return err
	}

	table.mutex.Lock()
	defer table.mutex.Unlock()

	// Check if player is at table
	if !table.IsPlayerAtTable(req.PlayerID) && !table.IsObserver(req.PlayerID) {
		return ErrPlayerNotAtTable
	}

	// Remove from player slots
	var mode TableJoinMode
	for i, slot := range table.PlayerSlots {
		if slot.PlayerID == req.PlayerID {
			table.PlayerSlots[i] = PlayerSlot{Position: i}
			mode = JoinModePlayer
			break
		}
	}

	// Remove from observers
	if mode == "" && tm.removeObserver(table, req.PlayerID) {
		mode = JoinModeObserver
	}

	table.Touch()

	// Notify webhook handlers
	for _, handler := range tm.webhookHandlers {
		go handler.OnPlayerLeft(table, req.PlayerID, mode)
	}

	// Check if table should be closed (no players left)
	if table.GetPlayerCount() == 0 && table.GetObserverCount() == 0 {
		tm.closeTable(table)
	}

	return nil
}

// removeObserver removes a player from observers list
func (tm *TableManager) removeObserver(table *GameTable, playerID string) bool {
	for i, observer := range table.Observers {
		if observer.PlayerID == playerID {
			table.Observers = append(table.Observers[:i], table.Observers[i+1:]...)
			return true
		}
	}
	return false
}

// tryStartGame attempts to start the game if conditions are met
func (tm *TableManager) tryStartGame(table *GameTable) error {
	if table.Status != TableStatusWaiting {
		return nil
	}

	if table.GetPlayerCount() < table.MinPlayers {
		return nil
	}

	// Initialize game engine if available
	if table.GameEngine != nil {
		// Add players to game engine
		for _, slot := range table.PlayerSlots {
			if slot.PlayerID != "" {
				player := &Player{
					ID:       slot.PlayerID,
					Name:     slot.Username,
					Position: slot.Position,
					IsActive: true,
				}
				// Initialize player data for buy-in chips
				player.Data = map[string]interface{}{
					"chips": table.Settings.BuyIn,
				}
				table.GameEngine.AddPlayer(player)
			}
		}

		// Start the game
		if err := table.GameEngine.Start(); err != nil {
			return fmt.Errorf("failed to start game: %w", err)
		}
	}

	table.Status = TableStatusActive
	table.Touch()

	// Notify webhook handlers
	for _, handler := range tm.webhookHandlers {
		go handler.OnGameStarted(table)
	}

	return nil
}

// CloseTable closes a table and removes it from the manager
func (tm *TableManager) CloseTable(tableID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	table, exists := tm.tables[tableID]
	if !exists {
		return ErrTableNotFound
	}

	tm.closeTable(table)
	return nil
}

// closeTable closes a table (internal method, assumes lock is held)
func (tm *TableManager) closeTable(table *GameTable) {
	table.Status = TableStatusClosed
	table.Touch()

	// Notify webhook handlers
	for _, handler := range tm.webhookHandlers {
		go handler.OnTableClosed(table)
	}

	// Remove from tables map
	delete(tm.tables, table.ID)
}

// GetTableCount returns the total number of active tables
func (tm *TableManager) GetTableCount() int {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	return len(tm.tables)
}

// GetStats returns overall table statistics
func (tm *TableManager) GetStats() map[string]interface{} {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	totalTables := len(tm.tables)
	totalPlayers := 0
	totalObservers := 0
	statusCounts := make(map[TableStatus]int)
	gameTypeCounts := make(map[GameType]int)

	for _, table := range tm.tables {
		totalPlayers += table.GetPlayerCount()
		totalObservers += table.GetObserverCount()
		statusCounts[table.Status]++
		gameTypeCounts[table.GameType]++
	}

	return map[string]interface{}{
		"total_tables":     totalTables,
		"total_players":    totalPlayers,
		"total_observers":  totalObservers,
		"status_counts":    statusCounts,
		"game_type_counts": gameTypeCounts,
	}
}
