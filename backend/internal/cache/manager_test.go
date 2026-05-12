package cache

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestManagerResponseCache(t *testing.T) {
	logger := zap.NewNop()

	t.Run("NewManager creates responseCache with correct TTL", func(t *testing.T) {
		manager := NewManager(
			100,
			5*time.Minute,
			10*time.Minute,
			30*time.Second, // responseCacheTTL
			logger,
		)
		defer manager.Close()

		if manager == nil {
			t.Fatal("expected manager to be created, got nil")
		}
	})

	t.Run("GetResponseCache returns cached response for valid key", func(t *testing.T) {
		manager := NewManager(100, 5*time.Minute, 10*time.Minute, 30*time.Second, logger)
		defer manager.Close()

		key := "test-key"
		value := []byte(`{"status":"ok"}`)

		// Set the value
		manager.SetResponseCache(key, value)

		// Get the value
		got, found := manager.GetResponseCache(key)
		if !found {
			t.Error("expected to find cached response")
		}
		if string(got) != string(value) {
			t.Errorf("expected %s, got %s", value, got)
		}
	})

	t.Run("SetResponseCache stores response correctly", func(t *testing.T) {
		manager := NewManager(100, 5*time.Minute, 10*time.Minute, 30*time.Second, logger)
		defer manager.Close()

		key := "store-test-key"
		value := []byte(`{"data":"stored"}`)

		manager.SetResponseCache(key, value)

		got, found := manager.GetResponseCache(key)
		if !found {
			t.Error("expected to find stored response")
		}
		if string(got) != string(value) {
			t.Errorf("expected %s, got %s", value, got)
		}
	})

	t.Run("ClearResponseCache removes entry", func(t *testing.T) {
		manager := NewManager(100, 5*time.Minute, 10*time.Minute, 30*time.Second, logger)
		defer manager.Close()

		key := "clear-test-key"
		value := []byte(`{"to":"clear"}`)

		manager.SetResponseCache(key, value)
		manager.ClearResponseCache(key)

		_, found := manager.GetResponseCache(key)
		if found {
			t.Error("expected response to be cleared")
		}
	})

	t.Run("GetResponseMetrics returns cache metrics", func(t *testing.T) {
		manager := NewManager(100, 5*time.Minute, 10*time.Minute, 30*time.Second, logger)
		defer manager.Close()

		key := "metrics-test-key"
		value := []byte(`{"metrics":"test"}`)

		// Set and get to generate hit
		manager.SetResponseCache(key, value)
		_, _ = manager.GetResponseCache(key)

		metrics := manager.GetResponseMetrics()
		if metrics.Hits != 1 {
			t.Errorf("expected 1 hit, got %d", metrics.Hits)
		}
	})

	t.Run("ManagerMetrics includes ResponseCacheMetrics", func(t *testing.T) {
		manager := NewManager(100, 5*time.Minute, 10*time.Minute, 30*time.Second, logger)
		defer manager.Close()

		// Generate some activity
		manager.SetResponseCache("key1", []byte("value1"))
		_, _ = manager.GetResponseCache("key1")
		_, _ = manager.GetResponseCache("nonexistent")

		metrics := manager.GetMetrics()

		// Check that ResponseCacheMetrics has recorded activity
		if metrics.ResponseCacheMetrics.Hits != 1 {
			t.Errorf("expected 1 response cache hit, got %d", metrics.ResponseCacheMetrics.Hits)
		}
		if metrics.ResponseCacheMetrics.Misses != 1 {
			t.Errorf("expected 1 response cache miss, got %d", metrics.ResponseCacheMetrics.Misses)
		}
	})

	t.Run("Clear clears response cache", func(t *testing.T) {
		manager := NewManager(100, 5*time.Minute, 10*time.Minute, 30*time.Second, logger)
		defer manager.Close()

		manager.SetResponseCache("key1", []byte("value1"))
		manager.SetResponseCache("key2", []byte("value2"))

		manager.Clear()

		_, found1 := manager.GetResponseCache("key1")
		_, found2 := manager.GetResponseCache("key2")

		if found1 || found2 {
			t.Error("expected response cache to be cleared")
		}
	})
}
