package cache

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Cache provides file-based caching for scraper results
type Cache struct {
	baseDir string
	ttl     time.Duration
}

// NewCache creates a new cache instance
func NewCache(baseDir string, ttl time.Duration) *Cache {
	return &Cache{
		baseDir: baseDir,
		ttl:     ttl,
	}
}

// CacheKey generates a cache key from service and regions
func (c *Cache) CacheKey(service string, regions []string) string {
	data := fmt.Sprintf("%s:%v", service, regions)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%s_%x.json", service, hash)
}

// Get retrieves cached data if it exists and is not expired
func (c *Cache) Get(key string, dest interface{}) (bool, error) {
	cachePath := filepath.Join(c.baseDir, key)
	
	// Check if file exists
	info, err := os.Stat(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	
	// Check if cache is expired
	if time.Since(info.ModTime()) > c.ttl {
		return false, nil
	}
	
	// Read and unmarshal cache file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return false, err
	}
	
	if err := json.Unmarshal(data, dest); err != nil {
		return false, err
	}
	
	return true, nil
}

// Set stores data in the cache
func (c *Cache) Set(key string, data interface{}) error {
	// Ensure cache directory exists
	if err := os.MkdirAll(c.baseDir, 0755); err != nil {
		return err
	}
	
	// Marshal data
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	
	// Write to cache file
	cachePath := filepath.Join(c.baseDir, key)
	return os.WriteFile(cachePath, jsonData, 0644)
}

// Clear removes all cached data
func (c *Cache) Clear() error {
	return os.RemoveAll(c.baseDir)
}