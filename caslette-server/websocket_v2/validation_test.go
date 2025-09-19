package websocket_v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInputValidationDirect(t *testing.T) {
	// Test simple inputs that should work
	inputs := []struct {
		input      string
		inputType  string
		shouldPass bool
		name       string
	}{
		{"test", "room", true, "simple room name"},
		{"room123", "room", true, "alphanumeric room"},
		{"test-room", "room", true, "hyphenated room"},
		{"test_room", "room", true, "underscore room"},
		{"user123", "username", true, "simple username"},
		{"'; DROP TABLE users; --", "room", false, "SQL injection"},
		{"<script>", "room", false, "XSS script tag"},
		{"", "room", false, "empty input"},
		{"verylongroomnamethatexceedsthefiftycharlimit123456789", "room", false, "too long"},
	}

	for _, test := range inputs {
		t.Run(test.name, func(t *testing.T) {
			result, err := validateInput(test.input, test.inputType)
			if test.shouldPass {
				assert.NoError(t, err, "Should pass validation: %s", test.input)
				assert.NotEmpty(t, result, "Should return non-empty result")
			} else {
				assert.Error(t, err, "Should fail validation: %s", test.input)
			}
		})
	}
}
