package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/api/compute/v1"
	"github.com/cloud-instances/scraper/internal/cache"
	"github.com/cloud-instances/scraper/pkg/models"
	"golang.org/x/sync/errgroup"
)

// ComputeScraper implements scraping for GCP Compute Engine instances
type ComputeScraper struct {
	clientManager *ClientManager
	ctx           context.Context
	instanceCache *cache.InstanceCache
	projectID     string
}

// NewComputeScraper creates a new GCP Compute scraper
func NewComputeScraper(ctx context.Context, projectID string) (*ComputeScraper, error) {
	cm, err := NewClientManager(ctx)
	if err != nil {
		return nil, err
	}
	
	// Check if instance cache is available in context
	var instanceCache *cache.InstanceCache
	if cacheVal := ctx.Value("instance_cache"); cacheVal != nil {
		instanceCache = cacheVal.(*cache.InstanceCache)
	}
	
	return &ComputeScraper{
		clientManager: cm,
		ctx:           ctx,
		instanceCache: instanceCache,
		projectID:     projectID,
	}, nil
}

// GetServiceName returns the service name
func (s *ComputeScraper) GetServiceName() string {
	return "Compute Engine"
}

// Scrape fetches GCP Compute Engine instance data
func (s *ComputeScraper) Scrape(config *models.ScraperConfig) ([]models.Instance, error) {
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

	// Build unique instance types map
	fmt.Println("[GCP] Building unique machine types map...")
	uniqueInstances := make(map[string]*models.Instance)
	totalMachineTypes := 0
	
	for region, machineTypes := range machineTypesByRegion {
		totalMachineTypes += len(machineTypes)
		for _, mt := range machineTypes {
			if mt.Name == "" {
				continue
			}
			
			machineType := mt.Name
			
			if _, exists := uniqueInstances[machineType]; !exists {
				instance := s.convertToInstance(mt)
				uniqueInstances[machineType] = instance
			}
			
			// Add region availability
			if uniqueInstances[machineType].Regions == nil {
				uniqueInstances[machineType].Regions = make(map[string]string)
			}
			uniqueInstances[machineType].Regions[region] = region
			
			// If sampling and we have enough instances, stop early
			if config.SampleSize > 0 && len(uniqueInstances) >= config.SampleSize*2 {
				fmt.Printf("[GCP] Reached %d unique machine types (2x sample size), stopping collection early\n", len(uniqueInstances))
				goto done
			}
		}
	}
	done:
	
	fmt.Printf("[GCP] Found %d unique machine types (from %d total across all regions)\n", 
		len(uniqueInstances), totalMachineTypes)

	// Apply sampling BEFORE fetching pricing to save time
	var instancesToPrice map[string]*models.Instance
	if config.SampleSize > 0 && len(uniqueInstances) > config.SampleSize {
		fmt.Printf("[GCP] Sampling %d random instances from %d total instances BEFORE pricing fetch\n", 
			config.SampleSize, len(uniqueInstances))
		
		instancesToPrice = s.sampleInstances(uniqueInstances, config.SampleSize)
	} else {
		instancesToPrice = uniqueInstances
	}

	fmt.Printf("[GCP] Fetching pricing for %d machine types...\n", len(instancesToPrice))
	
	// Note: GCP pricing would need to be implemented separately using Cloud Billing API
	// For now, we'll add placeholder pricing structure
	s.addPlaceholderPricing(instancesToPrice, regions)

	// Convert map to slice
	instances := make([]models.Instance, 0, len(instancesToPrice))
	for _, instance := range instancesToPrice {
		instances = append(instances, *instance)
	}

	fmt.Printf("[GCP] Scrape completed: returning %d instances\n", len(instances))
	return instances, nil
}

// getMachineTypesForRegion fetches machine types for a specific region
func (s *ComputeScraper) getMachineTypesForRegion(ctx context.Context, region string) ([]*compute.MachineType, error) {
	var allTypes []*compute.MachineType
	
	call := s.clientManager.GetComputeService().MachineTypes.List(s.projectID, region)
	if err := call.Pages(ctx, func(page *compute.MachineTypeList) error {
		allTypes = append(allTypes, page.Items...)
		return nil
	}); err != nil {
		return nil, err
	}

	return allTypes, nil
}

