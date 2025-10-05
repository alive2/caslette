package handlers

import (
	"caslette-server/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SecureDiamondHandler struct {
	db        *gorm.DB
	validator *SecurityValidator
}

func NewSecureDiamondHandler(db *gorm.DB) *SecureDiamondHandler {
	return &SecureDiamondHandler{db: db, validator: NewSecurityValidator()}
}

func NewDiamondHandler(db *gorm.DB) *SecureDiamondHandler {
	return NewSecureDiamondHandler(db)
}

func (h *SecureDiamondHandler) GetUserDiamonds(c *gin.Context) {
	userID, err := h.validator.ValidateIDParam(c, "userId")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Placeholder implementation - return 0 diamonds for now
	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"diamonds": 0,
	})
}

func (h *SecureDiamondHandler) AddDiamonds(c *gin.Context) {
	// Basic request validation structure
	var request struct {
		UserID uint   `json:"user_id" binding:"required"`
		Amount int    `json:"amount" binding:"required,min=1"`
		Reason string `json:"reason" binding:"max=200"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate and sanitize reason
	if request.Reason != "" {
		sanitizedReason, err := h.validator.ValidateAndSanitizeString(request.Reason, "reason", 200)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reason: " + err.Error()})
			return
		}
		request.Reason = sanitizedReason
	}

	// Placeholder implementation
	c.JSON(http.StatusOK, gin.H{
		"message": "diamonds added successfully",
		"user_id": request.UserID,
		"amount":  request.Amount,
	})
}

func (h *SecureDiamondHandler) DeductDiamonds(c *gin.Context) {
	// Basic request validation structure
	var request struct {
		UserID uint   `json:"user_id" binding:"required"`
		Amount int    `json:"amount" binding:"required,min=1"`
		Reason string `json:"reason" binding:"max=200"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate and sanitize reason
	if request.Reason != "" {
		sanitizedReason, err := h.validator.ValidateAndSanitizeString(request.Reason, "reason", 200)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reason: " + err.Error()})
			return
		}
		request.Reason = sanitizedReason
	}

	// Placeholder implementation
	c.JSON(http.StatusOK, gin.H{
		"message": "diamonds deducted successfully",
		"user_id": request.UserID,
		"amount":  request.Amount,
	})
}

func (h *SecureDiamondHandler) GetAllTransactions(c *gin.Context) {
	requestID, _ := c.Get("request_id")

	// Parse pagination parameters
	page := 1
	limit := 50

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := h.validator.ValidatePositiveInt(pageStr, "page"); err == nil {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := h.validator.ValidatePositiveInt(limitStr, "limit"); err == nil && l <= 100 {
			limit = l
		}
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get transactions with pagination
	var transactions []models.Diamond
	var total int64

	// Count total transactions
	h.db.Model(&models.Diamond{}).Count(&total)

	// Fetch transactions with user info
	if err := h.db.Preload("User").
		Order("created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to fetch transactions",
			"request_id": requestID,
		})
		return
	}

	// Calculate pagination info
	totalPages := (int(total) + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"transactions": transactions,
			"pagination": gin.H{
				"page":        page,
				"limit":       limit,
				"total":       total,
				"total_pages": totalPages,
			},
		},
		"success":    true,
		"request_id": requestID,
	})
}
