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

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(router.Run(":" + cfg.Port))
}
