package cache

import (
	"os"
	"testing"
	"time"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	// Create temp directory for cache
	tempDir, err := os.MkdirTemp("", "cache_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Create cache with 1 second TTL for testing
	cache := NewCache(tempDir, 1*time.Second)
	
	t.Run("CacheKey", func(t *testing.T) {
		key1 := cache.CacheKey("ec2", []string{"us-east-1", "us-west-2"})
		key3 := cache.CacheKey("rds", []string{"us-east-1", "us-west-2"})
		
		// Keys should start with service name
		assert.Contains(t, key1, "ec2_")
		assert.Contains(t, key3, "rds_")
		
		// Keys should end with .json
		assert.Contains(t, key1, ".json")
		
		// Different service should generate different key
		assert.NotEqual(t, key1, key3)
	})
	
	t.Run("SetAndGet", func(t *testing.T) {
		type testData struct {
			Name  string
			Value int
		}
		
		data := testData{
			Name:  "test",
			Value: 42,
		}
		
		key := "test_key"
		
		// Set data
		err := cache.Set(key, data)
		require.NoError(t, err)
		
		// Get data
		var retrieved testData
		hit, err := cache.Get(key, &retrieved)
		require.NoError(t, err)
		assert.True(t, hit)
		assert.Equal(t, data.Name, retrieved.Name)
		assert.Equal(t, data.Value, retrieved.Value)
		
		// Get non-existent key
		var notFound testData
		hit, err = cache.Get("nonexistent", &notFound)
		require.NoError(t, err)
		assert.False(t, hit)
	})
	
	t.Run("CacheExpiration", func(t *testing.T) {
		data := map[string]string{
			"key": "value",
		}
		
		key := "expiring_key"
		
		// Set data
		err := cache.Set(key, data)
		require.NoError(t, err)
		
		// Verify it exists
		var retrieved map[string]string
		hit, err := cache.Get(key, &retrieved)
		require.NoError(t, err)
		assert.True(t, hit)
		
		// Wait for expiration
		time.Sleep(1100 * time.Millisecond)
		
		// Should be expired
		hit, err = cache.Get(key, &retrieved)
		require.NoError(t, err)
		assert.False(t, hit)
	})
	
	t.Run("ComplexData", func(t *testing.T) {
		// Test with slice of structs
		type Instance struct {
			Type   string
			Memory string
		}
		
		instances := []Instance{
			{Type: "t2.micro", Memory: "1.0"},
			{Type: "t2.small", Memory: "2.0"},
			{Type: "t2.medium", Memory: "4.0"},
		}
		
		key := "instances"
		
		// Set data
		err := cache.Set(key, instances)
		require.NoError(t, err)
		
		// Get data
		var retrieved []Instance
		hit, err := cache.Get(key, &retrieved)
		require.NoError(t, err)
		assert.True(t, hit)
		assert.Equal(t, len(instances), len(retrieved))
		assert.Equal(t, instances[0].Type, retrieved[0].Type)
		assert.Equal(t, instances[2].Memory, retrieved[2].Memory)
	})
	
	t.Run("ConcurrentAccess", func(t *testing.T) {
		// Test concurrent reads and writes
		done := make(chan bool)
		
		// Writer goroutine
		go func() {
			for i := 0; i < 10; i++ {
				cache.Set("concurrent", i)
				time.Sleep(10 * time.Millisecond)
			}
			done <- true
		}()
		
		// Reader goroutine
		go func() {
			for i := 0; i < 10; i++ {
				var value int
				cache.Get("concurrent", &value)
				time.Sleep(10 * time.Millisecond)
			}
			done <- true
		}()
		
		// Wait for both goroutines
		<-done
		<-done
		
		// Final read should work
		var finalValue int
		hit, err := cache.Get("concurrent", &finalValue)
		if hit {
			require.NoError(t, err)
			assert.GreaterOrEqual(t, finalValue, 0)
			assert.LessOrEqual(t, finalValue, 9)
		}
	})

	t.Run("Clear", func(t *testing.T) {
		cache := NewCache(t.TempDir(), time.Hour)
		
		// Add some data
		cache.Set("key1", "value1")
		cache.Set("key2", "value2")
		
		// Verify data exists
		var value string
		hit, err := cache.Get("key1", &value)
		require.NoError(t, err)
		assert.True(t, hit)
		
		// Clear cache
		cache.Clear()
		
		// Verify data is gone
		hit, err = cache.Get("key1", &value)
		require.NoError(t, err)
		assert.False(t, hit)
		
		hit, err = cache.Get("key2", &value)
		require.NoError(t, err)
		assert.False(t, hit)
	})
}