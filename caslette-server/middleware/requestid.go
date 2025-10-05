package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is already set (e.g., from a load balancer header)
		requestID := c.GetHeader("X-Request-ID")

		// If not set, generate a new UUID
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set the request ID in the context
		c.Set("request_id", requestID)

		// Add the request ID to the response headers for debugging
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
