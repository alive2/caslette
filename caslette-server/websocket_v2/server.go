package websocket_v2

import (
	"caslette-server/auth"
	"context"
	"encoding/json"
	"log"
	"net/http"
)

// Server wraps the WebSocket hub with additional functionality
type Server struct {
	hub         HubInterface
	authService *auth.AuthService
}

// NewServer creates a new WebSocket server
func NewServer(authService *auth.AuthService) *Server {
	hub := NewActorHub()
	server := &Server{
		hub:         hub,
		authService: authService,
	}

	// Set up authentication handler once
	hub.SetAuthHandler(CreateWebSocketAuthHandler(authService))
	log.Printf("WebSocket server created with authentication handler")

	// Register built-in handlers
	server.registerBuiltinHandlers()

	return server
}

// Run starts the hub (should be called in a goroutine)
func (s *Server) Run() {
	s.hub.Start()
}

// ServeHTTP implements http.Handler for Gin integration
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.HandleWebSocket(w, r)
}

// GetHub returns the underlying hub
func (s *Server) GetHub() HubInterface {
	return s.hub
}

// GetConnectionCount returns the number of connected clients
func (s *Server) GetConnectionCount() int {
	return s.hub.GetConnectionCount()
}

// GetConnectedUsers returns a map of connected users
func (s *Server) GetConnectedUsers() map[string]string {
	// For now, return empty map since we don't have direct access to connections
	// This would need to be implemented as a method in the HubInterface if needed
	return make(map[string]string)
}

// GetActiveRooms returns a list of active room names
func (s *Server) GetActiveRooms() []string {
	// For now, return empty slice since we don't have direct access to rooms
	// This would need to be implemented as a method in the HubInterface if needed
	return []string{}
}

// HandleWebSocket handles WebSocket connections
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := NewConnection(s.hub, w, r)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	log.Printf("New WebSocket connection established: %s", conn.ID)

	// Register the connection
	s.hub.Register(conn)

	// Start the connection
	conn.Start()
}

// RegisterHandler registers a custom message handler
func (s *Server) RegisterHandler(messageType string, handler MessageHandler) {
	s.hub.RegisterMessageHandler(messageType, handler)
}

// SetAuthHandler sets the authentication handler
func (s *Server) SetAuthHandler(handler AuthHandler) {
	s.hub.SetAuthHandler(handler)
}

// BroadcastToRoom broadcasts a message to all users in a room
func (s *Server) BroadcastToRoom(room, messageType string, data interface{}) {
	msg := &Message{
		Type: messageType,
		Data: data,
		Room: room,
	}
	s.hub.BroadcastToRoom(room, msg)
}

// BroadcastToUser sends a message to a specific user
func (s *Server) BroadcastToUser(userID, messageType string, data interface{}) {
	msg := &Message{
		Type: messageType,
		Data: data,
	}
	s.hub.BroadcastToUser(userID, msg)
}

// GetRoomUsers returns users in a specific room
func (s *Server) GetRoomUsers(room string) []map[string]interface{} {
	// For now, return empty slice since this would need to be implemented
	// as a method in the HubInterface if needed
	return []map[string]interface{}{}
}

// registerBuiltinHandlers registers built-in message handlers
func (s *Server) registerBuiltinHandlers() {
	// Echo handler for testing
	s.RegisterHandler("echo", func(ctx context.Context, conn *Connection, msg *Message) *Message {
		return &Message{
			Type:      "echo_response",
			RequestID: msg.RequestID,
			Success:   true,
			Data:      msg.Data,
		}
	})

	// Get room info handler
	s.RegisterHandler("get_room_info", func(ctx context.Context, conn *Connection, msg *Message) *Message {
		room, ok := msg.Data.(string)
		if !ok {
			if dataMap, ok := msg.Data.(map[string]interface{}); ok {
				if roomStr, ok := dataMap["room"].(string); ok {
					room = roomStr
				}
			}
		}

		if room == "" {
			return &Message{
				Type:      "get_room_info_response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Room name is required",
			}
		}

		return &Message{
			Type:      "get_room_info_response",
			RequestID: msg.RequestID,
			Success:   true,
			Data: map[string]interface{}{
				"room":      room,
				"users":     s.GetRoomUsers(room),
				"userCount": len(s.GetRoomUsers(room)),
			},
		}
	})

	// Get user list handler
	s.RegisterHandler("get_users", func(ctx context.Context, conn *Connection, msg *Message) *Message {
		return &Message{
			Type:      "get_users_response",
			RequestID: msg.RequestID,
			Success:   true,
			Data: map[string]interface{}{
				"users":            s.GetConnectedUsers(),
				"totalUsers":       len(s.GetConnectedUsers()),
				"totalConnections": s.GetConnectionCount(),
			},
		}
	})

	// Send message to room handler
	s.RegisterHandler("send_to_room", func(ctx context.Context, conn *Connection, msg *Message) *Message {
		var data map[string]interface{}
		if dataBytes, err := json.Marshal(msg.Data); err == nil {
			json.Unmarshal(dataBytes, &data)
		}

		room, ok := data["room"].(string)
		if !ok || room == "" {
			return &Message{
				Type:      "send_to_room_response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Room name is required",
			}
		}

		message, ok := data["message"]
		if !ok {
			return &Message{
				Type:      "send_to_room_response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Message is required",
			}
		}

		// Check if user is in the room
		if !conn.IsInRoom(room) {
			return &Message{
				Type:      "send_to_room_response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "You are not in this room",
			}
		}

		// Broadcast the message to the room
		broadcastMsg := &Message{
			Type: "room_message",
			Data: map[string]interface{}{
				"room":     room,
				"message":  message,
				"userID":   conn.UserID,
				"username": conn.Username,
			},
			Room: room,
		}
		s.hub.BroadcastToRoom(room, broadcastMsg)

		return &Message{
			Type:      "send_to_room_response",
			RequestID: msg.RequestID,
			Success:   true,
			Data: map[string]interface{}{
				"room":    room,
				"message": message,
			},
		}
	})

	// Request-response pattern handler
	s.RegisterHandler("request", func(ctx context.Context, conn *Connection, msg *Message) *Message {
		// This is a generic request handler that other handlers can override
		// Extract the action from the data
		var data map[string]interface{}
		if dataBytes, err := json.Marshal(msg.Data); err == nil {
			json.Unmarshal(dataBytes, &data)
		}

		action, ok := data["action"].(string)
		if !ok {
			return &Message{
				Type:      "response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Action is required",
			}
		}

		// Route to specific action handlers
		switch action {
		case "ping":
			return &Message{
				Type:      "response",
				RequestID: msg.RequestID,
				Success:   true,
				Data: map[string]interface{}{
					"action": "pong",
					"time":   msg.Timestamp,
				},
			}
		default:
			return &Message{
				Type:      "response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Unknown action: " + action,
			}
		}
	})
}

// HealthStatus returns server health information
func (s *Server) HealthStatus() map[string]interface{} {
	return map[string]interface{}{
		"status":            "healthy",
		"connected_users":   len(s.GetConnectedUsers()),
		"total_connections": s.GetConnectionCount(),
		"active_rooms":      len(s.GetActiveRooms()),
	}
}
