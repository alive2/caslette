package main

import (
	"fmt"
	"regexp"
)

func main() {
	// Test the regex pattern
	tableNameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-_\.!?']{3,100}$`)

	testNames := []string{
		"Texas Hold'em - High Stakes",
		"tablename",
	}

	for _, name := range testNames {
		matches := tableNameRegex.MatchString(name)
		fmt.Printf("Name: '%s' matches: %v\n", name, matches)
	}
}
