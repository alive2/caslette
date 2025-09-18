package websocket_v2

import (
	"caslette-server/auth"
	"context"
	"errors"
	"log"
	"strconv"
)

// AuthService wraps our existing authentication service for WebSocket use
type AuthService struct {
	authService *auth.AuthService
}

// NewAuthService creates a new WebSocket auth service
func NewAuthService(authService *auth.AuthService) *AuthService {
	return &AuthService{
		authService: authService,
	}
}

// AuthenticateToken validates a JWT token and returns user information
func (a *AuthService) AuthenticateToken(token string) (*AuthResult, error) {
	if a.authService == nil {
		return &AuthResult{
			Success: false,
			Error:   "Authentication service not available",
		}, errors.New("authentication service not available")
	}

	// Validate the JWT token
	claims, err := a.authService.ValidateToken(token)
	if err != nil {
		return &AuthResult{
			Success: false,
			Error:   "Invalid token",
		}, err
	}

	// For now, we'll trust the token since we validated it
	// In a production environment, you might want to check the database
	// to ensure the user still exists and is active
	return &AuthResult{
		UserID:   strconv.FormatUint(uint64(claims.UserID), 10), // Convert uint to string
		Username: claims.Username,
		Success:  true,
	}, nil
}

// CreateWebSocketAuthHandler creates an auth handler for the WebSocket hub
func CreateWebSocketAuthHandler(authService *auth.AuthService) AuthHandler {
	log.Printf("Creating WebSocket auth handler")
	wsAuthService := NewAuthService(authService)

	return func(token string) (*AuthResult, error) {
		log.Printf("Auth handler called with token: %s", token)
		result, err := wsAuthService.AuthenticateToken(token)
		log.Printf("Auth result: %+v, error: %v", result, err)
		return result, err
	}
}

// RequireAuth is a middleware that ensures a connection is authenticated
func RequireAuth(handler MessageHandler) MessageHandler {
	return func(ctx context.Context, conn *Connection, msg *Message) *Message {
		if conn.UserID == "" {
			return &Message{
				Type:      "error",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Authentication required",
			}
		}
		return handler(ctx, conn, msg)
	}
}

// extractRoomFromMessage extracts room information from a message
func extractRoomFromMessage(msg *Message) string {
	if msg.Room != "" {
		return msg.Room
	}

	// Try to extract from data
	if dataMap, ok := msg.Data.(map[string]interface{}); ok {
		if room, ok := dataMap["room"].(string); ok {
			return room
		}
	}

	// Try to extract from string data
	if room, ok := msg.Data.(string); ok {
		return room
	}

	return ""
}
