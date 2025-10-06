package game

import (
	"encoding/json"
	"fmt"
)

// ConvertMapToStruct converts a map[string]interface{} to a struct using JSON marshaling
func ConvertMapToStruct(data map[string]interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal to target: %v", err)
	}

	return nil
}

// BroadcastGameEvent is a method that should be added to ActorTableManager
// For now, this is a placeholder implementation since the webhook system is not fully implemented
func (tm *ActorTableManager) BroadcastGameEvent(table *GameTable, event *GameEvent) {
	// This would broadcast game events to WebSocket clients
	// Currently a no-op since the webhook handler system is not fully implemented
	// TODO: Implement proper event broadcasting when needed
}

// GameEventBroadcaster interface for broadcasting game events
type GameEventBroadcaster interface {
	OnGameEvent(table *GameTable, event *GameEvent)
}
