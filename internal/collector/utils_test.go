package collector

import (
	"testing"
	"time"
)

// TestBoolToFloat64 tests the boolToFloat64 function.
func TestBoolToFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    bool
		expected float64
	}{
		{
			name:     "True to 1.0",
			input:    true,
			expected: 1.0,
		},
		{
			name:     "False to 0.0",
			input:    false,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := boolToFloat64(tt.input)
			if result != tt.expected {
				t.Errorf(
					"boolToFloat64(%v) = %v; expected %v",
					tt.input,
					result,
					tt.expected,
				)
			}
		})
	}
}

// TestMapClientState tests the MapClientState function.
func TestMapClientState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Run state",
			input:    "client-status-run",
			expected: 2,
		},
		{
			name:     "Authenticated state",
			input:    "client-status-authenticated",
			expected: 1,
		},
		{
			name:     "Unknown state",
			input:    "unknown",
			expected: 0,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "Invalid state",
			input:    "client-status-invalid",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := MapClientState(tt.input)
			if result != tt.expected {
				t.Errorf(
					"MapClientState(%q) = %d; expected %d",
					tt.input,
					result,
					tt.expected,
				)
			}
		})
	}
}

// TestStringToUint64 tests the stringToUint64 function.
func TestStringToUint64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected uint64
	}{
		{
			name:     "Valid number",
			input:    "12345",
			expected: 12345,
		},
		{
			name:     "Zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "Large number",
			input:    "9876543210",
			expected: 9876543210,
		},
		{
			name:     "Max uint64",
			input:    "18446744073709551615",
			expected: 18446744073709551615,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "Invalid string",
			input:    "abc",
			expected: 0,
		},
		{
			name:     "Negative number",
			input:    "-123",
			expected: 0,
		},
		{
			name:     "Float number",
			input:    "123.45",
			expected: 0,
		},
		{
			name:     "Number with spaces",
			input:    " 123 ",
			expected: 0,
		},
		{
			name:     "Hexadecimal",
			input:    "0x1A",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := stringToUint64(tt.input)
			if result != tt.expected {
				t.Errorf(
					"stringToUint64(%q) = %d; expected %d",
					tt.input,
					result,
					tt.expected,
				)
			}
		})
	}
}

// TestMapWirelessProtocol covers every value of ms-phy-radio-type, the typedef of
// the phyType argument, so that the table can be diffed against the model.
//
// confirmed records whether the spelling was seen coming back from a controller. The
// rest are taken from the model alone, and the model is not always right about them,
// so their mapping is a best effort: a spelling that turns out to be wrong falls
// through to the unknown protocol rather than producing a wrong one. A value that has
// been confirmed must always map to something, which the subtest below enforces.
func TestMapWirelessProtocol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		phyType   string
		want      WirelessProtocol
		confirmed bool
	}{
		{"client-unknown-prot", ProtocolUnknown, false},
		{"client-dot11b", Protocol11B, false},
		{"client-dot11g", Protocol11G, false},
		{"client-dot11a", Protocol11A, false},
		{"client-dot11n-24-ghz-prot", ProtocolN, true},
		{"client-dot11n-5-ghz-prot", ProtocolN, true},
		{"client-dot11ac", ProtocolAC, true},
		{"client-phy-type-notappl", ProtocolUnknown, false},
		{"client-ethernet", ProtocolUnknown, false},
		{"client-dot11ax-5ghz-prot", ProtocolAX, true},
		{"client-dot11ax-24ghz-prot", ProtocolAX, true},
		{"client-802-3", ProtocolUnknown, false},
		{"client-dot11ax-6ghz-prot", ProtocolAX, true},
		{"client-dot11be-24ghz-prot", ProtocolBE, false},
		{"client-dot11be-5ghz-prot", ProtocolBE, false},
		{"client-dot11be-6ghz-prot", ProtocolBE, false},
	}

	if len(tests) != 16 {
		t.Fatalf("the table covers %d values, want every value of ms-phy-radio-type", len(tests))
	}

	for _, tt := range tests {
		t.Run(tt.phyType, func(t *testing.T) {
			t.Parallel()

			got := MapWirelessProtocol(tt.phyType, "")
			if got != tt.want {
				t.Errorf("MapWirelessProtocol(%q, \"\") = %v, want %v", tt.phyType, got, tt.want)
			}
			if tt.confirmed && got == ProtocolUnknown {
				t.Errorf("MapWirelessProtocol(%q, \"\") = %v, but a controller does send this value",
					tt.phyType, got)
			}
		})
	}
}

