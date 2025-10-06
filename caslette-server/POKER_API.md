# Multiplayer Poker WebSocket API Documentation

## Overview

This document describes the complete WebSocket API for the multiplayer poker system. The backend provides a full-featured Texas Hold'em poker game with table management, real-time gameplay, and observer functionality.

## Connection

Connect to: `ws://localhost:8081/ws`

All messages require authentication. Send an auth message first:

```json
{
  "type": "auth",
  "request_id": "unique_request_id",
  "data": {
    "token": "your_jwt_token"
  }
}
```

## Message Format

All messages follow this format:

```json
{
  "type": "message_type",
  "request_id": "unique_request_id",
  "data": {
    /* message specific data */
  },
  "success": true,
  "error": "error message if any"
}
```

## Table Management API

### Create Table

Create a new poker table.

**Request:**

```json
{
  "type": "table_create",
  "request_id": "req123",
  "data": {
    "name": "High Stakes Table",
    "game_type": "texas_holdem",
    "description": "High stakes Texas Hold'em",
    "tags": ["high-stakes", "texas-holdem"],
    "settings": {
      "small_blind": 10,
      "big_blind": 20,
      "buy_in": 1000,
      "max_buy_in": 2000,
      "auto_start": false,
      "time_limit": 30,
      "tournament_mode": false,
      "observers_allowed": true,
      "private": false,
      "password": ""
    }
  }
}
```

**Response:**

```json
{
  "type": "table_created",
  "request_id": "req123",
  "success": true,
  "data": {
    "id": "table_uuid",
    "name": "High Stakes Table",
    "game_type": "texas_holdem",
    "status": "waiting",
    "created_by": "user_id",
    "max_players": 8,
    "min_players": 2,
    "player_slots": [],
    "observers": [],
    "settings": {
      /* table settings */
    },
    "room_id": "room_uuid"
  }
}
```

### Join Table

Join a table as a player or observer.

**Request:**

```json
{
  "type": "table_join",
  "request_id": "req124",
  "data": {
    "table_id": "table_uuid",
    "mode": "player", // or "observer"
    "seat": 1 // optional, specific seat number
  }
}
```

**Response:**

```json
{
  "type": "table_joined",
  "request_id": "req124",
  "success": true,
  "data": {
    "table": {
      /* updated table info */
    },
    "mode": "player"
  }
}
```

### Leave Table

Leave a table.

**Request:**

```json
{
  "type": "table_leave",
  "request_id": "req125",
  "data": {
    "table_id": "table_uuid"
  }
}
```

### List Tables

Get list of available tables.

**Request:**

```json
{
  "type": "table_list",
  "request_id": "req126",
  "data": {
    "status": "waiting", // optional filter
    "game_type": "texas_holdem", // optional filter
    "limit": 20 // optional
  }
}
```

**Response:**

```json
{
  "type": "table_list",
  "request_id": "req126",
  "success": true,
  "data": [
    {
      "id": "table_uuid",
      "name": "Table Name",
      "game_type": "texas_holdem",
      "status": "waiting",
      "player_count": 2,
      "max_players": 8,
      "settings": {
        /* basic settings */
      }
    }
  ]
}
```

### Get Table Info

Get detailed information about a specific table.

**Request:**

```json
{
  "type": "table_get",
  "request_id": "req127",
  "data": {
    "table_id": "table_uuid"
  }
}
```

### Set Ready Status

Set player ready status.

**Request:**

```json
{
  "type": "table_set_ready",
  "request_id": "req128",
  "data": {
    "table_id": "table_uuid",
    "ready": true
  }
}
```

### Start Game

Manually start a game (table creator only).

**Request:**

```json
{
  "type": "table_start_game",
  "request_id": "req129",
  "data": {
    "table_id": "table_uuid"
  }
}
```

### Close Table

Close a table (creator only).

**Request:**

```json
{
  "type": "table_close",
  "request_id": "req130",
  "data": {
    "table_id": "table_uuid"
  }
}
```

## Game Play API

### Poker Actions

Perform poker actions during gameplay.

**Request:**

```json
{
  "type": "poker_action",
  "request_id": "req131",
  "data": {
    "table_id": "table_uuid",
    "action": "fold", // fold, call, raise, check, bet, all_in
    "amount": 100 // for raise/bet actions
  }
}
```

**Response:**

```json
{
  "type": "poker_action_response",
  "request_id": "req131",
  "success": true,
  "data": {
    "action": "fold",
    "processed": true,
    "event": {
      "type": "player_folded",
      "player_id": "user_id",
      "data": {
        /* action details */
      },
      "timestamp": "2025-10-06T..."
    }
  }
}
```

### Get Game State

Get current state of the game.

**Request:**

```json
{
  "type": "get_game_state",
  "request_id": "req132",
  "data": {
    "table_id": "table_uuid"
  }
}
```

