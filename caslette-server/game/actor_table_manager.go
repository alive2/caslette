package game

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
)

// ActorTableManager manages tables using the actor pattern
type ActorTableManager struct {
	actors            map[string]*TableActor
	gameEngineFactory GameEngineFactory
	rateLimiter       *ActorRateLimiter
	validator         *TableValidator
	mu                sync.RWMutex // Protects the actors map only
}

// NewActorTableManager creates a new actor-based table manager
func NewActorTableManager(factory GameEngineFactory) *ActorTableManager {
	return &ActorTableManager{
		actors:            make(map[string]*TableActor),
		gameEngineFactory: factory,
		rateLimiter:       NewActorRateLimiter(),
		validator:         NewTableValidator(),
	}
}

// generateTableID generates a unique table ID
func (tm *ActorTableManager) generateTableID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// CreateTable creates a new table with an actor
func (tm *ActorTableManager) CreateTable(ctx context.Context, req *TableCreateRequest) (*GameTable, error) {
	// Check rate limits first
	if err := tm.rateLimiter.CanCreateTable(req.CreatedBy); err != nil {
		return nil, err
	}

	// Validate request
	if err := tm.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Generate table ID
	tableID := tm.generateTableID()

	// Create the table using the existing NewGameTable function
	table := NewGameTable(tableID, req.Name, req.GameType, req.CreatedBy, req.Settings)
	table.Description = req.Description
	table.Tags = req.Tags

	// Create game engine
	if tm.gameEngineFactory != nil {
		engine, err := tm.gameEngineFactory.CreateEngine(req.GameType, req.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to create game engine: %w", err)
		}
		table.GameEngine = engine
	}

	// Create actor for this table
	actor := NewTableActor(table)

	tm.mu.Lock()
	tm.actors[table.ID] = actor
	tm.mu.Unlock()

	return table, nil
}

// JoinTable handles a player joining a table
func (tm *ActorTableManager) JoinTable(ctx context.Context, req *TableJoinRequest) error {
	// Rate limiting check
	if err := tm.rateLimiter.CanJoinTable(req.PlayerID, req.TableID); err != nil {
		return err
	}

	// Get table actor
	tm.mu.RLock()
	actor, exists := tm.actors[req.TableID]
	tm.mu.RUnlock()

	if !exists {
		return ErrTableNotFound
	}

	// Check password for private tables
	table, err := tm.GetTable(req.TableID)
	if err != nil {
		return err
	}

	if table.Settings.Private && table.Settings.Password != "" {
		if req.Password != table.Settings.Password {
			return &TableError{"INVALID_PASSWORD", "Incorrect password for private table"}
		}
	}

	// Send command to actor based on join mode
	switch req.Mode {
	case JoinModePlayer:
		return actor.JoinPlayer(ctx, req.PlayerID, req.Username, req.Position)
	case JoinModeObserver:
		return actor.JoinObserver(ctx, req.PlayerID, req.Username)
	default:
		return &TableError{"INVALID_JOIN_MODE", "Invalid join mode"}
	}
}

// LeaveTable handles a player leaving a table
func (tm *ActorTableManager) LeaveTable(ctx context.Context, req *TableLeaveRequest) error {
	// Get table actor
	tm.mu.RLock()
	actor, exists := tm.actors[req.TableID]
	tm.mu.RUnlock()

	if !exists {
		return ErrTableNotFound
	}

	return actor.LeavePlayer(ctx, req.PlayerID)
}

// GetTable returns table information
func (tm *ActorTableManager) GetTable(tableID string) (*GameTable, error) {
	tm.mu.RLock()
	actor, exists := tm.actors[tableID]
	tm.mu.RUnlock()

	if !exists {
		return nil, ErrTableNotFound
	}

	// Return the table directly from the actor
	// Note: This breaks encapsulation a bit, but is needed for compatibility
	// In a full implementation, you might want to return info differently
	return actor.table, nil
}

// GetTables returns all tables (for listing)
func (tm *ActorTableManager) GetTables() []*GameTable {
	tm.mu.RLock()
	tables := make([]*GameTable, 0, len(tm.actors))
	for _, actor := range tm.actors {
		tables = append(tables, actor.table)
	}
	tm.mu.RUnlock()
	return tables
}

