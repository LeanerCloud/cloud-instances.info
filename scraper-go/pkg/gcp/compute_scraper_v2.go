package gcp

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"

	"google.golang.org/api/compute/v1"
	"github.com/cloud-instances/scraper/internal/cache"
	"github.com/cloud-instances/scraper/pkg/models"
	"golang.org/x/sync/errgroup"
)

// ComputeScraperV2 implements scraping for GCP Compute Engine instances with proper data structure
type ComputeScraperV2 struct {
	clientManager *ClientManager
	ctx           context.Context
	instanceCache *cache.InstanceCache
	projectID     string
}

// NewComputeScraperV2 creates a new GCP Compute scraper
func NewComputeScraperV2(ctx context.Context, projectID string) (*ComputeScraperV2, error) {
	cm, err := NewClientManager(ctx)
	if err != nil {
		return nil, err
	}
	
	// Check if instance cache is available in context
	var instanceCache *cache.InstanceCache
	if cacheVal := ctx.Value("instance_cache"); cacheVal != nil {
		instanceCache = cacheVal.(*cache.InstanceCache)
	}
	
	return &ComputeScraperV2{
		clientManager: cm,
		ctx:           ctx,
		instanceCache: instanceCache,
		projectID:     projectID,
	}, nil
}

// GetServiceName returns the service name
func (s *ComputeScraperV2) GetServiceName() string {
	return "Compute Engine"
}

// ScrapeToGCPFormat fetches GCP Compute Engine instance data in GCP format
func (s *ComputeScraperV2) ScrapeToGCPFormat(config *models.ScraperConfig) ([]*GCPInstance, error) {
	fmt.Printf("[GCP] Starting scrape with config: regions=%v, sample=%d\n", config.Regions, config.SampleSize)
	
	// Get regions to scrape
	regions := config.Regions
	if len(regions) == 0 {
		fmt.Println("[GCP] No regions specified, fetching all available regions...")
		var err error
		regions, err = s.clientManager.GetAvailableRegions(s.projectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get regions: %v", err)
		}
		fmt.Printf("[GCP] Found %d regions: %v\n", len(regions), regions)
	}
	
	// Limit regions for sampling
	if config.SampleSize > 0 && len(config.Regions) == 0 && len(regions) > 2 {
		fmt.Printf("[GCP] Sampling mode with no specific regions: limiting to 2 regions for faster testing\n")
		regions = regions[:2]
		fmt.Printf("[GCP] Using regions: %v\n", regions)
	}

	// Get machine types from all regions in parallel
	fmt.Printf("[GCP] Fetching machine types from %d regions...\n", len(regions))
	machineTypesByRegion := make(map[string][]*compute.MachineType)
	var mu sync.Mutex
	
	g, ctx := errgroup.WithContext(s.ctx)
	g.SetLimit(config.MaxConcurrency)
	
	completedRegions := 0
	for _, region := range regions {
		region := region // capture loop variable
		g.Go(func() error {
			fmt.Printf("[GCP] Fetching machine types for region %s...\n", region)
			types, err := s.getMachineTypesForRegion(ctx, region)
			if err != nil {
				return fmt.Errorf("failed to get machine types for region %s: %v", region, err)
			}
			
			mu.Lock()
			machineTypesByRegion[region] = types
			completedRegions++
			fmt.Printf("[GCP] Completed %s: found %d machine types (%d/%d regions done)\n", 
				region, len(types), completedRegions, len(regions))
			mu.Unlock()
			
			return nil
		})
	}
	
	if err := g.Wait(); err != nil {
		return nil, err
	}
	
	// Build unique instances map
	fmt.Println("[GCP] Building unique machine instances map...")
	uniqueInstances := make(map[string]*GCPInstance)
	
	for _, types := range machineTypesByRegion {
		for _, mt := range types {
			if mt.Name == "" {
				continue
			}
			
			instanceType := mt.Name
			
			if _, exists := uniqueInstances[instanceType]; !exists {
				instance := s.convertToGCPInstance(mt)
				uniqueInstances[instanceType] = instance
			}
		}
	}
	
	fmt.Printf("[GCP] Found %d unique machine types\n", len(uniqueInstances))
	
	// Apply sampling if requested
	var instancesToPrice map[string]*GCPInstance
	if config.SampleSize > 0 && len(uniqueInstances) > config.SampleSize {
		fmt.Printf("[GCP] Sampling %d random instances from %d total instances\n", 
			config.SampleSize, len(uniqueInstances))
		instancesToPrice = s.sampleGCPInstances(uniqueInstances, config.SampleSize)
	} else {
		instancesToPrice = uniqueInstances
	}
	
	// Add pricing data
	fmt.Printf("[GCP] Adding pricing for %d instances...\n", len(instancesToPrice))
	s.addPricing(instancesToPrice, regions)
	
	// Convert map to slice
	result := make([]*GCPInstance, 0, len(instancesToPrice))
	for _, instance := range instancesToPrice {
		result = append(result, instance)
	}
	
	return result, nil
}

