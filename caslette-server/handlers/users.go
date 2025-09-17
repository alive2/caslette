package handlers

import (
	"caslette-server/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

type UpdateUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	IsActive  *bool  `json:"is_active"`
}

type AssignRoleRequest struct {
	RoleIDs []uint `json:"role_ids" binding:"required"`
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	var users []models.User
	query := h.db.Preload("Roles").Preload("Permissions")

	// Add pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	h.db.Model(&models.User{}).Count(&total)

	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := h.db.Preload("Roles").Preload("Diamonds").First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get current diamond balance
	var currentBalance int64
	h.db.Model(&models.Diamond{}).Where("user_id = ?", id).Order("created_at desc").Limit(1).Pluck("balance", &currentBalance)

	c.JSON(http.StatusOK, gin.H{
		"user":            user,
		"diamond_balance": currentBalance,
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := h.db.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func (h *UserHandler) AssignRoles(c *gin.Context) {
	id := c.Param("id")
	var req AssignRoleRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get roles
	var roles []models.Role
	if err := h.db.Find(&roles, req.RoleIDs).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role IDs"})
		return
	}

	// Replace user roles
	if err := h.db.Model(&user).Association("Roles").Replace(roles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Roles assigned successfully"})
}

// AssignPermissionRequest represents the request to assign permissions to a user
type AssignPermissionRequest struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

// AssignPermissions assigns permissions directly to a user
func (h *UserHandler) AssignPermissions(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req AssignPermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get permissions
	var permissions []models.Permission
	if err := h.db.Find(&permissions, req.PermissionIDs).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission IDs"})
		return
	}

	// Replace user permissions
	if err := h.db.Model(&user).Association("Permissions").Replace(permissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permissions assigned successfully"})
}

// GetUserPermissions gets all permissions assigned to a user (both from roles and direct assignment)
func (h *UserHandler) GetUserPermissions(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user with roles and permissions
	var user models.User
	if err := h.db.Preload("Roles.Permissions").Preload("Permissions").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Collect all permissions (from roles and direct assignments)
	permissionMap := make(map[uint]models.Permission)

	// Add permissions from roles
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			permissionMap[permission.ID] = permission
		}
	}

	// Add direct user permissions
	for _, permission := range user.Permissions {
		permissionMap[permission.ID] = permission
	}

	// Convert map to slice
	var allPermissions []models.Permission
	for _, permission := range permissionMap {
		allPermissions = append(allPermissions, permission)
	}

	c.JSON(http.StatusOK, gin.H{
		"user_permissions": user.Permissions, // Direct user permissions
		"role_permissions": permissionMap,    // All permissions from roles
		"all_permissions":  allPermissions,   // Combined unique permissions
	})
}

// RemoveUserPermission removes a specific permission from a user
func (h *UserHandler) RemoveUserPermission(c *gin.Context) {
	id := c.Param("id")
	permissionID := c.Param("permission_id")

	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	permID, err := strconv.ParseUint(permissionID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	// Check if user exists
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if permission exists
	var permission models.Permission
	if err := h.db.First(&permission, permID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
		return
	}

	// Remove the permission from user
	if err := h.db.Model(&user).Association("Permissions").Delete(&permission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission removed successfully"})
}
