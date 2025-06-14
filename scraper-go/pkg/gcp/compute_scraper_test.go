package gcp

import (
	"context"
	"encoding/json"
	"testing"

	"google.golang.org/api/compute/v1"
	"github.com/cloud-instances/scraper/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeScraper(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"
	
	// Note: These tests don't require actual GCP credentials as we're testing utility functions
	scraper := &ComputeScraper{
		ctx:       ctx,
		projectID: projectID,
	}

	t.Run("GetServiceName", func(t *testing.T) {
		assert.Equal(t, "Compute Engine", scraper.GetServiceName())
	})

	t.Run("ConvertToInstance", func(t *testing.T) {
		machineType := &compute.MachineType{
			Name:      "n1-standard-1",
			GuestCpus: 1,
			MemoryMb:  3840, // 3.75 GB
		}
		
		instance := scraper.convertToInstance(machineType)
		
		assert.Equal(t, "n1-standard-1", instance.InstanceType)
		assert.Equal(t, "3.8", instance.Memory) // 3840 MB / 1024 = 3.75 GB, formatted to 3.8
		assert.Equal(t, "1", string(instance.VCPU))
		assert.Equal(t, "General purpose", instance.Family)
		assert.Equal(t, "N1 Standard", instance.PrettyName)
		assert.True(t, instance.VPCOnly)
		assert.True(t, instance.IPv6Support)
		assert.Equal(t, "current", instance.Generation)
	})

	t.Run("GetFamilyName", func(t *testing.T) {
		testCases := []struct {
			machineType string
			expected    string
		}{
			{"n1-standard-1", "General purpose"},
			{"n2-standard-2", "General purpose"},
			{"e2-medium", "General purpose"},
			{"c2-standard-4", "Compute optimized"},
			{"c3-standard-8", "Compute optimized"},
			{"m1-ultramem-40", "Memory optimized"},
			{"m2-megamem-416", "Memory optimized"},
			{"a2-highgpu-1g", "GPU instance"},
			{"g2-standard-4", "GPU instance"},
			{"unknown-type", "Unknown"},
		}
		
		for _, tc := range testCases {
			t.Run(tc.machineType, func(t *testing.T) {
				result := scraper.getFamilyName(tc.machineType)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("GeneratePrettyName", func(t *testing.T) {
		testCases := []struct {
			machineType string
			expected    string
		}{
			{"n1-standard-1", "N1 Standard"},
			{"e2-micro", "E2 Micro"},
			{"e2-small", "E2 Small"},
			{"e2-medium", "E2 Medium"},
			{"n1-highmem-2", "N1 High Memory"},
			{"n1-highcpu-16", "N1 High CPU"},
			{"m1-ultramem-40", "M1 Ultra Memory"},
			{"m1-megamem-96", "M1 Mega Memory"},
			{"c2-standard-4", "C2 Standard"},
			{"invalid", "INVALID"},
		}
		
		for _, tc := range testCases {
			t.Run(tc.machineType, func(t *testing.T) {
				result := scraper.generatePrettyName(tc.machineType)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("GetStorageInfo", func(t *testing.T) {
		testCases := []struct {
			name            string
			machineType     *compute.MachineType
			expectedSSD     bool
			expectedDevices int
		}{
			{
				name:            "Standard machine type with no local storage",
				machineType:     &compute.MachineType{Name: "n1-standard-1"},
				expectedSSD:     false,
				expectedDevices: 0,
			},
			{
				name:            "Machine type with local SSD",
				machineType:     &compute.MachineType{Name: "n1-standard-1-local-ssd"},
				expectedSSD:     true,
				expectedDevices: 1, // Has local SSD
			},
			{
				name:            "C2D machine type (typically has local storage)",
				machineType:     &compute.MachineType{Name: "c2d-standard-4"},
				expectedSSD:     true,
				expectedDevices: 1, // Has local SSD
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				storage := scraper.getStorageInfo(tc.machineType)
				
				var storageData map[string]interface{}
				err := json.Unmarshal(storage, &storageData)
				require.NoError(t, err)
				
				assert.Equal(t, tc.expectedSSD, storageData["ssd"])
				assert.Equal(t, float64(tc.expectedDevices), storageData["devices"])
			})
		}
	})

	t.Run("SampleInstances", func(t *testing.T) {
		instances := map[string]*models.Instance{
			"n1-standard-1": {InstanceType: "n1-standard-1"},
			"n1-standard-2": {InstanceType: "n1-standard-2"},
			"n1-standard-4": {InstanceType: "n1-standard-4"},
			"e2-medium":     {InstanceType: "e2-medium"},
			"c2-standard-4": {InstanceType: "c2-standard-4"},
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
			"n1-standard-1": {
				InstanceType: "n1-standard-1",
				VCPU:         "1",
				Memory:       "3.8",
				Pricing:      make(map[string]models.RegionPricing),
				Regions:      map[string]string{"us-central1": "us-central1"},
			},
		}
		
		regions := []string{"us-central1"}
		scraper.addPlaceholderPricing(instances, regions)
		
		instance := instances["n1-standard-1"]
		pricing, exists := instance.Pricing["us-central1"]
		require.True(t, exists)
		
		assert.Greater(t, pricing.OnDemand, 0.0)
		assert.Contains(t, pricing.Reserved, "1yr")
		assert.Contains(t, pricing.Reserved, "3yr")
		assert.Less(t, pricing.Reserved["1yr"], pricing.OnDemand)
		assert.Less(t, pricing.Reserved["3yr"], pricing.Reserved["1yr"])
	})
}

func TestComputeScraperIntegration(t *testing.T) {
	ctx := context.Background()
	projectID := "test-project"
	
	t.Run("NewComputeScraper", func(t *testing.T) {
		// This will fail without proper GCP credentials
		_, err := NewComputeScraper(ctx, projectID)
		
		if err != nil {
			t.Logf("Expected error without GCP credentials: %v", err)
		}
	})

	t.Run("ScraperConfig", func(t *testing.T) {
		config := &models.ScraperConfig{
			Regions:        []string{"us-central1", "us-east1"},
			MaxConcurrency: 5,
			OutputPath:     "/tmp/test",
			Service:        "gcp",
			SampleSize:     100,
		}

		assert.Len(t, config.Regions, 2)
		assert.Equal(t, 5, config.MaxConcurrency)
		assert.Equal(t, "gcp", config.Service)
		assert.Equal(t, 100, config.SampleSize)
	})
}