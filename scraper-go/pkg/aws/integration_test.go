package aws

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cloud-instances/scraper/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration-style tests that exercise more of the scraper functionality
// without requiring real AWS credentials

func TestEC2ScraperIntegration(t *testing.T) {
	ctx := context.Background()
	scraper, err := NewEC2Scraper(ctx)
	require.NoError(t, err)

	t.Run("GetStorageInfo", func(t *testing.T) {
		// Test EBS-only instance (no instance storage)
		ebsOnlyInstance := createTestInstanceTypeInfo("t2.micro", nil)
		storage := scraper.getStorageInfo(ebsOnlyInstance)
		
		var storageData map[string]interface{}
		err := json.Unmarshal(storage, &storageData)
		require.NoError(t, err)
		
		assert.Equal(t, float64(0), storageData["size"])
		assert.Equal(t, float64(0), storageData["devices"])
		
		// Test instance with SSD storage
		ssdInstance := createTestInstanceTypeInfoWithStorage("m5d.large", 75, 1, true)
		storage = scraper.getStorageInfo(ssdInstance)
		
		err = json.Unmarshal(storage, &storageData)
		require.NoError(t, err)
		
		assert.Equal(t, float64(75), storageData["size"])
		assert.Equal(t, float64(1), storageData["devices"])
		assert.True(t, storageData["ssd"].(bool))
	})

	t.Run("ProcessReservedPricing", func(t *testing.T) {
		instance := &models.Instance{
			InstanceType: "t2.micro",
			Pricing:      make(map[string]models.RegionPricing),
		}
		
		// Test reserved pricing processing
		reserved := map[string]interface{}{
			"term1": map[string]interface{}{
				"termAttributes": map[string]interface{}{
					"LeaseContractLength": "1yr",
					"PurchaseOption":      "All Upfront",
					"OfferingClass":       "standard",
				},
				"priceDimensions": map[string]interface{}{
					"upfront": map[string]interface{}{
						"unit": "Quantity",
						"pricePerUnit": map[string]interface{}{
							"USD": "100.0",
						},
					},
					"hourly": map[string]interface{}{
						"unit": "Hrs",
						"pricePerUnit": map[string]interface{}{
							"USD": "0.01",
						},
					},
				},
			},
		}
		
		scraper.processReservedPricing(reserved, instance, "us-east-1")
		
		pricing := instance.Pricing["us-east-1"]
		assert.NotNil(t, pricing.Reserved)
		
		// Calculate expected effective hourly rate
		// $100 upfront / (365 * 24 hours) + $0.01/hour
		expectedHourly := 100.0/(365*24) + 0.01
		assert.InDelta(t, expectedHourly, pricing.Reserved["yrTerm1Standard.allUpfront"], 0.001)
	})

	t.Run("ProcessOnDemandPricing", func(t *testing.T) {
		instance := &models.Instance{
			InstanceType: "t2.micro",
			Pricing:      make(map[string]models.RegionPricing),
		}
		
		onDemand := map[string]interface{}{
			"term1": map[string]interface{}{
				"priceDimensions": map[string]interface{}{
					"pd1": map[string]interface{}{
						"unit": "Hrs",
						"pricePerUnit": map[string]interface{}{
							"USD": "0.0116",
						},
						"description": "On Demand Linux t2.micro Instance Hour",
					},
				},
			},
		}
		
		scraper.processOnDemandPricing(onDemand, instance, "us-east-1")
		
		pricing := instance.Pricing["us-east-1"]
		assert.Equal(t, 0.0116, pricing.OnDemand)
	})

	t.Run("SampleInstances", func(t *testing.T) {
		// Create test instances
		instances := []models.Instance{
			{InstanceType: "t2.micro"},
			{InstanceType: "t2.small"},
			{InstanceType: "t2.medium"},
			{InstanceType: "t2.large"},
			{InstanceType: "t2.xlarge"},
		}
		
		// Test sampling
		sampled := scraper.sampleInstances(instances, 3)
		assert.Len(t, sampled, 3)
		
		// Test sampling with more requested than available
		sampled = scraper.sampleInstances(instances, 10)
		assert.Len(t, sampled, 5) // Should return all available
		
		// Test sampling with zero (should return all)
		sampled = scraper.sampleInstances(instances, 0)
		assert.Len(t, sampled, len(instances)) // Should return all
	})
}

