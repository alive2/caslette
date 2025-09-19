package handlers

import (
	"testing"
)

func TestBasicSecurityValidator(t *testing.T) {
	validator := NewSecurityValidator()

	// Test SQL injection detection
	_, err := validator.ValidateAndSanitizeString("'; DROP TABLE users; --", "username", 50)
	if err == nil {
		t.Error("Should have detected SQL injection")
	}

	// Test valid input
	result, err := validator.ValidateAndSanitizeString("john_doe", "username", 50)
	if err != nil {
		t.Errorf("Valid input failed: %v", err)
	}
	if result != "john_doe" {
		t.Errorf("Expected 'john_doe', got '%s'", result)
	}

	// Test positive int validation
	num, err := validator.ValidatePositiveInt("123", "id")
	if err != nil {
		t.Errorf("Valid positive int failed: %v", err)
	}
	if num != 123 {
		t.Errorf("Expected 123, got %d", num)
	}

	// Test invalid positive int
	_, err = validator.ValidatePositiveInt("-1", "id")
	if err == nil {
		t.Error("Should have rejected negative number")
	}
}
