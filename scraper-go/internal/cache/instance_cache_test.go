package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
	
	"github.com/cloud-instances/scraper/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstanceCache(t *testing.T) {
	// Create temp directory for cache
	tempDir, err := os.MkdirTemp("", "instance_cache_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)
	
	// Create cache with 1 second TTL for testing
	cache := NewInstanceCache(tempDir, 1*time.Second)
	
	t.Run("CacheKeyGeneration", func(t *testing.T) {
		key1 := cache.CacheKeyForInstance("ec2", "t2.micro", "us-east-1")
		key2 := cache.CacheKeyForInstance("ec2", "t2.micro", "us-east-1")
		key3 := cache.CacheKeyForInstance("ec2", "t2.micro", "us-west-2")
		
		assert.Equal(t, key1, key2, "Same inputs should generate same key")
		assert.NotEqual(t, key1, key3, "Different inputs should generate different keys")
	})
	
	t.Run("SetAndGetInstance", func(t *testing.T) {
		instance := &models.Instance{
			InstanceType: "t2.micro",
			Memory:       "1.0",
			VCPU:         "1",
			Family:       "General purpose",
		}
		
		// Set instance
		err := cache.SetInstance("ec2", "t2.micro", "us-east-1", instance)
		require.NoError(t, err)
		
		// Get instance
		retrieved, hit, err := cache.GetInstance("ec2", "t2.micro", "us-east-1")
		require.NoError(t, err)
		assert.True(t, hit, "Should be a cache hit")
		assert.Equal(t, instance.InstanceType, retrieved.InstanceType)
		assert.Equal(t, instance.Memory, retrieved.Memory)
		
		// Get non-existent instance
		_, hit, err = cache.GetInstance("ec2", "t2.large", "us-east-1")
		require.NoError(t, err)
		assert.False(t, hit, "Should be a cache miss")
	})
	
	t.Run("CacheExpiration", func(t *testing.T) {
		instance := &models.Instance{
			InstanceType: "t3.micro",
		}
		
		// Set instance
		err := cache.SetInstance("ec2", "t3.micro", "us-east-1", instance)
		require.NoError(t, err)
		
		// Verify it exists
		_, hit, err := cache.GetInstance("ec2", "t3.micro", "us-east-1")
		require.NoError(t, err)
		assert.True(t, hit)
		
		// Wait for expiration
		time.Sleep(1100 * time.Millisecond)
		
		// Should be expired
		_, hit, err = cache.GetInstance("ec2", "t3.micro", "us-east-1")
		require.NoError(t, err)
		assert.False(t, hit, "Cache should be expired")
	})
	
	t.Run("SetAndGetPricing", func(t *testing.T) {
		pricing := map[string]models.RegionPricing{
			"us-east-1": {
				OnDemand: 0.0116,
				Reserved: map[string]float64{
					"yrTerm1Standard.allUpfront": 0.0072,
				},
			},
			"us-west-2": {
				OnDemand: 0.0116,
			},
		}
		
		// Set pricing
		err := cache.SetPricing("ec2", "t2.micro", pricing)
		require.NoError(t, err)
		
		// Get pricing
		retrieved, hit, err := cache.GetPricing("ec2", "t2.micro")
		require.NoError(t, err)
		assert.True(t, hit)
		assert.Equal(t, len(pricing), len(retrieved))
		assert.Equal(t, pricing["us-east-1"].OnDemand, retrieved["us-east-1"].OnDemand)
	})
	
	t.Run("ClearExpired", func(t *testing.T) {
		// Create a new cache with very short TTL
		shortCache := NewInstanceCache(tempDir, 100*time.Millisecond)
		
		instance := &models.Instance{
			InstanceType: "m5.large",
		}
		
		// Set multiple instances
		err := shortCache.SetInstance("ec2", "m5.large", "us-east-1", instance)
		require.NoError(t, err)
		err = shortCache.SetInstance("ec2", "m5.xlarge", "us-east-1", instance)
		require.NoError(t, err)
		
		// Wait for expiration
		time.Sleep(200 * time.Millisecond)
		
		// Clear expired
		err = shortCache.ClearExpired()
		require.NoError(t, err)
		
		// Verify files are gone
		cacheDir := filepath.Join(tempDir, "ec2", "instances")
		files, err := os.ReadDir(cacheDir)
		if err == nil {
			for _, file := range files {
				info, err := os.Stat(filepath.Join(cacheDir, file.Name()))
				if err == nil {
					assert.True(t, time.Since(info.ModTime()) < 200*time.Millisecond, 
						"Old files should have been removed")
				}
			}
		}
	})
}