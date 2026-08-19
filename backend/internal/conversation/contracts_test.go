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

func TestNormalizeAddressWithServerGroupVsPersonal(t *testing.T) {
	// A bare numeric ID must normalize to different addresses depending on the
	// server parameter so that group and personal contacts remain distinct.
	personal, err := NormalizeAddressWithServer("120363", "s.whatsapp.net")
	if err != nil {
		t.Fatalf("personal: %v", err)
	}
	group, err := NormalizeAddressWithServer("120363", "g.us")
	if err != nil {
		t.Fatalf("group: %v", err)
	}
	if personal == group {
		t.Errorf("personal and group addresses must not collide: both %q", personal)
	}
	if personal != "120363@s.whatsapp.net" {
		t.Errorf("personal = %q, want 120363@s.whatsapp.net", personal)
	}
	if group != "120363@g.us" {
		t.Errorf("group = %q, want 120363@g.us", group)
	}
}

func TestNormalizeAddressRejectsEmptyAndMalformed(t *testing.T) {
	for _, input := range []string{"", "@", "@user", "user@"} {
		if _, err := NormalizeAddress(input); err == nil {
			t.Errorf("NormalizeAddress(%q) did not reject malformed address", input)
		}
	}
}
