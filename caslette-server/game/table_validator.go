package game

import (
	"fmt"
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

// TableValidator handles input validation for table operations
type TableValidator struct{}

// Constants for validation limits
const (
	MaxTableNameLength = 100
	MaxDescriptionLength = 500
	MaxPasswordLength = 50
	MaxUsernameLength = 30
	MaxTagsCount = 10
	MaxTagLength = 20
	
	MinTableNameLength = 3
	MinPasswordLength = 4
	MinBuyIn = 1
	MaxBuyIn = 1000000
	MinBlind = 1
	MaxBlind = 100000
	MaxTimeLimit = 300 // 5 minutes max per turn
)

var (
	// Valid characters for table names (alphanumeric, spaces, basic punctuation)
	tableNameRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\-_\.!?']{3,100}$`)
	// Valid characters for usernames (alphanumeric and underscore)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)
	// Valid characters for tags (alphanumeric and hyphen)
	tagRegex = regexp.MustCompile(`^[a-zA-Z0-9\-]{1,20}$`)
)

// NewTableValidator creates a new table validator
func NewTableValidator() *TableValidator {
	return &TableValidator{}
}

// ValidateTableCreateRequest validates a table creation request
func (v *TableValidator) ValidateTableCreateRequest(req *TableCreateRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	
	// Validate table name
	if err := v.ValidateTableName(req.Name); err != nil {
		return fmt.Errorf("invalid table name: %w", err)
	}
	
	// Validate username
	if err := v.ValidateUsername(req.Username); err != nil {
		return fmt.Errorf("invalid username: %w", err)
	}
	
	// Validate creator ID
	if err := v.ValidateUserID(req.CreatedBy); err != nil {
		return fmt.Errorf("invalid creator ID: %w", err)
	}
	
	// Validate game type
	if err := v.ValidateGameType(req.GameType); err != nil {
		return fmt.Errorf("invalid game type: %w", err)
	}
	
	// Validate description
	if err := v.ValidateDescription(req.Description); err != nil {
		return fmt.Errorf("invalid description: %w", err)
	}
	
	// Validate tags
	if err := v.ValidateTags(req.Tags); err != nil {
		return fmt.Errorf("invalid tags: %w", err)
	}
	
	// Validate settings
	if err := v.ValidateTableSettings(req.Settings); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}
	
	return nil
}

// ValidateTableJoinRequest validates a table join request
func (v *TableValidator) ValidateTableJoinRequest(req *TableJoinRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}
	
	// Validate table ID
	if err := v.ValidateTableID(req.TableID); err != nil {
		return fmt.Errorf("invalid table ID: %w", err)
	}
	
	// Validate username
	if err := v.ValidateUsername(req.Username); err != nil {
		return fmt.Errorf("invalid username: %w", err)
	}
	
	// Validate user ID
	if err := v.ValidateUserID(req.PlayerID); err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}
	
	// Validate join mode
	if err := v.ValidateJoinMode(req.Mode); err != nil {
		return fmt.Errorf("invalid join mode: %w", err)
	}
	
	// Validate position if specified (note: Position is int, not pointer)
	if req.Position < 0 {
		if err := v.ValidatePosition(req.Position); err != nil {
			return fmt.Errorf("invalid position: %w", err)
		}
	}
	
	// Validate password (sanitize HTML)
	if req.Password != "" {
		req.Password = v.SanitizeInput(req.Password)
		if len(req.Password) > MaxPasswordLength {
			return fmt.Errorf("password too long (max %d characters)", MaxPasswordLength)
		}
	}
	
	return nil
}

