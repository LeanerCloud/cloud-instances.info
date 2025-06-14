package azure

import (
	"context"
	"fmt"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/cloud-instances/scraper/internal/cache"
	"github.com/cloud-instances/scraper/pkg/models"
	"golang.org/x/sync/errgroup"
)

// RealComputeScraper implements scraping using real Azure APIs
type RealComputeScraper struct {
	clientManager *RealClientManager
	ctx           context.Context
	instanceCache *cache.InstanceCache
}

// NewRealComputeScraper creates a new Azure Compute scraper using real Azure APIs
func NewRealComputeScraper(ctx context.Context, subscriptionID string) (*RealComputeScraper, error) {
	cm, err := NewRealClientManager(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}
	
	// Check if instance cache is available in context
	var instanceCache *cache.InstanceCache
	if cacheVal := ctx.Value("instance_cache"); cacheVal != nil {
		instanceCache = cacheVal.(*cache.InstanceCache)
	}
	
	return &RealComputeScraper{
		clientManager: cm,
		ctx:           ctx,
		instanceCache: instanceCache,
	}, nil
}

// GetServiceName returns the service name
func (s *RealComputeScraper) GetServiceName() string {
	return "Compute"
}

// Scrape fetches Azure Virtual Machine instance data using real Azure APIs
func (s *RealComputeScraper) Scrape(config *models.ScraperConfig) ([]VantageAzureInstance, error) {
	fmt.Printf("[Azure] Starting real Azure API scrape with subscription: %s\n", s.clientManager.GetSubscriptionID())
	fmt.Printf("[Azure] Scrape config: regions=%v, sample=%d\n", config.Regions, config.SampleSize)
	
	// Get regions to scrape
	regions := config.Regions
	if len(regions) == 0 {
		fmt.Println("[Azure] No regions specified, using common regions...")
		regions = []string{"eastus", "westus2", "westeurope", "southeastasia"}
		fmt.Printf("[Azure] Using regions: %v\n", regions)
	}
	
	// Limit regions for sampling
	if config.SampleSize > 0 && len(config.Regions) == 0 && len(regions) > 2 {
		fmt.Printf("[Azure] Sampling mode: limiting to 2 regions for faster testing\n")
		regions = regions[:2]
		fmt.Printf("[Azure] Using regions: %v\n", regions)
	}

	// Get VM sizes from all regions in parallel
	fmt.Printf("[Azure] Fetching VM sizes from %d regions using Azure APIs...\n", len(regions))
	sizesByRegion := make(map[string][]*armcompute.VirtualMachineSize)
	var mu sync.Mutex
	
	g, _ := errgroup.WithContext(s.ctx)
	g.SetLimit(config.MaxConcurrency)
	
	completedRegions := 0
	for _, region := range regions {
		region := region // capture loop variable
		g.Go(func() error {
			fmt.Printf("[Azure] Fetching VM sizes for region %s...\n", region)
			sizes, err := s.clientManager.GetVMSizes(region)
			if err != nil {
				return fmt.Errorf("failed to get VM sizes for region %s: %v", region, err)
			}
			
			mu.Lock()
			sizesByRegion[region] = sizes
			completedRegions++
			fmt.Printf("[Azure] Completed %s: found %d VM sizes (%d/%d regions done)\n", 
				region, len(sizes), completedRegions, len(regions))
			mu.Unlock()
			
			return nil
		})
	}
	
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Build unique instance types map
	fmt.Println("[Azure] Building unique VM instances map...")
	uniqueInstances := make(map[string]*VantageAzureInstance)
	totalSizes := 0
	
	for region, sizes := range sizesByRegion {
		totalSizes += len(sizes)
		for _, size := range sizes {
			if size.Name == nil || *size.Name == "" {
				continue
			}
			
			vantageInstanceType := s.convertToVantageNaming(*size.Name)
			
			if _, exists := uniqueInstances[vantageInstanceType]; !exists {
				instance := s.convertRealVMSizeToInstance(size)
				uniqueInstances[vantageInstanceType] = instance
			}
			
			// Add region availability to AvailabilityZones (Vantage format)
			if uniqueInstances[vantageInstanceType].AvailabilityZones == nil {
				uniqueInstances[vantageInstanceType].AvailabilityZones = make(map[string]interface{})
			}
			uniqueInstances[vantageInstanceType].AvailabilityZones[region] = true
			
			// If sampling and we have enough instances, stop early
			if config.SampleSize > 0 && len(uniqueInstances) >= config.SampleSize*2 {
				fmt.Printf("[Azure] Reached %d unique VM types (2x sample size), stopping collection early\n", len(uniqueInstances))
				goto done
			}
		}
	}
	done:
	
	fmt.Printf("[Azure] Found %d unique VM types (from %d total across all regions)\n", 
		len(uniqueInstances), totalSizes)

	// Apply sampling BEFORE pricing fetch to save time
	var instancesToPrice map[string]*VantageAzureInstance
	if config.SampleSize > 0 && len(uniqueInstances) > config.SampleSize {
		fmt.Printf("[Azure] Sampling %d random instances from %d total instances\n", 
			config.SampleSize, len(uniqueInstances))
		
		instancesToPrice = s.sampleVantageInstances(uniqueInstances, config.SampleSize)
	} else {
		instancesToPrice = uniqueInstances
	}

	fmt.Printf("[Azure] Adding placeholder pricing for %d VM types...\n", len(instancesToPrice))
	
	// Add Vantage-style pricing structure
	s.addVantageStylePricing(instancesToPrice, regions)

	// Convert map to slice
	instances := make([]VantageAzureInstance, 0, len(instancesToPrice))
	for _, instance := range instancesToPrice {
		instances = append(instances, *instance)
	}

	fmt.Printf("[Azure] Real Azure API scrape completed: returning %d instances\n", len(instances))
	return instances, nil
}

