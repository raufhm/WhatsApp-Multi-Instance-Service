package whatsapp

import (
	"testing"
)

func TestSanitizePhoneNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean international format with plus",
			input:    "+6282141428746",
			expected: "6282141428746",
		},
		{
			name:     "with spaces",
			input:    "+62 82141428746",
			expected: "6282141428746",
		},
		{
			name:     "multiple spaces",
			input:    "+62 821 4142 8746",
			expected: "6282141428746",
		},
		{
			name:     "with dashes and spaces",
			input:    "+62-821-4142-8746",
			expected: "6282141428746",
		},
		{
			name:     "with parentheses and spaces",
			input:    "+1 (555) 123-4567",
			expected: "15551234567",
		},
		{
			name:     "with jid suffix",
			input:    "6282141428746@s.whatsapp.net",
			expected: "6282141428746",
		},
		{
			name:     "already digits only",
			input:    "6282141428746",
			expected: "6282141428746",
		},
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizePhoneNumber(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizePhoneNumber(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}
