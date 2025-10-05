package handlers

import (
	"caslette-server/auth"
	"caslette-server/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SecureAuthHandler struct {
	db          *gorm.DB
	authService *auth.AuthService
	validator   *SecurityValidator
}

type SecureLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SecureRegisterRequest struct {
	Username  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type SecureAuthResponse struct {
	Token     string     `json:"token"`
	User      SecureUser `json:"user"`
	RequestID string     `json:"request_id"`
}

type SecureUser struct {
	ID        uint       `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	IsActive  bool       `json:"is_active"`
	Roles     []UserRole `json:"roles"`
	// Note: Password and sensitive data excluded
}

type UserRole struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewSecureAuthHandler(db *gorm.DB, authService *auth.AuthService) *SecureAuthHandler {
	return &SecureAuthHandler{
		db:          db,
		authService: authService,
		validator:   NewSecurityValidator(),
	}
}

// Backward compatibility alias
func NewAuthHandler(db *gorm.DB, authService *auth.AuthService) *SecureAuthHandler {
	return NewSecureAuthHandler(db, authService)
}

func (h *SecureAuthHandler) Register(c *gin.Context) {
	requestID, _ := c.Get("request_id")

	var req SecureRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request format",
			"request_id": requestID,
		})
		return
	}

	// Validate and sanitize inputs
	username, err := h.validator.ValidateAndSanitizeString(req.Username, "username", 30)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      err.Error(),
			"request_id": requestID,
		})
		return
	}

	email, err := h.validator.ValidateAndSanitizeString(req.Email, "email", 255)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      err.Error(),
			"request_id": requestID,
		})
		return
	}

	firstName, err := h.validator.ValidateAndSanitizeString(req.FirstName, "name", 50)
	if err != nil && req.FirstName != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      err.Error(),
			"request_id": requestID,
		})
		return
	}

	lastName, err := h.validator.ValidateAndSanitizeString(req.LastName, "name", 50)
	if err != nil && req.LastName != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      err.Error(),
			"request_id": requestID,
		})
		return
	}

	// Additional password validation
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Password must be at least 8 characters",
			"request_id": requestID,
		})
		return
	}

	// Check if user already exists using prepared statement pattern
	var existingUser models.User
	if err := h.db.Where("username = ? OR email = ?", username, email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":      "User already exists",
			"request_id": requestID,
		})
		return
	}

	// Hash password
	hashedPassword, err := h.authService.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Registration failed",
			"request_id": requestID,
		})
		return
	}

	// Create user with validated data
	user := models.User{
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		FirstName: firstName,
		LastName:  lastName,
		IsActive:  true,
	}

	// Use transaction for data consistency
	tx := h.db.Begin()
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Registration failed",
			"request_id": requestID,
		})
		return
	}

	// Assign default user role
	var defaultRole models.Role
	if err := tx.Where("name = ?", "user").First(&defaultRole).Error; err == nil {
		tx.Model(&user).Association("Roles").Append(&defaultRole)
	}

	// Create initial diamond balance (1000 starting diamonds)
	diamond := models.Diamond{
		UserID:      user.ID,
		Amount:      1000,
		Balance:     1000,
		Type:        "bonus",
		Description: "Welcome bonus",
	}
	if err := tx.Create(&diamond).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Registration failed",
			"request_id": requestID,
		})
		return
	}

	tx.Commit()

	// Generate token
	token, err := h.authService.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Registration completed but login failed",
			"request_id": requestID,
		})
		return
	}

	// Return secure response (no sensitive data)
	secureUser := SecureUser{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsActive:  user.IsActive,
	}

	c.JSON(http.StatusCreated, SecureAuthResponse{
		Token:     token,
		User:      secureUser,
		RequestID: requestID.(string),
	})
}

func (h *SecureAuthHandler) Login(c *gin.Context) {
	requestID, _ := c.Get("request_id")

	var req SecureLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid request format",
			"request_id": requestID,
		})
		return
	}

	// Validate and sanitize inputs
	username, err := h.validator.ValidateAndSanitizeString(req.Username, "username", 255)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid credentials",
			"request_id": requestID,
		})
		return
	}

	// Basic password validation (don't reveal too much in error)
	if len(req.Password) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":      "Invalid credentials",
			"request_id": requestID,
		})
		return
	}

	// Find user using prepared statement
	var user models.User
	if err := h.db.Preload("Roles").Where("username = ? OR email = ?", username, username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "Invalid credentials",
			"request_id": requestID,
		})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "Account disabled",
			"request_id": requestID,
		})
		return
	}

	// Verify password
	if err := h.authService.CheckPassword(user.Password, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "Invalid credentials",
			"request_id": requestID,
		})
		return
	}

	// Generate token
	token, err := h.authService.GenerateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Login failed",
			"request_id": requestID,
		})
		return
	}

	// Return secure response
	// Convert roles to secure format
	secureRoles := make([]UserRole, len(user.Roles))
	for i, role := range user.Roles {
		secureRoles[i] = UserRole{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
		}
	}

	secureUser := SecureUser{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsActive:  user.IsActive,
		Roles:     secureRoles,
	}

	c.JSON(http.StatusOK, SecureAuthResponse{
		Token:     token,
		User:      secureUser,
		RequestID: requestID.(string),
	})
}

func (h *SecureAuthHandler) GetProfile(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":      "Authentication required",
			"request_id": requestID,
		})
		return
	}

	var user models.User
	if err := h.db.Preload("Roles").Preload("Diamonds").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":      "User not found",
			"request_id": requestID,
		})
		return
	}

	// Calculate current diamond balance securely
	var currentBalance int64
	h.db.Model(&models.Diamond{}).Where("user_id = ?", userID).Order("created_at desc").Limit(1).Pluck("balance", &currentBalance)

	// Return secure response
	// Convert roles to secure format
	secureRoles := make([]UserRole, len(user.Roles))
	for i, role := range user.Roles {
		secureRoles[i] = UserRole{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
		}
	}

	secureUser := SecureUser{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsActive:  user.IsActive,
		Roles:     secureRoles,
	}

	response := gin.H{
		"user":            secureUser,
		"diamond_balance": currentBalance,
		"request_id":      requestID,
	}

	c.JSON(http.StatusOK, response)
}
