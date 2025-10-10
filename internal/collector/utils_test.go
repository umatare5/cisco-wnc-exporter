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

// TestMapRadioSlotToBand tests the MapRadioSlotToBand function.
func TestMapRadioSlotToBand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "Slot 0 (2.4GHz)",
			input:    0,
			expected: "2.4",
		},
		{
			name:     "Slot 1 (5GHz)",
			input:    1,
			expected: "5",
		},
		{
			name:     "Slot 2 (6GHz)",
			input:    2,
			expected: "6",
		},
		{
			name:     "Unknown slot 3",
			input:    3,
			expected: "unknown",
		},
		{
			name:     "Unknown slot 10",
			input:    10,
			expected: "unknown",
		},
		{
			name:     "Negative slot -1",
			input:    -1,
			expected: "unknown",
		},
		{
			name:     "Negative slot -10",
			input:    -10,
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := MapRadioSlotToBand(tt.input)
			if result != tt.expected {
				t.Errorf(
					"MapRadioSlotToBand(%d) = %q; expected %q",
					tt.input,
					result,
					tt.expected,
				)
			}
		})
	}
}

// TestMapRadioTypeToSlot tests the MapRadioTypeToSlot function.
func TestMapRadioTypeToSlot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		// 2.4GHz radio types
		{
			name:     "dot11bg",
			input:    "dot11bg",
			expected: 0,
		},
		{
			name:     "24GHz ax protocol",
			input:    "client-dot11ax-24ghz-prot",
			expected: 0,
		},
		{
			name:     "24GHz n protocol",
			input:    "client-dot11n-24-ghz-prot",
			expected: 0,
		},
		{
			name:     "24GHz bg protocol",
			input:    "client-dot11bg-24-ghz-prot",
			expected: 0,
		},

		// 5GHz radio types
		{
			name:     "dot11a",
			input:    "dot11a",
			expected: 1,
		},
		{
			name:     "5GHz ax protocol",
			input:    "client-dot11ax-5ghz-prot",
			expected: 1,
		},
		{
			name:     "5GHz ac protocol",
			input:    "client-dot11ac-5-ghz-prot",
			expected: 1,
		},
		{
			name:     "5GHz n protocol",
			input:    "client-dot11n-5-ghz-prot",
			expected: 1,
		},
		{
			name:     "5GHz a protocol",
			input:    "client-dot11a-5-ghz-prot",
			expected: 1,
		},

		// 6GHz radio types
		{
			name:     "6GHz ax protocol",
			input:    "client-dot11ax-6ghz-prot",
			expected: 2,
		},

		// Unknown/Edge cases
		{
			name:     "Unknown type",
			input:    "unknown",
			expected: 0,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "Invalid format",
			input:    "invalid-radio-type",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := MapRadioTypeToSlot(tt.input)
			if result != tt.expected {
				t.Errorf(
					"MapRadioTypeToSlot(%q) = %d; expected %d",
					tt.input,
					result,
					tt.expected,
				)
			}
		})
	}
}

// TestDetermineBandFromRadioInfo tests the DetermineBandFromRadioInfo function.
func TestDetermineBandFromRadioInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		radioSlotID int
		radioType   string
		expected    string
	}{
		// Valid slot IDs
		{
			name:        "Slot 0 (2.4GHz)",
			radioSlotID: 0,
			radioType:   "",
			expected:    "2.4",
		},
		{
			name:        "Slot 1 (5GHz)",
			radioSlotID: 1,
			radioType:   "",
			expected:    "5",
		},
		{
			name:        "Slot 2 (6GHz)",
			radioSlotID: 2,
			radioType:   "",
			expected:    "6",
		},

		// Unknown slot with valid radio type (fallback to type)
		{
			name:        "Unknown slot with dot11a type",
			radioSlotID: -1,
			radioType:   "dot11a",
			expected:    "5",
		},
		{
			name:        "Unknown slot with 24GHz type",
			radioSlotID: 99,
			radioType:   "client-dot11ax-24ghz-prot",
			expected:    "2.4",
		},
		{
			name:        "Unknown slot with 6GHz type",
			radioSlotID: -1,
			radioType:   "client-dot11ax-6ghz-prot",
			expected:    "6",
		},

		// Valid slot with type override (type takes precedence when provided)
		{
			name:        "Slot 0 with conflicting 5GHz type",
			radioSlotID: 0,
			radioType:   "dot11a",
			expected:    "5",
		},
		{
			name:        "Slot 1 with conflicting 2.4GHz type",
			radioSlotID: 1,
			radioType:   "dot11bg",
			expected:    "2.4",
		},

		// Unknown slot and unknown type
		{
			name:        "Unknown slot and type",
			radioSlotID: -1,
			radioType:   "unknown",
			expected:    "2.4",
		},
		{
			name:        "Unknown slot with empty type",
			radioSlotID: 99,
			radioType:   "",
			expected:    "2.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := DetermineBandFromRadioInfo(tt.radioSlotID, tt.radioType)
			if result != tt.expected {
				t.Errorf(
					"DetermineBandFromRadioInfo(%d, %q) = %q; expected %q",
					tt.radioSlotID,
					tt.radioType,
					result,
					tt.expected,
				)
			}
		})
	}
}

