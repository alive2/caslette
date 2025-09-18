package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"unique;not null"`
	Email     string         `json:"email" gorm:"unique;not null"`
	Password  string         `json:"-" gorm:"not null"` // Hidden from JSON responses
	FirstName string         `json:"first_name"`
	LastName  string         `json:"last_name"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Roles       []Role       `json:"roles" gorm:"many2many:user_roles;"`
	Permissions []Permission `json:"permissions" gorm:"many2many:user_permissions;"`
	Diamonds    []Diamond    `json:"diamonds" gorm:"foreignKey:UserID"`
}

type Role struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"unique;not null"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Users       []User       `json:"users" gorm:"many2many:user_roles;"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
}

type Permission struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"unique;not null"`
	Description string         `json:"description"`
	Resource    string         `json:"resource"` // e.g., "users", "games", "transactions"
	Action      string         `json:"action"`   // e.g., "create", "read", "update", "delete"
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Roles []Role `json:"roles" gorm:"many2many:role_permissions;"`
}

type Diamond struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null"`
	Amount        int64          `json:"amount" gorm:"not null"`                // Amount of diamonds (can be negative for deductions)
	Balance       int64          `json:"balance" gorm:"not null"`               // Running balance after this transaction
	TransactionID string         `json:"transaction_id" gorm:"unique;not null"` // Unique transaction identifier
	Type          string         `json:"type" gorm:"not null"`                  // "credit", "debit", "bonus", "purchase", etc.
	Description   string         `json:"description"`
	Metadata      string         `json:"metadata" gorm:"type:json"` // Additional data as JSON
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID"`
}

// BeforeCreate hook for Diamond to generate transaction ID
func (d *Diamond) BeforeCreate(tx *gorm.DB) error {
	if d.TransactionID == "" {
		d.TransactionID = generateTransactionID()
	}
	return nil
}

// generateTransactionID creates a unique transaction ID
func generateTransactionID() string {
	timestamp := time.Now().Unix()
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("TXN_%d_%s", timestamp, hex.EncodeToString(bytes))
}

// UserRole junction table for many-to-many relationship
type UserRole struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

// RolePermission junction table for many-to-many relationship
type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}

// UserPermission junction table for user-level permissions
type UserPermission struct {
	UserID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}
