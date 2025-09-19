package game

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestTableSecurityValidation tests input validation and sanitization
func TestTableSecurityValidation(t *testing.T) {
	validator := NewTableValidator()

	// Test SQL injection attempts in table names
	sqlInjectionNames := []string{
		"'; DROP TABLE users; --",
		"<script>alert('xss')</script>",
		"test' OR '1'='1",
		"table" + string(rune(0)) + "name", // Null byte
		strings.Repeat("a", 1000),          // Too long
	}

	for _, name := range sqlInjectionNames {
		err := validator.ValidateTableName(name)
		if err == nil {
			t.Errorf("Expected validation error for malicious table name: %s", name)
		}
	}

	// Test valid table names
	validNames := []string{
		"My Poker Table",
		"Texas Hold'em - High Stakes",
		"Tournament_2024",
	}

	for _, name := range validNames {
		err := validator.ValidateTableName(name)
		if err != nil {
			t.Errorf("Expected valid table name to pass: %s, got error: %v", name, err)
		}
	}

	// Test username validation
	invalidUsernames := []string{
		"",
		"ab",                    // too short
		strings.Repeat("a", 50), // too long
		"user<script>",
		"user';--",
	}

	for _, username := range invalidUsernames {
		err := validator.ValidateUsername(username)
		if err == nil {
			t.Errorf("Expected validation error for invalid username: %s", username)
		}
	}
}

// TestTableRateLimiting tests rate limiting functionality
func TestTableRateLimiting(t *testing.T) {
	rateLimiter := NewActorRateLimiterWithLimits(map[string]interface{}{
		"max_tables_per_user":    10, // High enough to not interfere with window-based test
		"max_creates_per_window": 3,
		"max_joins_per_window":   5,
		"max_observer_tables":    3,
	})
	defer rateLimiter.Stop()

	userID := "test_user"

	// Test table creation limits
	for i := 0; i < 3; i++ {
		err := rateLimiter.CanCreateTable(userID)
		if err != nil {
			t.Errorf("Expected to allow table creation %d, got error: %v", i+1, err)
		}
		// Record that the table was actually created
		rateLimiter.RecordTableCreated(userID, fmt.Sprintf("table_%d", i))
	}

	// Should hit rate limit on 4th attempt
	err := rateLimiter.CanCreateTable(userID)
	if err == nil {
		t.Error("Expected rate limit error on 4th table creation attempt")
	}

	// Test join limits
	for i := 0; i < 5; i++ {
		tableID := fmt.Sprintf("join_table_%d", i)
		err := rateLimiter.CanJoinTable(userID, tableID)
		if err != nil {
			t.Errorf("Expected to allow join attempt %d, got error: %v", i+1, err)
		}
		// Record that the player actually joined
		rateLimiter.RecordPlayerJoined(userID, tableID)
	}

	// Should hit rate limit on 6th attempt
	err = rateLimiter.CanJoinTable(userID, "final_join_table")
	if err == nil {
		t.Error("Expected rate limit error on 6th join attempt")
	}

	// Test observer limits
	for i := 0; i < 3; i++ {
		tableID := fmt.Sprintf("table_%d", i)
		err := rateLimiter.CanObserveTable(userID, tableID)
		if err != nil {
			t.Errorf("Expected to allow observer %d, got error: %v", i+1, err)
		}
		rateLimiter.RecordObserverJoined(userID, tableID)
	}

	// Should hit observer limit on 4th table
	err = rateLimiter.CanObserveTable(userID, "table_4")
	if err == nil {
		t.Error("Expected rate limit error on 4th observed table")
	}
}

