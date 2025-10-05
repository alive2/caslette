package main

import (
	"caslette-server/auth"
	"caslette-server/config"
	"caslette-server/database"
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