**Response:**

```json
{
  "type": "game_state_response",
  "request_id": "req132",
  "success": true,
  "data": {
    "game_id": "game_uuid",
    "state": "inprogress",
    "current_turn": 1,
    "players": [
      {
        "id": "player1",
        "name": "Alice",
        "position": 0,
        "chips": 980,
        "current_bet": 20,
        "has_folded": false,
        "is_all_in": false,
        "hand": {
          "cards": [
            /* only visible to player */
          ]
        }
      }
    ],
    "community_cards": [
      { "suit": "hearts", "rank": "A" },
      { "suit": "spades", "rank": "K" },
      { "suit": "diamonds", "rank": "Q" }
    ],
    "pot": 60,
    "current_bet": 20,
    "round_state": "flop",
    "action_pos": 1
  }
}
```

### Get Hand History

Get hand history for a table.

**Request:**

```json
{
  "type": "get_hand_history",
  "request_id": "req133",
  "data": {
    "table_id": "table_uuid",
    "limit": 10
  }
}
```

### Get Player Stats

Get player statistics.

**Request:**

```json
{
  "type": "get_player_stats",
  "request_id": "req134",
  "data": {
    "table_id": "table_uuid",
    "player_id": "user_id" // optional, defaults to requesting user
  }
}
```

### Join Table Room

Join table room for real-time updates.

**Request:**

```json
{
  "type": "join_table_room",
  "request_id": "req135",
  "data": {
    "table_id": "table_uuid"
  }
}
```

## Real-time Events (Broadcasts)

These events are broadcasted to all users in a table room:

### Player Joined

```json
{
  "type": "player_joined",
  "data": {
    "player_id": "user_id",
    "username": "Alice",
    "mode": "player",
    "table": {
      /* updated table info */
    }
  }
}
```

### Player Left

```json
{
  "type": "player_left",
  "data": {
    "player_id": "user_id",
    "mode": "player",
    "table": {
      /* updated table info */
    }
  }
}
```

### Game Started

```json
{
  "type": "game_started",
  "data": {
    "table_id": "table_uuid",
    "table": {
      /* updated table info */
    }
  }
}
```

### Game Event

```json
{
  "type": "game_event",
  "data": {
    "table_id": "table_uuid",
    "event": {
      "type": "player_action", // card_dealt, round_started, etc.
      "player_id": "user_id",
      "data": {
        /* event specific data */
      },
      "timestamp": "2025-10-06T..."
    }
  }
}
```

### Player Ready Changed

```json
{
  "type": "player_ready_changed",
  "data": {
    "player_id": "user_id",
    "position": 0,
    "ready": true
  }
}
```

### Table Closed

```json
{
  "type": "table_closed",
  "data": {
    "table_id": "table_uuid",
    "reason": "closed"
  }
}
```

## Error Handling

Errors are returned in the standard message format:

```json
{
  "type": "error",
  "request_id": "req123",
  "success": false,
  "error": "[ERROR_CODE] Error description"
}
```

Common error codes:

- `INVALID_DATA`: Invalid request data
- `NOT_AUTHORIZED`: User not authorized for action
- `TABLE_NOT_FOUND`: Table doesn't exist
- `JOIN_FAILED`: Failed to join table
- `INVALID_ACTION`: Invalid poker action
- `NOT_AT_TABLE`: Player not at the table
- `GAME_NOT_ACTIVE`: Game is not currently active

## Authentication

All requests require authentication. Include JWT token in auth message after connection:

```json
{
  "type": "auth",
  "request_id": "auth_req",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

## Game Flow Example

1. **Connect and authenticate**
2. **Create or join table**
3. **Set ready status**
4. **Start game** (when all players ready)
5. **Receive game state updates**
6. **Perform poker actions** in turn
7. **Receive real-time game events**
8. **Game completes and returns to waiting state**

## Frontend Integration Tips

1. **Maintain WebSocket connection** throughout the session
2. **Listen for broadcasts** after joining table room
3. **Handle connection drops** with reconnection logic
4. **Cache table state** locally and update with events
5. **Validate actions** on frontend before sending
6. **Show loading states** during async operations
7. **Display real-time updates** from broadcast events

## Testing

Use the provided integration test files:

- `poker_integration_test.go` - Full integration test suite
- `validate_poker_backend.go` - Basic validation script

Run validation: `go run validate_poker_backend.go`
Run tests: `go test -v ./poker_integration_test.go`

## Next Steps for Frontend

1. Implement WebSocket connection management
2. Create table list/lobby UI
3. Build table/game UI components
4. Implement poker action buttons
5. Add real-time game state visualization
6. Handle observer mode
7. Add hand history display
8. Implement player statistics

The backend is fully functional and ready for frontend integration!
