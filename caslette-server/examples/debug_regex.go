package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	// Test the regex pattern
	tableNameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-_\.!?']{3,100}$`)

	testNames := []string{
		"Texas Hold'em Table",
		"tablename",
		"Test Table",
	}

	for _, name := range testNames {
		matches := tableNameRegex.MatchString(name)
		fmt.Printf("Name: '%s' matches: %v\n", name, matches)

		// Check for dangerous patterns
		input := strings.ToLower(name)
		dangerousPatterns := []string{
			"; drop", "; delete", "; insert", "; update", "; create", "; alter",
			"' or '1'='1", "\" or \"1\"=\"1", "'; --", "\"; --",
			"/*", "*/", "xp_", "sp_",
			"<script", "javascript:", "vbscript:", "onload=", "onerror=", "onclick=",
			"\x00", // null byte
		}

		hasPattern := false
		for _, pattern := range dangerousPatterns {
			if strings.Contains(input, pattern) {
				fmt.Printf("  Contains dangerous pattern: %s\n", pattern)
				hasPattern = true
			}
		}
		if !hasPattern {
			fmt.Printf("  No dangerous patterns found\n")
		}
	}
}
