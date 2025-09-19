package websocket_v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestActorHubBasics(t *testing.T) {
	// Create actor hub
	hub := NewActorHub()
	
	// Start the hub
	hub.Start()
	defer hub.Stop()
	
	// Test connection count
	assert.Equal(t, 0, hub.GetConnectionCount())
}

func TestActorHubSecureIDGeneration(t *testing.T) {
	hub := NewActorHub()
	hub.Start()
	defer hub.Stop()
	
	// Generate multiple IDs and ensure they're unique
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := hub.generateSecureConnectionID()
		assert.NotEmpty(t, id)
		assert.False(t, ids[id], "Generated duplicate ID: %s", id)
		ids[id] = true
	}
}

func TestActorHubInputValidation(t *testing.T) {
	// Test SQL injection patterns are blocked
	sqlInjections := []string{
		"'; DROP TABLE users; --",
		"' OR '1'='1",
		"UNION SELECT * FROM secrets",
		"<script>alert('xss')</script>",
	}
	
	for _, injection := range sqlInjections {
		_, err := validateInput(injection, "room")
		assert.Error(t, err, "Should reject dangerous pattern: %s", injection)
	}
	
	// Test safe input passes
	safeInputs := []string{
		"hello",
		"room123",
		"my-room",
		"test_room",
	}
	
	for _, safe := range safeInputs {
		result, err := validateInput(safe, "room")
		assert.NoError(t, err, "Should allow safe input: %s", safe)
		assert.NotEmpty(t, result)
	}
}

func TestActorHubRateLimiting(t *testing.T) {
	hub := NewActorHub()
	hub.Start()
	defer hub.Stop()
	
	connectionID := "test-conn-123"
	
	// Send messages within rate limit (first 10 should be allowed)
	var firstError error
	for i := 0; i < 10; i++ {
		err := hub.checkRateLimit(connectionID)
		if err != nil && firstError == nil {
			firstError = err
		}
	}
	assert.NoError(t, firstError, "Should allow messages within rate limit")
	
	// This should trigger rate limiting
	err := hub.checkRateLimit(connectionID)
	assert.Error(t, err, "Should block messages exceeding rate limit")
}