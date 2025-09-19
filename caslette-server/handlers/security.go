package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityValidator provides comprehensive input validation and sanitization
type SecurityValidator struct {
	// Rate limiting
	requestCounts map[string]*RequestLimit
	rateMutex     sync.RWMutex

	// Request tracking
	requestCounter int64
}

// RequestLimit tracks rate limiting per IP/user
type RequestLimit struct {
	Count      int64
	LastReset  time.Time
	Blocked    bool
	BlockUntil time.Time
	Violations int
	mutex      sync.Mutex
}

// Security patterns to detect malicious input
var (
	// SQL injection patterns
	sqlInjectionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`),
		regexp.MustCompile(`(?i)(--|\/\*|\*\/|;|'|"|\||&)`),
		regexp.MustCompile(`(?i)(or\s+1\s*=\s*1|and\s+1\s*=\s*1)`),
		regexp.MustCompile(`(?i)(information_schema|sys\.|master\.)`),
	}

	// XSS patterns
	xssPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(<script|</script|javascript:|vbscript:|onload|onerror|onclick)`),
		regexp.MustCompile(`(?i)(<iframe|<object|<embed|<form|<input)`),
		regexp.MustCompile(`(?i)(expression\(|@import|eval\(|setTimeout\()`),
	}

	// Path traversal patterns
	pathTraversalPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(\.\./|\.\.\\|%2e%2e%2f|%2e%2e%5c)`),
		regexp.MustCompile(`(\/etc\/|\/proc\/|\/sys\/|c:\\|\.\.)`),
	}

	// Command injection patterns
	commandInjectionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(\||&|;|\$\(|\x60)`),
		regexp.MustCompile(`(?i)(wget|curl|nc|netcat|bash|sh|cmd|powershell)`),
	}

	// Valid patterns
	validIDPattern       = regexp.MustCompile(`^[1-9]\d{0,9}$`)        // 1-10 digits, no leading zeros
	validUsernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`) // 3-30 alphanumeric chars
	validEmailPattern    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	validNamePattern     = regexp.MustCompile(`^[a-zA-Z\s'-]{1,50}$`)
)

// Rate limiting constants
const (
	MaxRequestsPerMinute = 60
	MaxRequestsPerHour   = 1000
	BlockDuration        = 15 * time.Minute
	MaxViolations        = 3
)

// NewSecurityValidator creates a new security validator with rate limiting
func NewSecurityValidator() *SecurityValidator {
	validator := &SecurityValidator{
		requestCounts: make(map[string]*RequestLimit),
	}

	// Start cleanup goroutine for old rate limit entries
	go validator.cleanupRoutine()

	return validator
}

// ValidateAndSanitizeString validates and sanitizes string input
func (s *SecurityValidator) ValidateAndSanitizeString(input, inputType string, maxLength int) (string, error) {
	if input == "" {
		return "", fmt.Errorf("%s cannot be empty", inputType)
	}

	if len(input) > maxLength {
		return "", fmt.Errorf("%s exceeds maximum length of %d characters", inputType, maxLength)
	}

	// Check for malicious patterns
	if err := s.checkMaliciousPatterns(input); err != nil {
		return "", fmt.Errorf("%s contains dangerous content: %v", inputType, err)
	}

	// Validate specific input types
	switch inputType {
	case "username":
		if !validUsernamePattern.MatchString(input) {
			return "", fmt.Errorf("username must be 3-30 characters (letters, numbers, _, -)")
		}
	case "email":
		if !validEmailPattern.MatchString(input) {
			return "", fmt.Errorf("invalid email format")
		}
	case "name":
		if !validNamePattern.MatchString(input) {
			return "", fmt.Errorf("name must contain only letters, spaces, apostrophes, and hyphens")
		}
	}

	// HTML escape and trim
	sanitized := html.EscapeString(strings.TrimSpace(input))
	return sanitized, nil
}

// ValidateID validates and converts string ID to uint
func (s *SecurityValidator) ValidateID(idStr string) (uint, error) {
	if idStr == "" {
		return 0, fmt.Errorf("ID cannot be empty")
	}

	if !validIDPattern.MatchString(idStr) {
		return 0, fmt.Errorf("invalid ID format")
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid ID format")
	}

	if id == 0 {
		return 0, fmt.Errorf("ID cannot be zero")
	}

	return uint(id), nil
}

// CheckRateLimit implements rate limiting per client
func (s *SecurityValidator) CheckRateLimit(clientID string) error {
	s.rateMutex.Lock()
	defer s.rateMutex.Unlock()

	now := time.Now()
	limit, exists := s.requestCounts[clientID]

	if !exists {
		s.requestCounts[clientID] = &RequestLimit{
			Count:      1,
			LastReset:  now,
			Blocked:    false,
			Violations: 0,
		}
		return nil
	}

	limit.mutex.Lock()
	defer limit.mutex.Unlock()

	// Check if client is currently blocked
	if limit.Blocked && now.Before(limit.BlockUntil) {
		return fmt.Errorf("client temporarily blocked due to rate limiting")
	}

	// Reset block status if block period expired
	if limit.Blocked && now.After(limit.BlockUntil) {
		limit.Blocked = false
		limit.Violations = 0
		limit.Count = 0
		limit.LastReset = now
	}

	// Reset counter if minute has passed
	if now.Sub(limit.LastReset) >= time.Minute {
		limit.Count = 0
		limit.LastReset = now
	}

	// Check rate limit
	limit.Count++
	if limit.Count > MaxRequestsPerMinute {
		limit.Violations++

		if limit.Violations >= MaxViolations {
			limit.Blocked = true
			limit.BlockUntil = now.Add(BlockDuration)
			return fmt.Errorf("rate limit exceeded: client blocked for %v", BlockDuration)
		}

		return fmt.Errorf("rate limit exceeded: max %d requests per minute", MaxRequestsPerMinute)
	}

	return nil
}

// checkMaliciousPatterns checks input for malicious patterns
func (s *SecurityValidator) checkMaliciousPatterns(input string) error {
	// Check SQL injection
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(input) {
			return fmt.Errorf("SQL injection attempt detected")
		}
	}

	// Check XSS
	for _, pattern := range xssPatterns {
		if pattern.MatchString(input) {
			return fmt.Errorf("XSS attempt detected")
		}
	}

	// Check path traversal
	for _, pattern := range pathTraversalPatterns {
		if pattern.MatchString(input) {
			return fmt.Errorf("path traversal attempt detected")
		}
	}

	// Check command injection
	for _, pattern := range commandInjectionPatterns {
		if pattern.MatchString(input) {
			return fmt.Errorf("command injection attempt detected")
		}
	}

	return nil
}

