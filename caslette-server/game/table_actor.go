package game

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TableCommand represents a command sent to a table actor
type TableCommand interface {
	Execute(table *GameTable) interface{}
}

// JoinPlayerCommand represents a player joining request
type JoinPlayerCommand struct {
	PlayerID string
	Username string
	Position int
	Response chan interface{}
}

func (cmd *JoinPlayerCommand) Execute(table *GameTable) interface{} {
	// All the join logic here, no locks needed since only one goroutine accesses table
	if table.Status != TableStatusWaiting && table.Status != TableStatusPaused {
		return &TableError{"TABLE_NOT_JOINABLE", "Table is not in a joinable state"}
	}

	// Check if player already at table
	for _, slot := range table.PlayerSlots {
		if slot.PlayerID == cmd.PlayerID {
			return &TableError{"PLAYER_ALREADY_AT_TABLE", "Player is already at this table"}
		}
	}

	// Find available position
	position := cmd.Position
	if position <= 0 { // Use <= 0 for auto-assign (position 0 or negative)
		// Auto-assign to first available slot
		position = -1 // Reset to indicate not found yet
		for _, slot := range table.PlayerSlots {
			if slot.PlayerID == "" {
				position = slot.Position
				break
			}
		}
		if position == -1 {
			return &TableError{"TABLE_FULL", "No available positions"}
		}
	} else {
		// Check if requested position is available (1-based from client, 0-based internally)
		adjustedPos := position - 1
		if adjustedPos < 0 || adjustedPos >= len(table.PlayerSlots) {
			return &TableError{"INVALID_POSITION", "Invalid position"}
		}
		if table.PlayerSlots[adjustedPos].PlayerID != "" {
			return &TableError{"POSITION_OCCUPIED", "Position is already occupied"}
		}
		position = adjustedPos // Use 0-based position internally
	}

	// Add player
	for i := range table.PlayerSlots {
		if table.PlayerSlots[i].Position == position {
			table.PlayerSlots[i] = PlayerSlot{
				Position: position,
				PlayerID: cmd.PlayerID,
				Username: cmd.Username,
				IsReady:  false,
				JoinedAt: time.Now(),
			}
			break
		}
	}

	table.UpdatedAt = time.Now()
	return nil // Success
}

// LeavePlayerCommand represents a player leaving request
type LeavePlayerCommand struct {
	PlayerID string
	Response chan interface{}
}

func (cmd *LeavePlayerCommand) Execute(table *GameTable) interface{} {
	// Find and remove player
	found := false
	for i := range table.PlayerSlots {
		if table.PlayerSlots[i].PlayerID == cmd.PlayerID {
			table.PlayerSlots[i] = PlayerSlot{Position: table.PlayerSlots[i].Position}
			found = true
			break
		}
	}

	if !found {
		return &TableError{"PLAYER_NOT_AT_TABLE", "Player is not at this table"}
	}

	table.UpdatedAt = time.Now()
	return nil
}

// JoinObserverCommand represents an observer joining request
type JoinObserverCommand struct {
	PlayerID string
	Username string
	Response chan interface{}
}

func (cmd *JoinObserverCommand) Execute(table *GameTable) interface{} {
	// Check if observers are allowed
	if !table.Settings.ObserversAllowed {
		return &TableError{"OBSERVERS_NOT_ALLOWED", "Observers are not allowed at this table"}
	}

	// Check if table is closed
	if table.Status == TableStatusClosed {
		return &TableError{"TABLE_CLOSED", "Table is closed"}
	}

	// Check if already observing or playing
	for _, slot := range table.PlayerSlots {
		if slot.PlayerID == cmd.PlayerID {
			return &TableError{"PLAYER_ALREADY_AT_TABLE", "Player is already at this table"}
		}
	}

	for _, observer := range table.Observers {
		if observer.PlayerID == cmd.PlayerID {
			return &TableError{"PLAYER_ALREADY_OBSERVING", "Player is already observing this table"}
		}
	}

	// Add to observers
	observer := TableObserver{
		PlayerID: cmd.PlayerID,
		Username: cmd.Username,
		JoinedAt: time.Now(),
	}
	table.Observers = append(table.Observers, observer)
	table.UpdatedAt = time.Now()
	return nil
}

