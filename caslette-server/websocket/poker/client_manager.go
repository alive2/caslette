package poker

// Client represents a WebSocket client connected to the poker system
type Client interface {
	// GetUserID returns the authenticated user ID for this client
	GetUserID() uint

	// SendError sends an error message to the client
	SendError(message string)

	// SendSuccess sends a success response with a message type and data
	SendSuccess(messageType string, data interface{})

	// SendMessage sends a raw poker message to the client
	SendMessage(msg PokerMessage)

	// GetConnection returns the underlying connection (for broadcasting)
	GetConnection() interface{}

	// IsConnected returns whether the client is still connected
	IsConnected() bool

	// Close closes the client connection
	Close()
}

// ClientManager handles client connections and broadcasting
type ClientManager struct {
	clients map[uint]Client // userID -> Client
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients: make(map[uint]Client),
	}
}

func (cm *ClientManager) AddClient(userID uint, client Client) {
	cm.clients[userID] = client
}

func (cm *ClientManager) RemoveClient(userID uint) {
	delete(cm.clients, userID)
}

func (cm *ClientManager) GetClient(userID uint) (Client, bool) {
	client, exists := cm.clients[userID]
	return client, exists
}

func (cm *ClientManager) GetAllClients() []Client {
	clients := make([]Client, 0, len(cm.clients))
	for _, client := range cm.clients {
		if client.IsConnected() {
			clients = append(clients, client)
		}
	}
	return clients
}

func (cm *ClientManager) GetTableClients(tableID uint, db interface{}) []Client {
	// This will be implemented when we connect to the database layer
	// For now, return empty slice
	return []Client{}
}

func (cm *ClientManager) BroadcastToAll(msg PokerMessage) {
	for _, client := range cm.clients {
		if client.IsConnected() {
			client.SendMessage(msg)
		}
	}
}

func (cm *ClientManager) BroadcastToTable(tableID uint, msg PokerMessage, db interface{}) {
	clients := cm.GetTableClients(tableID, db)
	for _, client := range clients {
		if client.IsConnected() {
			client.SendMessage(msg)
		}
	}
}

func (cm *ClientManager) BroadcastToUser(userID uint, msg PokerMessage) {
	if client, exists := cm.clients[userID]; exists && client.IsConnected() {
		client.SendMessage(msg)
	}
}