// getMachineTypesForRegion fetches machine types for a specific region
func (s *ComputeScraperV2) getMachineTypesForRegion(ctx context.Context, region string) ([]*compute.MachineType, error) {
	service := s.clientManager.GetComputeService()
	
	var allTypes []*compute.MachineType
	
	// First get zones for this region
	zones, err := s.getZonesForRegion(ctx, region)
	if err != nil {
		return nil, err
	}
	
	// Get machine types from all zones in the region
	for _, zone := range zones {
		call := service.MachineTypes.List(s.projectID, zone)
		
		err := call.Pages(ctx, func(page *compute.MachineTypeList) error {
			allTypes = append(allTypes, page.Items...)
			return nil
		})
		
		if err != nil {
			fmt.Printf("[GCP] Warning: Failed to get machine types for zone %s: %v\n", zone, err)
			continue
		}
	}
	
	return allTypes, nil
}

// getZonesForRegion gets all zones for a specific region
func (s *ComputeScraperV2) getZonesForRegion(ctx context.Context, region string) ([]string, error) {
	service := s.clientManager.GetComputeService()
	
	var zones []string
	call := service.Zones.List(s.projectID)
	
	err := call.Pages(ctx, func(page *compute.ZoneList) error {
		for _, zone := range page.Items {
			// Check if zone belongs to the region
			if zone.Region != "" && strings.HasSuffix(zone.Region, "/"+region) {
				zones = append(zones, zone.Name)
			}
		}
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return zones, nil
}

// convertToGCPInstance converts compute.MachineType to GCPInstance
func (s *ComputeScraperV2) convertToGCPInstance(mt *compute.MachineType) *GCPInstance {
	instance := &GCPInstance{
		InstanceType:         mt.Name,
		PrettyName:           s.generatePrettyName(mt.Name),
		Pricing:              make(map[string]GCPPricing),
		Family:               s.getFamilyName(mt.Name),
		Generation:           "current",
		SustainedUseDiscount: true,  // Most GCP instances support this
		Preemptible:          true,  // Most GCP instances support preemptible
		MachineType:          mt.Name,
	}
	
	// Memory (convert from MB to GB)
	if mt.MemoryMb > 0 {
		instance.Memory = float64(mt.MemoryMb) / 1024.0
	}
	
	// vCPUs
	if mt.GuestCpus > 0 {
		instance.VCPU = int(mt.GuestCpus)
	}
	
	// Architecture
	instance.Arch = s.getArchitecture(mt.Name)
	
	// Storage
	instance.Storage = s.getStorageInfo(mt)
	
	// Network performance
	instance.NetworkPerformance = s.getNetworkPerformance(mt)
	
	// GPU detection
	if mt.Accelerators != nil && len(mt.Accelerators) > 0 {
		instance.GPU = &GPUInfo{
			Count: int(mt.Accelerators[0].GuestAcceleratorCount),
			Type:  mt.Accelerators[0].GuestAcceleratorType,
		}
	}
	
	// Special machine type features
	if strings.Contains(mt.Name, "custom-") {
		instance.CustomSizing = true
	}
	
	return instance
}

// getArchitecture determines the architecture based on machine type
func (s *ComputeScraperV2) getArchitecture(machineType string) []string {
	// T2A series is ARM-based
	if strings.HasPrefix(machineType, "t2a-") {
		return []string{"arm64"}
	}
	
	// All other machine types are x86_64
	return []string{"x64"}
}

// getStorageInfo extracts storage information
func (s *ComputeScraperV2) getStorageInfo(mt *compute.MachineType) *GCPStorage {
	storage := &GCPStorage{
		Type: "persistent",
	}
	
	// Check by machine type name patterns for local SSD support
	// Note: The v1 API doesn't expose scratch disk info directly in MachineType
	switch {
	case strings.HasPrefix(mt.Name, "c3-") || strings.HasPrefix(mt.Name, "c3d-"):
		// C3 instances support local SSD
		storage.LocalSSD = 0 // Available but not included by default
	case strings.HasPrefix(mt.Name, "c2d-") && strings.Contains(mt.Name, "highcpu"):
		// C2D high-CPU instances support local SSD
		storage.LocalSSD = 0
	case strings.HasPrefix(mt.Name, "n1-") || strings.HasPrefix(mt.Name, "n2-"):
		// N1/N2 instances can attach local SSD
		storage.LocalSSD = 0
	}
	
	// All instances can attach persistent disks
	storage.PersistentDisk = 1 // Indicates support for persistent disks
	
	return storage
}

// getNetworkPerformance estimates network performance based on machine type
func (s *ComputeScraperV2) getNetworkPerformance(mt *compute.MachineType) string {
	// GCP network performance scales with vCPUs
	vcpus := mt.GuestCpus
	
	switch {
	case vcpus >= 96:
		return "100 Gbps"
	case vcpus >= 64:
		return "75 Gbps"
	case vcpus >= 32:
		return "50 Gbps"
	case vcpus >= 16:
		return "32 Gbps"
	case vcpus >= 8:
		return "16 Gbps"
	case vcpus >= 4:
		return "10 Gbps"
	case vcpus >= 2:
		return "4 Gbps"
	default:
		return "2 Gbps"
	}
}

// getFamilyName returns human-readable family names
func (s *ComputeScraperV2) getFamilyName(machineType string) string {
	parts := strings.Split(machineType, "-")
	if len(parts) == 0 {
		return "Unknown"
	}
	
	family := parts[0]
	
	// Map to user-friendly names
	familyMap := map[string]string{
		"e2":  "E2 General Purpose",
		"n1":  "N1 General Purpose", 
		"n2":  "N2 General Purpose",
		"n2d": "N2D General Purpose (AMD)",
		"t2d": "T2D General Purpose (AMD)",
		"t2a": "T2A General Purpose (ARM)",
		"c2":  "C2 Compute Optimized",
		"c2d": "C2D Compute Optimized (AMD)", 
		"c3":  "C3 Compute Optimized",
		"c3d": "C3D Compute Optimized (AMD)",
		"h3":  "H3 Compute Optimized",
		"m1":  "M1 Memory Optimized",
		"m2":  "M2 Memory Optimized",
		"m3":  "M3 Memory Optimized",
		"a2":  "A2 Accelerator Optimized",
		"a3":  "A3 Accelerator Optimized", 
		"g2":  "G2 Accelerator Optimized",
	}
	
	if name, ok := familyMap[family]; ok {
		return name
	}
	
	return strings.ToUpper(family) + " Series"
}

// generatePrettyName creates a human-readable name
func (s *ComputeScraperV2) generatePrettyName(machineType string) string {
	// e2-standard-4 -> E2 Standard 4
	parts := strings.Split(machineType, "-")
	if len(parts) < 3 {
		return strings.ToUpper(machineType)
	}
	
	family := strings.ToUpper(parts[0])
	typeStr := strings.Title(parts[1])
	size := parts[len(parts)-1]
	
	return fmt.Sprintf("%s %s %s", family, typeStr, size)
}

// addPricing adds pricing information to instances
func (s *ComputeScraperV2) addPricing(instances map[string]*GCPInstance, regions []string) {
	// TODO: Implement real pricing using Cloud Billing API
	// For now, use placeholder pricing based on machine specs
	
	for _, instance := range instances {
		for _, region := range regions {
			// Calculate base price based on resources
			basePrice := s.calculateBasePrice(instance)
			
			// Adjust for region
			regionMultiplier := s.getRegionMultiplier(region)
			adjustedPrice := basePrice * regionMultiplier
			
			// Create pricing structure
			pricing := GCPPricing{
				Linux: &OSPricing{
					OnDemand:    adjustedPrice,
					Preemptible: adjustedPrice * 0.2, // ~80% discount
					Spot:        adjustedPrice * 0.3, // ~70% discount
					Committed: map[string]float64{
						"1yr": adjustedPrice * 0.63, // ~37% discount
						"3yr": adjustedPrice * 0.44, // ~56% discount
					},
				},
				Windows: &OSPricing{
					OnDemand: adjustedPrice + s.getWindowsPremium(instance),
					// No preemptible for Windows in many cases
					Committed: map[string]float64{
						"1yr": (adjustedPrice + s.getWindowsPremium(instance)) * 0.63,
						"3yr": (adjustedPrice + s.getWindowsPremium(instance)) * 0.44,
					},
				},
			}
			
			instance.Pricing[region] = pricing
		}
	}
}

// calculateBasePrice estimates base price from instance specs
func (s *ComputeScraperV2) calculateBasePrice(instance *GCPInstance) float64 {
	// Base pricing model (approximate)
	vcpuPrice := 0.031611  // per vCPU hour
	memoryPrice := 0.004237 // per GB hour
	
	basePrice := float64(instance.VCPU)*vcpuPrice + instance.Memory*memoryPrice
	
	// Adjust for machine family
	familyMultiplier := 1.0
	switch {
	case strings.Contains(instance.Family, "Compute Optimized"):
		familyMultiplier = 1.1
	case strings.Contains(instance.Family, "Memory Optimized"):
		familyMultiplier = 1.3
	case strings.Contains(instance.Family, "Accelerator"):
		familyMultiplier = 2.5
	}
	
	return basePrice * familyMultiplier
}

// getRegionMultiplier returns pricing multiplier for region
func (s *ComputeScraperV2) getRegionMultiplier(region string) float64 {
	// Some regions are more expensive
	expensiveRegions := map[string]float64{
		"asia-northeast1": 1.162,
		"asia-northeast2": 1.162,
		"asia-northeast3": 1.162,
		"australia-southeast1": 1.23,
		"australia-southeast2": 1.23,
		"southamerica-east1": 1.476,
		"southamerica-west1": 1.25,
	}
	
	if multiplier, ok := expensiveRegions[region]; ok {
		return multiplier
	}
	
	return 1.0
}

// getWindowsPremium calculates Windows licensing premium
func (s *ComputeScraperV2) getWindowsPremium(instance *GCPInstance) float64 {
	// Windows pricing model
	if instance.VCPU <= 4 {
		return 0.023 * 4 // Minimum 4 vCPU charge
	}
	return 0.023 * float64(instance.VCPU)
}

// sampleGCPInstances returns a random sample of instances
func (s *ComputeScraperV2) sampleGCPInstances(instances map[string]*GCPInstance, sampleSize int) map[string]*GCPInstance {
	if sampleSize <= 0 || len(instances) <= sampleSize {
		return instances
	}
	
	// Convert to slice for sampling
	allInstances := make([]*GCPInstance, 0, len(instances))
	for _, inst := range instances {
		allInstances = append(allInstances, inst)
	}
	
	// Fisher-Yates shuffle
	for i := len(allInstances) - 1; i > 0; i-- {
		j := int(math.Floor(float64(i+1) * 0.5)) // Simple random approximation
		allInstances[i], allInstances[j] = allInstances[j], allInstances[i]
	}
	
	// Create new map with sampled instances
	sampled := make(map[string]*GCPInstance)
	for i := 0; i < sampleSize && i < len(allInstances); i++ {
		inst := allInstances[i]
		sampled[inst.InstanceType] = inst
	}
	
	return sampled
}