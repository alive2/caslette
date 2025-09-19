package game

import (
	"time"
)

// RateLimiterCommand represents commands for the actor-based rate limiter
type RateLimiterCommand interface {
	Execute(rl *RateLimiterState) interface{}
}

// CanCreateTableCommand checks if a user can create a table
type CanCreateTableCommand struct {
	UserID string
	Result chan error
}

func (cmd *CanCreateTableCommand) Execute(rl *RateLimiterState) interface{} {
	// Check if user has reached max tables limit
	if len(rl.getUserState(cmd.UserID).CreatedTables) >= rl.maxTablesPerUser {
		cmd.Result <- &TableError{"RATE_LIMIT_EXCEEDED", "Maximum number of tables reached"}
		return nil
	}

	// Check create attempts in time window
	userState := rl.getUserState(cmd.UserID)
	userState.CreateAttempts = rl.filterRecentAttempts(userState.CreateAttempts, rl.createTableWindow)

	if len(userState.CreateAttempts) >= rl.maxCreatesPerWindow {
		cmd.Result <- &TableError{"RATE_LIMIT_EXCEEDED", "Too many table creation attempts"}
		return nil
	}

	// Record the attempt
	userState.CreateAttempts = append(userState.CreateAttempts, time.Now())
	cmd.Result <- nil
	return nil
}

// CanJoinTableCommand checks if a user can join a table
type CanJoinTableCommand struct {
	UserID  string
	TableID string
	Result  chan error
}

func (cmd *CanJoinTableCommand) Execute(rl *RateLimiterState) interface{} {
	// Check join attempts in time window
	userState := rl.getUserState(cmd.UserID)
	userState.JoinAttempts = rl.filterRecentAttempts(userState.JoinAttempts, rl.joinAttemptWindow)

	if len(userState.JoinAttempts) >= rl.maxJoinsPerWindow {
		cmd.Result <- &TableError{"RATE_LIMIT_EXCEEDED", "Too many join attempts"}
		return nil
	}

	// Record the attempt
	userState.JoinAttempts = append(userState.JoinAttempts, time.Now())
	cmd.Result <- nil
	return nil
}

// CanObserveTableCommand checks if a user can observe a table
type CanObserveTableCommand struct {
	UserID  string
	TableID string
	Result  chan error
}

func (cmd *CanObserveTableCommand) Execute(rl *RateLimiterState) interface{} {
	userState := rl.getUserState(cmd.UserID)
	if len(userState.ObservedTables) >= rl.maxObserverTables {
		cmd.Result <- &TableError{"RATE_LIMIT_EXCEEDED", "Maximum number of observed tables reached"}
		return nil
	}

	cmd.Result <- nil
	return nil
}

// RecordTableCreatedCommand records that a user created a table
type RecordTableCreatedCommand struct {
	UserID  string
	TableID string
}

func (cmd *RecordTableCreatedCommand) Execute(rl *RateLimiterState) interface{} {
	userState := rl.getUserState(cmd.UserID)
	userState.CreatedTables = append(userState.CreatedTables, cmd.TableID)
	userState.LastActivity = time.Now()
	return nil
}

// RecordTableClosedCommand records that a user closed a table
type RecordTableClosedCommand struct {
	UserID  string
	TableID string
}

func (cmd *RecordTableClosedCommand) Execute(rl *RateLimiterState) interface{} {
	userState := rl.getUserState(cmd.UserID)
	userState.CreatedTables = rl.removeFromSlice(userState.CreatedTables, cmd.TableID)
	userState.LastActivity = time.Now()
	return nil
}

// RecordPlayerJoinedCommand records that a user joined a table as player
type RecordPlayerJoinedCommand struct {
	UserID  string
	TableID string
}

func (cmd *RecordPlayerJoinedCommand) Execute(rl *RateLimiterState) interface{} {
	userState := rl.getUserState(cmd.UserID)
	if !rl.containsString(userState.ActiveTables, cmd.TableID) {
		userState.ActiveTables = append(userState.ActiveTables, cmd.TableID)
	}
	userState.LastActivity = time.Now()
	return nil
}

// RecordPlayerLeftCommand records that a user left a table as player
type RecordPlayerLeftCommand struct {
	UserID  string
	TableID string
}

func (cmd *RecordPlayerLeftCommand) Execute(rl *RateLimiterState) interface{} {
	userState := rl.getUserState(cmd.UserID)
	userState.ActiveTables = rl.removeFromSlice(userState.ActiveTables, cmd.TableID)
	userState.LastActivity = time.Now()
	return nil
}

// RecordObserverJoinedCommand records that a user joined a table as observer
type RecordObserverJoinedCommand struct {
	UserID  string
	TableID string
}

func (cmd *RecordObserverJoinedCommand) Execute(rl *RateLimiterState) interface{} {
	userState := rl.getUserState(cmd.UserID)
	if !rl.containsString(userState.ObservedTables, cmd.TableID) {
		userState.ObservedTables = append(userState.ObservedTables, cmd.TableID)
	}
	userState.LastActivity = time.Now()
	return nil
}

