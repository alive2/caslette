package websocket_v2

// HubInterface defines the interface that both Hub and ActorHub implement
type HubInterface interface {
	// Connection management
	Register(conn *Connection)
	Unregister(conn *Connection)

	// Message processing
	ProcessMessage(conn *Connection, msg *Message)

	// Room management
	JoinRoom(connectionID, room string) error
	LeaveRoom(connectionID, room string) error

	// Broadcasting
	BroadcastToRoom(room string, msg *Message)
	BroadcastToUser(userID string, msg *Message)
	BroadcastToAll(msg *Message)

	// Configuration
	SetAuthHandler(handler AuthHandler)
	RegisterMessageHandler(messageType string, handler MessageHandler)

	// Lifecycle
	Start()
	GetConnectionCount() int
}

// Ensure ActorHub satisfies the interface
var _ HubInterface = (*ActorHub)(nil)
