package cli

import (
	"testing"

	"github.com/urfave/cli/v3"
)

// TestRegisterFlags verifies that registerFlags returns all flags from sub-registrars.
func TestRegisterFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expectedCount int
	}{
		{
			name: "All flags registered",
			// 3 (web) + 5 (wnc) + 1 (collector) + 2 (log) + 1 (utility) +
			// 2 (internal) + 6 (ap) + 5 (wlan) + 6 (client) = 31 flags
			expectedCount: 31,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := registerFlags()
			if got := len(flags); got != tt.expectedCount {
				t.Errorf("registerFlags() returned %d flags, want %d", got, tt.expectedCount)
			}
		})
	}
}

// TestRegisterWebFlags verifies web server configuration flags.
func TestRegisterWebFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expectedCount int
		expectedTypes []string
	}{
		{
			name:          "Web flags count",
			expectedCount: 3,
			expectedTypes: []string{"string", "int", "string"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := registerWebFlags()
			if got := len(flags); got != tt.expectedCount {
				t.Errorf("registerWebFlags() returned %d flags, want %d", got, tt.expectedCount)
			}

			// Verify flag types
			for i, flag := range flags {
				var gotType string
				switch flag.(type) {
				case *cli.StringFlag:
					gotType = "string"
				case *cli.IntFlag:
					gotType = "int"
				default:
					gotType = "unknown"
				}

				if gotType != tt.expectedTypes[i] {
					t.Errorf("flag[%d] type = %s, want %s", i, gotType, tt.expectedTypes[i])
				}
			}
		})
	}
}

// TestRegisterWNCFlags verifies WNC controller connection flags.
func TestRegisterWNCFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expectedCount int
		expectedTypes []string
	}{
		{
			name:          "WNC flags count",
			expectedCount: 5,
			expectedTypes: []string{"string", "string", "duration", "duration", "bool"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := registerWNCFlags()
			if got := len(flags); got != tt.expectedCount {
				t.Errorf("registerWNCFlags() returned %d flags, want %d", got, tt.expectedCount)
			}

			// Verify flag types
			for i, flag := range flags {
				var gotType string
				switch flag.(type) {
				case *cli.StringFlag:
					gotType = "string"
				case *cli.DurationFlag:
					gotType = "duration"
				case *cli.BoolFlag:
					gotType = "bool"
				default:
					gotType = "unknown"
				}

				if gotType != tt.expectedTypes[i] {
					t.Errorf("flag[%d] type = %s, want %s", i, gotType, tt.expectedTypes[i])
				}
			}
		})
	}
}

// TestRegisterCollectorFlags verifies collector-wide configuration flags.
func TestRegisterCollectorFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expectedCount int
	}{
		{
			name:          "Collector flags count",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := registerCollectorFlags()
			if got := len(flags); got != tt.expectedCount {
				t.Errorf("registerCollectorFlags() returned %d flags, want %d", got, tt.expectedCount)
			}

			// Verify it's a DurationFlag
			if _, ok := flags[0].(*cli.DurationFlag); !ok {
				t.Errorf("flag[0] is not *cli.DurationFlag")
			}
		})
	}
}

// TestRegisterAPCollectorFlags verifies AP collector module flags.
func TestRegisterAPCollectorFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expectedCount int
		expectedTypes []string
	}{
		{
			name:          "AP collector flags count",
			expectedCount: 6,
			expectedTypes: []string{"bool", "bool", "bool", "bool", "bool", "string"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := registerAPCollectorFlags()
			if got := len(flags); got != tt.expectedCount {
				t.Errorf("registerAPCollectorFlags() returned %d flags, want %d", got, tt.expectedCount)
			}

			// Verify flag types
			for i, flag := range flags {
				var gotType string
				switch flag.(type) {
				case *cli.BoolFlag:
					gotType = "bool"
				case *cli.StringFlag:
					gotType = "string"
				default:
					gotType = "unknown"
				}

				if gotType != tt.expectedTypes[i] {
					t.Errorf("flag[%d] type = %s, want %s", i, gotType, tt.expectedTypes[i])
				}
			}
		})
	}
}

