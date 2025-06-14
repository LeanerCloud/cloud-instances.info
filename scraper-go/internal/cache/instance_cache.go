package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	
	"github.com/cloud-instances/scraper/pkg/models"
)

// InstanceCache provides per-instance caching with expiration
type InstanceCache struct {
	baseDir string
	ttl     time.Duration
}

// NewInstanceCache creates a new instance cache
func NewInstanceCache(baseDir string, ttl time.Duration) *InstanceCache {
	return &InstanceCache{
		baseDir: baseDir,
		ttl:     ttl,
	}
}

// CacheKeyForInstance generates a cache key for a specific instance
func (ic *InstanceCache) CacheKeyForInstance(service, instanceType, region string) string {
	key := fmt.Sprintf("%s_%s_%s", service, instanceType, region)
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}

// GetInstance retrieves a cached instance if it exists and is not expired
func (ic *InstanceCache) GetInstance(service, instanceType, region string) (*models.Instance, bool, error) {
	key := ic.CacheKeyForInstance(service, instanceType, region)
	cacheFile := filepath.Join(ic.baseDir, service, "instances", key+".json")
	
	// Check if file exists
	info, err := os.Stat(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil // Cache miss
		}
		return nil, false, err
	}
	
	// Check if expired
	if time.Since(info.ModTime()) > ic.ttl {
		return nil, false, nil // Cache expired
	}
	
	// Read and unmarshal
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, false, err
	}
	
	var instance models.Instance
	if err := json.Unmarshal(data, &instance); err != nil {
		return nil, false, err
	}
	
	return &instance, true, nil
}

// SetInstance stores an instance in the cache
func (ic *InstanceCache) SetInstance(service, instanceType, region string, instance *models.Instance) error {
	key := ic.CacheKeyForInstance(service, instanceType, region)
	cacheDir := filepath.Join(ic.baseDir, service, "instances")
	cacheFile := filepath.Join(cacheDir, key+".json")
	
	// Create directory if needed
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	
	// Marshal data
	data, err := json.MarshalIndent(instance, "", "  ")
	if err != nil {
		return err
	}
	
	// Write to file
	return os.WriteFile(cacheFile, data, 0644)
}

// GetPricing retrieves cached pricing data for an instance type
func (ic *InstanceCache) GetPricing(service, instanceType string) (map[string]models.RegionPricing, bool, error) {
	key := fmt.Sprintf("%s_%s_pricing", service, instanceType)
	hash := md5.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	
	cacheFile := filepath.Join(ic.baseDir, service, "pricing", hashStr+".json")
	
	// Check if file exists
	info, err := os.Stat(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil // Cache miss
		}
		return nil, false, err
	}
	
	// Check if expired
	if time.Since(info.ModTime()) > ic.ttl {
		return nil, false, nil // Cache expired
	}
	
	// Read and unmarshal
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, false, err
	}
	
	var pricing map[string]models.RegionPricing
	if err := json.Unmarshal(data, &pricing); err != nil {
		return nil, false, err
	}
	
	return pricing, true, nil
}

// SetPricing stores pricing data in the cache
func (ic *InstanceCache) SetPricing(service, instanceType string, pricing map[string]models.RegionPricing) error {
	key := fmt.Sprintf("%s_%s_pricing", service, instanceType)
	hash := md5.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	
	cacheDir := filepath.Join(ic.baseDir, service, "pricing")
	cacheFile := filepath.Join(cacheDir, hashStr+".json")
	
	// Create directory if needed
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}
	
	// Marshal data
	data, err := json.MarshalIndent(pricing, "", "  ")
	if err != nil {
		return err
	}
	
	// Write to file
	return os.WriteFile(cacheFile, data, 0644)
}

// ClearExpired removes expired cache entries
func (ic *InstanceCache) ClearExpired() error {
	return filepath.Walk(ic.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			if time.Since(info.ModTime()) > ic.ttl {
				os.Remove(path)
			}
		}
		
		return nil
	})
}