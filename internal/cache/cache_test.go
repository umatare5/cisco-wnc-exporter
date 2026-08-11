package cache_test

import (
	"errors"
	"testing"
	"time"

	"github.com/umatare5/cisco-wnc-exporter/internal/cache"
)

func TestCache_Get_InitialFetch(t *testing.T) {
	t.Parallel()
	c := cache.New[string](5*time.Second, "test")
	expectedData := "test-data"

	data, err := c.Get(func() (string, error) {
		return expectedData, nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if data != expectedData {
		t.Errorf("expected %q, got %q", expectedData, data)
	}
}

func TestCache_Get_FromCache(t *testing.T) {
	t.Parallel()
	c := cache.New[string](5*time.Second, "test")
	expectedData := "cached-data"

	// First call to populate cache
	_, err := c.Get(func() (string, error) {
		return expectedData, nil
	})
	if err != nil {
		t.Fatalf("failed to populate cache: %v", err)
	}

	// Second call should return from cache
	data, err := c.Get(func() (string, error) {
		return "should-not-be-called", nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if data != expectedData {
		t.Errorf("expected cached data %q, got %q", expectedData, data)
	}
}

//nolint:paralleltest // Cannot be parallel due to time.Sleep
func TestCache_Get_ExpiredCache(t *testing.T) {
	c := cache.New[string](1*time.Millisecond, "test")

	// Populate cache
	_, err := c.Get(func() (string, error) {
		return "old-data", nil
	})
	if err != nil {
		t.Fatalf("failed to populate cache: %v", err)
	}

	// Wait for cache to expire
	time.Sleep(2 * time.Millisecond)

	// Should fetch new data
	newData := "new-data"
	data, err := c.Get(func() (string, error) {
		return newData, nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if data != newData {
		t.Errorf("expected new data %q, got %q", newData, data)
	}
}

func TestCache_Get_RefreshError(t *testing.T) {
	t.Parallel()
	c := cache.New[string](5*time.Second, "test")
	expectedErr := errors.New("refresh failed")

	data, err := c.Get(func() (string, error) {
		return "", expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
	if data != "" {
		t.Errorf("expected empty string on error, got %q", data)
	}
}

// Test concurrent access (basic scenario).
//
//nolint:paralleltest // Cannot be parallel due to goroutines coordination
func TestCache_ConcurrentAccess(t *testing.T) {
	c := cache.New[int](100*time.Millisecond, "test")

	// Start multiple goroutines
	done := make(chan bool, 2)

	go func() {
		_, err := c.Get(func() (int, error) {
			time.Sleep(10 * time.Millisecond) // Simulate work
			return 42, nil
		})
		if err != nil {
			t.Errorf("goroutine 1 error: %v", err)
		}
		done <- true
	}()

	go func() {
		time.Sleep(5 * time.Millisecond) // Start slightly later
		// The first goroutine holds the write lock while it refreshes, so this read
		// waits and then finds the cache fresh instead of refreshing a second time.
		data, err := c.Get(func() (int, error) {
			return 0, errors.New("second refresh")
		})
		if err != nil {
			t.Errorf("goroutine 2 error: %v", err)
		}
		if data != 42 {
			t.Errorf("goroutine 2 got unexpected data: %d", data)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}
