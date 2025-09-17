package handlers

import (
	"caslette-server/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DiamondHandler struct {
	db *gorm.DB
}

type DiamondTransactionRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`
	Amount      int64  `json:"amount" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Description string `json:"description"`
}

func NewDiamondHandler(db *gorm.DB) *DiamondHandler {
	return &DiamondHandler{db: db}
}

func (h *DiamondHandler) GetUserDiamonds(c *gin.Context) {
	userID := c.Param("userId")

	var diamonds []models.Diamond
	if err := h.db.Where("user_id = ?", userID).Order("created_at desc").Find(&diamonds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch diamond transactions"})
		return
	}

	// Get current balance
	var currentBalance int64
	if len(diamonds) > 0 {
		currentBalance = diamonds[0].Balance
	}

	c.JSON(http.StatusOK, gin.H{
		"diamonds":        diamonds,
		"current_balance": currentBalance,
	})
}

func (h *DiamondHandler) AddDiamonds(c *gin.Context) {
	var req DiamondTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate user exists
	var user models.User
	if err := h.db.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get current balance
	var currentBalance int64
	h.db.Model(&models.Diamond{}).Where("user_id = ?", req.UserID).Order("created_at desc").Limit(1).Pluck("balance", &currentBalance)

	// Create transaction
	diamond := models.Diamond{
		UserID:      req.UserID,
		Amount:      req.Amount,
		Balance:     currentBalance + req.Amount,
		Type:        req.Type,
		Description: req.Description,
		Metadata:    "{}",
	}

	if err := h.db.Create(&diamond).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create diamond transaction"})
		return
	}

	c.JSON(http.StatusCreated, diamond)
}

func (h *DiamondHandler) DeductDiamonds(c *gin.Context) {
	var req DiamondTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate user exists
	var user models.User
	if err := h.db.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get current balance
	var currentBalance int64
	h.db.Model(&models.Diamond{}).Where("user_id = ?", req.UserID).Order("created_at desc").Limit(1).Pluck("balance", &currentBalance)

	// Check if user has sufficient balance
	if currentBalance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient diamond balance"})
		return
	}

	// Create deduction transaction (negative amount)
	diamond := models.Diamond{
		UserID:      req.UserID,
		Amount:      -req.Amount,
		Balance:     currentBalance - req.Amount,
		Type:        req.Type,
		Description: req.Description,
		Metadata:    "{}",
	}

	if err := h.db.Create(&diamond).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create diamond transaction"})
		return
	}

	c.JSON(http.StatusCreated, diamond)
}

func (h *DiamondHandler) GetAllTransactions(c *gin.Context) {
	var diamonds []models.Diamond
	query := h.db.Preload("User")

	// Add pagination
	page := 1
	limit := 50
	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			page = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	offset := (page - 1) * limit

	var total int64
	h.db.Model(&models.Diamond{}).Count(&total)

	if err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&diamonds).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch diamond transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"transactions": diamonds,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}
