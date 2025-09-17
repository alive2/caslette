// Package main provides a modular WebSocket handler for poker operations
// This file replaces the monolithic websocket.go implementation

package websocket

import (
	"caslette-server/websocket/poker"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// WebSocketClient implements the poker.Client interface
type WebSocketClient struct {
	conn   *websocket.Conn
	userID uint
	send   chan []byte
}

func (c *WebSocketClient) GetUserID() uint {
	return c.userID
}

func (c *WebSocketClient) SendError(message string) {
	response := map[string]interface{}{
		"type":    "error",
		"message": message,
	}
	data, _ := json.Marshal(response)
	select {
	case c.send <- data:
	default:
		close(c.send)
	}
}

func (c *WebSocketClient) SendSuccess(messageType string, data interface{}) {
	response := map[string]interface{}{
		"type": messageType,
		"data": data,
	}
	responseData, _ := json.Marshal(response)
	select {
	case c.send <- responseData:
	default:
		close(c.send)
	}
}

func (c *WebSocketClient) SendMessage(msg poker.PokerMessage) {
	data, _ := json.Marshal(msg)
	select {
	case c.send <- data:
	default:
		close(c.send)
	}
}

func (c *WebSocketClient) GetConnection() interface{} {
	return c.conn
}

func (c *WebSocketClient) IsConnected() bool {
	return c.conn != nil
}

func (c *WebSocketClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
	close(c.send)
}

// WebSocketManager handles WebSocket connections and routing
type WebSocketManager struct {
	db          *gorm.DB
	pokerRouter *poker.PokerRouter
	upgrader    websocket.Upgrader
}

func NewWebSocketManager(db *gorm.DB) *WebSocketManager {
	return &WebSocketManager{
		db:          db,
		pokerRouter: poker.NewPokerRouter(db),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
	}
}

// HandleWebSocket handles WebSocket connections
func (wsm *WebSocketManager) HandleWebSocket(c *gin.Context) {
	// Extract user ID from context (set by auth middleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(401, gin.H{"error": "Invalid user ID"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := wsm.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create client wrapper
	client := &WebSocketClient{
		conn:   conn,
		userID: userID,
		send:   make(chan []byte, 256),
	}

	// Register client with poker router
	wsm.pokerRouter.AddClient(userID, client)

	// Start goroutines for reading and writing
	go wsm.readPump(client)
	go wsm.writePump(client)

	log.Printf("WebSocket connection established for user %d", userID)
}

// readPump handles incoming messages from the WebSocket
func (wsm *WebSocketManager) readPump(client *WebSocketClient) {
	defer func() {
		wsm.pokerRouter.RemoveClient(client.userID)
		client.conn.Close()
	}()

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Route message to poker router
		wsm.pokerRouter.HandleMessage(client, message)
	}
}

// writePump handles outgoing messages to the WebSocket
func (wsm *WebSocketManager) writePump(client *WebSocketClient) {
	defer client.conn.Close()

	for {
		select {
		case message, ok := <-client.send:
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

// GetPokerStats returns poker system statistics
func (wsm *WebSocketManager) GetPokerStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := wsm.pokerRouter.GetStats()
		c.JSON(200, stats)
	}
}

// Example integration with existing server setup
func SetupWebSocketRoutes(r *gin.Engine, db *gorm.DB) {
	wsManager := NewWebSocketManager(db)

	// WebSocket endpoint
	r.GET("/ws", wsManager.HandleWebSocket)

	// Stats endpoint
	r.GET("/api/poker/stats", wsManager.GetPokerStats())
}
