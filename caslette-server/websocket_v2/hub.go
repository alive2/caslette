package websocket_v2

import (
	"context"
	"encoding/json"
	"log"
	"sync"
)

// Hub maintains the set of active connections and broadcasts messages to them
type Hub struct {
	// Registered connections
	Connections map[string]*Connection

	// Room management
	Rooms map[string]map[string]*Connection // room -> connectionID -> connection

	// User mapping
	Users map[string]*Connection // userID -> connection

	// Registration requests
	Register chan *Connection

	// Unregistration requests
	Unregister chan *Connection

	// Message handlers
	MessageHandlers map[string]MessageHandler

	// Authentication handler
	AuthHandler AuthHandler

	// Mutex for thread safety
	mu sync.RWMutex
}

// MessageHandler defines the signature for message handlers
type MessageHandler func(ctx context.Context, conn *Connection, msg *Message) *Message

// AuthHandler defines the signature for authentication
type AuthHandler func(token string) (*AuthResult, error)

// AuthResult contains authentication result
type AuthResult struct {
	UserID   string
	Username string
	Success  bool
	Error    string
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		Connections:     make(map[string]*Connection),
		Rooms:           make(map[string]map[string]*Connection),
		Users:           make(map[string]*Connection),
		Register:        make(chan *Connection),
		Unregister:      make(chan *Connection),
		MessageHandlers: make(map[string]MessageHandler),
	}
}

// Start begins the hub's main loop
func (h *Hub) Start() {
	for {
		select {
		case conn := <-h.Register:
			h.registerConnection(conn)

		case conn := <-h.Unregister:
			h.unregisterConnection(conn)
		}
	}
}

// RegisterMessageHandler registers a handler for a specific message type
func (h *Hub) RegisterMessageHandler(messageType string, handler MessageHandler) {
	h.MessageHandlers[messageType] = handler
}

// SetAuthHandler sets the authentication handler
func (h *Hub) SetAuthHandler(handler AuthHandler) {
	h.AuthHandler = handler
}

// ProcessMessage processes an incoming message from a connection
func (h *Hub) ProcessMessage(conn *Connection, msg *Message) {
	ctx := context.Background()

	log.Printf("Processing message type: %s from connection %s (UserID: %s)", msg.Type, conn.ID, conn.UserID)

	// Handle authentication messages
	if msg.Type == "auth" {
		h.handleAuth(conn, msg)
		return
	}

	// Handle built-in message types
	switch msg.Type {
	case "test_echo":
		// Simple echo test
		log.Printf("Received test_echo, sending test_echo_response")
		response := &Message{
			Type:      "test_echo_response",
			RequestID: msg.RequestID,
			Success:   true,
			Data:      "echo received",
		}
		conn.SendMessage(response)
		return
	case "create_room":
		h.handleCreateRoom(conn, msg)
		return
	case "join_room":
		h.handleJoinRoom(conn, msg)
		return
	case "leave_room":
		h.handleLeaveRoom(conn, msg)
		return
	case "list_rooms":
		h.handleListRooms(conn, msg)
		return
	case "send_to_room":
		h.handleSendToRoom(conn, msg)
		return
	case "ping":
		h.handlePing(conn, msg)
		return
	}

	// Handle custom message types
	if handler, exists := h.MessageHandlers[msg.Type]; exists {
		response := handler(ctx, conn, msg)
		if response != nil {
			conn.SendMessage(response)
		}
	} else {
		// Unknown message type
		response := &Message{
			Type:      "error",
			RequestID: msg.RequestID,
			Error:     "Unknown message type: " + msg.Type,
			Success:   false,
		}
		conn.SendMessage(response)
	}
}

