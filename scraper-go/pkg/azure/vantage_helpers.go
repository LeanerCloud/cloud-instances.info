package azure

import (
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// convertToVantageNaming converts Azure naming to Vantage format
// Standard_D4s_v3 -> d4sv3
func (s *RealComputeScraper) convertToVantageNaming(azureName string) string {
	// Remove Standard_ or Basic_ prefix
	name := strings.TrimPrefix(azureName, "Standard_")
	name = strings.TrimPrefix(name, "Basic_")
	
	// Convert to lowercase and remove underscores
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "")
	
	return name
}

// generateVantagePrettyName generates Vantage-style pretty names
func (s *RealComputeScraper) generateVantagePrettyName(azureName string) string {
	// Standard_D4s_v3 -> D4s v3
	name := strings.TrimPrefix(azureName, "Standard_")
	name = strings.TrimPrefix(name, "Basic_")
	
	// Handle version numbers: D4s_v3 -> D4s v3
	re := regexp.MustCompile(`(.+)_v(\d+)$`)
	if match := re.FindStringSubmatch(name); len(match) == 3 {
		return strings.ToUpper(match[1]) + " v" + match[2]
	}
	
	return strings.ToUpper(name)
}

// getVantageFamilyCode returns single letter family code like Vantage
func (s *RealComputeScraper) getVantageFamilyCode(azureName string) string {
	name := strings.ToLower(azureName)
	
	switch {
	case strings.Contains(name, "_a"):
		return "a"
	case strings.Contains(name, "_b"):
		return "bs"
	case strings.Contains(name, "_d"):
		return "d"
	case strings.Contains(name, "_e"):
		return "e"
	case strings.Contains(name, "_f"):
		return "f"
	case strings.Contains(name, "_g"):
		return "g"
	case strings.Contains(name, "_h"):
		return "h"
	case strings.Contains(name, "_l"):
		return "l"
	case strings.Contains(name, "_m"):
		return "m"
	case strings.Contains(name, "_n"):
		return "n"
	default:
		return "general"
	}
}

// getVantageCategory returns Vantage category names
func (s *RealComputeScraper) getVantageCategory(azureName string) string {
	name := strings.ToLower(azureName)
	
	switch {
	case strings.Contains(name, "_a"):
		return "generalpurpose"
	case strings.Contains(name, "_b"):
		return "generalpurpose"
	case strings.Contains(name, "_d"):
		return "generalpurpose"
	case strings.Contains(name, "_e"):
		return "memoryoptimized"
	case strings.Contains(name, "_f"):
		return "computeoptimized"
	case strings.Contains(name, "_g"):
		return "memoryoptimized"
	case strings.Contains(name, "_h"):
		return "computeoptimized"
	case strings.Contains(name, "_l"):
		return "storageoptimized"
	case strings.Contains(name, "_m"):
		return "memoryoptimized"
	case strings.Contains(name, "_n"):
		return "acceleratedcompute"
	default:
		return "generalpurpose"
	}
}

// estimateACU estimates Azure Compute Units based on vCPUs
func (s *RealComputeScraper) estimateACU(size *armcompute.VirtualMachineSize) int {
	if size.NumberOfCores == nil {
		return 50
	}
	
	// Rough ACU estimation: 100 ACU per vCPU for most instances
	cores := int(*size.NumberOfCores)
	return cores * 100
}

// supportsPremiumIO determines if instance supports premium storage
func (s *RealComputeScraper) supportsPremiumIO(azureName string) bool {
	name := strings.ToLower(azureName)
	// s suffix typically indicates SSD/Premium IO support
	return strings.Contains(name, "s_") || strings.Contains(name, "s)")
}

// supportsAcceleratedNetworking determines if instance supports accelerated networking
func (s *RealComputeScraper) supportsAcceleratedNetworking(azureName string) bool {
	name := strings.ToLower(azureName)
	// Most modern Azure VMs support accelerated networking
	return !strings.Contains(name, "_a1") && !strings.Contains(name, "_a0")
}

// getVantageStorageInfo creates Vantage-compatible storage structure
func (s *RealComputeScraper) getVantageStorageInfo(size *armcompute.VirtualMachineSize) interface{} {
	storage := map[string]interface{}{
		"devices":          nil,
		"max_write_disks":  nil,
		"nvme_ssd":        false,
		"size":            0,
	}

	// Resource disk information
	if size.ResourceDiskSizeInMB != nil && *size.ResourceDiskSizeInMB > 0 {
		sizeMB := int(*size.ResourceDiskSizeInMB)
		storage["size"] = sizeMB
		
		// Estimate if it's NVMe SSD based on naming
		if size.Name != nil {
			name := strings.ToLower(*size.Name)
			if strings.Contains(name, "s_") || strings.Contains(name, "s)") {
				storage["nvme_ssd"] = true
			}
		}
	}

	return storage
}

// addVantageStylePricing adds Vantage-compatible pricing structure
func (s *RealComputeScraper) addVantageStylePricing(instances map[string]*VantageAzureInstance, regions []string) {
	for _, instance := range instances {
		for _, region := range regions {
			// Estimate pricing based on vCPUs and memory
			vcpus := 1.0
			if instance.VCPU != nil {
				if v, ok := instance.VCPU.(int); ok {
					vcpus = float64(v)
				}
			}
			
			memory := 1.0
			if instance.Memory > 0 {
				memory = instance.Memory
			}
			
			// Base pricing estimation
			basePrice := vcpus*0.05 + memory*0.01
			
			// Create Vantage-style pricing structure
			linuxPricing := VantageAzureOSPricing{
				OnDemand: basePrice,
				Reserved: make(map[string]float64),
			}
			linuxPricing.Reserved["1yr"] = basePrice * 0.65
			linuxPricing.Reserved["3yr"] = basePrice * 0.45
			
			windowsPricing := VantageAzureOSPricing{
				OnDemand:      basePrice * 1.5, // Windows is typically more expensive
				HybridBenefit: basePrice,        // Hybrid benefit matches Linux pricing
				Reserved:      make(map[string]float64),
			}
			windowsPricing.Reserved["1yr"] = basePrice * 0.65 * 1.5
			windowsPricing.Reserved["3yr"] = basePrice * 0.45 * 1.5
			
			pricing := VantageAzurePricing{
				Linux:   linuxPricing,
				Windows: windowsPricing,
			}
			
			instance.Pricing[region] = pricing
		}
	}
}

// sampleVantageInstances returns a random sample of Vantage instances
func (s *RealComputeScraper) sampleVantageInstances(instances map[string]*VantageAzureInstance, sampleSize int) map[string]*VantageAzureInstance {
	if sampleSize <= 0 || len(instances) <= sampleSize {
		return instances
	}
	
	// Convert to slice for sampling
	allInstances := make([]*VantageAzureInstance, 0, len(instances))
	for _, inst := range instances {
		allInstances = append(allInstances, inst)
	}
	
	// Simple sampling (Fisher-Yates would be better but this works for demo)
	sampled := make(map[string]*VantageAzureInstance)
	for i := 0; i < sampleSize && i < len(allInstances); i++ {
		inst := allInstances[i]
		sampled[inst.InstanceType] = inst
	}
	
	return sampled
}