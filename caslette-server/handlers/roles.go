package handlers

import (
	"caslette-server/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleHandler struct {
	db *gorm.DB
}

func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{db: db}
}

// GetRoles returns all roles with their permissions
func (h *RoleHandler) GetRoles(c *gin.Context) {
	var roles []models.Role

	if err := h.db.Preload("Permissions").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// GetRole returns a specific role by ID with its permissions
func (h *RoleHandler) GetRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var role models.Role
	if err := h.db.Preload("Permissions").First(&role, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch role"})
		return
	}

	c.JSON(http.StatusOK, role)
}

// CreateRole creates a new role
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := models.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.db.Create(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
		return
	}

	c.JSON(http.StatusCreated, role)
}

// UpdateRole updates an existing role
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	if err := h.db.First(&role, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch role"})
		return
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	if err := h.db.Save(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, role)
}

// DeleteRole deletes a role
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	if err := h.db.Delete(&models.Role{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role deleted successfully"})
}

// AssignPermissions assigns permissions to a role
func (h *RoleHandler) AssignPermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
		return
	}

	var req struct {
		PermissionIDs []uint `json:"permission_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	if err := h.db.First(&role, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch role"})
		return
	}

	// Fetch permissions
	var permissions []models.Permission
	if err := h.db.Find(&permissions, req.PermissionIDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch permissions"})
		return
	}

	// Replace all permissions for this role
	if err := h.db.Model(&role).Association("Permissions").Replace(permissions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign permissions"})
		return
	}

	// Reload role with permissions
	if err := h.db.Preload("Permissions").First(&role, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload role"})
		return
	}

	c.JSON(http.StatusOK, role)
}