// BroadcastToRoom sends a message to all connections in a room
func (h *Hub) BroadcastToRoom(room string, msg *Message) {
	h.mu.RLock()
	roomConnections, exists := h.Rooms[room]
	h.mu.RUnlock()

	if !exists {
		return
	}

	for _, conn := range roomConnections {
		conn.SendMessage(msg)
	}
}

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID string, msg *Message) {
	h.mu.RLock()
	conn, exists := h.Users[userID]
	h.mu.RUnlock()

	if exists {
		conn.SendMessage(msg)
	}
}

// BroadcastToAll sends a message to all authenticated connections
func (h *Hub) BroadcastToAll(msg *Message) {
	h.mu.RLock()
	connections := make([]*Connection, 0, len(h.Users))
	for _, conn := range h.Users {
		connections = append(connections, conn)
	}
	h.mu.RUnlock()

	for _, conn := range connections {
		conn.SendMessage(msg)
	}
}

// JoinRoom adds a connection to a room
func (h *Hub) JoinRoom(connectionID, room string) {
	h.mu.Lock()

	conn, exists := h.Connections[connectionID]
	if !exists {
		h.mu.Unlock()
		return
	}

	if h.Rooms[room] == nil {
		h.Rooms[room] = make(map[string]*Connection)
	}
	h.Rooms[room][connectionID] = conn

	// Create the notification message while holding the lock
	msg := &Message{
		Type:  "user_joined_room",
		Event: "user_joined",
		Data: map[string]interface{}{
			"userID":   conn.UserID,
			"username": conn.Username,
			"room":     room,
		},
		Room: room,
	}

	h.mu.Unlock() // Release lock before broadcasting

	log.Printf("Connection %s (%s) joined room %s", connectionID, conn.Username, room)

	// Notify others in the room (outside the mutex lock)
	h.BroadcastToRoom(room, msg)
}

// LeaveRoom removes a connection from a room
func (h *Hub) LeaveRoom(connectionID, room string) {
	h.mu.Lock()

	conn, exists := h.Connections[connectionID]
	if !exists {
		h.mu.Unlock()
		return
	}

	if h.Rooms[room] != nil {
		delete(h.Rooms[room], connectionID)
		if len(h.Rooms[room]) == 0 {
			delete(h.Rooms, room)
		}
	}

	// Create the notification message while holding the lock
	msg := &Message{
		Type:  "user_left_room",
		Event: "user_left",
		Data: map[string]interface{}{
			"userID":   conn.UserID,
			"username": conn.Username,
			"room":     room,
		},
		Room: room,
	}

	h.mu.Unlock() // Release lock before broadcasting

	log.Printf("Connection %s (%s) left room %s", connectionID, conn.Username, room)

	// Notify others in the room (outside the mutex lock)
	h.BroadcastToRoom(room, msg)
}

// GetConnectedUsers returns a list of all connected users
func (h *Hub) GetConnectedUsers() []map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]map[string]interface{}, 0, len(h.Users))
	for userID, conn := range h.Users {
		users = append(users, map[string]interface{}{
			"userID":       userID,
			"username":     conn.Username,
			"connectionID": conn.ID,
			"rooms":        conn.GetRooms(),
		})
	}
	return users
}

// GetRoomUsers returns users in a specific room
func (h *Hub) GetRoomUsers(room string) []map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	roomConnections, exists := h.Rooms[room]
	if !exists {
		return []map[string]interface{}{}
	}

	users := make([]map[string]interface{}, 0, len(roomConnections))
	for _, conn := range roomConnections {
		users = append(users, map[string]interface{}{
			"userID":       conn.UserID,
			"username":     conn.Username,
			"connectionID": conn.ID,
		})
	}
	return users
}

// GetConnectionCount returns the total number of connections
func (h *Hub) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Connections)
}

// registerConnection registers a new connection
func (h *Hub) registerConnection(conn *Connection) {
	h.mu.Lock()
	h.Connections[conn.ID] = conn
	h.mu.Unlock()

	log.Printf("Connection %s registered", conn.ID)

	// Send welcome message
	welcome := &Message{
		Type:  "connected",
		Event: "welcome",
		Data: map[string]interface{}{
			"connectionID": conn.ID,
			"message":      "Connected to Caslette WebSocket server",
		},
	}
	conn.SendMessage(welcome)
}

