package main

import (
	"caslette-server/auth"
	"caslette-server/config"
	"caslette-server/database"
	"caslette-server/game"
	"caslette-server/handlers"
	"caslette-server/middleware"
	"caslette-server/models"
	"caslette-server/websocket_v2"
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Run database migrations
	database.Migrate(cfg.DB)

	// Initialize auth service
	authService := auth.NewAuthService(cfg.JWTSecret)

	// Initialize WebSocket server
	wsServer := websocket_v2.NewServer(authService)

	// Initialize poker table system
	setupPokerSystem(wsServer)

	// Register custom WebSocket message handlers

	// Handler for getting user balance
	wsServer.RegisterHandler("get_user_balance", func(ctx context.Context, conn *websocket_v2.Connection, msg *websocket_v2.Message) *websocket_v2.Message {
		log.Printf("WebSocket: get_user_balance request from connection %s", conn.ID)

		// Check if user is authenticated
		if conn.UserID == "" {
			return &websocket_v2.Message{
				Type:      "get_user_balance_response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Authentication required",
			}
		}

		// Parse request data
		var requestData map[string]interface{}
		if data, ok := msg.Data.(map[string]interface{}); ok {
			requestData = data
		} else {
			return &websocket_v2.Message{
				Type:      "get_user_balance_response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Invalid request data",
			}
		}

		// Get userId from request or use authenticated user's ID
		userID := conn.UserID
		if reqUserID, exists := requestData["userId"]; exists {
			if reqUserIDStr, ok := reqUserID.(string); ok {
				// For now, users can only get their own balance
				if reqUserIDStr != conn.UserID {
					return &websocket_v2.Message{
						Type:      "get_user_balance_response",
						RequestID: msg.RequestID,
						Success:   false,
						Error:     "Access denied: can only access own balance",
					}
				}
				userID = reqUserIDStr
			}
		}

		// Query user's current balance
		var currentBalance int
		err := cfg.DB.Model(&models.Diamond{}).Where("user_id = ?", userID).Order("created_at desc").Limit(1).Pluck("balance", &currentBalance).Error
		if err != nil {
			log.Printf("Error getting user balance: %v", err)
			return &websocket_v2.Message{
				Type:      "get_user_balance_response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Failed to retrieve balance",
			}
		}

		// Return success response
		return &websocket_v2.Message{
			Type:      "get_user_balance_response",
			RequestID: msg.RequestID,
			Success:   true,
			Data: map[string]interface{}{
				"userId":          userID,
				"current_balance": currentBalance,
			},
		}
	})

	// Handler for getting user profile
	wsServer.RegisterHandler("get_user_profile", func(ctx context.Context, conn *websocket_v2.Connection, msg *websocket_v2.Message) *websocket_v2.Message {
		log.Printf("WebSocket: get_user_profile request from connection %s", conn.ID)

		// Check if user is authenticated
		if conn.UserID == "" {
			return &websocket_v2.Message{
				Type:      "get_user_profile_response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Authentication required",
			}
		}

		// Parse request data
		var requestData map[string]interface{}
		if data, ok := msg.Data.(map[string]interface{}); ok {
			requestData = data
		} else {
			return &websocket_v2.Message{
				Type:      "get_user_profile_response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Invalid request data",
			}
		}

		// Get userId from request or use authenticated user's ID
		userID := conn.UserID
		if reqUserID, exists := requestData["userId"]; exists {
			if reqUserIDStr, ok := reqUserID.(string); ok {
				// For now, users can only get their own profile
				if reqUserIDStr != conn.UserID {
					return &websocket_v2.Message{
						Type:      "get_user_profile_response",
						RequestID: msg.RequestID,
						Success:   false,
						Error:     "Access denied: can only access own profile",
					}
				}
				userID = reqUserIDStr
			}
		}

		// Query user profile
		var user models.User
		err := cfg.DB.Where("id = ?", userID).First(&user).Error
		if err != nil {
			log.Printf("Error getting user profile: %v", err)
			return &websocket_v2.Message{
				Type:      "get_user_profile_response",
				RequestID: msg.RequestID,
				Success:   false,
				Error:     "Failed to retrieve user profile",
			}
		}

		// Return success response
		return &websocket_v2.Message{
			Type:      "get_user_profile_response",
			RequestID: msg.RequestID,
			Success:   true,
			Data: map[string]interface{}{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
		}
	})

	// Start WebSocket server in background
	go wsServer.Run()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg.DB, authService)
	userHandler := handlers.NewUserHandler(cfg.DB)
	diamondHandler := handlers.NewDiamondHandler(cfg.DB)
	roleHandler := handlers.NewRoleHandler(cfg.DB)
	permissionHandler := handlers.NewPermissionHandler(cfg.DB)

	// Setup Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORSMiddleware())

	// Add Request ID middleware
	router.Use(middleware.RequestIDMiddleware())

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/profile", middleware.AuthMiddleware(authService), authHandler.GetProfile)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// User routes
			users := protected.Group("/users")
			{
				users.GET("", userHandler.GetUsers)
				users.GET("/:id", userHandler.GetUser)
				users.PUT("/:id", userHandler.UpdateUser)
				users.DELETE("/:id", userHandler.DeleteUser)
				users.POST("/:id/roles", userHandler.AssignRoles)
				users.POST("/:id/permissions", userHandler.AssignPermissions)
				users.GET("/:id/permissions", userHandler.GetUserPermissions)
				users.DELETE("/:id/permissions/:permission_id", userHandler.RemoveUserPermission)
			}

			// Role routes
			roles := protected.Group("/roles")
			{
				roles.GET("", roleHandler.GetRoles)
				roles.GET("/:id", roleHandler.GetRole)
				roles.POST("", roleHandler.CreateRole)
				roles.PUT("/:id", roleHandler.UpdateRole)
				roles.DELETE("/:id", roleHandler.DeleteRole)
				roles.POST("/:id/permissions", roleHandler.AssignPermissions)
			}

			// Permission routes
			permissions := protected.Group("/permissions")
			{
				permissions.GET("", permissionHandler.GetPermissions)
				permissions.GET("/:id", permissionHandler.GetPermission)
				permissions.POST("", permissionHandler.CreatePermission)
				permissions.PUT("/:id", permissionHandler.UpdatePermission)
				permissions.DELETE("/:id", permissionHandler.DeletePermission)
			}

			// Diamond routes
			diamonds := protected.Group("/diamonds")
			{
				diamonds.GET("/user/:userId", diamondHandler.GetUserDiamonds)
				diamonds.POST("/credit", diamondHandler.AddDiamonds)
				diamonds.POST("/debit", diamondHandler.DeductDiamonds)
				diamonds.GET("/transactions", diamondHandler.GetAllTransactions)
			}
		}
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// WebSocket endpoint
	router.GET("/ws", gin.WrapH(wsServer))

	// WebSocket health check endpoint
	router.GET("/api/websocket/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":            "healthy",
			"connected_clients": wsServer.GetConnectionCount(),
			"connected_users":   len(wsServer.GetConnectedUsers()),
			"active_rooms":      wsServer.GetActiveRooms(),
		})
	})

	log.Printf("Server starting on port 8081")
	log.Printf("WebSocket endpoint available at ws://localhost:8081/ws")
	log.Fatal(http.ListenAndServe(":8081", router))
}