// TestTableDataFiltering tests data exposure protection
func TestTableDataFiltering(t *testing.T) {
	dataFilter := NewDataFilter()

	// Create a test table
	settings := DefaultTableSettings()
	settings.Private = true
	settings.Password = "secret123"

	table := NewGameTable("table123", "Private Table", GameTypeTexasHoldem, "creator", settings)
	table.Description = "Secret game details"

	// Add a player
	table.PlayerSlots[0].PlayerID = "player1"
	table.PlayerSlots[0].Username = "Player1"
	table.PlayerSlots[0].JoinedAt = time.Now()

	// Add an observer
	table.Observers = append(table.Observers, TableObserver{
		PlayerID: "observer1",
		Username: "Observer1",
		JoinedAt: time.Now(),
	})

	// Test data filtering for different user types

	// 1. Random user (not at table) should get minimal info
	publicInfo := dataFilter.FilterTableInfo(table, "random_user", "")

	// Should not expose sensitive data
	if _, hasPassword := publicInfo["settings"].(map[string]interface{})["password"]; hasPassword {
		t.Error("Password should not be exposed to non-participants")
	}

	if _, hasCreator := publicInfo["created_by"]; hasCreator {
		t.Error("Creator should not be exposed for private tables to non-participants")
	}

	if _, hasRoomID := publicInfo["room_id"]; hasRoomID {
		t.Error("Room ID should not be exposed to non-participants")
	}

	// 2. Player at table should get more info
	playerInfo := dataFilter.FilterTableInfo(table, "player1", "")

	// Should have access to room and creator
	if _, hasCreator := playerInfo["created_by"]; !hasCreator {
		t.Error("Creator should be exposed to participants")
	}

	if _, hasRoomID := playerInfo["room_id"]; !hasRoomID {
		t.Error("Room ID should be exposed to participants")
	}

	// Should still not expose password
	settingsMap := playerInfo["settings"].(map[string]interface{})
	if _, hasPassword := settingsMap["password"]; hasPassword {
		t.Error("Password should never be exposed, even to participants")
	}

	// 3. Observer should get similar access to player
	observerInfo := dataFilter.FilterTableInfo(table, "observer1", "")

	if _, hasCreator := observerInfo["created_by"]; !hasCreator {
		t.Error("Creator should be exposed to observers")
	}
}

// TestTableAccessControl tests access control validation
func TestTableAccessControl(t *testing.T) {
	dataFilter := NewDataFilter()

	// Create a private table
	settings := DefaultTableSettings()
	settings.Private = true
	settings.Password = "secret"

	table := NewGameTable("private_table", "Private Game", GameTypeTexasHoldem, "creator", settings)

	// Test access validation

	// 1. Random user should be denied access to private table
	err := dataFilter.ValidateTableAccess(table, "random_user", "view")
	if err == nil {
		t.Error("Expected access denied for random user to private table")
	}

	// 2. Creator should have access
	err = dataFilter.ValidateTableAccess(table, "creator", "view")
	if err != nil {
		t.Errorf("Expected creator to have access, got error: %v", err)
	}

	// 3. Management access should be restricted to creator
	err = dataFilter.ValidateTableAccess(table, "random_user", "manage")
	if err == nil {
		t.Error("Expected management access denied for non-creator")
	}

	err = dataFilter.ValidateTableAccess(table, "creator", "manage")
	if err != nil {
		t.Errorf("Expected creator to have management access, got error: %v", err)
	}

	// 4. Closed table should deny joins
	table.Status = TableStatusClosed
	err = dataFilter.ValidateTableAccess(table, "player1", "join")
	if err == nil {
		t.Error("Expected join denied for closed table")
	}
}

// TestTableManagerSecurity tests the security features of the table manager
func TestTableManagerSecurity(t *testing.T) {
	manager := NewTableManager(nil)
	ctx := context.Background()

	// Test input validation on table creation
	maliciousRequest := &TableCreateRequest{
		Name:        "'; DROP TABLE users; --",
		GameType:    GameTypeTexasHoldem,
		CreatedBy:   "attacker",
		Username:    "Attacker",
		Settings:    DefaultTableSettings(),
		Description: "<script>alert('xss')</script>",
	}

	_, err := manager.CreateTable(ctx, maliciousRequest)
	if err == nil {
		t.Error("Expected validation error for malicious table creation request")
	}

	// Test rate limiting on table creation
	userID := "test_user"

	// Create multiple tables quickly to test rate limiting
	for i := 0; i < 10; i++ {
		request := &TableCreateRequest{
			Name:      fmt.Sprintf("Table %d", i),
			GameType:  GameTypeTexasHoldem,
			CreatedBy: userID,
			Username:  "TestUser",
			Settings:  DefaultTableSettings(),
		}

		_, err := manager.CreateTable(ctx, request)
		if i >= 5 && err == nil {
			t.Errorf("Expected rate limit error after creating %d tables", i+1)
			break
		}
	}

	// Test filtered table listing
	validRequest := &TableCreateRequest{
		Name:      "Valid Table",
		GameType:  GameTypeTexasHoldem,
		CreatedBy: "creator",
		Username:  "Creator",
		Settings:  DefaultTableSettings(),
	}

	table, err := manager.CreateTable(ctx, validRequest)
	if err != nil {
		t.Fatalf("Failed to create valid table: %v", err)
	}

	// Test that listing is properly filtered
	allTables := manager.ListTables(map[string]interface{}{})
	dataFilter := NewDataFilter()
	tables := dataFilter.FilterTableList(allTables, "random_user")
	if len(tables) == 0 {
		t.Error("Expected at least one table in listing")
	}

	// Check that sensitive data is not exposed in list
	tableInfo := tables[0]
	if _, hasPassword := tableInfo["password"]; hasPassword {
		t.Error("Password should not be exposed in table listing")
	}

	// Test secure table info retrieval
	info, err := manager.GetTableInfo(table.ID, "random_user")
	if err != nil {
		t.Errorf("Expected to be able to get table info, got error: %v", err)
	}

	if _, hasPassword := info["password"]; hasPassword {
		t.Error("Password should not be exposed in table info")
	}
}

