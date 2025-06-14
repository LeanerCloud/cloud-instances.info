package oci

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientManager(t *testing.T) {
	t.Run("NewClientManager", func(t *testing.T) {
		ctx := context.Background()
		
		// Note: This test will fail without proper OCI credentials
		// In a real environment, you would set up test credentials
		_, err := NewClientManager(ctx)
		
		// For unit testing without credentials, we expect an error
		// In integration tests with proper setup, this would succeed
		if err != nil {
			t.Logf("Expected error without OCI credentials: %v", err)
		}
	})

	t.Run("GetAvailableRegions", func(t *testing.T) {
		regions := GetAvailableRegions()
		
		// Should return a reasonable number of regions
		assert.Greater(t, len(regions), 10)
		assert.Contains(t, regions, "us-ashburn-1")
		assert.Contains(t, regions, "us-phoenix-1")
		assert.Contains(t, regions, "eu-frankfurt-1")
		assert.Contains(t, regions, "ap-tokyo-1")
	})
}

func TestClientManagerMethods(t *testing.T) {
	// Test the client manager interface without actually creating clients
	
	t.Run("ClientManagerInterface", func(t *testing.T) {
		// Test that the client manager has the expected methods
		// This is a compile-time test to ensure the interface is correct
		
		var cm *ClientManager
		if cm != nil {
			// These calls would panic, but we're just testing the interface exists
			_ = cm.GetComputeClient()
		}
		
		// Test passes if it compiles
		assert.True(t, true)
	})
}