// setupPokerSystem initializes the poker table system with WebSocket integration
func setupPokerSystem(wsServer *websocket_v2.Server) {
	// Create WebSocket hub adapter
	hubAdapter := &WebSocketHubAdapter{server: wsServer}

	// Create table integration
	tableIntegration := game.NewTableGameIntegration(hubAdapter)

	// Register all table message handlers
	tableHandlers := tableIntegration.GetMessageHandlers()
	for messageType, handler := range tableHandlers {
		registerTableHandler(wsServer, messageType, handler)
	}

	// Register poker action handlers
	registerPokerActionHandlers(wsServer, tableIntegration.GetTableManager())

	log.Printf("Poker system initialized with %d message handlers", len(tableHandlers)+5)
}

// WebSocketHubAdapter adapts websocket_v2.Server to game.WebSocketHub
type WebSocketHubAdapter struct {
	server *websocket_v2.Server
}

func (w *WebSocketHubAdapter) BroadcastToRoom(roomID string, msg interface{}) error {
	// Convert interface{} to appropriate message format
	switch m := msg.(type) {
	case *game.WebSocketMessage:
		w.server.BroadcastToRoom(roomID, m.Type, m.Data)
	case map[string]interface{}:
		msgType := "game_event"
		if t, ok := m["type"].(string); ok {
			msgType = t
		}
		w.server.BroadcastToRoom(roomID, msgType, m)
	default:
		w.server.BroadcastToRoom(roomID, "unknown", msg)
	}
	return nil
}

