package models

import (
	"encoding/json"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstance(t *testing.T) {
	t.Run("JSONMarshaling", func(t *testing.T) {
		instance := Instance{
			InstanceType:       "t2.micro",
			Memory:             "1.0",
			VCPU:               json.Number("1"),
			Family:             "General purpose",
			PhysicalProcessor:  "Intel Xeon",
			ClockSpeed:         "3.3 GHz",
			NetworkPerformance: "Low to Moderate",
			Storage:            json.RawMessage(`{"ssd":false,"devices":0,"size":0}`),
			EnhancedNetworking: "required",
			VPCOnly:            true,
			IPv6Support:        true,
			PlacementGroupSupport: false,
			LinuxVirtualization: []string{"hvm"},
			Generation:         "current",
			PrettyName:         "T2 Micro",
			Pricing: map[string]RegionPricing{
				"us-east-1": {
					OnDemand: 0.0116,
					Reserved: map[string]float64{
						"yrTerm1Standard.allUpfront": 0.0072,
						"spot": 0.0035,
					},
				},
			},
			Regions: map[string]string{
				"us-east-1": "us-east-1",
				"us-west-2": "us-west-2",
			},
		}
		
		// Marshal to JSON
		data, err := json.MarshalIndent(instance, "", "  ")
		require.NoError(t, err)
		
		// Unmarshal back
		var decoded Instance
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		
		// Verify fields
		assert.Equal(t, instance.InstanceType, decoded.InstanceType)
		assert.Equal(t, instance.Memory, decoded.Memory)
		assert.Equal(t, instance.VCPU, decoded.VCPU)
		assert.Equal(t, instance.Family, decoded.Family)
		assert.Equal(t, instance.NetworkPerformance, decoded.NetworkPerformance)
		assert.Equal(t, instance.Generation, decoded.Generation)
		assert.Equal(t, instance.IPv6Support, decoded.IPv6Support)
		assert.Equal(t, len(instance.Pricing), len(decoded.Pricing))
		assert.Equal(t, instance.Pricing["us-east-1"].OnDemand, decoded.Pricing["us-east-1"].OnDemand)
		assert.Equal(t, len(instance.Regions), len(decoded.Regions))
	})
	
	t.Run("StorageHandling", func(t *testing.T) {
		testCases := []struct {
			name     string
			storage  json.RawMessage
			expected map[string]interface{}
		}{
			{
				name:    "No storage",
				storage: json.RawMessage(`{"ssd":false,"devices":0,"size":0}`),
				expected: map[string]interface{}{
					"ssd":     false,
					"devices": float64(0),
					"size":    float64(0),
				},
			},
			{
				name:    "SSD storage",
				storage: json.RawMessage(`{"ssd":true,"devices":2,"size":1900}`),
				expected: map[string]interface{}{
					"ssd":     true,
					"devices": float64(2),
					"size":    float64(1900),
				},
			},
		}
		
		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				instance := Instance{
					InstanceType: "i3.2xlarge",
					Storage:      tc.storage,
				}
				
				var storage map[string]interface{}
				err := json.Unmarshal(instance.Storage, &storage)
				require.NoError(t, err)
				
				assert.Equal(t, tc.expected["ssd"], storage["ssd"])
				assert.Equal(t, tc.expected["devices"], storage["devices"])
				assert.Equal(t, tc.expected["size"], storage["size"])
			})
		}
	})
}

func TestRegionPricing(t *testing.T) {
	t.Run("PricingStructure", func(t *testing.T) {
		pricing := RegionPricing{
			OnDemand: 0.0116,
			Reserved: map[string]float64{
				"yrTerm1Standard.allUpfront":    0.0072,
				"yrTerm1Standard.partialUpfront": 0.0074,
				"yrTerm1Standard.noUpfront":     0.0078,
				"yrTerm3Standard.allUpfront":    0.0048,
				"spot":                          0.0035,
			},
		}
		
		// Verify OnDemand pricing
		assert.Equal(t, 0.0116, pricing.OnDemand)
		
		// Verify Reserved pricing
		assert.Equal(t, 5, len(pricing.Reserved))
		assert.Equal(t, 0.0072, pricing.Reserved["yrTerm1Standard.allUpfront"])
		assert.Equal(t, 0.0035, pricing.Reserved["spot"])
	})
}

func TestScraperConfig(t *testing.T) {
	t.Run("ConfigDefaults", func(t *testing.T) {
		config := ScraperConfig{
			Service: "ec2",
		}
		
		// Should have default values
		assert.Equal(t, "ec2", config.Service)
		assert.Empty(t, config.Regions)
		assert.Equal(t, 0, config.MaxConcurrency)
	})
	
	t.Run("ConfigWithValues", func(t *testing.T) {
		config := ScraperConfig{
			Service:        "rds",
			Regions:        []string{"us-east-1", "us-west-2"},
			MaxConcurrency: 10,
		}
		
		assert.Equal(t, "rds", config.Service)
		assert.Equal(t, 2, len(config.Regions))
		assert.Equal(t, 10, config.MaxConcurrency)
	})
}