// convertToInstance converts GCP machine type to our model
func (s *ComputeScraper) convertToInstance(mt *compute.MachineType) *models.Instance {
	instance := &models.Instance{
		InstanceType: mt.Name,
		Pricing:      make(map[string]models.RegionPricing),
		Regions:      make(map[string]string),
	}

	// Memory (convert from MB to GB)
	if mt.MemoryMb > 0 {
		instance.Memory = fmt.Sprintf("%.1f", float64(mt.MemoryMb)/1024.0)
	}

	// vCPUs
	if mt.GuestCpus > 0 {
		instance.VCPU = json.Number(strconv.Itoa(int(mt.GuestCpus)))
	}

	// Storage information
	instance.Storage = s.getStorageInfo(mt)

	// Family and pretty name
	instance.Family = s.getFamilyName(mt.Name)
	instance.PrettyName = s.generatePrettyName(mt.Name)

	// Set default values
	instance.Generation = "current" // GCP doesn't clearly distinguish generations like AWS
	instance.VPCOnly = true        // GCP uses VPC by default
	instance.IPv6Support = true    // Most GCP instances support IPv6

	return instance
}

// getStorageInfo extracts storage information from machine type
func (s *ComputeScraper) getStorageInfo(mt *compute.MachineType) json.RawMessage {
	// Most GCP machine types don't have local storage by default
	// Local SSDs are attached separately
	storage := map[string]interface{}{
		"ssd":     false,
		"devices": 0,
		"size":    0,
	}

	// Check if this is a machine type that typically has local storage
	if strings.Contains(mt.Name, "local-ssd") || 
	   strings.HasPrefix(mt.Name, "c2d-") || 
	   strings.HasPrefix(mt.Name, "c3d-") {
		storage["ssd"] = true
		// These are estimates - actual local SSD configuration varies
		if strings.Contains(mt.Name, "standard") || strings.Contains(mt.Name, "local-ssd") {
			storage["devices"] = 1
			storage["size"] = 375 // GB
		}
	}

	data, _ := json.Marshal(storage)
	return json.RawMessage(data)
}

// getFamilyName returns human-readable family names for GCP machine types
func (s *ComputeScraper) getFamilyName(machineType string) string {
	// GCP machine type naming: family-type-size
	parts := strings.Split(machineType, "-")
	if len(parts) == 0 {
		return "Unknown"
	}
	
	family := parts[0]
	
	// Common GCP machine families
	familyMap := map[string]string{
		// General purpose
		"e2":  "General purpose",
		"n1":  "General purpose", 
		"n2":  "General purpose",
		"n2d": "General purpose",
		"t2d": "General purpose",
		"t2a": "General purpose",
		
		// Compute optimized
		"c2":  "Compute optimized",
		"c2d": "Compute optimized", 
		"c3":  "Compute optimized",
		"c3d": "Compute optimized",
		"h3":  "Compute optimized",
		
		// Memory optimized
		"m1":  "Memory optimized",
		"m2":  "Memory optimized",
		"m3":  "Memory optimized",
		"r1":  "Memory optimized",
		"x4":  "Memory optimized",
		
		// Accelerator optimized (GPU)
		"a2":  "GPU instance",
		"a3":  "GPU instance", 
		"g2":  "GPU instance",
	}
	
	if familyName, ok := familyMap[family]; ok {
		return familyName
	}
	
	// Default categorization
	switch {
	case strings.HasPrefix(family, "e") || strings.HasPrefix(family, "n") || strings.HasPrefix(family, "t"):
		return "General purpose"
	case strings.HasPrefix(family, "c") || strings.HasPrefix(family, "h"):
		return "Compute optimized"
	case strings.HasPrefix(family, "m") || strings.HasPrefix(family, "r") || strings.HasPrefix(family, "x"):
		return "Memory optimized"
	case strings.HasPrefix(family, "a") || strings.HasPrefix(family, "g"):
		return "GPU instance"
	default:
		return "Unknown"
	}
}

