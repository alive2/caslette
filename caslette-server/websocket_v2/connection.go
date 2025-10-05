package websocket_v2

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Connection represents a WebSocket connection
type Connection struct {
	ID       string
	UserID   string
	Username string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      HubInterface
	Rooms    map[string]bool
	mu       sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	Event     string      `json:"event,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Room      string      `json:"room,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
	Success   bool        `json:"success,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// AuthMessage represents authentication message
type AuthMessage struct {
	Token string `json:"token"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// NewConnection creates a new WebSocket connection
func NewConnection(hub HubInterface, w http.ResponseWriter, r *http.Request) (*Connection, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	connection := &Connection{
		Conn:  conn,
		Send:  make(chan []byte, 256),
		Hub:   hub,
		Rooms: make(map[string]bool),
	}

	return connection, nil
}

// Start begins the connection's read and write pumps
func (c *Connection) Start() {
	go c.writePump()
	go c.readPump()
}

// Close cleanly closes the connection
func (c *Connection) Close() {
	// Get the list of rooms to leave while holding the lock
	c.mu.Lock()
	roomsToLeave := make([]string, 0, len(c.Rooms))
	for room := range c.Rooms {
		roomsToLeave = append(roomsToLeave, room)
	}
	c.mu.Unlock()

	// Leave all rooms after releasing the lock to prevent deadlock
	for _, room := range roomsToLeave {
		c.Hub.LeaveRoom(c.ID, room)
	}

	// Close the connection
	c.Conn.Close()
	close(c.Send)
}

// SendMessage sends a message to this connection
func (c *Connection) SendMessage(msg *Message) {
	msg.Timestamp = time.Now().Unix()
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	log.Printf("SendMessage: Sending %s to connection %s (data: %s)", msg.Type, c.ID, string(data))

	select {
	case c.Send <- data:
		log.Printf("SendMessage: Successfully queued %s for connection %s", msg.Type, c.ID)
	default:
		log.Printf("Connection %s send channel full, closing connection", c.ID)
		c.Close()
	}
}

// JoinRoom adds the connection to a room
func (c *Connection) JoinRoom(room string) {
	c.mu.Lock()
	c.Rooms[room] = true
	c.mu.Unlock()
	// Call hub method after releasing connection lock to prevent deadlock
	c.Hub.JoinRoom(c.ID, room)
}

// LeaveRoom removes the connection from a room
func (c *Connection) LeaveRoom(room string) {
	c.mu.Lock()
	delete(c.Rooms, room)
	c.mu.Unlock()
	// Call hub method after releasing connection lock to prevent deadlock
	c.Hub.LeaveRoom(c.ID, room)
}

// IsInRoom checks if connection is in a specific room
func (c *Connection) IsInRoom(room string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Rooms[room]
}

// GetRooms returns a copy of the rooms this connection is in
func (c *Connection) GetRooms() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rooms := make([]string, 0, len(c.Rooms))
	for room := range c.Rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// readPump pumps messages from the websocket connection to the hub
func (c *Connection) readPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Close()
	}()

	c.Conn.SetReadLimit(4096) // Increased from 512 to handle larger messages like JWT tokens
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		log.Printf("Connection %s: Received raw message: %s", c.ID, string(messageBytes))

		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		log.Printf("Connection %s: Parsed message type: %s, requestId: %s", c.ID, msg.Type, msg.RequestID)

		msg.Timestamp = time.Now().Unix()
		c.Hub.ProcessMessage(c, &msg)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Connection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// generateConnectionID generates a unique connection ID
func generateConnectionID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
