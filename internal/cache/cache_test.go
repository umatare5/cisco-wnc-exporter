package cache_test

import (
	"errors"
	"testing"
	"time"

	"github.com/umatare5/cisco-wnc-exporter/internal/cache"
)

func TestNew(t *testing.T) {
	t.Parallel()
	ttl := 5 * time.Second
	name := "test-cache"

	c := cache.New[string](ttl, name)

	// Test that constructor doesn't panic and cache is initially empty
	data, ok := c.GetCached()
	if ok {
		t.Error("expected new cache to be empty")
	}
	if data != "" {
		t.Errorf("expected empty string for new cache, got %q", data)
	}
}

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

	// Verify cache is populated by checking GetCached
	cachedData, ok := c.GetCached()
	if !ok {
		t.Error("expected cache to be populated")
	}
	if cachedData != expectedData {
		t.Errorf("expected cached data %q, got %q", expectedData, cachedData)
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

func TestCache_GetCached_Fresh(t *testing.T) {
	t.Parallel()
	c := cache.New[string](5*time.Second, "test")
	expectedData := "cached-data"

	// Populate cache
	_, err := c.Get(func() (string, error) {
		return expectedData, nil
	})
	if err != nil {
		t.Fatalf("failed to populate cache: %v", err)
	}

	// Get cached data
	data, ok := c.GetCached()

	if !ok {
		t.Error("expected cache to be fresh")
	}
	if data != expectedData {
		t.Errorf("expected %q, got %q", expectedData, data)
	}
}

//nolint:paralleltest // Cannot be parallel due to time.Sleep
func TestCache_GetCached_Stale(t *testing.T) {
	c := cache.New[string](1*time.Millisecond, "test")

	// Populate cache
	_, err := c.Get(func() (string, error) {
		return "data", nil
	})
	if err != nil {
		t.Fatalf("failed to populate cache: %v", err)
	}

	// Wait for cache to expire
	time.Sleep(2 * time.Millisecond)

	// Should return stale
	data, ok := c.GetCached()

	if ok {
		t.Error("expected cache to be stale")
	}
	if data != "" {
		t.Errorf("expected empty string for stale cache, got %q", data)
	}
}

func TestCache_GetCached_Empty(t *testing.T) {
	t.Parallel()
	c := cache.New[string](5*time.Second, "test")

	data, ok := c.GetCached()

	if ok {
		t.Error("expected empty cache to return false")
	}
	if data != "" {
		t.Errorf("expected empty string for empty cache, got %q", data)
	}
}

func TestCache_Invalidate(t *testing.T) {
	t.Parallel()
	c := cache.New[string](5*time.Second, "test")

	// Populate cache
	_, err := c.Get(func() (string, error) {
		return "data", nil
	})
	if err != nil {
		t.Fatalf("failed to populate cache: %v", err)
	}

	// Invalidate cache
	c.Invalidate()

	// Should be empty now
	data, ok := c.GetCached()
	if ok {
		t.Error("expected invalidated cache to return false")
	}
	if data != "" {
		t.Errorf("expected empty string for invalidated cache, got %q", data)
	}
}

func TestCache_SetTTL(t *testing.T) {
	t.Parallel()
	c := cache.New[string](5*time.Second, "test")
	newTTL := 10 * time.Second

	c.SetTTL(newTTL)

	// Test that SetTTL doesn't panic (internal behavior not testable)
	// Behavior tested indirectly in TestCache_SetTTL_AffectsFreshness
}

//nolint:paralleltest // Cannot be parallel due to time.Sleep
func TestCache_SetTTL_AffectsFreshness(t *testing.T) {
	c := cache.New[string](1*time.Millisecond, "test")

	// Populate cache
	_, err := c.Get(func() (string, error) {
		return "data", nil
	})
	if err != nil {
		t.Fatalf("failed to populate cache: %v", err)
	}

	// Wait for cache to expire with current TTL
	time.Sleep(2 * time.Millisecond)

	// Extend TTL significantly
	c.SetTTL(5 * time.Second)

	// Should now be fresh because TTL was extended beyond elapsed time
	data, ok := c.GetCached()
	if !ok {
		t.Error("expected cache to be fresh after TTL extension")
	}
	if data != "data" {
		t.Errorf("expected cached data \"data\", got %q", data)
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
		data, ok := c.GetCached()
		if ok && data != 42 {
			t.Errorf("goroutine 2 got unexpected data: %d", data)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}
