package cli

import (
	"testing"
)

// TestGetVersion verifies the getVersion function returns the version variable value.
func TestGetVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "Default version",
			expected: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := getVersion()
			if got != tt.expected {
				t.Errorf("getVersion() = %q, want %q", got, tt.expected)
			}
		})
	}
}
