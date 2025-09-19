package game

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestActorTableManager(t *testing.T) {
	factory := &MockGameEngineFactory{}
	manager := NewActorTableManager(factory)
	defer manager.Stop()

	ctx := context.Background()

	// Test table creation
	req := &TableCreateRequest{
		Name:        "Test Actor Table",
		GameType:    GameTypeTexasHoldem,
		CreatedBy:   "user1",
		Username:    "User1",
		Settings:    DefaultTableSettings(),
		Description: "Test actor description",
		Tags:        []string{"actor", "test"},
	}

	table, err := manager.CreateTable(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	if table.Name != req.Name {
		t.Errorf("Expected table name '%s', got '%s'", req.Name, table.Name)
	}

	// Test joining a player
	joinReq := &TableJoinRequest{
		TableID:  table.ID,
		PlayerID: "player1",
		Username: "Player1",
		Mode:     JoinModePlayer,
		Position: 0, // Auto-assign
	}

	err = manager.JoinTable(ctx, joinReq)
	if err != nil {
		t.Fatalf("Failed to join table: %v", err)
	}

	// Test joining an observer
	observeReq := &TableJoinRequest{
		TableID:  table.ID,
		PlayerID: "observer1",
		Username: "Observer1",
		Mode:     JoinModeObserver,
	}

	err = manager.JoinTable(ctx, observeReq)
	if err != nil {
		t.Fatalf("Failed to join as observer: %v", err)
	}

	// Test leaving
	leaveReq := &TableLeaveRequest{
		TableID:  table.ID,
		PlayerID: "player1",
	}

	err = manager.LeaveTable(ctx, leaveReq)
	if err != nil {
		t.Fatalf("Failed to leave table: %v", err)
	}

	// Test concurrent operations (this should not deadlock)
	done := make(chan bool, 10)
	
	// Start multiple concurrent join/leave operations
	for i := 0; i < 5; i++ {
		go func(playerNum int) {
			playerID := fmt.Sprintf("player%d", playerNum)
			
			joinReq := &TableJoinRequest{
				TableID:  table.ID,
				PlayerID: playerID,
				Username: fmt.Sprintf("Player%d", playerNum),
				Mode:     JoinModePlayer,
				Position: 0,
			}
			
			// Join
			if err := manager.JoinTable(ctx, joinReq); err != nil {
				t.Logf("Join failed for %s: %v", playerID, err)
			}
			
			// Small delay
			time.Sleep(10 * time.Millisecond)
			
			// Leave
			leaveReq := &TableLeaveRequest{
				TableID:  table.ID,
				PlayerID: playerID,
			}
			
			if err := manager.LeaveTable(ctx, leaveReq); err != nil {
				t.Logf("Leave failed for %s: %v", playerID, err)
			}
			
			done <- true
		}(i)
	}

	// Wait for all operations to complete (with timeout)
	timeout := time.After(5 * time.Second)
	completed := 0
	
	for completed < 5 {
		select {
		case <-done:
			completed++
		case <-timeout:
			t.Fatal("Concurrent operations timed out - possible deadlock")
		}
	}
	
	t.Log("All concurrent operations completed successfully!")
}

func TestActorTableManagerConcurrentAccess(t *testing.T) {
	factory := &MockGameEngineFactory{}
	manager := NewActorTableManager(factory)
	defer manager.Stop()

	ctx := context.Background()

	// Create a table
	req := &TableCreateRequest{
		Name:     "Concurrent Test Table",
		GameType: GameTypeTexasHoldem,
		CreatedBy: "user1",
		Username: "User1",
		Settings: DefaultTableSettings(),
	}

	table, err := manager.CreateTable(ctx, req)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Test many concurrent operations on the same table
	numOperations := 50
	done := make(chan bool, numOperations)
	
	for i := 0; i < numOperations; i++ {
		go func(opNum int) {
			defer func() { done <- true }()
			
			// Get table info multiple times
			for j := 0; j < 5; j++ {
				_, err := manager.GetTable(table.ID)
				if err != nil {
					t.Errorf("Failed to get table info: %v", err)
					return
				}
			}
		}(i)
	}

	// Wait for all operations to complete
	timeout := time.After(10 * time.Second)
	completed := 0
	
	for completed < numOperations {
		select {
		case <-done:
			completed++
		case <-timeout:
			t.Fatal("Concurrent table access timed out")
		}
	}
	
	t.Logf("Completed %d concurrent operations successfully!", numOperations)
}