// unregisterConnection unregisters a connection
func (h *Hub) unregisterConnection(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.Connections[conn.ID]; exists {
		// Remove from connections
		delete(h.Connections, conn.ID)

		// Remove from user mapping
		if conn.UserID != "" {
			delete(h.Users, conn.UserID)
		}

		// Remove from all rooms
		for room := range conn.Rooms {
			if h.Rooms[room] != nil {
				delete(h.Rooms[room], conn.ID)
				if len(h.Rooms[room]) == 0 {
					delete(h.Rooms, room)
				}
			}
		}

		log.Printf("Connection %s (%s) unregistered", conn.ID, conn.Username)
	}
}

// handleAuth handles authentication messages
func (h *Hub) handleAuth(conn *Connection, msg *Message) {
	log.Printf("handleAuth called for connection %s", conn.ID)

	if h.AuthHandler == nil {
		log.Printf("AuthHandler is nil")
		response := &Message{
			Type:      "auth_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Authentication not configured",
		}
		conn.SendMessage(response)
		return
	}

	log.Printf("Received auth message data: %+v", msg.Data)

	var authMsg AuthMessage
	if dataBytes, err := jsonMarshal(msg.Data); err == nil {
		log.Printf("Marshaled data: %s", string(dataBytes))
		if err := jsonUnmarshal(dataBytes, &authMsg); err != nil {
			log.Printf("Failed to unmarshal auth message: %v", err)
			response := &Message{
				Type:      "auth_response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Invalid auth message format",
			}
			conn.SendMessage(response)
			return
		}
	} else {
		log.Printf("Failed to marshal message data: %v", err)
		response := &Message{
			Type:      "auth_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid message data",
		}
		conn.SendMessage(response)
		return
	}

	log.Printf("Extracted token: %s", authMsg.Token)

	authResult, err := h.AuthHandler(authMsg.Token)
	if err != nil {
		log.Printf("AuthHandler returned error: %v", err)
		response := &Message{
			Type:      "auth_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     err.Error(),
		}
		conn.SendMessage(response)
		return
	}

	log.Printf("AuthHandler result: %+v", authResult)
	if err != nil {
		response := &Message{
			Type:      "auth_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     err.Error(),
		}
		conn.SendMessage(response)
		return
	}

	if authResult.Success {
		// Update connection with user info
		conn.UserID = authResult.UserID
		conn.Username = authResult.Username

		// Add to user mapping
		h.mu.Lock()
		h.Users[authResult.UserID] = conn
		h.mu.Unlock()

		response := &Message{
			Type:      "auth_response",
			RequestID: msg.RequestID,
			Success:   true,
			Data: map[string]interface{}{
				"userID":   authResult.UserID,
				"username": authResult.Username,
			},
		}
		conn.SendMessage(response)

		log.Printf("User %s (%s) authenticated on connection %s", authResult.UserID, authResult.Username, conn.ID)
	} else {
		response := &Message{
			Type:      "auth_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     authResult.Error,
		}
		conn.SendMessage(response)
	}
}

