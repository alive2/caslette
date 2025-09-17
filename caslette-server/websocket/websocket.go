package websocket

import (
	"caslette-server/auth"
	"caslette-server/models"
	"caslette-server/websocket/poker"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from localhost for development
		return true
	},
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	wsHandler  *WebSocketHandler // Reference to handler for poker messages
	mu         sync.RWMutex
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	user     *models.User
	userID   uint
	username string
}

// Implement poker.Client interface for Client
func (c *Client) GetUserID() uint {
	return c.userID
}

func (c *Client) SendError(message string) {
	response := Message{
		Type: "error",
		Data: map[string]string{"error": message},
	}
	data, _ := json.Marshal(response)
	select {
	case c.send <- data:
	default:
		// Channel full, skip
	}
}

func (c *Client) SendSuccess(messageType string, data interface{}) {
	response := Message{
		Type: messageType,
		Data: data,
	}
	responseData, _ := json.Marshal(response)
	select {
	case c.send <- responseData:
	default:
		// Channel full, skip
	}
}

func (c *Client) SendMessage(msg poker.PokerMessage) {
	data, _ := json.Marshal(msg)
	select {
	case c.send <- data:
	default:
		// Channel full, skip
	}
}

func (c *Client) GetConnection() interface{} {
	return c.conn
}

func (c *Client) IsConnected() bool {
	return c.conn != nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

type Message struct {
	Type   string      `json:"type"`
	Data   interface{} `json:"data"`
	UserID uint        `json:"user_id,omitempty"`
	Target string      `json:"target,omitempty"` // "all", "admins", "user:id"
}

type WebSocketHandler struct {
	hub         *Hub
	db          *gorm.DB
	authService *auth.AuthService
	pokerRouter *poker.PokerRouter
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("User %s connected via WebSocket", client.username)

			// Register with poker router if handler exists
			if h.wsHandler != nil {
				h.wsHandler.pokerRouter.AddClient(client.userID, client)
			}

			// Send welcome message
			welcome := Message{
				Type: "welcome",
				Data: map[string]interface{}{
					"message": "Connected to Caslette WebSocket",
					"user_id": client.userID,
				},
			}
			data, _ := json.Marshal(welcome)
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, client)
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("User %s disconnected from WebSocket", client.username)

				// Unregister from poker router if handler exists
				if h.wsHandler != nil {
					h.wsHandler.pokerRouter.RemoveClient(client.userID)
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) BroadcastToAdmins(message Message) {
	data, err := json.Marshal(message)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if client.user != nil {
			for _, role := range client.user.Roles {
				if role.Name == "admin" || role.Name == "moderator" {
					select {
					case client.send <- data:
					default:
						close(client.send)
						delete(h.clients, client)
					}
					break
				}
			}
		}
	}
}

func (h *Hub) GetConnectedUsers() []map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]map[string]interface{}, 0)
	for client := range h.clients {
		if client.user != nil {
			users = append(users, map[string]interface{}{
				"id":       client.userID,
				"username": client.username,
				"roles":    client.user.Roles,
			})
		}
	}
	return users
}

func NewWebSocketHandler(hub *Hub, db *gorm.DB, authService *auth.AuthService) *WebSocketHandler {
	pokerRouter := poker.NewPokerRouter(db)

	handler := &WebSocketHandler{
		hub:         hub,
		db:          db,
		authService: authService,
		pokerRouter: pokerRouter,
	}

	// Set back-reference so hub can access poker router
	hub.wsHandler = handler

	return handler
}

func (wsh *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get token from query parameter
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
		return
	}

	// Validate token
	claims, err := wsh.authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Get user from database
	var user models.User
	if err := wsh.db.Preload("Roles").First(&user, claims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create client
	client := &Client{
		hub:      wsh.hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		user:     &user,
		userID:   user.ID,
		username: user.Username,
	}

	// Register client
	client.hub.register <- client

	// Start goroutines
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Add user info to message
		msg.UserID = c.userID

		// Handle different message types
		switch msg.Type {
		case "ping":
			response := Message{Type: "pong", Data: "pong"}
			data, _ := json.Marshal(response)
			select {
			case c.send <- data:
			default:
				return
			}
		case poker.MsgCreateTable, poker.MsgListTables, poker.MsgJoinTable, poker.MsgLeaveTable,
			poker.MsgPlayerAction, poker.MsgStartHand, poker.MsgGameAction, poker.MsgGetGameState:
			// Handle poker messages by converting to raw message and routing
			rawMessage, _ := json.Marshal(msg)
			if wsHandler := c.getWebSocketHandler(); wsHandler != nil {
				wsHandler.pokerRouter.HandleMessage(c, rawMessage)
			}
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

// Helper method to get WebSocketHandler (we'll need to store this reference)
func (c *Client) getWebSocketHandler() *WebSocketHandler {
	// This is a temporary solution - ideally we'd store the handler reference in Client
	// For now, we'll access it through the hub
	if c.hub.wsHandler != nil {
		return c.hub.wsHandler
	}
	return nil
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}