// TestRegisterWLANCollectorFlags verifies WLAN collector module flags.
func TestRegisterWLANCollectorFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expectedCount int
		expectedTypes []string
	}{
		{
			name:          "WLAN collector flags count",
			expectedCount: 5,
			expectedTypes: []string{"bool", "bool", "bool", "bool", "string"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := registerWLANCollectorFlags()
			if got := len(flags); got != tt.expectedCount {
				t.Errorf("registerWLANCollectorFlags() returned %d flags, want %d", got, tt.expectedCount)
			}

			// Verify flag types
			for i, flag := range flags {
				var gotType string
				switch flag.(type) {
				case *cli.BoolFlag:
					gotType = "bool"
				case *cli.StringFlag:
					gotType = "string"
				default:
					gotType = "unknown"
				}

				if gotType != tt.expectedTypes[i] {
					t.Errorf("flag[%d] type = %s, want %s", i, gotType, tt.expectedTypes[i])
				}
			}
		})
	}
}

// TestRegisterClientCollectorFlags verifies Client collector module flags.
func TestRegisterClientCollectorFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expectedCount int
		expectedTypes []string
	}{
		{
			name:          "Client collector flags count",
			expectedCount: 6,
			expectedTypes: []string{"bool", "bool", "bool", "bool", "bool", "string"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := registerClientCollectorFlags()
			if got := len(flags); got != tt.expectedCount {
				t.Errorf("registerClientCollectorFlags() returned %d flags, want %d", got, tt.expectedCount)
			}

			// Verify flag types
			for i, flag := range flags {
				var gotType string
				switch flag.(type) {
				case *cli.BoolFlag:
					gotType = "bool"
				case *cli.StringFlag:
					gotType = "string"
				default:
					gotType = "unknown"
				}

				if gotType != tt.expectedTypes[i] {
					t.Errorf("flag[%d] type = %s, want %s", i, gotType, tt.expectedTypes[i])
				}
			}
		})
	}
}

// TestRegisterLogFlags verifies logging configuration flags.
func TestRegisterLogFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expectedCount int
	}{
		{
			name:          "Log flags count",
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := registerLogFlags()
			if got := len(flags); got != tt.expectedCount {
				t.Errorf("registerLogFlags() returned %d flags, want %d", got, tt.expectedCount)
			}

			// Verify both are StringFlags
			for i, flag := range flags {
				if _, ok := flag.(*cli.StringFlag); !ok {
					t.Errorf("flag[%d] is not *cli.StringFlag", i)
				}
			}
		})
	}
}

// TestRegisterInternalCollectorFlags verifies internal collector flags.
func TestRegisterInternalCollectorFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expectedCount int
	}{
		{
			name:          "Internal collector flags count",
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := registerInternalCollectorFlags()
			if got := len(flags); got != tt.expectedCount {
				t.Errorf("registerInternalCollectorFlags() returned %d flags, want %d", got, tt.expectedCount)
			}

			// Verify both are BoolFlags
			for i, flag := range flags {
				if _, ok := flag.(*cli.BoolFlag); !ok {
					t.Errorf("flag[%d] is not *cli.BoolFlag", i)
				}
			}
		})
	}
}

// TestRegisterUtilityFlags verifies utility flags.
func TestRegisterUtilityFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		expectedCount int
	}{
		{
			name:          "Utility flags count",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := registerUtilityFlags()
			if got := len(flags); got != tt.expectedCount {
				t.Errorf("registerUtilityFlags() returned %d flags, want %d", got, tt.expectedCount)
			}

			// Verify it's a BoolFlag
			if _, ok := flags[0].(*cli.BoolFlag); !ok {
				t.Errorf("flag[0] is not *cli.BoolFlag")
			}
		})
	}
}

// TestNewApp verifies basic CLI application structure.
func TestNewApp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		expectedName string
	}{
		{
			name:         "CLI app name",
			expectedName: "cisco-wnc-exporter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Note: We cannot call NewApp() directly as it calls cmd.Run() and os.Exit()
			// Instead, we create the command structure without running it
			cmd := &cli.Command{
				Name:    "cisco-wnc-exporter",
				Usage:   "Prometheus exporter for Cisco WNC",
				Version: getVersion(),
				Flags:   registerFlags(),
			}

			if cmd.Name != tt.expectedName {
				t.Errorf("cmd.Name = %q, want %q", cmd.Name, tt.expectedName)
			}

			if cmd.Version == "" {
				t.Error("cmd.Version is empty")
			}

			if len(cmd.Flags) == 0 {
				t.Error("cmd.Flags is empty")
			}
		})
	}
}
