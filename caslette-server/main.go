package main

import (
	"caslette-server/auth"
	"caslette-server/config"
	"caslette-server/database"
	"caslette-server/handlers"
	"caslette-server/middleware"
	"caslette-server/websocket"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Run database migrations
	database.Migrate(cfg.DB)

	// Initialize auth service
	authService := auth.NewAuthService(cfg.JWTSecret)

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg.DB, authService)
	userHandler := handlers.NewUserHandler(cfg.DB)
	diamondHandler := handlers.NewDiamondHandler(cfg.DB)
	roleHandler := handlers.NewRoleHandler(cfg.DB)
	permissionHandler := handlers.NewPermissionHandler(cfg.DB)
	pokerHandler := handlers.NewPokerHandler(cfg.DB)
	wsHandler := websocket.NewWebSocketHandler(hub, cfg.DB, authService)

	// Setup Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(middleware.CORSMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

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

		// WebSocket endpoint (handles its own authentication)
		api.GET("/ws", wsHandler.HandleWebSocket)

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

			// Poker routes
			poker := protected.Group("/poker")
			{
				poker.GET("/tables", pokerHandler.ListTables)
				poker.GET("/tables/:id", pokerHandler.GetTable)
				poker.POST("/tables", middleware.PermissionMiddleware(cfg.DB, "poker.table.create"), pokerHandler.CreateTable)
				poker.POST("/tables/:id/join", pokerHandler.JoinTable)
				poker.POST("/tables/:id/leave", pokerHandler.LeaveTable)
			}
		}
	}

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(router.Run(":" + cfg.Port))
}
