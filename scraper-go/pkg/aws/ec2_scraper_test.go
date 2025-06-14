package aws

import (
	"context"
	"encoding/json"
	"testing"
	
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/cloud-instances/scraper/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEC2Scraper(t *testing.T) {
	ctx := context.Background()
	scraper, err := NewEC2Scraper(ctx)
	require.NoError(t, err)
	
	t.Run("GetServiceName", func(t *testing.T) {
		assert.Equal(t, "EC2", scraper.GetServiceName())
	})
	
	t.Run("ConvertToInstance", func(t *testing.T) {
		// Create test instance type info
		instanceType := types.InstanceType("t2.micro")
		memoryMiB := int64(1024)
		vcpus := int32(1)
		clockSpeed := float64(3.3)
		manufacturer := "Intel"
		
		it := types.InstanceTypeInfo{
			InstanceType: instanceType,
			MemoryInfo: &types.MemoryInfo{
				SizeInMiB: &memoryMiB,
			},
			VCpuInfo: &types.VCpuInfo{
				DefaultVCpus: &vcpus,
			},
			ProcessorInfo: &types.ProcessorInfo{
				SupportedArchitectures: []types.ArchitectureType{types.ArchitectureTypeX8664},
				Manufacturer:          &manufacturer,
				SustainedClockSpeedInGhz: &clockSpeed,
			},
			NetworkInfo: &types.NetworkInfo{
				NetworkPerformance: stringPtr("Low to Moderate"),
				EnaSupport:        types.EnaSupportRequired,
				Ipv6Supported:     boolPtr(true),
			},
			PlacementGroupInfo: &types.PlacementGroupInfo{
				SupportedStrategies: []types.PlacementGroupStrategy{
					types.PlacementGroupStrategyCluster,
				},
			},
			SupportedVirtualizationTypes: []types.VirtualizationType{
				types.VirtualizationTypeHvm,
			},
			CurrentGeneration: boolPtr(true),
			InstanceStorageInfo: &types.InstanceStorageInfo{
				TotalSizeInGB: int64Ptr(0),
				NvmeSupport:   types.EphemeralNvmeSupportUnsupported,
			},
		}
		
		instance := scraper.convertToInstance(it)
		
		assert.Equal(t, "t2.micro", instance.InstanceType)
		assert.Equal(t, "1.0", instance.Memory)
		assert.Equal(t, json.Number("1"), instance.VCPU)
		assert.Equal(t, "Intel x86_64", instance.PhysicalProcessor)
		assert.Equal(t, "3.3 GHz", instance.ClockSpeed)
		assert.Equal(t, "Low to Moderate", instance.NetworkPerformance)
		assert.Equal(t, "General purpose", instance.Family)
		assert.Equal(t, "required", instance.EnhancedNetworking)
		assert.True(t, instance.IPv6Support)
		assert.True(t, instance.PlacementGroupSupport)
		assert.Equal(t, []string{"hvm"}, instance.LinuxVirtualization)
		assert.Equal(t, "current", instance.Generation)
		assert.Equal(t, "T2 Micro", instance.PrettyName)
		
		// Check storage
		var storage map[string]interface{}
		err := json.Unmarshal(instance.Storage, &storage)
		require.NoError(t, err)
		assert.Equal(t, float64(0), storage["size"])
		assert.Equal(t, float64(0), storage["devices"])
		assert.False(t, storage["ssd"].(bool))
	})
	
	t.Run("GeneratePrettyName", func(t *testing.T) {
		testCases := []struct {
			instanceType string
			expected     string
		}{
			{"t2.micro", "T2 Micro"},
			{"t2.small", "T2 Small"},
			{"t2.medium", "T2 Medium"},
			{"t2.large", "T2 Large"},
			{"t2.xlarge", "T2 Extra Large"},
			{"t2.2xlarge", "T2 Double Extra Large"},
			{"m5.metal", "M5 High Memory"},
			{"m5.24xlarge", "M5 24xlarge"},
			{"c5n.metal", "C5N High Memory"},
			{"p3.16xlarge", "P3 16xlarge"},
		}
		
		for _, tc := range testCases {
			instance := &models.Instance{InstanceType: tc.instanceType}
			result := scraper.generatePrettyName(instance)
			assert.Equal(t, tc.expected, result, "Pretty name for %s", tc.instanceType)
		}
	})
	
	t.Run("GetFamilyName", func(t *testing.T) {
		testCases := []struct {
			familyCode string
			expected   string
		}{
			{"t2", "General purpose"},
			{"t3", "General purpose"},
			{"m5", "General purpose"},
			{"c5", "Compute optimized"},
			{"r5", "Memory optimized"},
			{"i3", "Storage optimized"},
			{"p3", "GPU instance"},
			{"f1", "FPGA instance"},
			{"q9", "Unknown"},
		}
		
		for _, tc := range testCases {
			result := scraper.getFamilyName(tc.familyCode)
			assert.Equal(t, tc.expected, result, "Family name for %s", tc.familyCode)
		}
	})
	
	t.Run("LocationToRegion", func(t *testing.T) {
		testCases := []struct {
			location string
			expected string
		}{
			{"US East (N. Virginia)", "us-east-1"},
			{"US East (Ohio)", "us-east-2"},
			{"US West (Oregon)", "us-west-2"},
			{"Europe (Ireland)", "eu-west-1"},
			{"EU (Ireland)", "eu-west-1"},
			{"Asia Pacific (Tokyo)", "ap-northeast-1"},
			{"South America (Sao Paulo)", "sa-east-1"},
			{"Unknown Location", ""},
		}
		
		for _, tc := range testCases {
			result := scraper.locationToRegion(tc.location)
			assert.Equal(t, tc.expected, result, "Region for %s", tc.location)
		}
	})

	t.Run("IsValidPricingDescription", func(t *testing.T) {
		testCases := []struct {
			name     string
			dim      map[string]interface{}
			expected bool
		}{
			{
				name:     "Valid description",
				dim:      map[string]interface{}{"description": "On Demand Linux Instance"},
				expected: true,
			},
			{
				name:     "Unused resource description",
				dim:      map[string]interface{}{"description": "Unused Reservation"},
				expected: false,
			},
			{
				name:     "No description",
				dim:      map[string]interface{}{},
				expected: true,
			},
			{
				name:     "Non-string description",
				dim:      map[string]interface{}{"description": 123},
				expected: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := IsValidPricingDescription(tc.dim)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("CalculateEstimatedSpotPrice", func(t *testing.T) {
		testCases := []struct {
			name           string
			onDemandPrice  float64
			instanceType   string
			region         string
			expectedRange  [2]float64 // min, max
		}{
			{
				name:          "Valid on-demand price",
				onDemandPrice: 1.0,
				instanceType:  "t2.micro",
				region:        "us-east-1",
				expectedRange: [2]float64{0.15, 0.25}, // 75-85% discount range
			},
			{
				name:          "Zero on-demand price",
				onDemandPrice: 0.0,
				instanceType:  "t2.micro",
				region:        "us-east-1",
				expectedRange: [2]float64{0.0, 0.0},
			},
			{
				name:          "Negative on-demand price",
				onDemandPrice: -1.0,
				instanceType:  "t2.micro",
				region:        "us-east-1",
				expectedRange: [2]float64{0.0, 0.0},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := CalculateEstimatedSpotPrice(tc.onDemandPrice, tc.instanceType, tc.region)
				assert.GreaterOrEqual(t, result, tc.expectedRange[0])
				assert.LessOrEqual(t, result, tc.expectedRange[1])
			})
		}
	})
	
	t.Run("ProcessPricingItem", func(t *testing.T) {
		// Create test instances map
		instances := map[string]*models.Instance{
			"t2.micro": {
				InstanceType: "t2.micro",
				Pricing:      make(map[string]models.RegionPricing),
			},
		}
		
		// Create test pricing item
		priceItem := `{
			"product": {
				"attributes": {
					"instanceType": "t2.micro",
					"location": "US East (N. Virginia)",
					"operatingSystem": "Linux",
					"preInstalledSw": "NA",
					"tenancy": "Shared"
				}
			},
			"terms": {
				"OnDemand": {
					"ABC123": {
						"priceDimensions": {
							"ABC123.XYZ": {
								"unit": "Hrs",
								"pricePerUnit": {
									"USD": "0.0116"
								},
								"description": "On Demand Linux t2.micro Instance Hour"
							}
						}
					}
				}
			}
		}`
		
		instanceType := scraper.processPricingItem(priceItem, instances)
		
		assert.Equal(t, "t2.micro", instanceType)
		assert.Equal(t, 0.0116, instances["t2.micro"].Pricing["us-east-1"].OnDemand)
	})
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}