// GenerateSecureID generates a cryptographically secure request ID
func (s *SecurityValidator) GenerateSecureID() string {
	// Increment atomic counter
	counter := atomic.AddInt64(&s.requestCounter, 1)

	// Add random bytes
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)

	return fmt.Sprintf("%d-%s", counter, hex.EncodeToString(randomBytes))
}

// cleanupRoutine periodically cleans up old rate limit entries
func (s *SecurityValidator) cleanupRoutine() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.rateMutex.Lock()
		now := time.Now()

		for clientID, limit := range s.requestCounts {
			limit.mutex.Lock()
			// Remove entries older than 1 hour
			if now.Sub(limit.LastReset) > time.Hour {
				delete(s.requestCounts, clientID)
			}
			limit.mutex.Unlock()
		}

		s.rateMutex.Unlock()
	}
}

// SecurityMiddleware creates a Gin middleware for security validation
func (s *SecurityValidator) SecurityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID for tracking
		requestID := s.GenerateSecureID()
		c.Set("request_id", requestID)

		// Get client identifier (IP + User-Agent for better tracking)
		clientID := c.ClientIP() + "|" + c.GetHeader("User-Agent")

		// Check rate limit
		if err := s.CheckRateLimit(clientID); err != nil {
			c.JSON(429, gin.H{
				"error":      "Rate limit exceeded",
				"request_id": requestID,
			})
			c.Abort()
			return
		}

		// Add security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// AuthorizationMiddleware checks if user has required permissions
func AuthorizationMiddleware(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// For now, just check if user is authenticated
		// Permission checking would require database access to check user permissions
		// This can be implemented when a database instance is available in this middleware
		_ = userID
		_ = requiredPermission

		c.Next()
	}
}

// ValidateIDParam validates ID parameter from URL
func (s *SecurityValidator) ValidateIDParam(c *gin.Context, paramName string) (uint, error) {
	idStr := c.Param(paramName)
	return s.ValidateID(idStr)
}

// ValidatePositiveInt validates and parses a positive integer from string
func (s *SecurityValidator) ValidatePositiveInt(value, fieldName string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 0, fmt.Errorf("%s cannot be empty", fieldName)
	}

	// Check for dangerous content
	if err := s.checkMaliciousPatterns(value); err != nil {
		return 0, fmt.Errorf("%s contains dangerous content: %v", fieldName, err)
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer", fieldName)
	}

	if num <= 0 {
		return 0, fmt.Errorf("%s must be positive", fieldName)
	}

	return num, nil
}

// ValidateAndSanitizeEmail validates and sanitizes email addresses
func (s *SecurityValidator) ValidateAndSanitizeEmail(email string) (string, error) {
	// Basic email validation using regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	email = strings.TrimSpace(email)
	if email == "" {
		return "", fmt.Errorf("email cannot be empty")
	}

	// Check for dangerous content
	if err := s.checkMaliciousPatterns(email); err != nil {
		return "", fmt.Errorf("email contains dangerous content: %v", err)
	}

	if !emailRegex.MatchString(email) {
		return "", fmt.Errorf("invalid email format")
	}

	return email, nil
}
