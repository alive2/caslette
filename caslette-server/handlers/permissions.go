package handlers

import (
	"caslette-server/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PermissionHandler struct {
	db *gorm.DB
}

func NewPermissionHandler(db *gorm.DB) *PermissionHandler {
	return &PermissionHandler{db: db}
}

// GetPermissions returns all permissions
func (h *PermissionHandler) GetPermissions(c *gin.Context) {
	var permissions []models.Permission

	if err := h.db.Find(&permissions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch permissions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"permissions": permissions})
}

// GetPermission returns a specific permission by ID
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	var permission models.Permission
	if err := h.db.First(&permission, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch permission"})
		return
	}

	c.JSON(http.StatusOK, permission)
}

// CreatePermission creates a new permission
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Resource    string `json:"resource" binding:"required"`
		Action      string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	permission := models.Permission{
		Name:        req.Name,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
	}

	if err := h.db.Create(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create permission"})
		return
	}

	c.JSON(http.StatusCreated, permission)
}

// UpdatePermission updates an existing permission
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Resource    string `json:"resource"`
		Action      string `json:"action"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var permission models.Permission
	if err := h.db.First(&permission, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch permission"})
		return
	}

	if req.Name != "" {
		permission.Name = req.Name
	}
	if req.Description != "" {
		permission.Description = req.Description
	}
	if req.Resource != "" {
		permission.Resource = req.Resource
	}
	if req.Action != "" {
		permission.Action = req.Action
	}

	if err := h.db.Save(&permission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permission"})
		return
	}

	c.JSON(http.StatusOK, permission)
}

// DeletePermission deletes a permission
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid permission ID"})
		return
	}

	if err := h.db.Delete(&models.Permission{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete permission"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Permission deleted successfully"})
}
