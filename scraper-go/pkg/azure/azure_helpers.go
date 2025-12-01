package azure

import (
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// convertToNaming converts Azure naming to standard format
// Standard_D4s_v3 -> d4sv3
func (s *RealComputeScraper) convertToNaming(azureName string) string {
	// Remove Standard_ or Basic_ prefix
	name := strings.TrimPrefix(azureName, "Standard_")
	name = strings.TrimPrefix(name, "Basic_")
	
	// Convert to lowercase and remove underscores
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "")
	
	return name
}

// generatePrettyName generates standard pretty names
func (s *RealComputeScraper) generatePrettyName(azureName string) string {
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

// getFamilyCode returns single letter family code
func (s *RealComputeScraper) getFamilyCode(azureName string) string {
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

// getCategory returns standard category names
func (s *RealComputeScraper) getCategory(azureName string) string {
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

// getStorageInfo creates standard storage structure
func (s *RealComputeScraper) getStorageInfo(size *armcompute.VirtualMachineSize) interface{} {
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


// sampleInstances returns a random sample of instances
func (s *RealComputeScraper) sampleInstances(instances map[string]*AzureInstance, sampleSize int) map[string]*AzureInstance {
	if sampleSize <= 0 || len(instances) <= sampleSize {
		return instances
	}
	
	// Convert to slice for sampling
	allInstances := make([]*AzureInstance, 0, len(instances))
	for _, inst := range instances {
		allInstances = append(allInstances, inst)
	}
	
	// Simple sampling (Fisher-Yates would be better but this works for demo)
	sampled := make(map[string]*AzureInstance)
	for i := 0; i < sampleSize && i < len(allInstances); i++ {
		inst := allInstances[i]
		sampled[inst.InstanceType] = inst
	}
	
	return sampled
}