// CloseTable closes a table and stops its actor
func (tm *ActorTableManager) CloseTable(tableID string) error {
	tm.mu.Lock()
	actor, exists := tm.actors[tableID]
	if !exists {
		tm.mu.Unlock()
		return ErrTableNotFound
	}

	// Stop the actor
	actor.Stop()

	// Remove from map
	delete(tm.actors, tableID)
	tm.mu.Unlock()

	return nil
}

// Stop gracefully stops all table actors
func (tm *ActorTableManager) Stop() {
	tm.mu.Lock()
	for _, actor := range tm.actors {
		actor.Stop()
	}
	tm.actors = make(map[string]*TableActor)
	tm.mu.Unlock()
}

func (tm *ActorTableManager) validateCreateRequest(req *TableCreateRequest) error {
	// Use the table validator for comprehensive validation
	return tm.validator.ValidateTableCreateRequest(req)
}

// ListTables returns a filtered list of tables
func (tm *ActorTableManager) ListTables(filters map[string]interface{}) []*GameTable {
	tables := tm.GetTables()

	if len(filters) == 0 {
		return tables
	}

	var filteredTables []*GameTable

	for _, table := range tables {
		matchesFilter := true

		// Check game_type filter
		if gameType, exists := filters["game_type"]; exists {
			if gameTypeStr, ok := gameType.(string); ok {
				if string(table.GameType) != gameTypeStr {
					matchesFilter = false
				}
			}
		}

		// Check created_by filter
		if createdBy, exists := filters["created_by"]; exists {
			if createdByStr, ok := createdBy.(string); ok {
				if table.CreatedBy != createdByStr {
					matchesFilter = false
				}
			}
		}

		// Check observers_allowed filter
		if observersAllowed, exists := filters["observers_allowed"]; exists {
			if observersAllowedBool, ok := observersAllowed.(bool); ok {
				if table.Settings.ObserversAllowed != observersAllowedBool {
					matchesFilter = false
				}
			}
		}

		if matchesFilter {
			filteredTables = append(filteredTables, table)
		}
	}

	return filteredTables
}

// GetStats returns statistics about the table manager
func (tm *ActorTableManager) GetStats() map[string]interface{} {
	tables := tm.GetTables()

	stats := map[string]interface{}{
		"total_tables":    len(tables),
		"active_tables":   0,
		"total_players":   0,
		"total_observers": 0,
	}

	for _, table := range tables {
		if table.Status == TableStatusActive {
			stats["active_tables"] = stats["active_tables"].(int) + 1
		}
		stats["total_players"] = stats["total_players"].(int) + table.GetPlayerCount()
		stats["total_observers"] = stats["total_observers"].(int) + table.GetObserverCount()
	}

	return stats
}

// AddWebhookHandler adds a webhook handler for table events
func (tm *ActorTableManager) AddWebhookHandler(handler interface{}) {
	// For now, this is a no-op to maintain compatibility
	// In the future, could implement event broadcasting if needed
}

// tryStartGame attempts to start a game on the given table
func (tm *ActorTableManager) tryStartGame(table *GameTable) error {
	// Check if game engine is available and can start
	if table.GameEngine == nil {
		return &TableError{"NO_ENGINE", "No game engine available"}
	}

	// Check if enough players
	if table.GetPlayerCount() < table.MinPlayers {
		return &TableError{"NOT_ENOUGH_PLAYERS", "Not enough players to start game"}
	}

	// Update table status to active
	table.Status = TableStatusActive

	// Initialize the game engine if needed
	// The actual game start logic would be handled by the game engine
	return nil
}

// GetTableCount returns the number of tables
func (tm *ActorTableManager) GetTableCount() int {
	return len(tm.GetTables())
}

// GetTableInfo returns information about a specific table
func (tm *ActorTableManager) GetTableInfo(tableID string, userID string) (map[string]interface{}, error) {
	table, err := tm.GetTable(tableID)
	if err != nil {
		return nil, err
	}

	// Check access permissions for private tables
	if table.Settings.Private {
		// Only the creator and players/observers at the table can access private table info
		if table.CreatedBy != userID && !table.IsPlayerAtTable(userID) && !table.IsObserver(userID) {
			return nil, &TableError{"ACCESS_DENIED", "Access denied to private table information"}
		}
	}

	// Return the table info
	return table.GetTableInfo(), nil
}
