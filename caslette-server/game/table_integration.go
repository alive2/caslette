package game

import (
	"context"
	"fmt"
)

// TexasHoldemEngineFactory implements GameEngineFactory for Texas Hold'em
type TexasHoldemEngineFactory struct{}

func (f *TexasHoldemEngineFactory) CreateEngine(gameType GameType, settings TableSettings) (GameEngine, error) {
	switch gameType {
	case GameTypeTexasHoldem:
		engine := NewTexasHoldemEngine("table_game")

		// Configure engine with table settings
		engine.SetSmallBlind(settings.SmallBlind)
		engine.SetBigBlind(settings.BigBlind)

		return engine, nil
	default:
		return nil, fmt.Errorf("unsupported game type: %s", gameType)
	}
}

// TableGameIntegration provides integration between tables and game engines
type TableGameIntegration struct {
	tableManager *ActorTableManager
	wsHandler    *TableWebSocketHandler
}

// NewTableGameIntegration creates a new table game integration
func NewTableGameIntegration(hub WebSocketHub) *TableGameIntegration {
	// Create engine factory
	engineFactory := &TexasHoldemEngineFactory{}

	// Create table manager
	tableManager := NewActorTableManager(engineFactory)

	// Create websocket handler
	wsHandler := NewTableWebSocketHandler(tableManager, hub)

	return &TableGameIntegration{
		tableManager: tableManager,
		wsHandler:    wsHandler,
	}
}

// GetTableManager returns the table manager
func (tgi *TableGameIntegration) GetTableManager() *ActorTableManager {
	return tgi.tableManager
}

// GetWebSocketHandler returns the websocket handler
func (tgi *TableGameIntegration) GetWebSocketHandler() *TableWebSocketHandler {
	return tgi.wsHandler
}

// GetMessageHandlers returns all websocket message handlers for tables
func (tgi *TableGameIntegration) GetMessageHandlers() map[string]func(ctx context.Context, conn WebSocketConnection, msg *WebSocketMessage) *WebSocketMessage {
	return tgi.wsHandler.GetMessageHandlers()
}

// Example usage and configuration helpers

// DefaultTableSettings returns default settings for Texas Hold'em
func DefaultTableSettings() TableSettings {
	return TableSettings{
		SmallBlind:       10,
		BigBlind:         20,
		BuyIn:            1000,
		MaxBuyIn:         2000,
		AutoStart:        false,
		TimeLimit:        30,
		TournamentMode:   false,
		ObserversAllowed: true,
		Private:          false,
	}
}

// QuickGameSettings returns settings for quick casual games
func QuickGameSettings() TableSettings {
	return TableSettings{
		SmallBlind:       5,
		BigBlind:         10,
		BuyIn:            500,
		MaxBuyIn:         1000,
		AutoStart:        true,
		TimeLimit:        20,
		TournamentMode:   false,
		ObserversAllowed: true,
		Private:          false,
	}
}

// TournamentSettings returns settings for tournament play
func TournamentSettings() TableSettings {
	return TableSettings{
		SmallBlind:       25,
		BigBlind:         50,
		BuyIn:            2000,
		MaxBuyIn:         2000,
		AutoStart:        false,
		TimeLimit:        45,
		TournamentMode:   true,
		ObserversAllowed: true,
		Private:          false,
	}
}

// PrivateTableSettings returns settings for private tables
func PrivateTableSettings(password string) TableSettings {
	settings := DefaultTableSettings()
	settings.Private = true
	settings.Password = password
	settings.ObserversAllowed = false
	return settings
}
