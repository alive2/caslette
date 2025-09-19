package game

import (
	"time"
)

// DataFilter handles secure data filtering for table information
type DataFilter struct{}

// NewDataFilter creates a new data filter
func NewDataFilter() *DataFilter {
	return &DataFilter{}
}

// FilterTableInfo returns filtered table information based on user context
func (df *DataFilter) FilterTableInfo(table *GameTable, requesterID string, requesterRole string) map[string]interface{} {
	// Base public information available to everyone
	info := map[string]interface{}{
		"id":             table.ID,
		"name":           table.Name,
		"game_type":      table.GameType,
		"status":         table.Status,
		"created_at":     table.CreatedAt,
		"updated_at":     table.UpdatedAt,
		"max_players":    table.MaxPlayers,
		"min_players":    table.MinPlayers,
		"player_count":   table.GetPlayerCount(),
		"observer_count": table.GetObserverCount(),
		"description":    table.Description,
		"tags":           table.Tags,
	}

	// Check if user is at the table
	isPlayer := table.IsPlayerAtTable(requesterID)
	isObserver := table.IsObserver(requesterID)
	isCreator := table.CreatedBy == requesterID
	isAtTable := isPlayer || isObserver

	// Add filtered settings (hide sensitive data like passwords)
	filteredSettings := df.filterTableSettings(table.Settings, isAtTable || isCreator)
	info["settings"] = filteredSettings

	// Add creator info only if public or user is at table
	if !table.Settings.Private || isAtTable || isCreator {
		info["created_by"] = table.CreatedBy
	}

	// Add room ID only for users at the table
	if isAtTable || isCreator {
		info["room_id"] = table.RoomID
	}

	// Add detailed player information only for users at the table
	if isAtTable || isCreator {
		info["player_slots"] = df.filterPlayerSlots(table.PlayerSlots, requesterID, isPlayer)
		info["observers"] = df.filterObservers(table.Observers, requesterID, isObserver)
	} else {
		// For non-participants, only show occupied slots count
		occupiedSlots := 0
		for _, slot := range table.PlayerSlots {
			if slot.PlayerID != "" {
				occupiedSlots++
			}
		}
		info["occupied_slots"] = occupiedSlots
	}

	return info
}

// FilterTableList returns filtered table list for browsing
func (df *DataFilter) FilterTableList(tables []*GameTable, requesterID string) []map[string]interface{} {
	var result []map[string]interface{}

	for _, table := range tables {
		// For table browsing, provide minimal public info
		tableInfo := map[string]interface{}{
			"id":           table.ID,
			"name":         table.Name,
			"game_type":    table.GameType,
			"status":       table.Status,
			"created_at":   table.CreatedAt,
			"max_players":  table.MaxPlayers,
			"player_count": table.GetPlayerCount(),
			"description":  table.Description,
			"tags":         table.Tags,
		}

		// Add buy-in info for public tables or if user is at table
		isAtTable := table.IsPlayerAtTable(requesterID) || table.IsObserver(requesterID)
		isCreator := table.CreatedBy == requesterID

		if !table.Settings.Private || isAtTable || isCreator {
			tableInfo["buy_in"] = table.Settings.BuyIn
			tableInfo["small_blind"] = table.Settings.SmallBlind
			tableInfo["big_blind"] = table.Settings.BigBlind
		}

		// Indicate if table requires password (but don't expose the password)
		if table.Settings.Private && table.Settings.Password != "" {
			tableInfo["requires_password"] = true
		}

		// Show if table has space for more players
		tableInfo["has_space"] = table.GetPlayerCount() < table.MaxPlayers
		tableInfo["observers_allowed"] = table.Settings.ObserversAllowed

		result = append(result, tableInfo)
	}

	return result
}

// filterTableSettings filters table settings based on permissions
func (df *DataFilter) filterTableSettings(settings TableSettings, hasAccess bool) map[string]interface{} {
	filtered := map[string]interface{}{
		"small_blind":       settings.SmallBlind,
		"big_blind":         settings.BigBlind,
		"buy_in":            settings.BuyIn,
		"auto_start":        settings.AutoStart,
		"time_limit":        settings.TimeLimit,
		"observers_allowed": settings.ObserversAllowed,
		"private":           settings.Private,
	}

	// Only add sensitive fields if user has access
	if hasAccess {
		filtered["max_buy_in"] = settings.MaxBuyIn
		filtered["tournament_mode"] = settings.TournamentMode

		// Never expose the actual password, only indicate if one exists
		if settings.Password != "" {
			filtered["has_password"] = true
		}
	}

	return filtered
}