// RecordObserverLeftCommand records that a user left a table as observer
type RecordObserverLeftCommand struct {
	UserID  string
	TableID string
}

func (cmd *RecordObserverLeftCommand) Execute(rl *RateLimiterState) interface{} {
	userState := rl.getUserState(cmd.UserID)
	userState.ObservedTables = rl.removeFromSlice(userState.ObservedTables, cmd.TableID)
	userState.LastActivity = time.Now()
	return nil
}

// GetUserStatsCommand gets stats for a user
type GetUserStatsCommand struct {
	UserID string
	Result chan map[string]interface{}
}

func (cmd *GetUserStatsCommand) Execute(rl *RateLimiterState) interface{} {
	userState := rl.getUserState(cmd.UserID)

	// Clean up old attempts
	userState.CreateAttempts = rl.filterRecentAttempts(userState.CreateAttempts, rl.createTableWindow)
	userState.JoinAttempts = rl.filterRecentAttempts(userState.JoinAttempts, rl.joinAttemptWindow)

	stats := map[string]interface{}{
		"tables_created":         len(userState.CreatedTables),
		"max_tables":             rl.maxTablesPerUser,
		"tables_joined":          len(userState.ActiveTables),
		"tables_observing":       len(userState.ObservedTables),
		"max_observer_tables":    rl.maxObserverTables,
		"recent_creates":         len(userState.CreateAttempts),
		"max_creates_per_window": rl.maxCreatesPerWindow,
		"recent_joins":           len(userState.JoinAttempts),
		"max_joins_per_window":   rl.maxJoinsPerWindow,
		"last_activity":          userState.LastActivity,
	}

	cmd.Result <- stats
	return nil
}

// CleanupCommand performs cleanup of old entries
type CleanupCommand struct{}

func (cmd *CleanupCommand) Execute(rl *RateLimiterState) interface{} {
	cutoff := time.Now().Add(-24 * time.Hour) // Remove entries older than 24 hours

	for userID, userState := range rl.userLimits {
		if userState.LastActivity.Before(cutoff) &&
			len(userState.CreatedTables) == 0 &&
			len(userState.ActiveTables) == 0 &&
			len(userState.ObservedTables) == 0 {
			delete(rl.userLimits, userID)
		} else {
			// Clean up old attempts
			userState.CreateAttempts = rl.filterRecentAttempts(userState.CreateAttempts, rl.createTableWindow)
			userState.JoinAttempts = rl.filterRecentAttempts(userState.JoinAttempts, rl.joinAttemptWindow)
		}
	}
	return nil
}

// RateLimiterState holds the state managed by the actor
type RateLimiterState struct {
	// Map of user ID to their rate limit state
	userLimits map[string]*UserLimitState

	// Configuration
	maxTablesPerUser    int           // Max tables a user can create
	createTableWindow   time.Duration // Time window for table creation limits
	maxCreatesPerWindow int           // Max table creates per window
	joinAttemptWindow   time.Duration // Time window for join attempt limits
	maxJoinsPerWindow   int           // Max join attempts per window
	maxObserverTables   int           // Max tables a user can observe simultaneously
	cleanupInterval     time.Duration // How often to clean up old entries
}

// getUserState returns the rate limit state for a user, creating it if needed
func (rl *RateLimiterState) getUserState(userID string) *UserLimitState {
	state, exists := rl.userLimits[userID]
	if !exists {
		state = &UserLimitState{
			CreatedTables:  make([]string, 0),
			CreateAttempts: make([]time.Time, 0),
			JoinAttempts:   make([]time.Time, 0),
			ObservedTables: make([]string, 0),
			ActiveTables:   make([]string, 0),
			LastActivity:   time.Now(),
		}
		rl.userLimits[userID] = state
	}
	return state
}

// filterRecentAttempts removes attempts outside the time window
func (rl *RateLimiterState) filterRecentAttempts(attempts []time.Time, window time.Duration) []time.Time {
	cutoff := time.Now().Add(-window)
	filtered := make([]time.Time, 0, len(attempts))

	for _, attempt := range attempts {
		if attempt.After(cutoff) {
			filtered = append(filtered, attempt)
		}
	}

	return filtered
}

// removeFromSlice removes a value from a string slice
func (rl *RateLimiterState) removeFromSlice(slice []string, value string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != value {
			result = append(result, item)
		}
	}
	return result
}

