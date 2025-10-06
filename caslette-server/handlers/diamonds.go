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

	// Verify user exists
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Calculate current balance (sum of all diamond transactions for this user)
	var currentBalance int64
	err = h.db.Model(&models.Diamond{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(amount), 0)").
		Row().Scan(&currentBalance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"diamonds": currentBalance,
		"username": user.Username,
	})
}

func (h *SecureDiamondHandler) AddDiamonds(c *gin.Context) {
	// Basic request validation structure
	var request struct {
		UserID      uint   `json:"user_id" binding:"required"`
		Amount      int    `json:"amount" binding:"required,min=1"`
		Type        string `json:"type"`
		Description string `json:"description" binding:"max=200"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate and sanitize description
	if request.Description != "" {
		sanitizedDescription, err := h.validator.ValidateAndSanitizeString(request.Description, "description", 200)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid description: " + err.Error()})
			return
		}
		request.Description = sanitizedDescription
	}

	// Default type if not provided
	if request.Type == "" {
		request.Type = "credit"
	}

	// Verify user exists
	var user models.User
	if err := h.db.First(&user, request.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get current balance (sum of all diamond transactions for this user)
	var currentBalance int64
	err := tx.Model(&models.Diamond{}).
		Where("user_id = ?", request.UserID).
		Select("COALESCE(SUM(amount), 0)").
		Row().Scan(&currentBalance)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate current balance"})
		return
	}

	// Create new diamond transaction
	newBalance := currentBalance + int64(request.Amount)
	diamond := models.Diamond{
		UserID:      request.UserID,
		Amount:      int64(request.Amount),
		Balance:     newBalance,
		Type:        request.Type,
		Description: request.Description,
		Metadata:    "{}",
	}

	if err := tx.Create(&diamond).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add diamonds"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "diamonds added successfully",
		"user_id":        request.UserID,
		"amount":         request.Amount,
		"new_balance":    newBalance,
		"transaction_id": diamond.TransactionID,
	})
}

func (h *SecureDiamondHandler) DeductDiamonds(c *gin.Context) {
	// Basic request validation structure
	var request struct {
		UserID      uint   `json:"user_id" binding:"required"`
		Amount      int    `json:"amount" binding:"required,min=1"`
		Type        string `json:"type"`
		Description string `json:"description" binding:"max=200"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate and sanitize description
	if request.Description != "" {
		sanitizedDescription, err := h.validator.ValidateAndSanitizeString(request.Description, "description", 200)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid description: " + err.Error()})
			return
		}
		request.Description = sanitizedDescription
	}

	// Default type if not provided
	if request.Type == "" {
		request.Type = "debit"
	}

	// Verify user exists
	var user models.User
	if err := h.db.First(&user, request.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get current balance (sum of all diamond transactions for this user)
	var currentBalance int64
	err := tx.Model(&models.Diamond{}).
		Where("user_id = ?", request.UserID).
		Select("COALESCE(SUM(amount), 0)").
		Row().Scan(&currentBalance)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to calculate current balance"})
		return
	}

	// Check if user has sufficient balance
	if currentBalance < int64(request.Amount) {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error":           "insufficient balance",
			"current_balance": currentBalance,
			"required":        request.Amount,
		})
		return
	}

	// Create new diamond transaction (negative amount for deduction)
	newBalance := currentBalance - int64(request.Amount)
	diamond := models.Diamond{
		UserID:      request.UserID,
		Amount:      -int64(request.Amount), // Negative for deduction
		Balance:     newBalance,
		Type:        request.Type,
		Description: request.Description,
		Metadata:    "{}",
	}

	if err := tx.Create(&diamond).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deduct diamonds"})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "diamonds deducted successfully",
		"user_id":        request.UserID,
		"amount":         request.Amount,
		"new_balance":    newBalance,
		"transaction_id": diamond.TransactionID,
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