// ValidateTableName validates table names
func (v *TableValidator) ValidateTableName(name string) error {
	if name == "" {
		return fmt.Errorf("table name cannot be empty")
	}
	
	// Sanitize HTML and trim spaces
	name = strings.TrimSpace(v.SanitizeInput(name))
	
	if len(name) < MinTableNameLength {
		return fmt.Errorf("table name too short (min %d characters)", MinTableNameLength)
	}
	
	if len(name) > MaxTableNameLength {
		return fmt.Errorf("table name too long (max %d characters)", MaxTableNameLength)
	}
	
	if !utf8.ValidString(name) {
		return fmt.Errorf("table name contains invalid UTF-8 characters")
	}
	
	if !tableNameRegex.MatchString(name) {
		return fmt.Errorf("table name contains invalid characters")
	}
	
	// Check for common injection patterns
	if v.containsSQLInjectionPatterns(name) {
		return fmt.Errorf("table name contains invalid patterns")
	}
	
	return nil
}

// ValidateUsername validates usernames
func (v *TableValidator) ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	
	username = strings.TrimSpace(username)
	
	if len(username) < 3 {
		return fmt.Errorf("username too short (min 3 characters)")
	}
	
	if len(username) > MaxUsernameLength {
		return fmt.Errorf("username too long (max %d characters)", MaxUsernameLength)
	}
	
	if !utf8.ValidString(username) {
		return fmt.Errorf("username contains invalid UTF-8 characters")
	}
	
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("username contains invalid characters")
	}
	
	return nil
}

