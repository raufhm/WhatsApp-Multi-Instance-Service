package domain

import "testing"

func TestNormalizeRole(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"admin", RoleAdmin},
		{"ADMIN", RoleAdmin},
		{" Admin ", RoleAdmin},
		{"operator", RoleOperator},
		{"OPERATOR", RoleOperator},
		{"  Operator  ", RoleOperator},
		{"viewer", RoleViewer},
		{"VIEWER", RoleViewer},
		{" Viewer ", RoleViewer},
		{"", RoleOperator},
		{"invalid", RoleOperator},
		{"UNKNOWN", RoleOperator},
	}

	for _, tt := range tests {
		got := NormalizeRole(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeRole(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
