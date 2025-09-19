package websocket_v2

import "context"

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
