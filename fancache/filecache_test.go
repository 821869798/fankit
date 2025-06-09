package fancache

import (
	"bytes"
	"os"
	"testing"
	"time"
)

// Helper function to create a temporary cache directory
func setupTestCache(t *testing.T, options ...Option) (*FileCache, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "filecache_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	defaultOptions := []Option{WithMaxItems(10), WithEvictPercent(0.3)}
	allOptions := append(defaultOptions, options...)

	fc, err := NewFileCache(tmpDir, allOptions...)
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

	fc, err := NewFileCache(tmpDir)
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
	fc, cleanup := setupTestCache(t)
	defer cleanup()

	key := "testKey"
	value := "testValue"
	duration := time.Minute

	// Test Set
	if err := fc.Set(key, value, duration); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Test Get
	var retrievedString string
	found, err := fc.Get(key, &retrievedString)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !found {
		t.Fatal("Get: key not found after Set")
	}

	if retrievedString != value {
		t.Errorf("Get: expected value %s, got %s", value, retrievedString)
	}

	// Test Get non-existent key
	found, err = fc.Get("nonExistentKey", &retrievedString)
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
	found, err = fc.Get(key, &retrievedString)
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
	fc, cleanup := setupTestCache(t)
	defer cleanup()

	key := "testKeyBytes"
	value := []byte("testValueBytes")
	duration := time.Minute

	// Test Set
	if err := fc.Set(key, value, duration); err != nil {
		t.Fatalf("Set (bytes) failed: %v", err)
	}

	// Test Get
	var retrievedValue []byte // Declare a variable to store the retrieved bytes
	found, err := fc.Get(key, &retrievedValue)
	if err != nil {
		t.Fatalf("Get (bytes) failed: %v", err)
	}
	if !found {
		t.Fatal("Get (bytes): key not found after Set")
	}
	// No type assertion needed as retrievedValue is already []byte if gob decoding is successful
	if !bytes.Equal(retrievedValue, value) {
		t.Errorf("Get (bytes): expected value %v, got %v", value, retrievedValue)
	}

	// Test Get non-existent key
	var nonExistentValue []byte // Declare a variable for the non-existent key
	found, err = fc.Get("nonExistentKeyBytes", &nonExistentValue)
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
	var valueAfterRemove []byte // Declare a variable for the value after remove
	found, err = fc.Get(key, &valueAfterRemove)
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
