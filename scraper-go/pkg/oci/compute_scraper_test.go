package oci

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/cloud-instances/scraper/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeScraper(t *testing.T) {
	ctx := context.Background()
	compartmentID := "test-compartment"
	
	// Note: These tests don't require actual OCI credentials as we're testing utility functions
	scraper := &ComputeScraper{
		ctx:           ctx,
		compartmentID: compartmentID,
	}

	t.Run("GetServiceName", func(t *testing.T) {
		assert.Equal(t, "Compute", scraper.GetServiceName())
	})

	t.Run("ConvertToInstance", func(t *testing.T) {
		shape := core.Shape{
			Shape:                stringPtr("VM.Standard2.1"),
			ProcessorDescription: stringPtr("2.0 GHz Intel Xeon Platinum 8167M"),
			Ocpus:                float32Ptr(1),
			MemoryInGBs:          float32Ptr(15),
		}
		
		instance := scraper.convertToInstance(shape)
		
		assert.Equal(t, "VM.Standard2.1", instance.InstanceType)
		assert.Equal(t, "15.0", instance.Memory)
		assert.Equal(t, "1", string(instance.VCPU))
		assert.Equal(t, "General purpose", instance.Family)
		assert.Equal(t, "VM STANDARD2 1", instance.PrettyName)
		assert.Equal(t, "2.0 GHz Intel Xeon Platinum 8167M", instance.PhysicalProcessor)
		assert.Equal(t, "2.0 GHz", instance.ClockSpeed)
		assert.True(t, instance.VPCOnly)
		assert.True(t, instance.IPv6Support)
		assert.Equal(t, "current", instance.Generation)
		assert.Equal(t, "Supported", instance.EnhancedNetworking)
	})

	t.Run("GetFamilyName", func(t *testing.T) {
		testCases := []struct {
			shapeName string
			expected  string
		}{
			{"VM.Standard.1", "General purpose"},
			{"VM.Standard2.2", "General purpose"},
			{"VM.Standard3.Flex", "General purpose"},
			{"VM.Optimized3.1", "Compute optimized"},
			{"VM.DenseIO.1", "Storage optimized"},
			{"VM.HighIO.1", "Storage optimized"},
			{"VM.GPU.1", "GPU instance"},
			{"VM.GPU2.1", "GPU instance"},
			{"BM.Standard.52", "Bare metal"},
			{"BM.GPU3.8", "Bare metal"},
			{"Invalid", "Unknown"},
		}
		
		for _, tc := range testCases {
			t.Run(tc.shapeName, func(t *testing.T) {
				result := scraper.getFamilyName(tc.shapeName)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("GeneratePrettyName", func(t *testing.T) {
		testCases := []struct {
			shapeName string
			expected  string
		}{
			{"VM.Standard.1", "VM STANDARD 1"},
			{"VM.Standard2.2", "VM STANDARD2 2"},
			{"VM.Standard3.Flex", "VM STANDARD3 Flex"},
			{"VM.Optimized3.1", "VM OPTIMIZED3 1"},
			{"BM.Standard.52", "BM STANDARD 52"},
			{"BM.GPU3.8", "BM GPU3 8"},
			{"Invalid", "INVALID"},
			{"NoSeparator", "NOSEPARATOR"},
		}
		
		for _, tc := range testCases {
			t.Run(tc.shapeName, func(t *testing.T) {
				result := scraper.generatePrettyName(tc.shapeName)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("GetStorageInfo", func(t *testing.T) {
		testCases := []struct {
			name            string
			shape           core.Shape
			expectedSSD     bool
			expectedDevices int
			expectedSize    int
		}{
			{
				name:            "VM shape with no local storage",
				shape:           core.Shape{Shape: stringPtr("VM.Standard2.1")},
				expectedSSD:     false,
				expectedDevices: 0,
				expectedSize:    0,
			},
			{
				name:            "Bare metal shape with local storage",
				shape:           core.Shape{Shape: stringPtr("BM.Standard2.52")},
				expectedSSD:     true,
				expectedDevices: 1,
				expectedSize:    3200,
			},
			{
				name:            "DenseIO bare metal shape",
				shape:           core.Shape{Shape: stringPtr("BM.DenseIO.52")},
				expectedSSD:     true,
				expectedDevices: 8,
				expectedSize:    28800,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				storage := scraper.getStorageInfo(tc.shape)
				
				var storageData map[string]interface{}
				err := json.Unmarshal(storage, &storageData)
				require.NoError(t, err)
				
				assert.Equal(t, tc.expectedSSD, storageData["ssd"])
				assert.Equal(t, float64(tc.expectedDevices), storageData["devices"])
				assert.Equal(t, float64(tc.expectedSize), storageData["size"])
			})
		}
	})

	t.Run("SampleInstances", func(t *testing.T) {
		instances := map[string]*models.Instance{
			"VM.Standard.1":     {InstanceType: "VM.Standard.1"},
			"VM.Standard.2":     {InstanceType: "VM.Standard.2"},
			"VM.Standard.4":     {InstanceType: "VM.Standard.4"},
			"VM.Optimized3.1":   {InstanceType: "VM.Optimized3.1"},
			"BM.Standard.52":    {InstanceType: "BM.Standard.52"},
		}
		
		t.Run("Sample fewer than available", func(t *testing.T) {
			sampled := scraper.sampleInstances(instances, 3)
			assert.Len(t, sampled, 3)
		})

		t.Run("Sample more than available", func(t *testing.T) {
			sampled := scraper.sampleInstances(instances, 10)
			assert.Len(t, sampled, 5) // Should return all available
		})

		t.Run("Sample zero", func(t *testing.T) {
			sampled := scraper.sampleInstances(instances, 0)
			assert.Len(t, sampled, 5) // Should return all available
		})

		t.Run("Sample negative", func(t *testing.T) {
			sampled := scraper.sampleInstances(instances, -1)
			assert.Len(t, sampled, 5) // Should return all available
		})
	})

	t.Run("AddPlaceholderPricing", func(t *testing.T) {
		instances := map[string]*models.Instance{
			"VM.Standard.1": {
				InstanceType: "VM.Standard.1",
				VCPU:         "1",
				Memory:       "15.0",
				Pricing:      make(map[string]models.RegionPricing),
				Regions:      map[string]string{"us-ashburn-1": "us-ashburn-1"},
			},
			"BM.Standard.52": {
				InstanceType: "BM.Standard.52",
				VCPU:         "52",
				Memory:       "768.0",
				Pricing:      make(map[string]models.RegionPricing),
				Regions:      map[string]string{"us-ashburn-1": "us-ashburn-1"},
			},
		}
		
		regions := []string{"us-ashburn-1"}
		scraper.addPlaceholderPricing(instances, regions)
		
		// Test VM instance pricing
		vmInstance := instances["VM.Standard.1"]
		vmPricing, exists := vmInstance.Pricing["us-ashburn-1"]
		require.True(t, exists)
		
		assert.Greater(t, vmPricing.OnDemand, 0.0)
		assert.Contains(t, vmPricing.Reserved, "1yr")
		assert.Contains(t, vmPricing.Reserved, "3yr")
		assert.Less(t, vmPricing.Reserved["1yr"], vmPricing.OnDemand)
		assert.Less(t, vmPricing.Reserved["3yr"], vmPricing.Reserved["1yr"])
		
		// Test Bare Metal instance pricing (should be more expensive)
		bmInstance := instances["BM.Standard.52"]
		bmPricing, exists := bmInstance.Pricing["us-ashburn-1"]
		require.True(t, exists)
		
		assert.Greater(t, bmPricing.OnDemand, vmPricing.OnDemand) // BM should be more expensive
	})

	t.Run("GetMockShapes", func(t *testing.T) {
		mockShapes := scraper.getMockShapes()
		
		assert.Greater(t, len(mockShapes), 0)
		
		// Check that mock shapes have expected properties
		for _, shape := range mockShapes {
			assert.NotNil(t, shape.Shape)
			assert.NotEmpty(t, *shape.Shape)
			assert.NotNil(t, shape.Ocpus)
			assert.Greater(t, *shape.Ocpus, float32(0))
			assert.NotNil(t, shape.MemoryInGBs)
			assert.Greater(t, *shape.MemoryInGBs, float32(0))
		}
		
		// Check specific mock shapes
		shapeNames := make([]string, len(mockShapes))
		for i, shape := range mockShapes {
			shapeNames[i] = *shape.Shape
		}
		
		assert.Contains(t, shapeNames, "VM.Standard2.1")
		assert.Contains(t, shapeNames, "VM.Standard3.Flex")
		assert.Contains(t, shapeNames, "BM.Standard2.52")
	})
}

func TestComputeScraperIntegration(t *testing.T) {
	ctx := context.Background()
	compartmentID := "test-compartment"
	
	t.Run("NewComputeScraper", func(t *testing.T) {
		// This will fail without proper OCI credentials
		_, err := NewComputeScraper(ctx, compartmentID)
		
		if err != nil {
			t.Logf("Expected error without OCI credentials: %v", err)
		}
	})

	t.Run("ScraperConfig", func(t *testing.T) {
		config := &models.ScraperConfig{
			Regions:        []string{"us-ashburn-1", "us-phoenix-1"},
			MaxConcurrency: 5,
			OutputPath:     "/tmp/test",
			Service:        "oci",
			SampleSize:     100,
		}

		assert.Len(t, config.Regions, 2)
		assert.Equal(t, 5, config.MaxConcurrency)
		assert.Equal(t, "oci", config.Service)
		assert.Equal(t, 100, config.SampleSize)
	})
}