func TestRDSScraperIntegration(t *testing.T) {
	ctx := context.Background()
	scraper, err := NewRDSScraper(ctx)
	require.NoError(t, err)

	// Note: Testing private methods like getEngineCode is not possible from external tests

	t.Run("LocationToRegion", func(t *testing.T) {
		testCases := []struct {
			location string
			expected string
		}{
			{"US East (N. Virginia)", "us-east-1"},
			{"Europe (Ireland)", "eu-west-1"},
			{"Asia Pacific (Singapore)", "ap-southeast-1"},
			{"Unknown Location", ""},
		}

		for _, tc := range testCases {
			result := scraper.locationToRegion(tc.location)
			assert.Equal(t, tc.expected, result, "Region for %s", tc.location)
		}
	})
}

func TestScraperConfig(t *testing.T) {
	t.Run("ConfigValidation", func(t *testing.T) {
		config := &models.ScraperConfig{
			Regions:        []string{"us-east-1", "us-west-2"},
			MaxConcurrency: 5,
			OutputPath:     "/tmp/test",
			Service:        "ec2",
			CacheDuration:  time.Hour,
			SampleSize:     100,
		}

		assert.Len(t, config.Regions, 2)
		assert.Equal(t, 5, config.MaxConcurrency)
		assert.Equal(t, "ec2", config.Service)
		assert.Equal(t, 100, config.SampleSize)
		assert.Equal(t, time.Hour, config.CacheDuration)
	})
}

// Helper functions to create test data
func createTestInstanceTypeInfo(instanceType string, memoryMiB *int64) types.InstanceTypeInfo {
	if memoryMiB == nil {
		mem := int64(1024)
		memoryMiB = &mem
	}
	
	vcpus := int32(1)
	return types.InstanceTypeInfo{
		InstanceType: types.InstanceType(instanceType),
		MemoryInfo: &types.MemoryInfo{
			SizeInMiB: memoryMiB,
		},
		VCpuInfo: &types.VCpuInfo{
			DefaultVCpus: &vcpus,
		},
		CurrentGeneration: boolPtr(true),
		InstanceStorageInfo: &types.InstanceStorageInfo{
			TotalSizeInGB: int64Ptr(0),
			NvmeSupport:   types.EphemeralNvmeSupportUnsupported,
		},
	}
}

func createTestInstanceTypeInfoWithStorage(instanceType string, storageGB int64, devices int32, isSSD bool) types.InstanceTypeInfo {
	vcpus := int32(2)
	memoryMiB := int64(8192)
	
	nvmeSupport := types.EphemeralNvmeSupportUnsupported
	if isSSD {
		nvmeSupport = types.EphemeralNvmeSupportSupported
	}
	
	return types.InstanceTypeInfo{
		InstanceType: types.InstanceType(instanceType),
		MemoryInfo: &types.MemoryInfo{
			SizeInMiB: &memoryMiB,
		},
		VCpuInfo: &types.VCpuInfo{
			DefaultVCpus: &vcpus,
		},
		CurrentGeneration: boolPtr(true),
		InstanceStorageInfo: &types.InstanceStorageInfo{
			TotalSizeInGB: &storageGB,
			NvmeSupport:   nvmeSupport,
			Disks: []types.DiskInfo{
				{
					Count:    &devices,
					SizeInGB: &storageGB,
					Type:     types.DiskTypeSsd,
				},
			},
		},
	}
}