// handleJoinRoom handles room join requests
func (h *Hub) handleJoinRoom(conn *Connection, msg *Message) {
	log.Printf("handleJoinRoom called - msg.Data: %+v, RequestID: %s", msg.Data, msg.RequestID)

	room, ok := msg.Data.(string)
	if !ok {
		if dataMap, ok := msg.Data.(map[string]interface{}); ok {
			if roomStr, ok := dataMap["room"].(string); ok {
				room = roomStr
			}
		}
	}

	if room == "" {
		log.Printf("handleJoinRoom: Invalid room name, sending error response")
		response := &Message{
			Type:      "join_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Room name is required",
		}
		conn.SendMessage(response)
		return
	}

	log.Printf("handleJoinRoom: About to join room '%s'", room)
	log.Printf("handleJoinRoom: Calling conn.JoinRoom(%s)...", room)
	conn.JoinRoom(room)
	log.Printf("handleJoinRoom: conn.JoinRoom returned successfully")
	log.Printf("handleJoinRoom: Successfully joined room '%s'", room)

	response := &Message{
		Type:      "join_room_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data: map[string]interface{}{
			"room":  room,
			"users": h.GetRoomUsers(room),
		},
	}
	log.Printf("Sending join_room_response: RequestID=%s, Success=%t, Room=%s", msg.RequestID, true, room)
	conn.SendMessage(response)
	log.Printf("handleJoinRoom: Response sent for RequestID=%s", msg.RequestID)
}

// handleLeaveRoom handles room leave requests
func (h *Hub) handleLeaveRoom(conn *Connection, msg *Message) {
	log.Printf("handleLeaveRoom called - msg.Data: %+v", msg.Data)

	room, ok := msg.Data.(string)
	if !ok {
		if dataMap, ok := msg.Data.(map[string]interface{}); ok {
			if roomStr, ok := dataMap["room"].(string); ok {
				room = roomStr
			}
		}
	}

	log.Printf("Extracted room name for leave: '%s'", room)

	if room == "" {
		response := &Message{
			Type:      "leave_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Room name is required",
		}
		conn.SendMessage(response)
		return
	}

	conn.LeaveRoom(room)

	response := &Message{
		Type:      "leave_room_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data: map[string]interface{}{
			"room": room,
		},
	}
	conn.SendMessage(response)
}

// handlePing handles ping messages
func (h *Hub) handlePing(conn *Connection, msg *Message) {
	response := &Message{
		Type:      "pong",
		RequestID: msg.RequestID,
		Success:   true,
		Data: map[string]interface{}{
			"timestamp": msg.Timestamp,
		},
	}
	conn.SendMessage(response)
}

// handleCreateRoom handles room creation requests
func (h *Hub) handleCreateRoom(conn *Connection, msg *Message) {
	log.Printf("handleCreateRoom called - msg.Data: %+v", msg.Data)

	room, ok := msg.Data.(string)
	if !ok {
		if dataMap, ok := msg.Data.(map[string]interface{}); ok {
			if roomStr, ok := dataMap["room"].(string); ok {
				room = roomStr
			}
		}
	}

	log.Printf("Extracted room name: '%s'", room)

	if room == "" {
		log.Printf("handleCreateRoom: Empty room name, sending error response")
		response := &Message{
			Type:      "create_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Room name is required",
		}
		conn.SendMessage(response)
		return
	}

	// Check if user is authenticated
	if conn.UserID == "" {
		log.Printf("handleCreateRoom: User not authenticated, sending error response")
		response := &Message{
			Type:      "create_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Authentication required to create rooms",
		}
		conn.SendMessage(response)
		return
	}

	log.Printf("handleCreateRoom: User authenticated (UserID: %s), proceeding with room creation", conn.UserID)

	h.mu.Lock()
	// Check if room already exists
	if _, exists := h.Rooms[room]; exists {
		h.mu.Unlock()
		log.Printf("handleCreateRoom: Room '%s' already exists, sending error response", room)
		response := &Message{
			Type:      "create_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Room already exists",
		}
		conn.SendMessage(response)
		return
	}

	// Create the room
	h.Rooms[room] = make(map[string]*Connection)
	h.mu.Unlock() // Release mutex before joining room to avoid deadlock

	log.Printf("Room created: %s by user %s", room, conn.UserID)

	// Don't auto-join for now - let client join explicitly
	log.Printf("handleCreateRoom: Skipping auto-join, preparing response")

	// Send success response
	log.Printf("handleCreateRoom: Preparing success response")
	response := &Message{
		Type:      "create_room_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data: map[string]interface{}{
			"room":    room,
			"creator": conn.UserID,
			"message": "Room created successfully",
			"joined":  false, // Not auto-joined
		},
	}
	log.Printf("handleCreateRoom: About to send create_room_response for RequestID=%s", msg.RequestID)
	conn.SendMessage(response)
	log.Printf("handleCreateRoom: create_room_response sent for RequestID=%s", msg.RequestID)

	// Broadcast room creation to all authenticated users
	log.Printf("handleCreateRoom: About to broadcast room_created event")
	h.BroadcastToAll(&Message{
		Type: "room_created",
		Data: map[string]interface{}{
			"room":    room,
			"creator": conn.Username,
		},
	})
	log.Printf("handleCreateRoom: Completed successfully")
}

