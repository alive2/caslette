package websocket

import (
	"caslette-server/auth"
	"caslette-server/models"
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
	return &WebSocketHandler{
		hub:         hub,
		db:          db,
		authService: authService,
	}
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
		case "game_action":
			// Handle game actions (to be implemented)
			log.Printf("Game action from user %s: %+v", c.username, msg.Data)
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
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