// filterPlayerSlots filters player slot information
func (df *DataFilter) filterPlayerSlots(slots []PlayerSlot, requesterID string, isPlayer bool) []map[string]interface{} {
	var filtered []map[string]interface{}

	for _, slot := range slots {
		slotInfo := map[string]interface{}{
			"position": slot.Position,
		}

		if slot.PlayerID != "" {
			slotInfo["player_id"] = slot.PlayerID
			slotInfo["username"] = slot.Username
			slotInfo["is_ready"] = slot.IsReady

			// Only show join time to the player themselves or other players
			if isPlayer || slot.PlayerID == requesterID {
				slotInfo["joined_at"] = slot.JoinedAt
			}
		}

		filtered = append(filtered, slotInfo)
	}

	return filtered
}

// filterObservers filters observer information
func (df *DataFilter) filterObservers(observers []TableObserver, requesterID string, isObserver bool) []map[string]interface{} {
	var filtered []map[string]interface{}

	for _, observer := range observers {
		observerInfo := map[string]interface{}{
			"player_id": observer.PlayerID,
			"username":  observer.Username,
		}

		// Only show join time to observers themselves or other observers
		if isObserver || observer.PlayerID == requesterID {
			observerInfo["joined_at"] = observer.JoinedAt
		}

		filtered = append(filtered, observerInfo)
	}

	return filtered
}

// FilterGameState returns filtered game state information
func (df *DataFilter) FilterGameState(table *GameTable, requesterID string) map[string]interface{} {
	isPlayer := table.IsPlayerAtTable(requesterID)
	isObserver := table.IsObserver(requesterID)
	isCreator := table.CreatedBy == requesterID

	// Only allow game state access to participants
	if !isPlayer && !isObserver && !isCreator {
		return map[string]interface{}{
			"error": "Access denied: not authorized to view game state",
		}
	}

	gameState := map[string]interface{}{
		"table_id": table.ID,
		"status":   table.Status,
	}

	// Add game-specific state if engine exists
	if table.GameEngine != nil {
		// Get public game state (community cards, pot, etc.)
		publicState := table.GameEngine.GetPublicGameState()
		for key, value := range publicState {
			gameState[key] = value
		}

		// Add private state only for players
		if isPlayer {
			privateState := table.GameEngine.GetPlayerState(requesterID)
			if privateState != nil {
				gameState["player_state"] = privateState
			}
		}
	}

	return gameState
}

// ValidateTableAccess checks if a user can access a table
func (df *DataFilter) ValidateTableAccess(table *GameTable, requesterID string, accessType string) error {
	if table == nil {
		return ErrTableNotFound
	}

	isPlayer := table.IsPlayerAtTable(requesterID)
	isObserver := table.IsObserver(requesterID)
	isCreator := table.CreatedBy == requesterID

	switch accessType {
	case "view":
		// Anyone can view public tables, only participants can view private ones
		if table.Settings.Private && !isPlayer && !isObserver && !isCreator {
			return &TableError{"ACCESS_DENIED", "Access denied to private table"}
		}
	case "join":
		// Basic join validation (additional checks in main join logic)
		if table.Status == TableStatusClosed {
			return &TableError{"TABLE_CLOSED", "Table is closed"}
		}
	case "manage":
		// Only creator can manage table
		if !isCreator {
			return ErrNotTableCreator
		}
	case "game_state":
		// Only participants can access game state
		if !isPlayer && !isObserver && !isCreator {
			return &TableError{"ACCESS_DENIED", "Access denied to game state"}
		}
	default:
		return &TableError{"INVALID_ACCESS_TYPE", "Invalid access type"}
	}

	return nil
}

// AuditLogEntry represents a security audit log entry
type AuditLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	TableID   string    `json:"table_id"`
	Action    string    `json:"action"`
	Result    string    `json:"result"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
	Details   string    `json:"details,omitempty"`
}

// SecurityAuditor handles security audit logging
type SecurityAuditor struct {
	logs []AuditLogEntry
}

// NewSecurityAuditor creates a new security auditor
func NewSecurityAuditor() *SecurityAuditor {
	return &SecurityAuditor{
		logs: make([]AuditLogEntry, 0),
	}
}

// LogAction logs a security-relevant action
func (sa *SecurityAuditor) LogAction(userID, tableID, action, result, details string) {
	entry := AuditLogEntry{
		Timestamp: time.Now(),
		UserID:    userID,
		TableID:   tableID,
		Action:    action,
		Result:    result,
		Details:   details,
	}

	sa.logs = append(sa.logs, entry)

	// In production, this would log to a proper audit system
	// For now, we just keep in memory (not suitable for production)
}

// GetAuditLogs returns recent audit logs (admin only)
func (sa *SecurityAuditor) GetAuditLogs(limit int) []AuditLogEntry {
	if limit <= 0 || limit > len(sa.logs) {
		limit = len(sa.logs)
	}

	// Return most recent entries
	start := len(sa.logs) - limit
	return sa.logs[start:]
}