// handleListRooms handles room listing requests
func (h *Hub) handleListRooms(conn *Connection, msg *Message) {
	log.Printf("handleListRooms called from connection %s", conn.ID)

	h.mu.RLock()
	roomList := make([]map[string]interface{}, 0, len(h.Rooms))

	for roomName, roomConnections := range h.Rooms {
		userList := make([]string, 0, len(roomConnections))
		for _, roomConn := range roomConnections {
			if roomConn.Username != "" {
				userList = append(userList, roomConn.Username)
			}
		}

		roomInfo := map[string]interface{}{
			"name":      roomName,
			"userCount": len(roomConnections),
			"users":     userList,
		}
		roomList = append(roomList, roomInfo)
	}
	h.mu.RUnlock()

	response := &Message{
		Type:      "list_rooms_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data: map[string]interface{}{
			"rooms": roomList,
			"total": len(roomList),
		},
	}
	conn.SendMessage(response)
}

// handleSendToRoom handles sending messages to a specific room
func (h *Hub) handleSendToRoom(conn *Connection, msg *Message) {
	log.Printf("handleSendToRoom called - msg.Data: %+v", msg.Data)

	// Extract room and message from data
	dataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		log.Printf("handleSendToRoom: Invalid data format")
		response := &Message{
			Type:      "send_to_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid data format",
		}
		conn.SendMessage(response)
		return
	}

	room, ok := dataMap["room"].(string)
	if !ok || room == "" {
		log.Printf("handleSendToRoom: Missing or invalid room name")
		response := &Message{
			Type:      "send_to_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Room name is required",
		}
		conn.SendMessage(response)
		return
	}

	message := dataMap["message"]
	if message == nil {
		log.Printf("handleSendToRoom: Missing message")
		response := &Message{
			Type:      "send_to_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Message is required",
		}
		conn.SendMessage(response)
		return
	}

	// Check if user is authenticated
	if conn.UserID == "" {
		log.Printf("handleSendToRoom: User not authenticated")
		response := &Message{
			Type:      "send_to_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Authentication required to send messages",
		}
		conn.SendMessage(response)
		return
	}

	// Check if user is in the room
	if !conn.IsInRoom(room) {
		log.Printf("handleSendToRoom: User %s not in room %s", conn.UserID, room)
		response := &Message{
			Type:      "send_to_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "You must join the room before sending messages",
		}
		conn.SendMessage(response)
		return
	}

	log.Printf("handleSendToRoom: Broadcasting message to room %s from user %s", room, conn.Username)

	// Broadcast the message to all users in the room
	roomMessage := &Message{
		Type: "room_message",
		Data: map[string]interface{}{
			"room":     room,
			"message":  message,
			"sender":   conn.Username,
			"senderID": conn.UserID,
		},
	}
	h.BroadcastToRoom(room, roomMessage)

	// Send success response to sender
	response := &Message{
		Type:      "send_to_room_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data: map[string]interface{}{
			"room":    room,
			"message": "Message sent successfully",
		},
	}
	log.Printf("handleSendToRoom: Sending success response for RequestID=%s", msg.RequestID)
	conn.SendMessage(response)
	log.Printf("handleSendToRoom: Completed successfully")
}

// Helper functions for JSON marshaling
func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
