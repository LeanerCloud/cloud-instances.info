package azure

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientManager(t *testing.T) {
	t.Run("NewClientManager", func(t *testing.T) {
		ctx := context.Background()
		
		cm, err := NewClientManager(ctx)
		
		assert.NoError(t, err)
		assert.NotNil(t, cm)
		assert.NotNil(t, cm.GetHTTPClient())
	})

	t.Run("GetAvailableRegions", func(t *testing.T) {
		regions := GetAvailableRegions()
		
		// Should return a reasonable number of regions
		assert.Greater(t, len(regions), 15)
		assert.Contains(t, regions, "eastus")
		assert.Contains(t, regions, "westus2")
		assert.Contains(t, regions, "westeurope")
		assert.Contains(t, regions, "southeastasia")
	})
}

func TestClientManagerMethods(t *testing.T) {
	t.Run("ClientManagerInterface", func(t *testing.T) {
		ctx := context.Background()
		cm, err := NewClientManager(ctx)
		
		assert.NoError(t, err)
		assert.NotNil(t, cm.GetHTTPClient())
		assert.Equal(t, 30000000000, int(cm.GetHTTPClient().Timeout)) // 30 seconds in nanoseconds
	})
}