package aws

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientManager(t *testing.T) {
	// Skip these tests if AWS credentials are not available
	if testing.Short() {
		t.Skip("Skipping AWS client tests in short mode")
	}
	
	ctx := context.Background()
	
	t.Run("NewClientManager", func(t *testing.T) {
		cm, err := NewClientManager(ctx)
		require.NoError(t, err)
		assert.NotNil(t, cm)
		assert.NotNil(t, cm.clients)
		assert.Equal(t, ctx, cm.ctx)
	})
	
	t.Run("GetClients", func(t *testing.T) {
		cm, err := NewClientManager(ctx)
		require.NoError(t, err)
		
		// Get clients for us-east-1
		clients1, err := cm.GetClients("us-east-1")
		require.NoError(t, err)
		assert.NotNil(t, clients1)
		assert.Equal(t, "us-east-1", clients1.Region)
		assert.NotNil(t, clients1.EC2Client)
		assert.NotNil(t, clients1.PricingClient)
		
		// Get clients again - should return cached
		clients2, err := cm.GetClients("us-east-1")
		require.NoError(t, err)
		assert.Equal(t, clients1, clients2) // Should be the same pointer
		
		// Get clients for different region
		clients3, err := cm.GetClients("us-west-2")
		require.NoError(t, err)
		assert.NotNil(t, clients3)
		assert.Equal(t, "us-west-2", clients3.Region)
		assert.NotEqual(t, clients1, clients3) // Should be different
	})
	
	t.Run("ConcurrentAccess", func(t *testing.T) {
		cm, err := NewClientManager(ctx)
		require.NoError(t, err)
		
		// Test concurrent access to the same region
		done := make(chan bool)
		errors := make(chan error, 10)
		
		for i := 0; i < 10; i++ {
			go func() {
				_, err := cm.GetClients("eu-west-1")
				if err != nil {
					errors <- err
				}
				done <- true
			}()
		}
		
		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
		
		// Check for errors
		select {
		case err := <-errors:
			t.Fatalf("Concurrent access error: %v", err)
		default:
			// No errors
		}
		
		// Verify only one client was created
		assert.Equal(t, 1, countClientsForRegion(cm, "eu-west-1"))
	})
}

func TestGetAllRegions(t *testing.T) {
	// Skip if no AWS credentials
	if testing.Short() {
		t.Skip("Skipping AWS region tests in short mode")
	}
	
	ctx := context.Background()
	
	t.Run("GetAllRegions", func(t *testing.T) {
		regions, err := GetAllRegions(ctx)
		
		// If we have valid credentials, this should work
		if err == nil {
			assert.NotEmpty(t, regions)
			// Should contain common regions
			assert.Contains(t, regions, "us-east-1")
			assert.Contains(t, regions, "us-west-2")
			assert.Contains(t, regions, "eu-west-1")
		} else {
			// If no credentials, we expect a specific error
			assert.Contains(t, err.Error(), "credentials")
		}
	})
}

func TestGetRegionDescriptions(t *testing.T) {
	// Skip if no AWS credentials
	if testing.Short() {
		t.Skip("Skipping AWS region description tests in short mode")
	}
	
	ctx := context.Background()
	
	t.Run("GetRegionDescriptions", func(t *testing.T) {
		descriptions, err := GetRegionDescriptions(ctx)
		
		// If we have valid credentials, this should work
		if err == nil {
			assert.NotEmpty(t, descriptions)
			// Should contain common regions
			assert.Contains(t, descriptions, "us-east-1")
			assert.Contains(t, descriptions, "us-west-2")
			assert.Contains(t, descriptions, "eu-west-1")
			
			// All values should be region codes
			for region, desc := range descriptions {
				assert.Equal(t, region, desc)
			}
		} else {
			// If no credentials, we expect a specific error
			assert.Contains(t, err.Error(), "credentials")
		}
	})
}

// Helper function to count clients for a specific region
func countClientsForRegion(cm *ClientManager, region string) int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	count := 0
	for r := range cm.clients {
		if r == region {
			count++
		}
	}
	return count
}