func (w *WebSocketHubAdapter) GetRoomUsers(roomID string) []map[string]interface{} {
	users := w.server.GetRoomUsers(roomID)
	result := make([]map[string]interface{}, len(users))
	for i, user := range users {
		result[i] = map[string]interface{}{
			"user_id":  user,
			"username": "", // Would need to fetch from user store
		}
	}
	return result
}

// WebSocketConnectionAdapter adapts websocket_v2.Connection to game.WebSocketConnection
type WebSocketConnectionAdapter struct {
	conn *websocket_v2.Connection
}

func (w *WebSocketConnectionAdapter) GetUserID() string {
	return w.conn.UserID
}

func (w *WebSocketConnectionAdapter) GetUsername() string {
	return w.conn.Username
}

func (w *WebSocketConnectionAdapter) SendMessage(msg interface{}) error {
	// Convert interface{} to *websocket_v2.Message
	switch m := msg.(type) {
	case *websocket_v2.Message:
		w.conn.SendMessage(m)
	case *game.WebSocketMessage:
		wsMsg := &websocket_v2.Message{
			Type:      m.Type,
			RequestID: m.RequestID,
			Success:   m.Success,
			Error:     m.Error,
			Data:      m.Data,
		}
		w.conn.SendMessage(wsMsg)
	default:
		// Create a generic message
		wsMsg := &websocket_v2.Message{
			Type: "unknown",
			Data: msg,
		}
		w.conn.SendMessage(wsMsg)
	}
	return nil
}

func (w *WebSocketConnectionAdapter) JoinRoom(roomID string) error {
	w.conn.JoinRoom(roomID)
	return nil
}

func (w *WebSocketConnectionAdapter) LeaveRoom(roomID string) error {
	w.conn.LeaveRoom(roomID)
	return nil
}

// registerTableHandler registers a table handler with WebSocket message conversion
func registerTableHandler(wsServer *websocket_v2.Server, messageType string, handler func(ctx context.Context, conn game.WebSocketConnection, msg *game.WebSocketMessage) *game.WebSocketMessage) {
	wsServer.RegisterHandler(messageType, func(ctx context.Context, conn *websocket_v2.Connection, msg *websocket_v2.Message) *websocket_v2.Message {
		log.Printf("registerTableHandler: Handling message type '%s' for user %s", messageType, conn.UserID)

		// Convert websocket types to game types
		tableConn := &WebSocketConnectionAdapter{conn: conn}
		tableMsg := &game.WebSocketMessage{
			Type:      msg.Type,
			RequestID: msg.RequestID,
			Data:      msg.Data,
		}

		log.Printf("registerTableHandler: Calling handler for '%s'", messageType)

		// Call the table handler
		response := handler(ctx, tableConn, tableMsg)
		if response == nil {
			log.Printf("registerTableHandler: Handler returned nil for '%s'", messageType)
			return nil
		}

		log.Printf("registerTableHandler: Handler returned success=%t, error='%s' for '%s'", response.Success, response.Error, messageType)

		// Convert response back to websocket types
		return &websocket_v2.Message{
			Type:      response.Type,
			RequestID: response.RequestID,
			Success:   response.Success,
			Error:     response.Error,
			Data:      response.Data,
		}
	})
}

// registerPokerActionHandlers registers poker-specific action handlers
func registerPokerActionHandlers(wsServer *websocket_v2.Server, tableManager *game.ActorTableManager) {
	// Register poker action handler
	wsServer.RegisterHandler("poker_action", func(ctx context.Context, conn *websocket_v2.Connection, msg *websocket_v2.Message) *websocket_v2.Message {
		return handlePokerAction(ctx, conn, msg, tableManager)
	})

	// Register hand history request handler
	wsServer.RegisterHandler("get_hand_history", func(ctx context.Context, conn *websocket_v2.Connection, msg *websocket_v2.Message) *websocket_v2.Message {
		return handleGetHandHistory(ctx, conn, msg, tableManager)
	})

	// Register player stats handler
	wsServer.RegisterHandler("get_player_stats", func(ctx context.Context, conn *websocket_v2.Connection, msg *websocket_v2.Message) *websocket_v2.Message {
		return handleGetPlayerStats(ctx, conn, msg, tableManager)
	})

	// Register table join room handler (for spectating)
	wsServer.RegisterHandler("join_table_room", func(ctx context.Context, conn *websocket_v2.Connection, msg *websocket_v2.Message) *websocket_v2.Message {
		return handleJoinTableRoom(ctx, conn, msg, tableManager)
	})
}

