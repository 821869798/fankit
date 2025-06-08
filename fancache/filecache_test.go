package fancache

import (
	"bytes"
	"os"
	"testing"
	"time"
)

// Helper function to create a temporary cache directory
func setupTestCache[V any](t *testing.T, options ...Option[V]) (*FileCache[V], func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "filecache_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Specify type V for default options
	defaultOptions := []Option[V]{WithMaxItems[V](10), WithEvictPercent[V](0.3)}
	allOptions := append(defaultOptions, options...)

	// Specify type V for NewFileCache
	fc, err := NewFileCache[V](tmpDir, allOptions...)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("NewFileCache failed: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}
	return fc, cleanup
}

func TestFileCache_NewFileCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "newcache_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Specify a type for V, e.g., string, when calling NewFileCache
	fc, err := NewFileCache[string](tmpDir)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}
	if fc == nil {
		t.Fatal("NewFileCache() returned nil")
	}
	if fc.dir != tmpDir {
		t.Errorf("fc.dir = %s; want %s", fc.dir, tmpDir)
	}
	if fc.maxItems != DefaultMaxItems {
		t.Errorf("fc.maxItems = %d; want %d", fc.maxItems, DefaultMaxItems)
	}
	if fc.evictPercent != DefaultEvictPercent {
		t.Errorf("fc.evictPercent = %f; want %f", fc.evictPercent, DefaultEvictPercent)
	}
}

func TestFileCache_SetGetRemove(t *testing.T) {
	// Specify a type for V, e.g., string, when calling setupTestCache
	fc, cleanup := setupTestCache[string](t)
	defer cleanup()

	key := "testKey"
	value := "testValue" // This will be of type string, matching V
	duration := time.Minute

	// Test Set
	if err := fc.Set(key, value, duration); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Test Get
	// retrievedValue will be of type string (V)
	retrievedValue, found, err := fc.Get(key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !found {
		t.Fatal("Get: key not found after Set")
	}
	// No type assertion needed as retrievedValue is already of type string
	if retrievedValue != value {
		t.Errorf("Get: expected value %s, got %s", value, retrievedValue)
	}

	// Test Get non-existent key
	_, found, err = fc.Get("nonExistentKey")
	if err != nil {
		t.Fatalf("Get non-existent key failed: %v", err)
	}
	if found {
		t.Error("Get: found non-existent key")
	}

	// Test Remove
	if err := fc.Remove(key); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Test Get after Remove
	_, found, err = fc.Get(key)
	if err != nil {
		t.Fatalf("Get after Remove failed: %v", err)
	}
	if found {
		t.Error("Get: key found after Remove")
	}

	// Test Remove non-existent key
	if err := fc.Remove("nonExistentKey"); err != nil {
		t.Fatalf("Remove non-existent key failed: %v", err)
	}
}

func TestFileCache_SetGetRemove_Bytes(t *testing.T) {
	// Specify []byte as the type for V when calling setupTestCache
	fc, cleanup := setupTestCache[[]byte](t)
	defer cleanup()

	key := "testKeyBytes"
	value := []byte("testValueBytes") // This will be of type []byte
	duration := time.Minute

	// Test Set
	if err := fc.Set(key, value, duration); err != nil {
		t.Fatalf("Set (bytes) failed: %v", err)
	}

	// Test Get
	retrievedValue, found, err := fc.Get(key)
	if err != nil {
		t.Fatalf("Get (bytes) failed: %v", err)
	}
	if !found {
		t.Fatal("Get (bytes): key not found after Set")
	}
	// Use bytes.Equal for comparing byte slices
	if !bytes.Equal(retrievedValue, value) {
		t.Errorf("Get (bytes): expected value %v, got %v", value, retrievedValue)
	}

	// Test Get non-existent key
	_, found, err = fc.Get("nonExistentKeyBytes")
	if err != nil {
		t.Fatalf("Get (bytes) non-existent key failed: %v", err)
	}
	if found {
		t.Error("Get (bytes): found non-existent key")
	}

	// Test Remove
	if err := fc.Remove(key); err != nil {
		t.Fatalf("Remove (bytes) failed: %v", err)
	}

	// Test Get after Remove
	_, found, err = fc.Get(key)
	if err != nil {
		t.Fatalf("Get (bytes) after Remove failed: %v", err)
	}
	if found {
		t.Error("Get (bytes): key found after Remove")
	}

	// Test Remove non-existent key
	if err := fc.Remove("nonExistentKeyBytes"); err != nil {
		t.Fatalf("Remove (bytes) non-existent key failed: %v", err)
	}
}
