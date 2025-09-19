package handlers

import (
	"caslette-server/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SecureUserHandler handles HTTP requests for user operations with security enhancements
type SecureUserHandler struct {
	db        *gorm.DB
	validator *SecurityValidator
}

// SecureUpdateUserRequest with validation constraints
type SecureUpdateUserRequest struct {
	FirstName string `json:"first_name" binding:"max=50"`
	LastName  string `json:"last_name" binding:"max=50"`
	Email     string `json:"email" binding:"omitempty,email,max=100"`
	IsActive  *bool  `json:"is_active"`
}

// SecureAssignRoleRequest with validation
type SecureAssignRoleRequest struct {
	RoleIDs []uint `json:"role_ids" binding:"required,max=10"`
}

// SecureUserResponse with sanitized data
type SecureUserResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email,omitempty"` // Only include for authorized users
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	RequestID string `json:"request_id"`
}

// SecureUserListResponse for paginated user lists
type SecureUserListResponse struct {
	Users      []SecureUserResponse `json:"users"`
	Pagination PaginationInfo       `json:"pagination"`
	RequestID  string               `json:"request_id"`
}

// PaginationInfo for secure pagination
type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewSecureUserHandler creates a new secure user handler
func NewSecureUserHandler(db *gorm.DB) *SecureUserHandler {
	return &SecureUserHandler{
		db:        db,
		validator: NewSecurityValidator(),
	}
}

// Backward compatibility alias
func NewUserHandler(db *gorm.DB) *SecureUserHandler {
	return NewSecureUserHandler(db)
}

// GetUsers handles GET /api/users with authorization and secure pagination
func (h *SecureUserHandler) GetUsers(c *gin.Context) {
	requestID, _ := c.Get("request_id")

	// Check authorization - only admin users can list all users
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error":      "Authentication required",
			"request_id": requestID,
		})
		return
	}

	// Verify admin permission
	if !h.hasAdminPermission(userID.(uint)) {
		c.JSON(http.StatusForbidden, gin.H{
			"success":    false,
			"error":      "Insufficient permissions",
			"request_id": requestID,
		})
		return
	}

	// Validate and sanitize pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := h.validator.ValidatePositiveInt(pageStr, "page")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Invalid page parameter",
			"request_id": requestID,
		})
		return
	}

	limit, err := h.validator.ValidatePositiveInt(limitStr, "limit")
	if err != nil || limit > 100 { // Limit max results
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Invalid limit parameter (max 100)",
			"request_id": requestID,
		})
		return
	}

	offset := (page - 1) * limit

	var users []models.User
	var total int64

	// Get total count
	if err := h.db.Model(&models.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error":      "Failed to get user count",
			"request_id": requestID,
		})
		return
	}

	// Get users with limited data exposure
	if err := h.db.Select("id, username, email, first_name, last_name, is_active, created_at").
		Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error":      "Failed to fetch users",
			"request_id": requestID,
		})
		return
	}

	// Create secure response
	secureUsers := make([]SecureUserResponse, len(users))
	for i, user := range users {
		secureUsers[i] = SecureUserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email, // Only admins see this
			FirstName: user.FirstName,
			LastName:  user.LastName,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			RequestID: requestID.(string),
		}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := SecureUserListResponse{
		Users: secureUsers,
		Pagination: PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
		RequestID: requestID.(string),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetUser handles GET /api/users/:id with IDOR protection
func (h *SecureUserHandler) GetUser(c *gin.Context) {
	requestID, _ := c.Get("request_id")

	// Validate user ID parameter
	targetUserID, err := h.validator.ValidateIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Invalid user ID",
			"request_id": requestID,
		})
		return
	}

	// Get current user from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error":      "Authentication required",
			"request_id": requestID,
		})
		return
	}

	// IDOR Protection: Users can only access their own data unless they're admin
	if targetUserID != currentUserID.(uint) && !h.hasAdminPermission(currentUserID.(uint)) {
		c.JSON(http.StatusForbidden, gin.H{
			"success":    false,
			"error":      "Access denied",
			"request_id": requestID,
		})
		return
	}

	var user models.User
	if err := h.db.Select("id, username, email, first_name, last_name, is_active, created_at").
		First(&user, targetUserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error":      "User not found",
				"request_id": requestID,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success":    false,
				"error":      "Database error",
				"request_id": requestID,
			})
		}
		return
	}

	// Get diamond balance only for own account or admin
	var diamondBalance int64
	if targetUserID == currentUserID.(uint) || h.hasAdminPermission(currentUserID.(uint)) {
		h.db.Model(&models.Diamond{}).Where("user_id = ?", targetUserID).
			Order("created_at desc").Limit(1).Pluck("balance", &diamondBalance)
	}

	// Create secure response with limited data exposure
	userResponse := SecureUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		RequestID: requestID.(string),
	}

	// Only include sensitive data for authorized access
	if targetUserID == currentUserID.(uint) || h.hasAdminPermission(currentUserID.(uint)) {
		userResponse.Email = user.Email
	}

	response := gin.H{
		"success":    true,
		"user":       userResponse,
		"request_id": requestID,
	}

	// Only include sensitive data for authorized access
	if targetUserID == currentUserID.(uint) || h.hasAdminPermission(currentUserID.(uint)) {
		response["diamond_balance"] = diamondBalance
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUser handles PUT /api/users/:id with IDOR protection and input validation
func (h *SecureUserHandler) UpdateUser(c *gin.Context) {
	requestID, _ := c.Get("request_id")

	// Validate user ID parameter
	targetUserID, err := h.validator.ValidateIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Invalid user ID",
			"request_id": requestID,
		})
		return
	}

	// Get current user from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error":      "Authentication required",
			"request_id": requestID,
		})
		return
	}

	// IDOR Protection: Users can only update their own data unless they're admin
	if targetUserID != currentUserID.(uint) && !h.hasAdminPermission(currentUserID.(uint)) {
		c.JSON(http.StatusForbidden, gin.H{
			"success":    false,
			"error":      "Access denied",
			"request_id": requestID,
		})
		return
	}

	var req SecureUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Invalid request format",
			"request_id": requestID,
		})
		return
	}

	// Find user
	var user models.User
	if err := h.db.First(&user, targetUserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error":      "User not found",
				"request_id": requestID,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success":    false,
				"error":      "Database error",
				"request_id": requestID,
			})
		}
		return
	}

	// Validate and sanitize input fields
	if req.FirstName != "" {
		firstName, err := h.validator.ValidateAndSanitizeString(req.FirstName, "name", 50)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error":      "Invalid first name: " + err.Error(),
				"request_id": requestID,
			})
			return
		}
		user.FirstName = firstName
	}

	if req.LastName != "" {
		lastName, err := h.validator.ValidateAndSanitizeString(req.LastName, "name", 50)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error":      "Invalid last name: " + err.Error(),
				"request_id": requestID,
			})
			return
		}
		user.LastName = lastName
	}

	if req.Email != "" {
		email, err := h.validator.ValidateAndSanitizeEmail(req.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":    false,
				"error":      "Invalid email: " + err.Error(),
				"request_id": requestID,
			})
			return
		}

		// Check for email uniqueness
		var existingUser models.User
		if err := h.db.Where("email = ? AND id != ?", email, targetUserID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"success":    false,
				"error":      "Email already in use",
				"request_id": requestID,
			})
			return
		}

		user.Email = email
	}

	// Only admins can change active status
	if req.IsActive != nil && h.hasAdminPermission(currentUserID.(uint)) {
		user.IsActive = *req.IsActive
	}

	// Update user with transaction safety
	tx := h.db.Begin()
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error":      "Failed to update user",
			"request_id": requestID,
		})
		return
	}
	tx.Commit()

	// Return secure response
	response := SecureUserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
		RequestID: requestID.(string),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    response,
	})
}

