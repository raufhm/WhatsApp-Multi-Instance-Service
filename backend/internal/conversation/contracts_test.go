package conversation

import "testing"

func TestNormalizeAddress(t *testing.T) {
	tests := map[string]string{
		"15551234567":                  "15551234567@s.whatsapp.net",
		"15551234567@s.whatsapp.net":   "15551234567@s.whatsapp.net",
		"15551234567:2@s.whatsapp.net": "15551234567@s.whatsapp.net",
		"  GROUP-1@g.us ":              "group-1@g.us",
	}
	for input, expected := range tests {
		actual, err := NormalizeAddress(input)
		if err != nil || actual != expected {
			t.Errorf("NormalizeAddress(%q) = %q, %v; want %q", input, actual, err, expected)
		}
	}
}

func TestNormalizeAddressRejectsEmptyAndMalformed(t *testing.T) {
	for _, input := range []string{"", "@", "@user", "user@"} {
		if _, err := NormalizeAddress(input); err == nil {
			t.Errorf("NormalizeAddress(%q) did not reject malformed address", input)
		}
	}
}