// TestMapWirelessProtocol tests the MapWirelessProtocol function.
func TestMapWirelessProtocol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		phyType     string
		radioType   string
		is11GClient bool
		expected    WirelessProtocol
	}{
		// 802.11n
		{
			name:        "802.11n by phyType",
			phyType:     "dot11n",
			radioType:   "",
			is11GClient: false,
			expected:    ProtocolN,
		},
		{
			name:        "802.11n with substring",
			phyType:     "client-dot11n-24-ghz-prot",
			radioType:   "",
			is11GClient: false,
			expected:    ProtocolN,
		},

		// 802.11ac
		{
			name:        "802.11ac by phyType",
			phyType:     "dot11ac",
			radioType:   "",
			is11GClient: false,
			expected:    ProtocolAC,
		},
		{
			name:        "802.11ac with substring",
			phyType:     "client-dot11ac-5-ghz-prot",
			radioType:   "",
			is11GClient: false,
			expected:    ProtocolAC,
		},

		// 802.11ax
		{
			name:        "802.11ax by phyType",
			phyType:     "dot11ax",
			radioType:   "",
			is11GClient: false,
			expected:    ProtocolAX,
		},
		{
			name:        "802.11ax 24GHz",
			phyType:     "client-dot11ax-24ghz-prot",
			radioType:   "",
			is11GClient: false,
			expected:    ProtocolAX,
		},
		{
			name:        "802.11ax 5GHz",
			phyType:     "client-dot11ax-5ghz-prot",
			radioType:   "",
			is11GClient: false,
			expected:    ProtocolAX,
		},
		{
			name:        "802.11ax 6GHz",
			phyType:     "client-dot11ax-6ghz-prot",
			radioType:   "",
			is11GClient: false,
			expected:    ProtocolAX,
		},

		// 802.11be
		{
			name:        "802.11be by phyType",
			phyType:     "dot11be",
			radioType:   "",
			is11GClient: false,
			expected:    ProtocolBE,
		},
		{
			name:        "802.11be EHT",
			phyType:     "eht",
			radioType:   "",
			is11GClient: false,
			expected:    ProtocolBE,
		},

		// 802.11bg with flag differentiation
		{
			name:        "802.11g client (flag set)",
			phyType:     "dot11bg",
			radioType:   "",
			is11GClient: true,
			expected:    Protocol11G,
		},
		{
			name:        "802.11b client (flag not set)",
			phyType:     "dot11bg",
			radioType:   "",
			is11GClient: false,
			expected:    Protocol11B,
		},

		// 802.11a
		{
			name:        "802.11a by phyType",
			phyType:     "dot11a",
			radioType:   "",
			is11GClient: false,
			expected:    Protocol11A,
		},
		{
			name:        "802.11a by radioType",
			phyType:     "",
			radioType:   "dot11-radio-type-a",
			is11GClient: false,
			expected:    Protocol11A,
		},
		{
			name:        "802.11a protocol substring",
			phyType:     "client-dot11a-5-ghz-prot",
			radioType:   "",
			is11GClient: false,
			expected:    Protocol11A,
		},

		// 802.11g
		{
			name:        "802.11g by phyType",
			phyType:     "dot11g",
			radioType:   "",
			is11GClient: false,
			expected:    Protocol11G,
		},
		{
			name:        "802.11g by flag",
			phyType:     "",
			radioType:   "",
			is11GClient: true,
			expected:    Protocol11G,
		},

		// Unknown
		{
			name:        "Unknown protocol",
			phyType:     "unknown",
			radioType:   "",
			is11GClient: false,
			expected:    ProtocolUnknown,
		},
		{
			name:        "Empty parameters",
			phyType:     "",
			radioType:   "",
			is11GClient: false,
			expected:    ProtocolUnknown,
		},

		// Edge cases - precedence tests
		{
			name:        "Multiple conditions - n takes precedence",
			phyType:     "dot11n",
			radioType:   "dot11-radio-type-a",
			is11GClient: true,
			expected:    ProtocolN,
		},
		{
			name:        "AC over ax substring match",
			phyType:     "dot11ac",
			radioType:   "client-dot11ax-5ghz-prot",
			is11GClient: false,
			expected:    ProtocolAC,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := MapWirelessProtocol(tt.phyType, tt.radioType, tt.is11GClient)
			if result != tt.expected {
				t.Errorf(
					"MapWirelessProtocol(%q, %q, %v) = %v; expected %v",
					tt.phyType,
					tt.radioType,
					tt.is11GClient,
					result,
					tt.expected,
				)
			}
		})
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