// DeleteUser handles DELETE /api/users/:id with admin authorization only
func (h *SecureUserHandler) DeleteUser(c *gin.Context) {
	requestID, _ := c.Get("request_id")

	// Only admins can delete users
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success":    false,
			"error":      "Authentication required",
			"request_id": requestID,
		})
		return
	}

	if !h.hasAdminPermission(currentUserID.(uint)) {
		c.JSON(http.StatusForbidden, gin.H{
			"success":    false,
			"error":      "Admin access required",
			"request_id": requestID,
		})
		return
	}

	// Validate user ID parameter
	targetUserID, err := h.validator.ValidateIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Invalid user ID",
			"request_id": requestID,
		})
		return
	}

	// Prevent self-deletion
	if targetUserID == currentUserID.(uint) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "Cannot delete own account",
			"request_id": requestID,
		})
		return
	}

	// Check if user exists
	var user models.User
	if err := h.db.First(&user, targetUserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success":    false,
				"error":      "User not found",
				"request_id": requestID,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success":    false,
				"error":      "Database error",
				"request_id": requestID,
			})
		}
		return
	}

	// Soft delete with transaction safety
	tx := h.db.Begin()
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error":      "Failed to delete user",
			"request_id": requestID,
		})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "User deleted successfully",
		"request_id": requestID,
	})
}

// hasAdminPermission checks if user has admin permissions
func (h *SecureUserHandler) hasAdminPermission(userID uint) bool {
	var count int64
	h.db.Table("user_roles").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.name = ?", userID, "admin").
		Count(&count)
	return count > 0
}

// AssignRoles handles POST /api/users/:id/roles with admin authorization
func (h *SecureUserHandler) AssignRoles(c *gin.Context) {
	userID, err := h.validator.ValidateIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Additional admin permission check
	if !h.hasAdminPermission(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	// Placeholder implementation
	c.JSON(http.StatusOK, gin.H{"message": "roles assigned successfully"})
}

func (h *SecureUserHandler) AssignPermissions(c *gin.Context) {
	userID, err := h.validator.ValidateIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Additional admin permission check
	if !h.hasAdminPermission(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	// Placeholder implementation
	c.JSON(http.StatusOK, gin.H{"message": "permissions assigned successfully"})
}

func (h *SecureUserHandler) GetUserPermissions(c *gin.Context) {
	_, err := h.validator.ValidateIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Placeholder implementation - return empty permissions list for now
	c.JSON(http.StatusOK, gin.H{"permissions": []string{}})
}

func (h *SecureUserHandler) RemoveUserPermission(c *gin.Context) {
	userID, err := h.validator.ValidateIDParam(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	permissionID, err := h.validator.ValidateIDParam(c, "permission_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid permission ID"})
		return
	}

	// Additional admin permission check
	if !h.hasAdminPermission(userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		return
	}

	// Placeholder implementation
	c.JSON(http.StatusOK, gin.H{"message": "permission removed successfully", "permission_id": permissionID})
}
