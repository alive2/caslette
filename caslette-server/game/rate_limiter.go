package game

import (
	"sync"
	"time"
)

// RateLimiter implements rate limiting for table operations
type RateLimiter struct {
	// Map of user ID to their rate limit state
	userLimits map[string]*UserLimitState
	mutex      sync.RWMutex

	// Configuration
	maxTablesPerUser    int           // Max tables a user can create
	createTableWindow   time.Duration // Time window for table creation limits
	maxCreatesPerWindow int           // Max table creates per window
	joinAttemptWindow   time.Duration // Time window for join attempt limits
	maxJoinsPerWindow   int           // Max join attempts per window
	maxObserverTables   int           // Max tables a user can observe simultaneously
	cleanupInterval     time.Duration // How often to clean up old entries
}

// UserLimitState tracks rate limiting state for a user
type UserLimitState struct {
	// Table creation limits
	CreatedTables  []string    // List of table IDs created by user
	CreateAttempts []time.Time // Timestamps of recent create attempts

	// Join attempt limits
	JoinAttempts []time.Time // Timestamps of recent join attempts

	// Current state
	ObservedTables []string // Tables currently being observed
	ActiveTables   []string // Tables currently playing in

	LastActivity time.Time // Last activity timestamp
}

// NewRateLimiter creates a new rate limiter with default settings
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		userLimits:          make(map[string]*UserLimitState),
		maxTablesPerUser:    5,           // Max 5 tables per user
		createTableWindow:   time.Hour,   // 1 hour window
		maxCreatesPerWindow: 10,          // Max 10 creates per hour
		joinAttemptWindow:   time.Minute, // 1 minute window
		maxJoinsPerWindow:   20,          // Max 20 join attempts per minute
		maxObserverTables:   10,          // Max 10 observed tables
		cleanupInterval:     time.Hour,   // Cleanup every hour
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// CanCreateTable checks if a user can create a new table
func (rl *RateLimiter) CanCreateTable(userID string) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	userState := rl.getUserState(userID)
	now := time.Now()

	// Check active table limit
	if len(userState.CreatedTables) >= rl.maxTablesPerUser {
		return &TableError{"RATE_LIMIT_TABLES", "Maximum number of active tables reached"}
	}

	// Clean old create attempts
	userState.CreateAttempts = rl.filterRecentAttempts(userState.CreateAttempts, rl.createTableWindow)

	// Check create attempts in window
	if len(userState.CreateAttempts) >= rl.maxCreatesPerWindow {
		return &TableError{"RATE_LIMIT_CREATES", "Too many table creation attempts"}
	}

	// Record this attempt
	userState.CreateAttempts = append(userState.CreateAttempts, now)
	userState.LastActivity = now

	return nil
}

// CanJoinTable checks if a user can join a table
func (rl *RateLimiter) CanJoinTable(userID string, tableID string) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	userState := rl.getUserState(userID)
	now := time.Now()

	// Clean old join attempts
	userState.JoinAttempts = rl.filterRecentAttempts(userState.JoinAttempts, rl.joinAttemptWindow)

	// Check join attempts in window
	if len(userState.JoinAttempts) >= rl.maxJoinsPerWindow {
		return &TableError{"RATE_LIMIT_JOINS", "Too many join attempts"}
	}

	// Record this attempt
	userState.JoinAttempts = append(userState.JoinAttempts, now)
	userState.LastActivity = now

	return nil
}

// CanObserveTable checks if a user can observe a table
func (rl *RateLimiter) CanObserveTable(userID string, tableID string) error {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	userState := rl.getUserState(userID)

	// Check if already observing this table
	for _, observedID := range userState.ObservedTables {
		if observedID == tableID {
			return nil // Already observing, allow
		}
	}

	// Check observer limit
	if len(userState.ObservedTables) >= rl.maxObserverTables {
		return &TableError{"RATE_LIMIT_OBSERVERS", "Maximum number of observed tables reached"}
	}

	return nil
}

// RecordTableCreated records that a user successfully created a table
func (rl *RateLimiter) RecordTableCreated(userID string, tableID string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	userState := rl.getUserState(userID)
	userState.CreatedTables = append(userState.CreatedTables, tableID)
	userState.LastActivity = time.Now()
}

// RecordTableClosed records that a user's table was closed
func (rl *RateLimiter) RecordTableClosed(userID string, tableID string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	userState := rl.getUserState(userID)
	userState.CreatedTables = rl.removeFromSlice(userState.CreatedTables, tableID)
}