// TestConcurrentTableAccess tests thread safety and race conditions
func TestConcurrentTableAccess(t *testing.T) {
	manager := NewTableManager(nil)
	ctx := context.Background()

	// Create a table
	request := &TableCreateRequest{
		Name:      "Concurrent Test Table",
		GameType:  GameTypeTexasHoldem,
		CreatedBy: "creator",
		Username:  "Creator",
		Settings:  DefaultTableSettings(),
	}

	table, err := manager.CreateTable(ctx, request)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Launch concurrent operations
	numGoroutines := 50
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Try to join and leave the table concurrently
			joinRequest := &TableJoinRequest{
				TableID:  table.ID,
				PlayerID: fmt.Sprintf("player_%d", id),
				Username: fmt.Sprintf("Player%d", id),
				Mode:     JoinModePlayer,
			}

			err := manager.JoinTable(ctx, joinRequest)
			errChan <- err
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		if err == nil {
			successCount++
		}
	}

	// Should not have more successes than available slots
	if successCount > table.MaxPlayers {
		t.Errorf("Too many players joined table: %d > %d", successCount, table.MaxPlayers)
	}

	// Verify table state consistency
	finalTable, err := manager.GetTable(table.ID)
	if err != nil {
		t.Fatalf("Failed to get table after concurrent access: %v", err)
	}

	playerCount := finalTable.GetPlayerCount()
	if playerCount != successCount {
		t.Errorf("Player count mismatch: expected %d, got %d", successCount, playerCount)
	}
}

// TestAuditLogging tests security audit logging
func TestAuditLogging(t *testing.T) {
	auditor := NewSecurityAuditor()

	// Log some actions
	auditor.LogAction("user1", "table1", "create_table", "success", "")
	auditor.LogAction("user2", "table1", "join_table", "failed", "table full")
	auditor.LogAction("user3", "table2", "get_table_info", "access_denied", "private table")

	// Get audit logs
	logs := auditor.GetAuditLogs(10)

	if len(logs) != 3 {
		t.Errorf("Expected 3 audit logs, got %d", len(logs))
	}

	// Check log contents
	if logs[0].UserID != "user1" || logs[0].Action != "create_table" {
		t.Error("First audit log has incorrect data")
	}

	if logs[1].Result != "failed" || logs[1].Details != "table full" {
		t.Error("Second audit log has incorrect data")
	}

	if logs[2].Action != "get_table_info" || logs[2].Result != "access_denied" {
		t.Error("Third audit log has incorrect data")
	}
}

// TestSecurityIntegration tests end-to-end security features
func TestSecurityIntegration(t *testing.T) {
	manager := NewTableManager(nil)
	ctx := context.Background()

	// Create a private table with password
	settings := DefaultTableSettings()
	settings.Private = true
	settings.Password = "secret123"

	request := &TableCreateRequest{
		Name:      "Secure Table",
		GameType:  GameTypeTexasHoldem,
		CreatedBy: "creator",
		Username:  "Creator",
		Settings:  settings,
	}

	table, err := manager.CreateTable(ctx, request)
	if err != nil {
		t.Fatalf("Failed to create secure table: %v", err)
	}

	// Test password-protected join
	joinRequest := &TableJoinRequest{
		TableID:  table.ID,
		PlayerID: "player1",
		Username: "Player1",
		Mode:     JoinModePlayer,
		Password: "wrong_password",
	}

	err = manager.JoinTable(ctx, joinRequest)
	if err == nil {
		t.Error("Expected join to fail with wrong password")
	}

	// Test correct password
	joinRequest.Password = "secret123"
	err = manager.JoinTable(ctx, joinRequest)
	if err != nil {
		t.Errorf("Expected join to succeed with correct password, got error: %v", err)
	}

	// Test that table info is properly filtered
	info, err := manager.GetTableInfo(table.ID, "random_user")
	if err == nil {
		t.Error("Expected access denied for random user to private table info")
	}

	// Creator should be able to access
	info, err = manager.GetTableInfo(table.ID, "creator")
	if err != nil {
		t.Errorf("Expected creator to access table info, got error: %v", err)
	}

	// Verify no sensitive data is exposed
	if settings, ok := info["settings"].(map[string]interface{}); ok {
		if _, hasPassword := settings["password"]; hasPassword {
			t.Error("Password should never be exposed in table info")
		}
	}
}
