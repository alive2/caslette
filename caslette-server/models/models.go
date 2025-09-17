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

// generateTransactionID creates a unique transaction ID
func generateTransactionID() string {
	timestamp := time.Now().Unix()
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("TXN_%d_%s", timestamp, hex.EncodeToString(bytes))
}

// PokerTable represents a poker table where games are played
type PokerTable struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	GameType    string         `json:"game_type" gorm:"not null;default:'texas_holdem'"` // texas_holdem, omaha, etc.
	MaxPlayers  int            `json:"max_players" gorm:"not null;default:9"`
	MinBuyIn    int64          `json:"min_buy_in" gorm:"not null"`                // Minimum diamonds to join
	MaxBuyIn    int64          `json:"max_buy_in" gorm:"not null"`                // Maximum diamonds to join
	SmallBlind  int64          `json:"small_blind" gorm:"not null"`               // Small blind amount
	BigBlind    int64          `json:"big_blind" gorm:"not null"`                 // Big blind amount
	RakePercent float64        `json:"rake_percent" gorm:"not null;default:0.05"` // House rake (5%)
	MaxRake     int64          `json:"max_rake" gorm:"not null"`                  // Maximum rake per hand
	Status      string         `json:"status" gorm:"not null;default:'waiting'"`  // waiting, playing, paused
	CreatedBy   uint           `json:"created_by" gorm:"not null"`                // User who created the table
	IsPrivate   bool           `json:"is_private" gorm:"default:false"`
	Password    string         `json:"-"` // Hidden from JSON, for private tables
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Creator      User          `json:"creator" gorm:"foreignKey:CreatedBy"`
	Players      []TablePlayer `json:"players" gorm:"foreignKey:TableID"`
	GameHands    []GameHand    `json:"game_hands" gorm:"foreignKey:TableID"`
	Transactions []Transaction `json:"transactions" gorm:"foreignKey:TableID"`
}

// TablePlayer represents a player sitting at a poker table
type TablePlayer struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	TableID    uint           `json:"table_id" gorm:"not null"`
	UserID     uint           `json:"user_id" gorm:"not null"`
	SeatNumber int            `json:"seat_number" gorm:"not null"`              // 1-9 seat position
	ChipCount  int64          `json:"chip_count" gorm:"not null"`               // Current chips at table
	Status     string         `json:"status" gorm:"not null;default:'sitting'"` // sitting, playing, sitting_out
	JoinedAt   time.Time      `json:"joined_at"`
	LeftAt     *time.Time     `json:"left_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Table User `json:"table" gorm:"foreignKey:TableID"`
	User  User `json:"user" gorm:"foreignKey:UserID"`
}

// GameHand represents a single hand of poker
type GameHand struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	TableID        uint           `json:"table_id" gorm:"not null"`
	HandNumber     int            `json:"hand_number" gorm:"not null"`     // Sequential hand number for this table
	DealerPosition int            `json:"dealer_position" gorm:"not null"` // Seat number of dealer
	SmallBlindSeat int            `json:"small_blind_seat" gorm:"not null"`
	BigBlindSeat   int            `json:"big_blind_seat" gorm:"not null"`
	CommunityCards string         `json:"community_cards"`                          // JSON array of cards
	PotAmount      int64          `json:"pot_amount" gorm:"not null"`               // Total pot amount
	RakeAmount     int64          `json:"rake_amount" gorm:"not null"`              // House rake taken
	Status         string         `json:"status" gorm:"not null;default:'preflop'"` // preflop, flop, turn, river, finished
	WinnerUserID   *uint          `json:"winner_user_id"`                           // Winner of the hand (if any)
	StartedAt      time.Time      `json:"started_at"`
	FinishedAt     *time.Time     `json:"finished_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Table        PokerTable    `json:"table" gorm:"foreignKey:TableID"`
	Winner       *User         `json:"winner" gorm:"foreignKey:WinnerUserID"`
	PlayerHands  []PlayerHand  `json:"player_hands" gorm:"foreignKey:GameHandID"`
	Bets         []Bet         `json:"bets" gorm:"foreignKey:GameHandID"`
	Transactions []Transaction `json:"transactions" gorm:"foreignKey:GameHandID"`
}

// PlayerHand represents a player's cards and actions in a specific hand
type PlayerHand struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	GameHandID uint           `json:"game_hand_id" gorm:"not null"`
	UserID     uint           `json:"user_id" gorm:"not null"`
	SeatNumber int            `json:"seat_number" gorm:"not null"`
	HoleCards  string         `json:"hole_cards"`                              // JSON array of 2 cards
	BestHand   string         `json:"best_hand"`                               // Best 5-card hand
	HandRank   int            `json:"hand_rank"`                               // Hand strength ranking
	TotalBet   int64          `json:"total_bet" gorm:"not null;default:0"`     // Total amount bet in this hand
	Status     string         `json:"status" gorm:"not null;default:'active'"` // active, folded, all_in
	Position   string         `json:"position"`                                // sb, bb, utg, mp, co, btn
	LastAction string         `json:"last_action"`                             // fold, call, raise, check, all_in
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	GameHand GameHand `json:"game_hand" gorm:"foreignKey:GameHandID"`
	User     User     `json:"user" gorm:"foreignKey:UserID"`
}

// Bet represents a betting action in a poker hand
type Bet struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	GameHandID   uint           `json:"game_hand_id" gorm:"not null"`
	UserID       uint           `json:"user_id" gorm:"not null"`
	BettingRound string         `json:"betting_round" gorm:"not null"`    // preflop, flop, turn, river
	Action       string         `json:"action" gorm:"not null"`           // fold, call, raise, check, all_in
	Amount       int64          `json:"amount" gorm:"not null;default:0"` // Amount bet/raised
	TotalBet     int64          `json:"total_bet" gorm:"not null"`        // Total bet amount in this round
	Sequence     int            `json:"sequence" gorm:"not null"`         // Order of action in betting round
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	GameHand GameHand `json:"game_hand" gorm:"foreignKey:GameHandID"`
	User     User     `json:"user" gorm:"foreignKey:UserID"`
}

// Transaction represents all diamond movements related to poker games
type Transaction struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null"`
	TableID       *uint          `json:"table_id"`               // Optional, for game-related transactions
	GameHandID    *uint          `json:"game_hand_id"`           // Optional, for hand-specific transactions
	Amount        int64          `json:"amount" gorm:"not null"` // Amount (positive for credits, negative for debits)
	Type          string         `json:"type" gorm:"not null"`   // buy_in, cash_out, bet, win, rake, refund
	Description   string         `json:"description"`
	TransactionID string         `json:"transaction_id" gorm:"unique;not null"`      // Unique transaction identifier
	Status        string         `json:"status" gorm:"not null;default:'completed'"` // pending, completed, failed, cancelled
	Metadata      string         `json:"metadata" gorm:"type:json"`                  // Additional data as JSON
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User     User        `json:"user" gorm:"foreignKey:UserID"`
	Table    *PokerTable `json:"table" gorm:"foreignKey:TableID"`
	GameHand *GameHand   `json:"game_hand" gorm:"foreignKey:GameHandID"`
}

// BeforeCreate hook for Transaction to generate transaction ID
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.TransactionID == "" {
		t.TransactionID = generateGameTransactionID()
	}
	return nil
}

// generateGameTransactionID creates a unique transaction ID for game transactions
func generateGameTransactionID() string {
	timestamp := time.Now().Unix()
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("GAME_%d_%s", timestamp, hex.EncodeToString(bytes))
}
