# Game Package Organization

This package contains the core game logic for the poker server. Files are organized by functional areas:

## Core Components

### Engine System

- `engine.go` - Base game engine interface and implementation
- `engine_test.go` - Engine system tests

### Card System

- `cards.go` - Card, deck, and hand management
- `cards_test.go` - Card system tests

### Poker Game Logic

- `poker.go` - General poker logic and hand evaluation
- `poker_test.go` - Poker logic tests
- `texas_holdem.go` - Texas Hold'em specific implementation
- `texas_holdem_test.go` - Texas Hold'em tests

### Table Management (Actor-Based)

- `table.go` - Game table data structures
- `table_actor.go` - Actor-based table implementation
- `table_integration.go` - Table integration components
- `table_validator.go` - Table validation logic
- `table_websocket.go` - WebSocket handlers for tables
- `actor_table_manager.go` - Lock-free table manager using actor pattern
- `table_test.go` - Basic table tests
- `table_simple_test.go` - Simple table operation tests
- `actor_table_test.go` - Actor-based table tests

### Rate Limiting (Actor-Based)

- `actor_rate_limiter.go` - Lock-free rate limiter using actor pattern

### Security & Validation

- `security.go` - Security utilities and validation
- `table_security_test.go` - Table security tests
- `anti_exploitation_test.go` - Anti-exploitation tests
- `advanced_anti_exploitation_test.go` - Advanced anti-exploitation tests
- `advanced_security_exploits_test.go` - Advanced security exploit tests
- `sophisticated_exploits_test.go` - Sophisticated exploit prevention tests

### Architecture & Performance

- `types.go` - Shared type definitions
- `lockfree_test.go` - Lock-free architecture validation tests

## Architecture Notes

This package has been migrated from a mutex-based architecture to a lock-free actor pattern:

- Individual tables are managed by `TableActor` instances
- Table management uses `ActorTableManager` with minimal synchronization
- Rate limiting uses `ActorRateLimiter` for thread-safe operations
- All table-level operations are lock-free, preventing deadlocks

## Testing

Tests follow Go conventions and are kept in the same directory as source code. Test files are grouped by the component they test and can be run individually or as a complete suite.