// TestMapWirelessProtocol_RadioTypeFallback covers ms-radio-type, which names the
// band rather than the generation and is only consulted when the PHY type is absent.
func TestMapWirelessProtocol_RadioTypeFallback(t *testing.T) {
	t.Parallel()

	// Only dot11-radio-type-none was never seen coming back from a controller.
	tests := []struct {
		radioType string
		want      WirelessProtocol
	}{
		{"dot11-radio-type-none", ProtocolUnknown},
		{"dot11-radio-type-a", Protocol11A},
		// The band alone does not name a generation for these two.
		{"dot11-radio-type-bg", ProtocolUnknown},
		{"dot11-radio-type-6ghz", ProtocolUnknown},
		{"", ProtocolUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.radioType, func(t *testing.T) {
			t.Parallel()
			if got := MapWirelessProtocol("", tt.radioType); got != tt.want {
				t.Errorf("MapWirelessProtocol(\"\", %q) = %v, want %v", tt.radioType, got, tt.want)
			}
		})
	}
}

// TestMapWirelessProtocol_EveryValueIsReachable pins that no declared protocol is
// dead code. Protocol11B was unreachable because the mapping looked for a combined
// bg spelling that the typedef does not use.
func TestMapWirelessProtocol_EveryValueIsReachable(t *testing.T) {
	t.Parallel()

	phyTypes := []string{
		"client-unknown-prot", "client-dot11b", "client-dot11g", "client-dot11a",
		"client-dot11n-24-ghz-prot", "client-dot11n-5-ghz-prot", "client-dot11ac",
		"client-phy-type-notappl", "client-ethernet", "client-dot11ax-5ghz-prot",
		"client-dot11ax-24ghz-prot", "client-802-3", "client-dot11ax-6ghz-prot",
		"client-dot11be-24ghz-prot", "client-dot11be-5ghz-prot", "client-dot11be-6ghz-prot",
	}
	radioTypes := []string{
		"", "dot11-radio-type-none", "dot11-radio-type-a",
		"dot11-radio-type-bg", "dot11-radio-type-6ghz",
	}

	seen := make(map[WirelessProtocol]bool)
	for _, phyType := range append([]string{""}, phyTypes...) {
		for _, radioType := range radioTypes {
			seen[MapWirelessProtocol(phyType, radioType)] = true
		}
	}

	for protocol := ProtocolUnknown; protocol <= ProtocolBE; protocol++ {
		if !seen[protocol] {
			t.Errorf("protocol %v is never returned for any documented input", protocol)
		}
	}
}

// TestCalculateUptimeFromBootTime tests the CalculateUptimeFromBootTime function.
func TestCalculateUptimeFromBootTime(t *testing.T) {
	t.Parallel()

	// Generate a timestamp 1 hour ago dynamically
	oneHourAgo := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)

	tests := []struct {
		name        string
		bootTimeStr string
		expectError bool
		expectedMin int64 // For range checking
		expectedMax int64 // For range checking
		description string
	}{
		{
			name:        "Valid timestamp - 1 hour ago",
			bootTimeStr: oneHourAgo,
			expectError: false,
			expectedMin: 3500,
			expectedMax: 3700,
			description: "Should return approximately 3600 seconds (1 hour)",
		},
		{
			name:        "Invalid timestamp format",
			bootTimeStr: "2024-01-01 12:00:00",
			expectError: true,
			expectedMin: 0,
			expectedMax: 0,
			description: "Should return error for invalid format",
		},
		{
			name:        "Empty string",
			bootTimeStr: "",
			expectError: true,
			expectedMin: 0,
			expectedMax: 0,
			description: "Should return error for empty string",
		},
		{
			name:        "Invalid RFC3339 - missing timezone",
			bootTimeStr: "2024-01-01T12:00:00",
			expectError: true,
			expectedMin: 0,
			expectedMax: 0,
			description: "Should return error for missing timezone",
		},
		{
			name:        "Valid old timestamp",
			bootTimeStr: "2020-01-01T00:00:00Z",
			expectError: false,
			expectedMin: 157680000,
			expectedMax: 999999999,
			description: "Should return large positive value for old timestamp",
		},
		{
			name:        "Garbage input",
			bootTimeStr: "not-a-timestamp",
			expectError: true,
			expectedMin: 0,
			expectedMax: 0,
			description: "Should return error for garbage input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := CalculateUptimeFromBootTime(tt.bootTimeStr)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf(
					"CalculateUptimeFromBootTime(%q) expected error but got nil",
					tt.bootTimeStr,
				)
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf(
					"CalculateUptimeFromBootTime(%q) unexpected error: %v",
					tt.bootTimeStr,
					err,
				)
				return
			}

			// For valid cases, check range
			if !tt.expectError {
				if result < tt.expectedMin || result > tt.expectedMax {
					t.Errorf(
						"CalculateUptimeFromBootTime(%q) = %d; expected range [%d, %d] - %s",
						tt.bootTimeStr,
						result,
						tt.expectedMin,
						tt.expectedMax,
						tt.description,
					)
				}
			}
		})
	}
}