// ValidateUserID validates user IDs
func (v *TableValidator) ValidateUserID(userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID cannot be empty")
	}
	
	userID = strings.TrimSpace(userID)
	
	if len(userID) < 1 || len(userID) > 100 {
		return fmt.Errorf("user ID length invalid")
	}
	
	if !utf8.ValidString(userID) {
		return fmt.Errorf("user ID contains invalid UTF-8 characters")
	}
	
	// Basic alphanumeric validation for user IDs
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-_]+$`, userID); !matched {
		return fmt.Errorf("user ID contains invalid characters")
	}
	
	return nil
}

// ValidateTableID validates table IDs
func (v *TableValidator) ValidateTableID(tableID string) error {
	if tableID == "" {
		return fmt.Errorf("table ID cannot be empty")
	}
	
	tableID = strings.TrimSpace(tableID)
	
	// Table IDs should be hex strings of specific length
	if matched, _ := regexp.MatchString(`^[a-f0-9]{16}$`, tableID); !matched {
		return fmt.Errorf("table ID format invalid")
	}
	
	return nil
}

// ValidateGameType validates game types
func (v *TableValidator) ValidateGameType(gameType GameType) error {
	switch gameType {
	case GameTypeTexasHoldem:
		return nil
	default:
		return fmt.Errorf("unsupported game type: %s", gameType)
	}
}

// ValidateDescription validates table descriptions
func (v *TableValidator) ValidateDescription(description string) error {
	if description == "" {
		return nil // Description is optional
	}
	
	// Sanitize HTML
	description = v.SanitizeInput(description)
	
	if len(description) > MaxDescriptionLength {
		return fmt.Errorf("description too long (max %d characters)", MaxDescriptionLength)
	}
	
	if !utf8.ValidString(description) {
		return fmt.Errorf("description contains invalid UTF-8 characters")
	}
	
	// Check for injection patterns
	if v.containsSQLInjectionPatterns(description) {
		return fmt.Errorf("description contains invalid patterns")
	}
	
	return nil
}

// ValidateTags validates table tags
func (v *TableValidator) ValidateTags(tags []string) error {
	if len(tags) > MaxTagsCount {
		return fmt.Errorf("too many tags (max %d)", MaxTagsCount)
	}
	
	for i, tag := range tags {
		tag = strings.TrimSpace(v.SanitizeInput(tag))
		
		if len(tag) == 0 {
			return fmt.Errorf("tag %d is empty", i+1)
		}
		
		if len(tag) > MaxTagLength {
			return fmt.Errorf("tag %d too long (max %d characters)", i+1, MaxTagLength)
		}
		
		if !utf8.ValidString(tag) {
			return fmt.Errorf("tag %d contains invalid UTF-8 characters", i+1)
		}
		
		if !tagRegex.MatchString(tag) {
			return fmt.Errorf("tag %d contains invalid characters", i+1)
		}
	}
	
	return nil
}

// ValidateTableSettings validates table settings
func (v *TableValidator) ValidateTableSettings(settings TableSettings) error {
	// Validate blinds
	if settings.SmallBlind < MinBlind || settings.SmallBlind > MaxBlind {
		return fmt.Errorf("small blind out of range (%d-%d)", MinBlind, MaxBlind)
	}
	
	if settings.BigBlind < MinBlind || settings.BigBlind > MaxBlind {
		return fmt.Errorf("big blind out of range (%d-%d)", MinBlind, MaxBlind)
	}
	
	if settings.BigBlind <= settings.SmallBlind {
		return fmt.Errorf("big blind must be greater than small blind")
	}
	
	// Validate buy-in
	if settings.BuyIn < MinBuyIn || settings.BuyIn > MaxBuyIn {
		return fmt.Errorf("buy-in out of range (%d-%d)", MinBuyIn, MaxBuyIn)
	}
	
	if settings.MaxBuyIn > 0 && settings.MaxBuyIn < settings.BuyIn {
		return fmt.Errorf("max buy-in must be greater than or equal to buy-in")
	}
	
	// Validate time limit
	if settings.TimeLimit < 0 || settings.TimeLimit > MaxTimeLimit {
		return fmt.Errorf("time limit out of range (0-%d seconds)", MaxTimeLimit)
	}
	
	// Validate password
	if settings.Password != "" {
		settings.Password = v.SanitizeInput(settings.Password)
		if len(settings.Password) < MinPasswordLength {
			return fmt.Errorf("password too short (min %d characters)", MinPasswordLength)
		}
		if len(settings.Password) > MaxPasswordLength {
			return fmt.Errorf("password too long (max %d characters)", MaxPasswordLength)
		}
	}
	
	return nil
}

// ValidateJoinMode validates join modes
func (v *TableValidator) ValidateJoinMode(mode TableJoinMode) error {
	switch mode {
	case JoinModePlayer, JoinModeObserver:
		return nil
	default:
		return fmt.Errorf("invalid join mode: %s", mode)
	}
}

// ValidatePosition validates player positions
func (v *TableValidator) ValidatePosition(position int) error {
	if position < 0 || position >= 10 { // Assuming max 10 players
		return fmt.Errorf("position out of range (0-9)")
	}
	return nil
}

// SanitizeInput sanitizes user input to prevent XSS
func (v *TableValidator) SanitizeInput(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)
	
	// Escape HTML entities
	input = html.EscapeString(input)
	
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	return input
}

// containsSQLInjectionPatterns checks for common SQL injection patterns
func (v *TableValidator) containsSQLInjectionPatterns(input string) bool {
	input = strings.ToLower(input)
	
	// Common SQL injection patterns (excluding single quotes for normal text)
	patterns := []string{
		"\"", ";", "--", "/*", "*/", "xp_", "sp_",
		"exec", "execute", "select", "insert", "update", "delete",
		"drop", "create", "alter", "union", "script", "javascript",
		"vbscript", "onload", "onerror", "onclick", "\x00",
	}
	
	for _, pattern := range patterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}
	
	return false
}

// ValidateFilterRequest validates table filtering requests
func (v *TableValidator) ValidateFilterRequest(filters map[string]interface{}) error {
	allowedFilters := map[string]bool{
		"game_type":   true,
		"status":      true,
		"has_space":   true,
		"created_by":  true,
		"max_players": true,
		"min_buy_in":  true,
		"max_buy_in":  true,
		"tags":        true,
	}
	
	for key := range filters {
		if !allowedFilters[key] {
			return fmt.Errorf("invalid filter key: %s", key)
		}
	}
	
	return nil
}