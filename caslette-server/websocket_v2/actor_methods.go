package websocket_v2

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Actor methods that handle the actual hub operations
// These run in the single actor goroutine, ensuring thread safety

// actorRegisterConnection registers a connection (actor method)
func (h *ActorHub) actorRegisterConnection(conn *Connection, response chan interface{}) {
	h.connections[conn.ID] = conn
	log.Printf("ActorHub: Connection %s registered", conn.ID)

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

	response <- nil
}

// actorUnregisterConnection unregisters a connection (actor method)
func (h *ActorHub) actorUnregisterConnection(conn *Connection, response chan interface{}) {
	if _, exists := h.connections[conn.ID]; exists {
		// Remove from connections
		delete(h.connections, conn.ID)

		// Remove from user mapping
		if conn.UserID != "" {
			delete(h.users, conn.UserID)
		}

		// Remove from all rooms
		for room := range conn.Rooms {
			if h.rooms[room] != nil {
				delete(h.rooms[room], conn.ID)
				if len(h.rooms[room]) == 0 {
					delete(h.rooms, room)
				}
			}
		}

		log.Printf("ActorHub: Connection %s (%s) unregistered", conn.ID, conn.Username)
	}

	response <- nil
}

// actorProcessMessage processes an incoming message (actor method)
func (h *ActorHub) actorProcessMessage(conn *Connection, msg *Message, response chan interface{}) {
	log.Printf("ActorHub: actorProcessMessage started for connection %s, message type: %s", conn.ID, msg.Type)

	// Check rate limiting first - call actor method directly to avoid deadlock
	log.Printf("ActorHub: About to check rate limit for connection %s", conn.ID)
	rateLimitResponse := make(chan interface{}, 1)
	h.actorCheckRateLimit(conn.ID, rateLimitResponse)
	if rateLimitResult := <-rateLimitResponse; rateLimitResult != nil {
		if err, ok := rateLimitResult.(error); ok {
			log.Printf("ActorHub: Rate limit exceeded for connection %s: %v", conn.ID, err)
			errorResponse := &Message{
				Type:      "error",
				RequestID: msg.RequestID,
				Error:     err.Error(),
				Success:   false,
			}
			conn.SendMessage(errorResponse)
			response <- err
			return
		}
	}
	log.Printf("ActorHub: Rate limit check passed for connection %s", conn.ID)

	ctx := context.Background()
	log.Printf("ActorHub: Processing message type: %s from connection %s (UserID: %s)", msg.Type, conn.ID, conn.UserID)

	// Handle authentication messages
	if msg.Type == "auth" {
		h.actorHandleAuth(conn, msg)
		response <- nil
		return
	}

	// Handle built-in message types with input validation
	switch msg.Type {
	case "logout":
		h.actorHandleLogout(conn, msg)
		response <- nil
		return

	case "test_echo":
		log.Printf("ActorHub: Received test_echo, sending test_echo_response")
		echoResponse := &Message{
			Type:      "test_echo_response",
			RequestID: msg.RequestID,
			Success:   true,
			Data:      "echo received",
		}
		conn.SendMessage(echoResponse)
		response <- nil
		return

	case "create_room":
		h.actorHandleCreateRoom(conn, msg)
		response <- nil
		return

	case "join_room":
		h.actorHandleJoinRoom(conn, msg)
		response <- nil
		return

	case "leave_room":
		h.actorHandleLeaveRoom(conn, msg)
		response <- nil
		return

	case "list_rooms":
		h.actorHandleListRooms(conn, msg)
		response <- nil
		return

	default:
		// Check for custom message handlers
		if handler, exists := h.messageHandlers[msg.Type]; exists {
			handlerResponse := handler(ctx, conn, msg)
			if handlerResponse != nil {
				conn.SendMessage(handlerResponse)
			}
			response <- nil
			return
		}

		// Unknown message type
		log.Printf("ActorHub: Unknown message type: %s", msg.Type)
		errorResponse := &Message{
			Type:      "error",
			RequestID: msg.RequestID,
			Error:     "Unknown message type: " + msg.Type,
			Success:   false,
		}
		conn.SendMessage(errorResponse)
		response <- fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// actorHandleAuth handles authentication (actor method)
func (h *ActorHub) actorHandleAuth(conn *Connection, msg *Message) {
	log.Printf("ActorHub: handleAuth called for connection %s", conn.ID)

	if h.authHandler == nil {
		log.Printf("ActorHub: AuthHandler is nil")
		response := &Message{
			Type:      "auth_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Authentication not configured",
		}
		conn.SendMessage(response)
		return
	}

	log.Printf("ActorHub: Received auth message data: %+v", msg.Data)

	var authMsg AuthMessage
	if dataBytes, err := json.Marshal(msg.Data); err == nil {
		log.Printf("ActorHub: Marshaled data: %s", string(dataBytes))
		if err := json.Unmarshal(dataBytes, &authMsg); err != nil {
			log.Printf("ActorHub: Failed to unmarshal auth message: %v", err)
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
		log.Printf("ActorHub: Failed to marshal message data: %v", err)
		response := &Message{
			Type:      "auth_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid message data",
		}
		conn.SendMessage(response)
		return
	}

	log.Printf("ActorHub: Extracted token: %s", authMsg.Token)

	authResult, err := h.authHandler(authMsg.Token)
	if err != nil {
		log.Printf("ActorHub: AuthHandler returned error: %v", err)
		response := &Message{
			Type:      "auth_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     err.Error(),
		}
		conn.SendMessage(response)
		return
	}

	log.Printf("ActorHub: AuthHandler result: %+v", authResult)

	if authResult.Success {
		// Validate username
		validatedUsername, err := validateInput(authResult.Username, "username")
		if err != nil {
			log.Printf("ActorHub: Invalid username: %v", err)
			response := &Message{
				Type:      "auth_response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Invalid username: " + err.Error(),
			}
			conn.SendMessage(response)
			return
		}

		// Update connection with user info
		conn.UserID = authResult.UserID
		conn.Username = validatedUsername

		// Add to user mapping
		h.users[authResult.UserID] = conn

		response := &Message{
			Type:      "auth_response",
			RequestID: msg.RequestID,
			Success:   true,
			Data: map[string]interface{}{
				"userID":   authResult.UserID,
				"username": validatedUsername,
			},
		}
		conn.SendMessage(response)

		log.Printf("ActorHub: User %s (%s) authenticated on connection %s", authResult.UserID, validatedUsername, conn.ID)
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

// actorHandleLogout handles user logout (actor method)
func (h *ActorHub) actorHandleLogout(conn *Connection, msg *Message) {
	log.Printf("ActorHub: handleLogout called for connection %s (UserID: %s)", conn.ID, conn.UserID)

	// Clear user authentication
	if conn.UserID != "" {
		// Remove from user mapping
		delete(h.users, conn.UserID)
		log.Printf("ActorHub: Removed user %s from user mapping", conn.UserID)
	}

	// Clear connection authentication info
	conn.UserID = ""
	conn.Username = ""

	// Send logout response
	response := &Message{
		Type:      "logout_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data:      map[string]interface{}{"message": "Logged out successfully"},
	}
	conn.SendMessage(response)

	log.Printf("ActorHub: User logged out from connection %s", conn.ID)
}

// actorHandleCreateRoom handles room creation (actor method)
func (h *ActorHub) actorHandleCreateRoom(conn *Connection, msg *Message) {
	log.Printf("ActorHub: handleCreateRoom called - msg.Data: %+v", msg.Data)

	// Extract room name
	roomData, ok := msg.Data.(map[string]interface{})
	if !ok {
		log.Printf("ActorHub: Invalid create_room message data format")
		response := &Message{
			Type:      "create_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid message format",
		}
		conn.SendMessage(response)
		return
	}

	roomName, ok := roomData["room"].(string)
	if !ok {
		log.Printf("ActorHub: Room name not provided or invalid type")
		response := &Message{
			Type:      "create_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Room name is required",
		}
		conn.SendMessage(response)
		return
	}

	log.Printf("ActorHub: Extracted room name: '%s'", roomName)

	// Validate and sanitize room name
	validatedRoomName, err := validateInput(roomName, "room")
	if err != nil {
		log.Printf("ActorHub: Invalid room name: %v", err)
		response := &Message{
			Type:      "create_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid room name: " + err.Error(),
		}
		conn.SendMessage(response)
		return
	}

	// Check if user is authenticated
	if conn.UserID == "" {
		log.Printf("ActorHub: User not authenticated, cannot create room")
		response := &Message{
			Type:      "create_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Authentication required to create room",
		}
		conn.SendMessage(response)
		return
	}

	log.Printf("ActorHub: User authenticated (UserID: %s), proceeding with room creation", conn.UserID)

	// Check if room already exists
	if _, exists := h.rooms[validatedRoomName]; exists {
		log.Printf("ActorHub: Room '%s' already exists", validatedRoomName)
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
	h.rooms[validatedRoomName] = make(map[string]*Connection)
	log.Printf("ActorHub: Room created: %s by user %s", validatedRoomName, conn.UserID)

	// Send success response
	response := &Message{
		Type:      "create_room_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data: map[string]interface{}{
			"room":    validatedRoomName,
			"creator": conn.UserID,
			"message": "Room created successfully",
			"joined":  false,
		},
	}
	conn.SendMessage(response)

	// Broadcast room creation event to all authenticated users
	roomCreatedEvent := &Message{
		Type: "room_created",
		Data: map[string]interface{}{
			"room":    validatedRoomName,
			"creator": conn.Username,
		},
	}
	h.actorBroadcastToAll(roomCreatedEvent, nil)

	log.Printf("ActorHub: Room creation completed successfully")
}

// actorHandleJoinRoom handles joining a room (actor method)
func (h *ActorHub) actorHandleJoinRoom(conn *Connection, msg *Message) {
	log.Printf("ActorHub: handleJoinRoom called - msg.Data: %+v, RequestID: %s", msg.Data, msg.RequestID)

	// Extract room name
	roomData, ok := msg.Data.(map[string]interface{})
	if !ok {
		response := &Message{
			Type:      "join_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid message format",
		}
		conn.SendMessage(response)
		return
	}

	roomName, ok := roomData["room"].(string)
	if !ok {
		response := &Message{
			Type:      "join_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Room name is required",
		}
		conn.SendMessage(response)
		return
	}

	// Validate room name
	validatedRoomName, err := validateInput(roomName, "room")
	if err != nil {
		response := &Message{
			Type:      "join_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid room name: " + err.Error(),
		}
		conn.SendMessage(response)
		return
	}

	log.Printf("ActorHub: About to join room '%s'", validatedRoomName)
	h.actorJoinRoom(conn.ID, validatedRoomName, nil)

	// Get room users for response
	users := []map[string]interface{}{}
	if roomConnections, exists := h.rooms[validatedRoomName]; exists {
		for _, roomConn := range roomConnections {
			users = append(users, map[string]interface{}{
				"userID":       roomConn.UserID,
				"username":     roomConn.Username,
				"connectionID": roomConn.ID,
			})
		}
	}

	log.Printf("ActorHub: Sending join_room_response: RequestID=%s, Success=true, Room=%s", msg.RequestID, validatedRoomName)
	response := &Message{
		Type:      "join_room_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data: map[string]interface{}{
			"room":  validatedRoomName,
			"users": users,
		},
	}
	conn.SendMessage(response)
	log.Printf("ActorHub: Response sent for RequestID=%s", msg.RequestID)
}

// actorHandleLeaveRoom handles leaving a room (actor method)
func (h *ActorHub) actorHandleLeaveRoom(conn *Connection, msg *Message) {
	roomData, ok := msg.Data.(map[string]interface{})
	if !ok {
		response := &Message{
			Type:      "leave_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid message format",
		}
		conn.SendMessage(response)
		return
	}

	roomName, ok := roomData["room"].(string)
	if !ok {
		response := &Message{
			Type:      "leave_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Room name is required",
		}
		conn.SendMessage(response)
		return
	}

	// Validate room name
	validatedRoomName, err := validateInput(roomName, "room")
	if err != nil {
		response := &Message{
			Type:      "leave_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid room name: " + err.Error(),
		}
		conn.SendMessage(response)
		return
	}

	h.actorLeaveRoom(conn.ID, validatedRoomName, nil)

	response := &Message{
		Type:      "leave_room_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data: map[string]interface{}{
			"room": validatedRoomName,
		},
	}
	conn.SendMessage(response)
}

// actorHandleListRooms handles listing rooms (actor method)
func (h *ActorHub) actorHandleListRooms(conn *Connection, msg *Message) {
	log.Printf("ActorHub: handleListRooms called from connection %s", conn.ID)

	roomList := []map[string]interface{}{}
	for roomName, roomConnections := range h.rooms {
		usernames := []string{}
		for _, roomConn := range roomConnections {
			if roomConn.Username != "" {
				usernames = append(usernames, roomConn.Username)
			}
		}

		roomList = append(roomList, map[string]interface{}{
			"name":      roomName,
			"userCount": len(roomConnections),
			"users":     usernames,
		})
	}

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

// Actor operations for room management

// actorJoinRoom joins a connection to a room (actor method)
func (h *ActorHub) actorJoinRoom(connectionID, room string, response chan interface{}) {
	conn, exists := h.connections[connectionID]
	if !exists {
		if response != nil {
			response <- fmt.Errorf("connection not found")
		}
		return
	}

	// Validate room name
	validatedRoom, err := validateInput(room, "room")
	if err != nil {
		if response != nil {
			response <- err
		}
		return
	}

	if h.rooms[validatedRoom] == nil {
		h.rooms[validatedRoom] = make(map[string]*Connection)
	}

	h.rooms[validatedRoom][connectionID] = conn
	conn.Rooms[validatedRoom] = true

	log.Printf("ActorHub: Connection %s (%s) joined room %s", connectionID, conn.Username, validatedRoom)

	// Notify other users in the room
	userJoinedEvent := &Message{
		Type:  "user_joined_room",
		Event: "user_joined",
		Room:  validatedRoom,
		Data: map[string]interface{}{
			"userID":   conn.UserID,
			"username": conn.Username,
			"room":     validatedRoom,
		},
	}

	// Send to all connections in the room
	for _, roomConn := range h.rooms[validatedRoom] {
		roomConn.SendMessage(userJoinedEvent)
	}

	if response != nil {
		response <- nil
	}
}

// actorLeaveRoom removes a connection from a room (actor method)
func (h *ActorHub) actorLeaveRoom(connectionID, room string, response chan interface{}) {
	conn, exists := h.connections[connectionID]
	if !exists {
		if response != nil {
			response <- fmt.Errorf("connection not found")
		}
		return
	}

	// Validate room name
	validatedRoom, err := validateInput(room, "room")
	if err != nil {
		if response != nil {
			response <- err
		}
		return
	}

	if h.rooms[validatedRoom] != nil {
		delete(h.rooms[validatedRoom], connectionID)
		delete(conn.Rooms, validatedRoom)

		if len(h.rooms[validatedRoom]) == 0 {
			delete(h.rooms, validatedRoom)
		}

		log.Printf("ActorHub: Connection %s (%s) left room %s", connectionID, conn.Username, validatedRoom)

		// Notify other users in the room
		if len(h.rooms[validatedRoom]) > 0 {
			userLeftEvent := &Message{
				Type:  "user_left_room",
				Event: "user_left",
				Room:  validatedRoom,
				Data: map[string]interface{}{
					"userID":   conn.UserID,
					"username": conn.Username,
					"room":     validatedRoom,
				},
			}

			for _, roomConn := range h.rooms[validatedRoom] {
				roomConn.SendMessage(userLeftEvent)
			}
		}
	}

	if response != nil {
		response <- nil
	}
}

// Broadcasting methods

// actorBroadcastToRoom broadcasts to all connections in a room (actor method)
func (h *ActorHub) actorBroadcastToRoom(room string, msg *Message, response chan interface{}) {
	roomConnections, exists := h.rooms[room]
	if !exists {
		if response != nil {
			response <- nil
		}
		return
	}

	for _, conn := range roomConnections {
		conn.SendMessage(msg)
	}

	if response != nil {
		response <- nil
	}
}

// actorBroadcastToUser broadcasts to a specific user (actor method)
func (h *ActorHub) actorBroadcastToUser(userID string, msg *Message, response chan interface{}) {
	conn, exists := h.users[userID]
	if exists {
		conn.SendMessage(msg)
	}

	if response != nil {
		response <- nil
	}
}

// actorBroadcastToAll broadcasts to all authenticated connections (actor method)
func (h *ActorHub) actorBroadcastToAll(msg *Message, response chan interface{}) {
	for _, conn := range h.users {
		conn.SendMessage(msg)
	}

	if response != nil {
		response <- nil
	}
}

// actorGetConnectionCount returns connection count (actor method)
func (h *ActorHub) actorGetConnectionCount(response chan interface{}) {
	response <- len(h.connections)
}

// actorListRooms returns room information (actor method)
func (h *ActorHub) actorListRooms(response chan interface{}) {
	roomList := make(map[string]int)
	for roomName, roomConnections := range h.rooms {
		roomList[roomName] = len(roomConnections)
	}
	response <- roomList
}

// actorCheckRateLimit performs rate limiting check (actor method)
func (h *ActorHub) actorCheckRateLimit(connectionID string, response chan interface{}) {
	limit := h.rateLimiter.connectionLimits[connectionID]
	now := time.Now()

	if limit == nil {
		// First message from this connection
		h.rateLimiter.connectionLimits[connectionID] = &ConnectionLimit{
			messageCount:    1,
			lastMessageTime: now,
			violations:      0,
			blocked:         false,
		}
		response <- nil
		return
	}

	// Check if connection is currently blocked
	if limit.blocked && now.Before(limit.blockUntil) {
		response <- fmt.Errorf("connection temporarily blocked due to rate limiting")
		return
	}

	// Reset block status if block period has expired
	if limit.blocked && now.After(limit.blockUntil) {
		limit.blocked = false
		limit.violations = 0
		limit.messageCount = 0
	}

	// Check rate limiting
	timeSinceLastMessage := now.Sub(limit.lastMessageTime)
	if timeSinceLastMessage < time.Second {
		limit.messageCount++
		if limit.messageCount > MaxMessagesPerSecond {
			limit.violations++
			log.Printf("Rate limit violation for connection %s (violation %d)", connectionID, limit.violations)

			if limit.violations >= MaxViolations {
				limit.blocked = true
				limit.blockUntil = now.Add(BlockDuration)
				response <- fmt.Errorf("connection blocked for %v due to repeated rate limit violations", BlockDuration)
				return
			}

			response <- fmt.Errorf("rate limit exceeded: max %d messages per second", MaxMessagesPerSecond)
			return
		}
	} else {
		// Reset message count after a second has passed
		limit.messageCount = 1
		limit.lastMessageTime = now
	}

	limit.lastMessageTime = now
	response <- nil
}
