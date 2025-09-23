package main

import (
	"os"
	"testing"
)

// TestMain_CanCall verifies that the main function can be called.
// This test uses dry-run mode to avoid actual server startup.
func TestMain_CanCall(t *testing.T) {
	t.Parallel()

	// Save original args and environment
	originalArgs := os.Args
	originalController := os.Getenv("WNC_HOST")
	originalToken := os.Getenv("WNC_TOKEN")

	defer func() {
		// Restore original state
		os.Args = originalArgs
		if originalController == "" {
			os.Unsetenv("WNC_HOST")
		} else {
			os.Setenv("WNC_HOST", originalController)
		}
		if originalToken == "" {
			os.Unsetenv("WNC_TOKEN")
		} else {
			os.Setenv("WNC_TOKEN", originalToken)
		}

		// Recover from potential panic or os.Exit
		if r := recover(); r != nil {
			t.Fatalf("main() panic: %v", r)
		}
	}()

	// Set up minimal required environment for dry-run
	os.Setenv("WNC_HOST", "test.example.com")
	os.Setenv("WNC_TOKEN", "dGVzdDp0ZXN0") // base64 encoded "test:test"
	os.Args = []string{"cisco-wnc-exporter", "--dry-run"}

	// Call main function directly
	main()
}
