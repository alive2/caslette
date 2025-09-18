# Socket.IO Implementation Documentation

## Overview

This document describes the Socket.IO implementation for the Caslette project, providing real-time communication between the Go server and Flutter client.

## Architecture

### Server Side (Go)

**Location**: `caslette-server/socketio/`

**Components**:

- `manager.go` - Connection management and event handling
- `server.go` - Socket.IO server setup and HTTP integration
- `handlers.go` - Default event handlers and authentication

**Key Features**:

- Connection tracking per user
- Authentication integration
- Request-response pattern support
- Room management for games
- Event broadcasting

### Client Side (Flutter)

**Location**: `caslette-flutter/lib/services/socketio/` and `caslette-flutter/lib/providers/`

**Components**:

- `socket_service.dart` - Core Socket.IO client service
- `models.dart` - Data models for Socket.IO communication
- `socket_provider.dart` - Riverpod providers for state management

**Key Features**:

- Automatic connection management
- Authentication flow integration
- Request-response pattern with timeout handling
- Room join/leave functionality
- Event streaming

## API Reference

### Server Events

#### Authentication

```go
// Event: "authenticate"
// Description: Authenticate user session
// Data: {"token": "jwt_token"}
// Response: {"success": bool, "error": string?, "data": user_info?}
```

#### Request-Response Pattern

```go
// Event: "request"
// Description: Generic request handler
// Data: {"type": "request_type", "data": {}}
// Response: {"type": "response_type", "success": bool, "error": string?, "data": {}}
```

#### Supported Requests

- `ping` - Server ping test
- `get_user_info` - Get current user information
- `join_room` - Join a game room
- `leave_room` - Leave current room

### Client Events

#### Game Events

- `game_update` - Real-time game state updates
- `player_joined` - Player joined notification
- `player_left` - Player left notification
- `room_update` - Room state changes

## Usage Examples

### Flutter Client

```dart
// Initialize Socket.IO service
final socketService = ref.read(socketServiceProvider);

// Connect and authenticate
await socketService.connect();
await socketService.authenticate(userToken);

// Join a game room
final success = await ref.read(gameRoomNotifierProvider.notifier)
    .joinRoom(roomId: 'poker_123', gameType: 'poker');

// Listen to game events
socketService.onEvent('game_update', (event) {
  print('Game update: ${event.data}');
});

// Send a ping request
final response = await socketService.ping();
if (response.success) {
  print('Ping successful: ${response.data}');
}
```

### Go Server

```go
// Register custom event handler
type GameHandler struct{}

func (h *GameHandler) GetEventTypes() []string {
    return []string{"game_action"}
}

func (h *GameHandler) HandleEvent(ctx context.Context, clientID string, eventType string, data []byte) Response {
    // Handle game action
    return Response{
        Type:    "game_response",
        Success: true,
        Data:    map[string]interface{}{"action": "processed"},
    }
}

// Register the handler
socketServer.RegisterHandler(&GameHandler{})

// Broadcast to room
manager.BroadcastToRoom("poker_123", "game_update", gameState)
```

## Configuration

### Server Configuration

**Port**: 8080 (default)
**Endpoint**: `/socket.io/*any`
**Health Check**: `/api/socketio/health`

### Client Configuration

**Server URL**: `http://localhost:8080`
**Auto Connect**: `false` (manual connection after login)
**Reconnection**: Automatic with exponential backoff

## Error Handling

### Connection Errors

- Network failures trigger automatic reconnection
- Authentication failures require re-login
- Room join failures are reported to the UI

### Request Timeouts

- Default timeout: 10 seconds
- Configurable per request type
- Automatic retry for network errors

## Security

### Authentication

- JWT token validation on connection
- User session tracking
- Permission-based room access

### CORS

- Configured for development (`*` allowed)
- Should be restricted in production

## Testing

Use the Socket.IO test widget in the Flutter app:

1. Log in to the application
2. Navigate to the developer icon (⚙️) in the header
3. Test various Socket.IO operations
4. Monitor connection status and event logs

## Development Notes

### Adding New Events

1. **Server**: Add event handler in `handlers.go` or create custom handler
2. **Client**: Add event listener in the service or provider
3. **Models**: Update `models.dart` with new data structures

### Room Management

Rooms are automatically managed:

- Users join rooms via `join_room` request
- Only one room per user at a time
- Automatic cleanup on disconnect

### Performance Considerations

- Connection pooling handled automatically
- Event data should be kept minimal
- Large payloads should use separate API endpoints

## Troubleshooting

### Common Issues

1. **Connection Refused**: Check server is running on port 8080
2. **Authentication Failed**: Verify JWT token is valid
3. **Room Join Failed**: Check user permissions and room existence
4. **Events Not Received**: Verify event listener registration

### Debug Logging

Enable debug logs:

- **Server**: Set log level in main.go
- **Client**: Use Socket.IO test widget for real-time monitoring