// handlePokerAction handles poker actions (fold, call, raise, etc.)
func handlePokerAction(ctx context.Context, conn *websocket_v2.Connection, msg *websocket_v2.Message, tableManager *game.ActorTableManager) *websocket_v2.Message {
	if conn.UserID == "" {
		return &websocket_v2.Message{
			Type:      "poker_action_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Authentication required",
		}
	}

	// Parse poker action data
	var actionData struct {
		TableID string `json:"table_id"`
		Action  string `json:"action"` // fold, call, raise, check, bet, all_in
		Amount  int    `json:"amount"` // for raise/bet actions
	}

	if err := parseMessageData(msg.Data, &actionData); err != nil {
		return &websocket_v2.Message{
			Type:      "poker_action_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid action data: " + err.Error(),
		}
	}

	// Get table
	table, err := tableManager.GetTable(actionData.TableID)
	if err != nil {
		return &websocket_v2.Message{
			Type:      "poker_action_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Table not found",
		}
	}

	// Check if player is at table and game is active
	playerID := conn.UserID
	if !table.IsPlayerAtTable(playerID) {
		return &websocket_v2.Message{
			Type:      "poker_action_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Player not at table",
		}
	}

	if table.Status != game.TableStatusActive {
		return &websocket_v2.Message{
			Type:      "poker_action_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Game not active",
		}
	}

	// Create game action
	gameAction := &game.GameAction{
		Type:     actionData.Action,
		PlayerID: playerID,
		Data: map[string]interface{}{
			"action": actionData.Action,
			"amount": actionData.Amount,
		},
	}

	// Validate action
	if err := table.GameEngine.IsValidAction(gameAction); err != nil {
		return &websocket_v2.Message{
			Type:      "poker_action_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid action: " + err.Error(),
		}
	}

	// Process action
	event, err := table.GameEngine.ProcessAction(ctx, gameAction)
	if err != nil {
		return &websocket_v2.Message{
			Type:      "poker_action_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Failed to process action: " + err.Error(),
		}
	}

	// Broadcast game event to all players at table
	tableManager.BroadcastGameEvent(table, event)

	return &websocket_v2.Message{
		Type:      "poker_action_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data: map[string]interface{}{
			"action":    actionData.Action,
			"processed": true,
			"event":     event,
		},
	}
}

// handleGetGameState returns current game state for a table
func handleGetGameState(ctx context.Context, conn *websocket_v2.Connection, msg *websocket_v2.Message, tableManager *game.ActorTableManager) *websocket_v2.Message {
	if conn.UserID == "" {
		return &websocket_v2.Message{
			Type:      "game_state_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Authentication required",
		}
	}

	var requestData struct {
		TableID string `json:"table_id"`
	}

	if err := parseMessageData(msg.Data, &requestData); err != nil {
		return &websocket_v2.Message{
			Type:      "game_state_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid request data",
		}
	}

	table, err := tableManager.GetTable(requestData.TableID)
	if err != nil {
		return &websocket_v2.Message{
			Type:      "game_state_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Table not found",
		}
	}

	// Check if user can view game state (player or observer)
	playerID := conn.UserID
	if !table.IsPlayerAtTable(playerID) && !table.IsObserver(playerID) {
		return &websocket_v2.Message{
			Type:      "game_state_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Access denied",
		}
	}

	// Get game state
	gameState := table.GameEngine.GetGameState()

	return &websocket_v2.Message{
		Type:      "game_state_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data:      gameState,
	}
}

// handleGetHandHistory returns hand history for a table
func handleGetHandHistory(ctx context.Context, conn *websocket_v2.Connection, msg *websocket_v2.Message, tableManager *game.ActorTableManager) *websocket_v2.Message {
	if conn.UserID == "" {
		return &websocket_v2.Message{
			Type:      "hand_history_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Authentication required",
		}
	}

	var requestData struct {
		TableID string `json:"table_id"`
		Limit   int    `json:"limit"`
	}

	if err := parseMessageData(msg.Data, &requestData); err != nil {
		return &websocket_v2.Message{
			Type:      "hand_history_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid request data",
		}
	}

	if requestData.Limit == 0 {
		requestData.Limit = 10 // Default limit
	}

	table, err := tableManager.GetTable(requestData.TableID)
	if err != nil {
		return &websocket_v2.Message{
			Type:      "hand_history_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Table not found",
		}
	}

	// Check access permissions
	playerID := conn.UserID
	if !table.IsPlayerAtTable(playerID) && !table.IsObserver(playerID) {
		return &websocket_v2.Message{
			Type:      "hand_history_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Access denied",
		}
	}

	// Get hand history
	history := table.GameEngine.GetHandHistory(requestData.Limit)

	return &websocket_v2.Message{
		Type:      "hand_history_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data: map[string]interface{}{
			"table_id": requestData.TableID,
			"history":  history,
		},
	}
}