// GetTableInfoCommand represents a request for table information
type GetTableInfoCommand struct {
	Response chan interface{}
}

func (cmd *GetTableInfoCommand) Execute(table *GameTable) interface{} {
	// Count players
	playerCount := 0
	for _, slot := range table.PlayerSlots {
		if slot.PlayerID != "" {
			playerCount++
		}
	}

	return map[string]interface{}{
		"id":             table.ID,
		"name":           table.Name,
		"game_type":      table.GameType,
		"status":         table.Status,
		"created_by":     table.CreatedBy,
		"created_at":     table.CreatedAt,
		"updated_at":     table.UpdatedAt,
		"max_players":    table.MaxPlayers,
		"min_players":    table.MinPlayers,
		"player_count":   playerCount,
		"observer_count": len(table.Observers),
		"settings":       table.Settings,
		"description":    table.Description,
		"tags":           table.Tags,
		"room_id":        table.RoomID,
	}
}

// TableActor manages a single table's state through message passing
type TableActor struct {
	table    *GameTable
	commands chan TableCommand
	quit     chan struct{}
	wg       sync.WaitGroup
}

// NewTableActor creates a new table actor
func NewTableActor(table *GameTable) *TableActor {
	actor := &TableActor{
		table:    table,
		commands: make(chan TableCommand, 100), // Buffered channel for commands
		quit:     make(chan struct{}),
	}

	actor.wg.Add(1)
	go actor.run()

	return actor
}

// run is the main loop of the table actor
func (ta *TableActor) run() {
	defer ta.wg.Done()

	for {
		select {
		case cmd := <-ta.commands:
			result := cmd.Execute(ta.table)

			// Send response back if the command has a response channel
			switch typedCmd := cmd.(type) {
			case *JoinPlayerCommand:
				typedCmd.Response <- result
			case *JoinObserverCommand:
				typedCmd.Response <- result
			case *LeavePlayerCommand:
				typedCmd.Response <- result
			case *GetTableInfoCommand:
				typedCmd.Response <- result
			}

		case <-ta.quit:
			return
		}
	}
}

// JoinPlayer sends a join command to the table actor
func (ta *TableActor) JoinPlayer(ctx context.Context, playerID, username string, position int) error {
	cmd := &JoinPlayerCommand{
		PlayerID: playerID,
		Username: username,
		Position: position,
		Response: make(chan interface{}, 1),
	}

	select {
	case ta.commands <- cmd:
		// Command sent successfully
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case result := <-cmd.Response:
		if err, ok := result.(*TableError); ok {
			return err
		}
		return nil // Success
	case <-ctx.Done():
		return ctx.Err()
	}
}

// LeavePlayer sends a leave command to the table actor
func (ta *TableActor) LeavePlayer(ctx context.Context, playerID string) error {
	cmd := &LeavePlayerCommand{
		PlayerID: playerID,
		Response: make(chan interface{}, 1),
	}

	select {
	case ta.commands <- cmd:
		// Command sent successfully
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case result := <-cmd.Response:
		if err, ok := result.(*TableError); ok {
			return err
		}
		return nil // Success
	case <-ctx.Done():
		return ctx.Err()
	}
}

// JoinObserver sends a join observer command to the table actor
func (ta *TableActor) JoinObserver(ctx context.Context, playerID, username string) error {
	cmd := &JoinObserverCommand{
		PlayerID: playerID,
		Username: username,
		Response: make(chan interface{}, 1),
	}

	select {
	case ta.commands <- cmd:
		// Command sent successfully
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case result := <-cmd.Response:
		if err, ok := result.(*TableError); ok {
			return err
		}
		return nil // Success
	case <-ctx.Done():
		return ctx.Err()
	}
}

// GetTableInfo gets table information via the actor
func (ta *TableActor) GetTableInfo(ctx context.Context) (map[string]interface{}, error) {
	cmd := &GetTableInfoCommand{
		Response: make(chan interface{}, 1),
	}

	select {
	case ta.commands <- cmd:
		// Command sent successfully
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case result := <-cmd.Response:
		if info, ok := result.(map[string]interface{}); ok {
			return info, nil
		}
		return nil, fmt.Errorf("unexpected response type")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Stop gracefully stops the table actor
func (ta *TableActor) Stop() {
	close(ta.quit)
	ta.wg.Wait()
}
