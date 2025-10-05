package websocket_v2

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"log"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

// HubMessage represents different types of messages sent to the Hub actor
type HubMessage struct {
	Type       string
	Connection *Connection
	Message    *Message
	Response   chan interface{}
	Room       string
	UserID     string
	Data       interface{}
}

// ActorHub implements the Hub using actor pattern with goroutines and channels
type ActorHub struct {
	// Actor communication channel
	hubChannel chan HubMessage

	// Internal state (only accessed by the actor goroutine)
	connections map[string]*Connection
	rooms       map[string]map[string]*Connection
	users       map[string]*Connection

	// Message handlers
	messageHandlers map[string]MessageHandler
	authHandler     AuthHandler

	// Rate limiting
	rateLimiter *RateLimiter

	// Connection counter for unique IDs
	connectionCounter int64

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// RateLimiter tracks message rates per connection
type RateLimiter struct {
	connectionLimits map[string]*ConnectionLimit
	globalCounter    int64
	cleanupTicker    *time.Ticker
}

// ConnectionLimit tracks limits for a specific connection
type ConnectionLimit struct {
	messageCount    int64
	lastMessageTime time.Time
	violations      int
	blocked         bool
	blockUntil      time.Time
}

// Rate limiting constants
const (
	MaxMessagesPerSecond = 10
	MaxViolations        = 3
	BlockDuration        = time.Minute * 5
	CleanupInterval      = time.Minute * 10
)

// Input validation patterns
var (
	validRoomName     = regexp.MustCompile(`^[a-zA-Z0-9_\-]{1,50}$`)
	validUsername     = regexp.MustCompile(`^[a-zA-Z0-9_\-]{1,30}$`)
	dangerousPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(script|javascript|onload|onerror)`),
		regexp.MustCompile(`(?i)(drop|delete|insert|update|select|union|--|;)`),
		regexp.MustCompile(`(?i)(\$\(|\$\{|<%|%>)`),
		regexp.MustCompile(`(?i)(exec|eval|system|cmd)`),
	}
)

// NewActorHub creates a new actor-based hub
func NewActorHub() *ActorHub {
	ctx, cancel := context.WithCancel(context.Background())

	hub := &ActorHub{
		hubChannel:        make(chan HubMessage, 1000), // Buffered channel for performance
		connections:       make(map[string]*Connection),
		rooms:             make(map[string]map[string]*Connection),
		users:             make(map[string]*Connection),
		messageHandlers:   make(map[string]MessageHandler),
		connectionCounter: 0,
		ctx:               ctx,
		cancel:            cancel,
		rateLimiter:       newRateLimiter(),
	}

	// Start the actor goroutine
	go hub.actorLoop()

	return hub
}

// newRateLimiter creates a new rate limiter
func newRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		connectionLimits: make(map[string]*ConnectionLimit),
		cleanupTicker:    time.NewTicker(CleanupInterval),
	}

	// Start cleanup goroutine
	go rl.cleanupLoop()

	return rl
}

// actorLoop is the main actor goroutine that processes all hub operations
func (h *ActorHub) actorLoop() {
	log.Printf("ActorHub: Starting actor loop")

	for {
		select {
		case <-h.ctx.Done():
			log.Printf("ActorHub: Shutting down")
			h.rateLimiter.cleanupTicker.Stop()
			return

		case msg := <-h.hubChannel:
			h.handleActorMessage(msg)
		}
	}
}

// handleActorMessage processes a message sent to the actor
func (h *ActorHub) handleActorMessage(msg HubMessage) {
	log.Printf("ActorHub: handleActorMessage called with type: %s", msg.Type)
	switch msg.Type {
	case "register":
		h.actorRegisterConnection(msg.Connection, msg.Response)
	case "unregister":
		h.actorUnregisterConnection(msg.Connection, msg.Response)
	case "process_message":
		log.Printf("ActorHub: About to call actorProcessMessage for connection %s", msg.Connection.ID)
		h.actorProcessMessage(msg.Connection, msg.Message, msg.Response)
	case "join_room":
		h.actorJoinRoom(msg.Connection.ID, msg.Room, msg.Response)
	case "leave_room":
		h.actorLeaveRoom(msg.Connection.ID, msg.Room, msg.Response)
	case "broadcast_to_room":
		h.actorBroadcastToRoom(msg.Room, msg.Message, msg.Response)
	case "broadcast_to_user":
		h.actorBroadcastToUser(msg.UserID, msg.Message, msg.Response)
	case "broadcast_to_all":
		h.actorBroadcastToAll(msg.Message, msg.Response)
	case "get_connection_count":
		h.actorGetConnectionCount(msg.Response)
	case "list_rooms":
		h.actorListRooms(msg.Response)
	case "check_rate_limit":
		h.actorCheckRateLimit(msg.UserID, msg.Response)
	default:
		log.Printf("ActorHub: Unknown message type: %s", msg.Type)
		if msg.Response != nil {
			msg.Response <- fmt.Errorf("unknown message type: %s", msg.Type)
			close(msg.Response)
		}
	}
}

// generateSecureConnectionID generates a cryptographically secure connection ID
func (h *ActorHub) generateSecureConnectionID() string {
	// Use atomic counter for uniqueness
	counter := atomic.AddInt64(&h.connectionCounter, 1)

	// Add random bytes for security
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)

	// Create timestamp for debugging
	timestamp := time.Now().Format("20060102150405")

	return fmt.Sprintf("%s-%d-%s", timestamp, counter, hex.EncodeToString(randomBytes))
}

// validateInput validates and sanitizes input to prevent injection attacks
func validateInput(input, inputType string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("%s cannot be empty", inputType)
	}

	// Check for dangerous patterns
	for _, pattern := range dangerousPatterns {
		if pattern.MatchString(input) {
			return "", fmt.Errorf("%s contains dangerous characters", inputType)
		}
	}

	switch inputType {
	case "room":
		if !validRoomName.MatchString(input) {
			return "", fmt.Errorf("room name must be 1-50 alphanumeric characters, underscores, or hyphens")
		}
	case "username":
		if !validUsername.MatchString(input) {
			return "", fmt.Errorf("username must be 1-30 alphanumeric characters, underscores, or hyphens")
		}
	}

	// HTML escape to prevent XSS
	return html.EscapeString(strings.TrimSpace(input)), nil
}

// checkRateLimit checks if a connection has exceeded rate limits using actor pattern
func (h *ActorHub) checkRateLimit(connectionID string) error {
	response := make(chan interface{}, 1)
	msg := HubMessage{
		Type:     "check_rate_limit",
		UserID:   connectionID,
		Response: response,
	}

	select {
	case h.hubChannel <- msg:
		result := <-response
		if err, ok := result.(error); ok {
			return err
		}
		return nil
	case <-h.ctx.Done():
		return fmt.Errorf("hub is shutting down")
	}
}

// cleanupLoop periodically cleans up old rate limit entries
func (rl *RateLimiter) cleanupLoop() {
	for range rl.cleanupTicker.C {
		now := time.Now()
		for connID, limit := range rl.connectionLimits {
			// Remove entries older than cleanup interval
			if now.Sub(limit.lastMessageTime) > CleanupInterval {
				delete(rl.connectionLimits, connID)
			}
		}
	}
}

// Public API methods (these send messages to the actor)

// Register registers a new connection
func (h *ActorHub) Register(conn *Connection) {
	// Generate secure connection ID
	conn.ID = h.generateSecureConnectionID()

	response := make(chan interface{})
	h.hubChannel <- HubMessage{
		Type:       "register",
		Connection: conn,
		Response:   response,
	}
	<-response // Wait for completion
	close(response)
}

// Unregister unregisters a connection
func (h *ActorHub) Unregister(conn *Connection) {
	response := make(chan interface{})
	h.hubChannel <- HubMessage{
		Type:       "unregister",
		Connection: conn,
		Response:   response,
	}
	<-response // Wait for completion
	close(response)
}

// ProcessMessage processes an incoming message
func (h *ActorHub) ProcessMessage(conn *Connection, msg *Message) {
	log.Printf("ActorHub: ProcessMessage called for connection %s, message type: %s", conn.ID, msg.Type)
	response := make(chan interface{})
	h.hubChannel <- HubMessage{
		Type:       "process_message",
		Connection: conn,
		Message:    msg,
		Response:   response,
	}
	<-response // Wait for completion
	close(response)
}

// JoinRoom adds a connection to a room
func (h *ActorHub) JoinRoom(connectionID, room string) error {
	response := make(chan interface{})
	h.hubChannel <- HubMessage{
		Type:     "join_room",
		Room:     room,
		Response: response,
	}
	h.hubChannel <- HubMessage{
		Type:       "join_room",
		Connection: &Connection{ID: connectionID},
		Room:       room,
		Response:   response,
	}
	result := <-response
	close(response)

	if err, ok := result.(error); ok {
		return err
	}
	return nil
}

// LeaveRoom removes a connection from a room
func (h *ActorHub) LeaveRoom(connectionID, room string) error {
	response := make(chan interface{})
	h.hubChannel <- HubMessage{
		Type:       "leave_room",
		Connection: &Connection{ID: connectionID},
		Room:       room,
		Response:   response,
	}
	result := <-response
	close(response)

	if err, ok := result.(error); ok {
		return err
	}
	return nil
}

// BroadcastToRoom sends a message to all connections in a room
func (h *ActorHub) BroadcastToRoom(room string, msg *Message) {
	response := make(chan interface{})
	h.hubChannel <- HubMessage{
		Type:     "broadcast_to_room",
		Room:     room,
		Message:  msg,
		Response: response,
	}
	<-response // Wait for completion
	close(response)
}

// BroadcastToUser sends a message to a specific user
func (h *ActorHub) BroadcastToUser(userID string, msg *Message) {
	response := make(chan interface{})
	h.hubChannel <- HubMessage{
		Type:     "broadcast_to_user",
		UserID:   userID,
		Message:  msg,
		Response: response,
	}
	<-response // Wait for completion
	close(response)
}

// BroadcastToAll sends a message to all connections
func (h *ActorHub) BroadcastToAll(msg *Message) {
	response := make(chan interface{})
	h.hubChannel <- HubMessage{
		Type:     "broadcast_to_all",
		Message:  msg,
		Response: response,
	}
	<-response // Wait for completion
	close(response)
}

// GetConnectionCount returns the number of active connections
func (h *ActorHub) GetConnectionCount() int {
	response := make(chan interface{})
	h.hubChannel <- HubMessage{
		Type:     "get_connection_count",
		Response: response,
	}
	result := <-response
	close(response)

	if count, ok := result.(int); ok {
		return count
	}
	return 0
}

// SetAuthHandler sets the authentication handler
func (h *ActorHub) SetAuthHandler(handler AuthHandler) {
	h.authHandler = handler
}

// RegisterMessageHandler registers a message handler
func (h *ActorHub) RegisterMessageHandler(messageType string, handler MessageHandler) {
	h.messageHandlers[messageType] = handler
}

// Start starts the hub (actor is already running)
func (h *ActorHub) Start() {
	// Actor is already started in NewActorHub
	log.Printf("ActorHub: Hub is ready")
}

// Stop gracefully stops the hub
func (h *ActorHub) Stop() {
	h.cancel()
}
