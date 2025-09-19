package game

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// TableValidator handles input validation for table operations
type TableValidator struct{}

// Constants for validation limits
const (
	MaxTableNameLength   = 100
	MaxDescriptionLength = 500
	MaxPasswordLength    = 50
	MaxUsernameLength    = 30
	MaxTagsCount         = 10
	MaxTagLength         = 20

	MinTableNameLength = 3
	MinPasswordLength  = 4
	MinBuyIn           = 1
	MaxBuyIn           = 1000000
	MinBlind           = 1
	MaxBlind           = 100000
	MaxTimeLimit       = 300 // 5 minutes max per turn
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

	// Trim spaces but don't HTML escape yet (we'll do that when storing)
	originalName := strings.TrimSpace(name)

	if len(originalName) < MinTableNameLength {
		return fmt.Errorf("table name too short (min %d characters)", MinTableNameLength)
	}

	if len(originalName) > MaxTableNameLength {
		return fmt.Errorf("table name too long (max %d characters)", MaxTableNameLength)
	}

	if !utf8.ValidString(originalName) {
		return fmt.Errorf("table name contains invalid UTF-8 characters")
	}

	if !tableNameRegex.MatchString(originalName) {
		return fmt.Errorf("table name contains invalid characters")
	}

	// Check for common injection patterns on the original name
	if v.containsSQLInjectionPatterns(originalName) {
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

	// Table IDs should be reasonable alphanumeric strings (8-32 chars)
	if len(tableID) < 8 || len(tableID) > 32 {
		return fmt.Errorf("table ID length invalid (must be 8-32 characters)")
	}

	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-_]+$`, tableID); !matched {
		return fmt.Errorf("table ID contains invalid characters")
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

	// Store original for pattern checking
	originalDescription := description

	// Check for injection patterns BEFORE sanitizing
	if v.containsSQLInjectionPatterns(originalDescription) {
		return fmt.Errorf("description contains invalid patterns")
	}

	// Sanitize HTML
	description = v.SanitizeInput(description)

	if len(description) > MaxDescriptionLength {
		return fmt.Errorf("description too long (max %d characters)", MaxDescriptionLength)
	}

	if !utf8.ValidString(description) {
		return fmt.Errorf("description contains invalid UTF-8 characters")
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
	originalInput := input
	input = strings.ToLower(input)

	// Remove URL encoding to catch encoded attacks
	decodedInput := input
	if decoded, err := url.QueryUnescape(input); err == nil {
		decodedInput = strings.ToLower(decoded)
	}

	// Check both original and decoded input
	inputs := []string{input, decodedInput}

	for _, checkInput := range inputs {
		// SQL injection keywords and patterns
		sqlKeywords := []string{
			"select", "union", "insert", "update", "delete", "drop", "create", "alter",
			"exec", "execute", "sp_", "xp_", "waitfor", "delay", "sleep",
			"information_schema", "sysobjects", "syscolumns", "msysaccessobjects",
			"pg_", "mysql", "oracle", "mssql", "sqlite", "substr", "substring",
			"concat", "char", "ascii", "benchmark", "extractvalue", "updatexml",
			"load_file", "into outfile", "into dumpfile", "sqlmap", "havij",
			// NoSQL injection patterns
			"$where", "$gt", "$ne", "$regex", "$exists", "$in", "$nin", "$or", "$and",
			"function()", "emit(", "map(", "reduce(", "finalize(",
		}

		// Dangerous operator patterns
		dangerousPatterns := []string{
			"; drop", "; delete", "; insert", "; update", "; create", "; alter",
			"' or ", "\" or ", "' and ", "\" and ",
			"' union", "\" union", "'union", "\"union",
			"1=1", "1=2", "'='", "\"=\"", "'1'='1", "\"1\"=\"1",
			"'; --", "\"; --", "'/*", "\"/*", "*/", "--",
			"'||'", "\"||\"", "'+'", "\"+\"",
			"0x", "char(", "ascii(", "substr(", "substring(",
			"concat(", "group_concat(", "hex(", "unhex(",
			"<script", "javascript:", "vbscript:", "onload=", "onerror=", "onclick=",
			"onmouseover=", "onfocus=", "onblur=", "onchange=", "onsubmit=",
			"ng-click=", "ng-app", "$event.view", "{{", "}}", "<%", "%>",
			"eval(", "expression(", "import(", "require(",
			"\x00", // null byte
			"${", "#{", "<%", "%>", "{{", "}}",
			"../", "..\\", "%2e%2e", "%252e%252e", "%252f", "%u002f",
			"..;/", "../;", "\\u002e", "\\u002f",
			"file://", "http://", "https://", "ftp://",
			"\\\\", "\\x", "\\u00", "\\u20", "%00", "%0a", "%0d",
			"'x'='x", "'x'='y", "\"x\"=\"x", "\"x\"=\"y", // Boolean injection patterns
			// Command injection patterns
			"; ls", "; dir", "; cat", "; type", "; echo", "; rm", "; del",
			"; wget", "; curl", "; nc", "; netcat", "; python", "; perl",
			"| ls", "| dir", "| cat", "| type", "| echo", "| nc", "| curl",
			"& whoami", "& echo", "& dir", "& ls", "&& rm", "&& del",
			"|| echo", "`ls", "`cat", "`whoami", "`echo", "`dir",
			"$(ls", "$(cat", "$(whoami", "$(echo", "$(rm", "$(python",
			"; python -c", "; perl -e", "wget ", "curl ",
			// LDAP injection patterns
			"*)((", "*)(", ")(&", "))(|", "objectclass=", "\\2a", "\\29", "\\28",
			// Format string patterns
			"%n", "%$", "%.*", "%.1000", "%1000", "%7$", "%8$", "%9$",
		}

		// Check for SQL keywords in suspicious contexts
		for _, keyword := range sqlKeywords {
			if strings.Contains(checkInput, keyword) {
				// Allow common words in normal contexts
				if keyword == "select" || keyword == "insert" || keyword == "update" || keyword == "delete" {
					// Only flag if followed by SQL-like patterns
					keywordIndex := strings.Index(checkInput, keyword)
					if keywordIndex != -1 && keywordIndex < len(checkInput)-len(keyword) {
						remaining := checkInput[keywordIndex+len(keyword):]
						if strings.Contains(remaining, " from") || strings.Contains(remaining, " into") ||
							strings.Contains(remaining, " where") || strings.Contains(remaining, " set") ||
							strings.Contains(remaining, "*") || strings.Contains(remaining, "password") ||
							strings.Contains(remaining, "user") {
							return true
						}
					}
				} else {
					return true
				}
			}
		}

		// Check for dangerous patterns
		for _, pattern := range dangerousPatterns {
			if strings.Contains(checkInput, pattern) {
				return true
			}
		}

		// Check for multiple single quotes (often used in SQL injection)
		if strings.Count(checkInput, "'") > 2 || strings.Count(checkInput, "\"") > 2 {
			return true
		}

		// Check for suspicious character combinations
		suspiciousRegexes := []*regexp.Regexp{
			regexp.MustCompile(`['"]\s*(\bor\b|\band\b|\bunion\b)\s*['"]?`),
			regexp.MustCompile(`['"]\s*;\s*\w+`),
			regexp.MustCompile(`\b(union|select|insert|update|delete|drop|create|alter)\s+.*\s+(from|into|where|set|table|database)`),
			regexp.MustCompile(`\b\d+\s*=\s*\d+\b`),
			regexp.MustCompile(`['"]\s*\+\s*['"]`),
			regexp.MustCompile(`['"]\s*\|\|\s*['"]`),
			regexp.MustCompile(`\s*--\s*$`),
			regexp.MustCompile(`/\*.*\*/`),
			regexp.MustCompile(`/\*\s*$`), // Comment start at end
			regexp.MustCompile(`\${.*}`),
			regexp.MustCompile(`#{.*}`),
			regexp.MustCompile(`<%.*%>`),
			regexp.MustCompile(`{{.*}}`),
			regexp.MustCompile(`'\s*or\s*'.*?'\s*=\s*'.*?'`), // 'x'='x pattern with any values
			regexp.MustCompile(`"\s*or\s*".*?"\s*=\s*".*?"`), // "x"="x pattern with any values
		}

		for _, regex := range suspiciousRegexes {
			if regex.MatchString(checkInput) {
				return true
			}
		}
	}

	// Check for format string attacks
	if strings.Count(originalInput, "%") > 5 {
		return true
	}

	// Check for excessive repetition (potential buffer overflow)
	for _, char := range []string{"A", "B", "C", "X", "1", "0"} {
		if strings.Count(originalInput, char) > 100 {
			return true
		}
	}

	return false
} // ValidateFilterRequest validates table filtering requests
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