// handleGetPlayerStats returns player statistics
func handleGetPlayerStats(ctx context.Context, conn *websocket_v2.Connection, msg *websocket_v2.Message, tableManager *game.ActorTableManager) *websocket_v2.Message {
	if conn.UserID == "" {
		return &websocket_v2.Message{
			Type:      "player_stats_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Authentication required",
		}
	}

	var requestData struct {
		TableID  string `json:"table_id"`
		PlayerID string `json:"player_id,omitempty"`
	}

	if err := parseMessageData(msg.Data, &requestData); err != nil {
		return &websocket_v2.Message{
			Type:      "player_stats_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid request data",
		}
	}

	// Default to requesting user's stats
	if requestData.PlayerID == "" {
		requestData.PlayerID = conn.UserID
	}

	table, err := tableManager.GetTable(requestData.TableID)
	if err != nil {
		return &websocket_v2.Message{
			Type:      "player_stats_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Table not found",
		}
	}

	// Get player stats
	stats := table.GameEngine.GetPlayerStats(requestData.PlayerID)

	return &websocket_v2.Message{
		Type:      "player_stats_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data: map[string]interface{}{
			"table_id":  requestData.TableID,
			"player_id": requestData.PlayerID,
			"stats":     stats,
		},
	}
}

// handleJoinTableRoom allows users to join table room for spectating
func handleJoinTableRoom(ctx context.Context, conn *websocket_v2.Connection, msg *websocket_v2.Message, tableManager *game.ActorTableManager) *websocket_v2.Message {
	log.Printf("handleJoinTableRoom: Starting for user %s", conn.UserID)

	if conn.UserID == "" {
		log.Printf("handleJoinTableRoom: Authentication required")
		return &websocket_v2.Message{
			Type:      "join_table_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Authentication required",
		}
	}

	log.Printf("handleJoinTableRoom: Parsing request data")
	var requestData struct {
		TableID string `json:"table_id"`
	}

	if err := parseMessageData(msg.Data, &requestData); err != nil {
		log.Printf("handleJoinTableRoom: Failed to parse request data: %v", err)
		return &websocket_v2.Message{
			Type:      "join_table_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Invalid request data",
		}
	}

	log.Printf("handleJoinTableRoom: Getting table %s", requestData.TableID)
	table, err := tableManager.GetTable(requestData.TableID)
	if err != nil {
		log.Printf("handleJoinTableRoom: Table not found: %v", err)
		return &websocket_v2.Message{
			Type:      "join_table_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Table not found",
		}
	}

	log.Printf("handleJoinTableRoom: Checking observer permissions for table %s", requestData.TableID)
	// Check if observers are allowed
	if !table.Settings.ObserversAllowed && !table.IsPlayerAtTable(conn.UserID) {
		log.Printf("handleJoinTableRoom: Observers not allowed")
		return &websocket_v2.Message{
			Type:      "join_table_room_response",
			RequestID: msg.RequestID,
			Success:   false,
			Error:     "Observers not allowed at this table",
		}
	}

	log.Printf("handleJoinTableRoom: Joining room %s", table.RoomID)
	// Join the table room
	conn.JoinRoom(table.RoomID)

	log.Printf("handleJoinTableRoom: Successfully joined room %s", table.RoomID)
	return &websocket_v2.Message{
		Type:      "join_table_room_response",
		RequestID: msg.RequestID,
		Success:   true,
		Data: map[string]interface{}{
			"table_id": requestData.TableID,
			"room_id":  table.RoomID,
		},
	}
}

// parseMessageData helper function to parse message data
func parseMessageData(data interface{}, target interface{}) error {
	if data == nil {
		return nil
	}

	// For now, simple type assertion
	if dataMap, ok := data.(map[string]interface{}); ok {
		// Convert to target using JSON marshal/unmarshal
		// This is a simple approach that works for basic cases
		// In a production system, you might want more sophisticated parsing
		if err := game.ConvertMapToStruct(dataMap, target); err != nil {
			return err
		}
	}
	return nil
}