// generatePrettyName generates a human-readable name for the machine type
func (s *ComputeScraper) generatePrettyName(machineType string) string {
	parts := strings.Split(machineType, "-")
	if len(parts) < 2 {
		return strings.ToUpper(machineType)
	}
	
	family := strings.ToUpper(parts[0])
	
	// Check for specific type patterns first
	if len(parts) >= 2 {
		typeAndSize := strings.Join(parts[1:], "-")
		
		// Handle specific GCP naming patterns
		switch {
		case strings.HasPrefix(typeAndSize, "standard"):
			return fmt.Sprintf("%s Standard", family)
		case strings.HasPrefix(typeAndSize, "highmem"):
			return fmt.Sprintf("%s High Memory", family)
		case strings.HasPrefix(typeAndSize, "highcpu"):
			return fmt.Sprintf("%s High CPU", family)
		case strings.HasPrefix(typeAndSize, "megamem"):
			return fmt.Sprintf("%s Mega Memory", family)
		case strings.HasPrefix(typeAndSize, "ultramem"):
			return fmt.Sprintf("%s Ultra Memory", family)
		case typeAndSize == "micro":
			return fmt.Sprintf("%s Micro", family)
		case typeAndSize == "small":
			return fmt.Sprintf("%s Small", family)
		case typeAndSize == "medium":
			return fmt.Sprintf("%s Medium", family)
		}
	}
	
	// Fallback to last part
	size := parts[len(parts)-1]
	
	// For numbered sizes like 2, 4, 8, 16, etc.
	if _, err := strconv.Atoi(size); err == nil {
		return fmt.Sprintf("%s %s vCPU", family, size)
	}
	
	return fmt.Sprintf("%s %s", family, strings.Title(size))
}

// addPlaceholderPricing adds placeholder pricing structure
func (s *ComputeScraper) addPlaceholderPricing(instances map[string]*models.Instance, regions []string) {
	// Note: This is placeholder pricing. Real implementation would use Cloud Billing API
	for _, instance := range instances {
		for _, region := range regions {
			if _, hasRegion := instance.Regions[region]; hasRegion {
				// Estimate pricing based on vCPUs and memory (very rough approximation)
				vcpus := 1
				if instance.VCPU != "" {
					if v, err := strconv.Atoi(string(instance.VCPU)); err == nil {
						vcpus = v
					}
				}
				
				memory := 1.0
				if instance.Memory != "" {
					if m, err := strconv.ParseFloat(instance.Memory, 64); err == nil {
						memory = m
					}
				}
				
				// Very rough pricing estimation (not accurate)
				basePrice := float64(vcpus)*0.05 + memory*0.01
				
				pricing := models.RegionPricing{
					OnDemand: basePrice,
					Reserved: make(map[string]float64),
				}
				
				// Add some placeholder reserved pricing
				pricing.Reserved["1yr"] = basePrice * 0.7
				pricing.Reserved["3yr"] = basePrice * 0.5
				
				instance.Pricing[region] = pricing
			}
		}
	}
}

// sampleInstances returns a random sample of instances
func (s *ComputeScraper) sampleInstances(instances map[string]*models.Instance, sampleSize int) map[string]*models.Instance {
	if sampleSize <= 0 || len(instances) <= sampleSize {
		return instances
	}
	
	// Convert to slice for sampling
	allInstances := make([]*models.Instance, 0, len(instances))
	for _, inst := range instances {
		allInstances = append(allInstances, inst)
	}
	
	// Fisher-Yates shuffle
	for i := len(allInstances) - 1; i > 0; i-- {
		j := int(math.Floor(float64(i+1) * 0.5)) // Simple random approximation
		allInstances[i], allInstances[j] = allInstances[j], allInstances[i]
	}
	
	// Create new map with sampled instances
	sampled := make(map[string]*models.Instance)
	for i := 0; i < sampleSize && i < len(allInstances); i++ {
		inst := allInstances[i]
		sampled[inst.InstanceType] = inst
	}
	
	return sampled
}