// containsString checks if a string slice contains a value
func (rl *RateLimiterState) containsString(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// ActorRateLimiter implements rate limiting using the actor pattern
type ActorRateLimiter struct {
	commands chan RateLimiterCommand
	state    *RateLimiterState
	done     chan struct{}
}

// getLimit extracts an integer limit from a limits map with a default value
func getLimit(limits map[string]interface{}, key string, defaultValue int) int {
	if limits == nil {
		return defaultValue
	}
	if val, exists := limits[key]; exists {
		if intVal, ok := val.(int); ok {
			return intVal
		}
	}
	return defaultValue
}

// NewActorRateLimiter creates a new actor-based rate limiter
func NewActorRateLimiter() *ActorRateLimiter {
	return NewActorRateLimiterWithLimits(map[string]interface{}{})
}

// NewActorRateLimiterWithLimits creates a new actor-based rate limiter with custom limits
func NewActorRateLimiterWithLimits(limits map[string]interface{}) *ActorRateLimiter {
	state := &RateLimiterState{
		userLimits:          make(map[string]*UserLimitState),
		maxTablesPerUser:    getLimit(limits, "max_tables_per_user", 10),
		createTableWindow:   time.Minute * 5,
		maxCreatesPerWindow: getLimit(limits, "max_creates_per_window", 5),
		joinAttemptWindow:   time.Minute,
		maxJoinsPerWindow:   getLimit(limits, "max_joins_per_window", 10),
		maxObserverTables:   getLimit(limits, "max_observer_tables", 20),
		cleanupInterval:     time.Hour,
	}

	arl := &ActorRateLimiter{
		commands: make(chan RateLimiterCommand, 1000),
		state:    state,
		done:     make(chan struct{}),
	}

	// Start the actor goroutine
	go arl.run()

	// Start cleanup routine
	go arl.cleanupRoutine()

	return arl
}

// run is the main actor loop
func (arl *ActorRateLimiter) run() {
	for {
		select {
		case cmd := <-arl.commands:
			cmd.Execute(arl.state)
		case <-arl.done:
			return
		}
	}
}

// cleanupRoutine periodically cleans up old entries
func (arl *ActorRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(arl.state.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			arl.commands <- &CleanupCommand{}
		case <-arl.done:
			return
		}
	}
}

// Stop stops the actor
func (arl *ActorRateLimiter) Stop() {
	close(arl.done)
}

// CanCreateTable checks if a user can create a table
func (arl *ActorRateLimiter) CanCreateTable(userID string) error {
	result := make(chan error, 1)
	cmd := &CanCreateTableCommand{
		UserID: userID,
		Result: result,
	}

	arl.commands <- cmd
	return <-result
}

// CanJoinTable checks if a user can join a table
func (arl *ActorRateLimiter) CanJoinTable(userID string, tableID string) error {
	result := make(chan error, 1)
	cmd := &CanJoinTableCommand{
		UserID:  userID,
		TableID: tableID,
		Result:  result,
	}

	arl.commands <- cmd
	return <-result
}

// CanObserveTable checks if a user can observe a table
func (arl *ActorRateLimiter) CanObserveTable(userID string, tableID string) error {
	result := make(chan error, 1)
	cmd := &CanObserveTableCommand{
		UserID:  userID,
		TableID: tableID,
		Result:  result,
	}

	arl.commands <- cmd
	return <-result
}

// RecordTableCreated records that a user created a table
func (arl *ActorRateLimiter) RecordTableCreated(userID string, tableID string) {
	cmd := &RecordTableCreatedCommand{
		UserID:  userID,
		TableID: tableID,
	}

	arl.commands <- cmd
}

// RecordTableClosed records that a user closed a table
func (arl *ActorRateLimiter) RecordTableClosed(userID string, tableID string) {
	cmd := &RecordTableClosedCommand{
		UserID:  userID,
		TableID: tableID,
	}

	arl.commands <- cmd
}

// RecordPlayerJoined records that a user joined a table as player
func (arl *ActorRateLimiter) RecordPlayerJoined(userID string, tableID string) {
	cmd := &RecordPlayerJoinedCommand{
		UserID:  userID,
		TableID: tableID,
	}

	arl.commands <- cmd
}

// RecordPlayerLeft records that a user left a table as player
func (arl *ActorRateLimiter) RecordPlayerLeft(userID string, tableID string) {
	cmd := &RecordPlayerLeftCommand{
		UserID:  userID,
		TableID: tableID,
	}

	arl.commands <- cmd
}

// RecordObserverJoined records that a user joined a table as observer
func (arl *ActorRateLimiter) RecordObserverJoined(userID string, tableID string) {
	cmd := &RecordObserverJoinedCommand{
		UserID:  userID,
		TableID: tableID,
	}

	arl.commands <- cmd
}

// RecordObserverLeft records that a user left a table as observer
func (arl *ActorRateLimiter) RecordObserverLeft(userID string, tableID string) {
	cmd := &RecordObserverLeftCommand{
		UserID:  userID,
		TableID: tableID,
	}

	arl.commands <- cmd
}

// GetUserStats returns statistics for a user
func (arl *ActorRateLimiter) GetUserStats(userID string) map[string]interface{} {
	result := make(chan map[string]interface{}, 1)
	cmd := &GetUserStatsCommand{
		UserID: userID,
		Result: result,
	}

	arl.commands <- cmd
	return <-result
}

// SetLimits provides a compatibility method for tests
func (arl *ActorRateLimiter) SetLimits(limits map[string]interface{}) {
	// For compatibility with existing tests, this is a no-op
	// The limits are set during construction and are immutable in the actor pattern
	// In the future, this could send a command to update limits if needed
}
