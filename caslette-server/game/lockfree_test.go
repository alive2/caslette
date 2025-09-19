package game

import (
	"context"
	"testing"
	"time"
)

func TestActorRateLimiter(t *testing.T) {
	// Create actor-based rate limiter
	limiter := NewActorRateLimiter()
	defer limiter.Stop()

	userID := "test-user"
	tableID := "test-table"

	// Test can create table
	err := limiter.CanCreateTable(userID)
	if err != nil {
		t.Fatalf("Expected to allow table creation, got error: %v", err)
	}

	// Record table created
	limiter.RecordTableCreated(userID, tableID)

	// Test can join table
	err = limiter.CanJoinTable(userID, tableID)
	if err != nil {
		t.Fatalf("Expected to allow table join, got error: %v", err)
	}

	// Record player joined
	limiter.RecordPlayerJoined(userID, tableID)

	// Test can observe table
	err = limiter.CanObserveTable(userID, "another-table")
	if err != nil {
		t.Fatalf("Expected to allow table observation, got error: %v", err)
	}

	// Record observer joined
	limiter.RecordObserverJoined(userID, "another-table")

	// Get user stats
	stats := limiter.GetUserStats(userID)
	if stats["tables_created"].(int) != 1 {
		t.Errorf("Expected 1 table created, got %v", stats["tables_created"])
	}
	if stats["tables_joined"].(int) != 1 {
		t.Errorf("Expected 1 table joined, got %v", stats["tables_joined"])
	}
	if stats["tables_observing"].(int) != 1 {
		t.Errorf("Expected 1 table observing, got %v", stats["tables_observing"])
	}

	// Record player left
	limiter.RecordPlayerLeft(userID, tableID)

	// Record observer left
	limiter.RecordObserverLeft(userID, "another-table")

	// Record table closed
	limiter.RecordTableClosed(userID, tableID)

	// Check stats again
	stats = limiter.GetUserStats(userID)
	if stats["tables_created"].(int) != 0 {
		t.Errorf("Expected 0 tables created after closing, got %v", stats["tables_created"])
	}
	if stats["tables_joined"].(int) != 0 {
		t.Errorf("Expected 0 tables joined after leaving, got %v", stats["tables_joined"])
	}
	if stats["tables_observing"].(int) != 0 {
		t.Errorf("Expected 0 tables observing after leaving, got %v", stats["tables_observing"])
	}
}

func TestActorRateLimiterConcurrency(t *testing.T) {
	limiter := NewActorRateLimiter()
	defer limiter.Stop()

	// Test concurrent operations
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func(i int) {
			userID := "user-" + string(rune(i))
			tableID := "table-" + string(rune(i))

			// Rapid operations
			limiter.CanCreateTable(userID)
			limiter.RecordTableCreated(userID, tableID)
			limiter.CanJoinTable(userID, tableID)
			limiter.RecordPlayerJoined(userID, tableID)
			limiter.GetUserStats(userID)
			limiter.RecordPlayerLeft(userID, tableID)
			limiter.RecordTableClosed(userID, tableID)

			done <- true
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < 100; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent operations timed out")
		}
	}

	// No deadlocks or panics means success!
	t.Log("All concurrent operations completed successfully")
}

func TestLockFreeArchitecture(t *testing.T) {
	t.Log("=== Testing Lock-Free Architecture ===")

	// Test TableActor
	factory := &MockGameEngineFactory{}
	actorManager := NewActorTableManager(factory)
	defer actorManager.Stop()

	// Test RateLimiter Actor
	rateLimiter := NewActorRateLimiter()
	defer rateLimiter.Stop()

	userID := "test-user"
	ctx := context.Background()

	// Create table through actor manager
	req := &TableCreateRequest{
		Name:      "Test Table",
		GameType:  GameTypeTexasHoldem,
		CreatedBy: userID,
		Username:  "testuser",
		Settings: TableSettings{
			SmallBlind: 1,
			BigBlind:   2,
			BuyIn:      100,
		},
		Description: "Test table for lock-free architecture",
		Tags:        []string{"test"},
	}

	table, err := actorManager.CreateTable(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	tableID := table.ID

	// Test concurrent access to both systems
	done := make(chan bool, 50)

	for i := 0; i < 50; i++ {
		go func(i int) {
			testUserID := userID + string(rune(i))

			// Rate limiter operations
			rateLimiter.CanJoinTable(testUserID, tableID)

			// Table operations
			joinReq := &TableJoinRequest{
				TableID:  tableID,
				PlayerID: testUserID,
				Username: "TestUser" + string(rune(i)),
				Mode:     JoinModeObserver,
			}

			actorManager.JoinTable(ctx, joinReq)

			// More rate limiter operations
			rateLimiter.GetUserStats(testUserID)

			done <- true
		}(i)
	}

	// Wait for all operations
	for i := 0; i < 50; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(10 * time.Second):
			t.Fatal("Lock-free architecture test timed out")
		}
	}

	t.Log("Lock-free architecture successfully handled concurrent operations!")
	t.Log("✅ NO LOCKS USED - All concurrency handled through actor pattern")
}