// convertRealVMSizeToInstance converts Azure VM size from real API to Vantage-compatible format
func (s *RealComputeScraper) convertRealVMSizeToInstance(size *armcompute.VirtualMachineSize) *VantageAzureInstance {
	// Convert Standard_D4s_v3 to d4sv3 format (Vantage naming)
	vantageInstanceType := s.convertToVantageNaming(*size.Name)
	
	instance := &VantageAzureInstance{
		InstanceType:         vantageInstanceType,
		PrettyName:          s.generateVantagePrettyName(*size.Name),
		PrettyNameAzure:     *size.Name,
		Pricing:             make(map[string]VantageAzurePricing),
		AvailabilityZones:   make(map[string]interface{}),
	}

	// Memory (convert from MB to GB as float)
	if size.MemoryInMB != nil {
		instance.Memory = float64(*size.MemoryInMB) / 1024
	}

	// vCPUs
	if size.NumberOfCores != nil {
		instance.VCPU = int(*size.NumberOfCores)
		instance.VCPUsAvailable = int(*size.NumberOfCores)
		instance.VCPUsPerCore = 1 // Default for most Azure VMs
	}

	// Storage information from real Azure data
	instance.Storage = s.getVantageStorageInfo(size)

	// Family and category based on Azure naming
	instance.Family = s.getVantageFamilyCode(*size.Name)
	instance.Category = s.getVantageCategory(*size.Name)

	// Azure-specific fields
	instance.ACU = s.estimateACU(size)
	instance.GPU = "0" // Default, would need GPU API to determine
	instance.PremiumIO = s.supportsPremiumIO(*size.Name)
	instance.AcceleratedNetworking = s.supportsAcceleratedNetworking(*size.Name)
	instance.Arch = []string{"x64"} // Most Azure VMs are x64
	instance.HypervGenerations = "V1,V2" // Most support both
	
	// Resource disk information
	if size.ResourceDiskSizeInMB != nil {
		instance.UncachedDisk = int(*size.ResourceDiskSizeInMB / 1024) // Convert to GB
		instance.Size = int(*size.ResourceDiskSizeInMB / 1024)
	}

	// Default values
	instance.CachedDisk = 0
	instance.CapacitySupport = false
	instance.Confidential = false
	instance.Encryption = false
	instance.LowPriority = false
	instance.MemoryMaintenance = true
	instance.RDMA = false
	instance.TrustedLaunch = false
	instance.UltraSSD = false

	return instance
}