// RecordPlayerJoined records that a user joined a table as a player
func (rl *RateLimiter) RecordPlayerJoined(userID string, tableID string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	userState := rl.getUserState(userID)

	// Remove from observers if present
	userState.ObservedTables = rl.removeFromSlice(userState.ObservedTables, tableID)

	// Add to active tables if not present
	for _, activeID := range userState.ActiveTables {
		if activeID == tableID {
			return // Already in active tables
		}
	}
	userState.ActiveTables = append(userState.ActiveTables, tableID)
	userState.LastActivity = time.Now()
}

// RecordPlayerLeft records that a user left a table
func (rl *RateLimiter) RecordPlayerLeft(userID string, tableID string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	userState := rl.getUserState(userID)
	userState.ActiveTables = rl.removeFromSlice(userState.ActiveTables, tableID)
}

// RecordObserverJoined records that a user joined a table as an observer
func (rl *RateLimiter) RecordObserverJoined(userID string, tableID string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	userState := rl.getUserState(userID)

	// Add to observed tables if not present
	for _, observedID := range userState.ObservedTables {
		if observedID == tableID {
			return // Already observing
		}
	}
	userState.ObservedTables = append(userState.ObservedTables, tableID)
	userState.LastActivity = time.Now()
}

// RecordObserverLeft records that a user stopped observing a table
func (rl *RateLimiter) RecordObserverLeft(userID string, tableID string) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	userState := rl.getUserState(userID)
	userState.ObservedTables = rl.removeFromSlice(userState.ObservedTables, tableID)
}

// GetUserStats returns current stats for a user
func (rl *RateLimiter) GetUserStats(userID string) map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	userState := rl.getUserState(userID)

	return map[string]interface{}{
		"created_tables":  len(userState.CreatedTables),
		"active_tables":   len(userState.ActiveTables),
		"observed_tables": len(userState.ObservedTables),
		"recent_creates":  len(rl.filterRecentAttempts(userState.CreateAttempts, rl.createTableWindow)),
		"recent_joins":    len(rl.filterRecentAttempts(userState.JoinAttempts, rl.joinAttemptWindow)),
		"last_activity":   userState.LastActivity,
	}
}

// getUserState gets or creates a user's rate limit state
func (rl *RateLimiter) getUserState(userID string) *UserLimitState {
	if state, exists := rl.userLimits[userID]; exists {
		return state
	}

	state := &UserLimitState{
		CreatedTables:  make([]string, 0),
		CreateAttempts: make([]time.Time, 0),
		JoinAttempts:   make([]time.Time, 0),
		ObservedTables: make([]string, 0),
		ActiveTables:   make([]string, 0),
		LastActivity:   time.Now(),
	}

	rl.userLimits[userID] = state
	return state
}

// filterRecentAttempts filters timestamps to only include recent ones
func (rl *RateLimiter) filterRecentAttempts(attempts []time.Time, window time.Duration) []time.Time {
	now := time.Now()
	cutoff := now.Add(-window)

	var recent []time.Time
	for _, attempt := range attempts {
		if attempt.After(cutoff) {
			recent = append(recent, attempt)
		}
	}

	return recent
}

// removeFromSlice removes a value from a string slice
func (rl *RateLimiter) removeFromSlice(slice []string, value string) []string {
	var result []string
	for _, item := range slice {
		if item != value {
			result = append(result, item)
		}
	}
	return result
}

// cleanupRoutine periodically cleans up old user state entries
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes old user state entries
func (rl *RateLimiter) cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-24 * time.Hour) // Keep entries for 24 hours

	for userID, state := range rl.userLimits {
		// Remove if user has been inactive for too long and has no active state
		if state.LastActivity.Before(cutoff) &&
			len(state.CreatedTables) == 0 &&
			len(state.ActiveTables) == 0 &&
			len(state.ObservedTables) == 0 {
			delete(rl.userLimits, userID)
		} else {
			// Clean old attempts
			state.CreateAttempts = rl.filterRecentAttempts(state.CreateAttempts, rl.createTableWindow)
			state.JoinAttempts = rl.filterRecentAttempts(state.JoinAttempts, rl.joinAttemptWindow)
		}
	}
}

// SetLimits allows configuring the rate limits
func (rl *RateLimiter) SetLimits(config map[string]interface{}) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if maxTables, ok := config["max_tables_per_user"].(int); ok {
		rl.maxTablesPerUser = maxTables
	}
	if maxCreates, ok := config["max_creates_per_window"].(int); ok {
		rl.maxCreatesPerWindow = maxCreates
	}
	if maxJoins, ok := config["max_joins_per_window"].(int); ok {
		rl.maxJoinsPerWindow = maxJoins
	}
	if maxObservers, ok := config["max_observer_tables"].(int); ok {
		rl.maxObserverTables = maxObservers
	}
}
