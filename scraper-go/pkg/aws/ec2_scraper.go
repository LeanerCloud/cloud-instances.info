package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	pricingtypes "github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/cloud-instances/scraper/internal/cache"
	"github.com/cloud-instances/scraper/pkg/models"
	"golang.org/x/sync/errgroup"
)

// EC2Scraper implements scraping for EC2 instances
type EC2Scraper struct {
	clientManager  *ClientManager
	ctx            context.Context
	instanceCache  *cache.InstanceCache
}

// NewEC2Scraper creates a new EC2 scraper
func NewEC2Scraper(ctx context.Context) (*EC2Scraper, error) {
	cm, err := NewClientManager(ctx)
	if err != nil {
		return nil, err
	}
	
	// Check if instance cache is available in context
	var instanceCache *cache.InstanceCache
	if cacheVal := ctx.Value("instance_cache"); cacheVal != nil {
		instanceCache = cacheVal.(*cache.InstanceCache)
	}
	
	return &EC2Scraper{
		clientManager: cm,
		ctx:           ctx,
		instanceCache: instanceCache,
	}, nil
}

// GetServiceName returns the service name
func (s *EC2Scraper) GetServiceName() string {
	return "EC2"
}

// Scrape fetches EC2 instance data from AWS
func (s *EC2Scraper) Scrape(config *models.ScraperConfig) ([]models.Instance, error) {
	fmt.Printf("[EC2] Starting scrape with config: regions=%v, sample=%d\n", config.Regions, config.SampleSize)
	
	// Get regions to scrape
	regions := config.Regions
	if len(regions) == 0 {
		fmt.Println("[EC2] No regions specified, fetching all available regions...")
		var err error
		regions, err = GetAllRegions(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get regions: %v", err)
		}
		fmt.Printf("[EC2] Found %d regions: %v\n", len(regions), regions)
	}
	
	// If sampling AND no specific regions provided, limit to 1 region for fastest testing
	if config.SampleSize > 0 && len(config.Regions) == 0 && len(regions) > 1 {
		fmt.Printf("[EC2] Sampling mode with no specific regions: limiting to 1 region for fastest testing\n")
		// Prefer us-east-1 as it has the most instance types
		selectedRegion := "us-east-1"
		found := false
		for _, region := range regions {
			if region == selectedRegion {
				found = true
				break
			}
		}
		if !found {
			selectedRegion = regions[0]
		}
		regions = []string{selectedRegion}
		fmt.Printf("[EC2] Using region: %s\n", selectedRegion)
	}

	// Get instance types from all regions in parallel
	fmt.Printf("[EC2] Fetching instance types from %d regions...\n", len(regions))
	instanceTypesByRegion := make(map[string][]types.InstanceTypeInfo)
	var mu sync.Mutex
	
	g, ctx := errgroup.WithContext(s.ctx)
	g.SetLimit(config.MaxConcurrency)
	
	completedRegions := 0
	for _, region := range regions {
		region := region // capture loop variable
		g.Go(func() error {
			fmt.Printf("[EC2] Fetching instance types for region %s...\n", region)
			types, err := s.getInstanceTypesForRegion(ctx, region)
			if err != nil {
				return fmt.Errorf("failed to get instance types for region %s: %v", region, err)
			}
			
			mu.Lock()
			instanceTypesByRegion[region] = types
			completedRegions++
			fmt.Printf("[EC2] Completed %s: found %d instance types (%d/%d regions done)\n", 
				region, len(types), completedRegions, len(regions))
			mu.Unlock()
			
			return nil
		})
	}
	
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Build unique instance types map
	fmt.Println("[EC2] Building unique instance types map...")
	uniqueInstances := make(map[string]*models.Instance)
	totalInstanceTypes := 0
	
	// If sampling, stop early once we have enough instances
	for region, instanceTypes := range instanceTypesByRegion {
		totalInstanceTypes += len(instanceTypes)
		for _, it := range instanceTypes {
			if it.InstanceType == "" {
				continue
			}
			
			instanceType := string(it.InstanceType)
			
			if _, exists := uniqueInstances[instanceType]; !exists {
				instance := s.convertToInstance(it)
				uniqueInstances[instanceType] = instance
			}
			
			// Add region availability
			if uniqueInstances[instanceType].Regions == nil {
				uniqueInstances[instanceType].Regions = make(map[string]string)
			}
			uniqueInstances[instanceType].Regions[region] = region
			
			// If sampling and we have enough instances, stop early
			if config.SampleSize > 0 && len(uniqueInstances) >= config.SampleSize*2 {
				fmt.Printf("[EC2] Reached %d unique instances (2x sample size), stopping collection early\n", len(uniqueInstances))
				goto done
			}
		}
	}
	done:
	
	fmt.Printf("[EC2] Found %d unique instance types (from %d total across all regions)\n", 
		len(uniqueInstances), totalInstanceTypes)

	// Apply sampling BEFORE fetching pricing to save time
	var instancesToPrice map[string]*models.Instance
	if config.SampleSize > 0 && len(uniqueInstances) > config.SampleSize {
		fmt.Printf("[EC2] Sampling %d random instances from %d total instances BEFORE pricing fetch\n", 
			config.SampleSize, len(uniqueInstances))
		
		// Convert to slice for sampling
		allInstances := make([]*models.Instance, 0, len(uniqueInstances))
		for _, inst := range uniqueInstances {
			allInstances = append(allInstances, inst)
		}
		
		// Sample
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for i := len(allInstances) - 1; i > 0; i-- {
			j := r.Intn(i + 1)
			allInstances[i], allInstances[j] = allInstances[j], allInstances[i]
		}
		
		// Create new map with sampled instances
		instancesToPrice = make(map[string]*models.Instance)
		for i := 0; i < config.SampleSize; i++ {
			inst := allInstances[i]
			instancesToPrice[inst.InstanceType] = inst
		}
	} else {
		instancesToPrice = uniqueInstances
	}

	fmt.Printf("[EC2] Fetching pricing for %d instances...\n", len(instancesToPrice))
	
	// Get pricing information - use the original context, not the errgroup context
	err := s.fetchPricing(s.ctx, instancesToPrice, config, regions)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pricing: %v", err)
	}

	// Convert map to slice
	instances := make([]models.Instance, 0, len(instancesToPrice))
	for _, instance := range instancesToPrice {
		instances = append(instances, *instance)
	}

	fmt.Printf("[EC2] Scrape completed: returning %d instances\n", len(instances))
	return instances, nil
}

// getInstanceTypesForRegion fetches instance types for a specific region
func (s *EC2Scraper) getInstanceTypesForRegion(ctx context.Context, region string) ([]types.InstanceTypeInfo, error) {
	clients, err := s.clientManager.GetClients(region)
	if err != nil {
		return nil, err
	}

	var allTypes []types.InstanceTypeInfo
	var nextToken *string

	for {
		input := &ec2.DescribeInstanceTypesInput{
			MaxResults: aws.Int32(100),
		}
		if nextToken != nil {
			input.NextToken = nextToken
		}

		output, err := clients.EC2Client.DescribeInstanceTypes(ctx, input)
		if err != nil {
			return nil, err
		}

		allTypes = append(allTypes, output.InstanceTypes...)

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return allTypes, nil
}

// convertToInstance converts AWS instance type to our model
func (s *EC2Scraper) convertToInstance(it types.InstanceTypeInfo) *models.Instance {
	instance := &models.Instance{
		InstanceType: string(it.InstanceType),
		Pricing:      make(map[string]models.RegionPricing),
		Regions:      make(map[string]string),
	}

	// Memory
	if it.MemoryInfo != nil && it.MemoryInfo.SizeInMiB != nil {
		instance.Memory = fmt.Sprintf("%.1f", float64(*it.MemoryInfo.SizeInMiB)/1024.0)
	}

	// vCPUs
	if it.VCpuInfo != nil && it.VCpuInfo.DefaultVCpus != nil {
		instance.VCPU = json.Number(strconv.Itoa(int(*it.VCpuInfo.DefaultVCpus)))
	}

	// Physical Processor
	if it.ProcessorInfo != nil && len(it.ProcessorInfo.SupportedArchitectures) > 0 {
		arch := string(it.ProcessorInfo.SupportedArchitectures[0])
		if it.ProcessorInfo.Manufacturer != nil {
			instance.PhysicalProcessor = fmt.Sprintf("%s %s", *it.ProcessorInfo.Manufacturer, arch)
		} else {
			instance.PhysicalProcessor = arch
		}
		
		if it.ProcessorInfo.SustainedClockSpeedInGhz != nil {
			instance.ClockSpeed = fmt.Sprintf("%.1f GHz", *it.ProcessorInfo.SustainedClockSpeedInGhz)
		}
	}

	// Storage
	storageInfo := s.getStorageInfo(it)
	if storageInfo != nil {
		instance.Storage = storageInfo
	}

	// Network Performance
	if it.NetworkInfo != nil && it.NetworkInfo.NetworkPerformance != nil {
		instance.NetworkPerformance = *it.NetworkInfo.NetworkPerformance
	}

	// Family - use human-readable names to match Vantage
	parts := strings.Split(string(it.InstanceType), ".")
	if len(parts) > 0 {
		familyCode := parts[0]
		instance.Family = s.getFamilyName(familyCode)
	}

	// Enhanced Networking
	if it.NetworkInfo != nil {
		instance.EnhancedNetworking = string(it.NetworkInfo.EnaSupport)
	}

	// VPC Only
	instance.VPCOnly = true // EC2-Classic is deprecated

	// IPv6 Support
	if it.NetworkInfo != nil && it.NetworkInfo.Ipv6Supported != nil {
		instance.IPv6Support = *it.NetworkInfo.Ipv6Supported
	}

	// Placement Group Support
	if it.PlacementGroupInfo != nil && len(it.PlacementGroupInfo.SupportedStrategies) > 0 {
		instance.PlacementGroupSupport = true
	}

	// Virtualization types
	if len(it.SupportedVirtualizationTypes) > 0 {
		virtTypes := make([]string, len(it.SupportedVirtualizationTypes))
		for i, vt := range it.SupportedVirtualizationTypes {
			virtTypes[i] = string(vt)
		}
		instance.LinuxVirtualization = virtTypes
	}

	// GPU Info
	if it.GpuInfo != nil && len(it.GpuInfo.Gpus) > 0 {
		gpu := it.GpuInfo.Gpus[0]
		if gpu.Name != nil {
			instance.GPUModel = *gpu.Name
		}
		if gpu.MemoryInfo != nil && gpu.MemoryInfo.SizeInMiB != nil {
			instance.GPUMemory = fmt.Sprintf("%d", *gpu.MemoryInfo.SizeInMiB)
		}
		if gpu.Count != nil {
			instance.GPUCount = json.Number(strconv.Itoa(int(*gpu.Count)))
		}
	}

	// Generation
	if it.CurrentGeneration != nil && *it.CurrentGeneration {
		instance.Generation = "current"
	} else {
		instance.Generation = "previous"
	}

	// Pretty name
	instance.PrettyName = s.generatePrettyName(instance)

	return instance
}

// getStorageInfo extracts storage information
func (s *EC2Scraper) getStorageInfo(it types.InstanceTypeInfo) json.RawMessage {
	if it.InstanceStorageInfo == nil || it.InstanceStorageInfo.TotalSizeInGB == nil {
		return json.RawMessage(`{"devices":0,"size":0}`)
	}

	storage := map[string]interface{}{
		"ssd":     it.InstanceStorageInfo.NvmeSupport != types.EphemeralNvmeSupportUnsupported,
		"devices": 0,
		"size":    *it.InstanceStorageInfo.TotalSizeInGB,
	}

	if it.InstanceStorageInfo.Disks != nil && len(it.InstanceStorageInfo.Disks) > 0 {
		totalDevices := 0
		for _, disk := range it.InstanceStorageInfo.Disks {
			if disk.Count != nil {
				totalDevices += int(*disk.Count)
			}
		}
		storage["devices"] = totalDevices
	}

	data, _ := json.Marshal(storage)
	return json.RawMessage(data)
}

// generatePrettyName generates a human-readable name for the instance to match Vantage format
func (s *EC2Scraper) generatePrettyName(instance *models.Instance) string {
	parts := strings.Split(instance.InstanceType, ".")
	if len(parts) != 2 {
		return instance.InstanceType
	}

	family := strings.ToUpper(parts[0])
	size := parts[1]

	// Handle size naming to match Vantage format
	switch size {
	case "nano":
		return fmt.Sprintf("%s Nano", family)
	case "micro":
		return fmt.Sprintf("%s Micro", family)
	case "small":
		return fmt.Sprintf("%s Small", family)
	case "medium":
		return fmt.Sprintf("%s Medium", family)
	case "large":
		return fmt.Sprintf("%s Large", family)
	case "xlarge":
		return fmt.Sprintf("%s Extra Large", family)
	case "2xlarge":
		return fmt.Sprintf("%s Double Extra Large", family)
	case "3xlarge":
		return fmt.Sprintf("%s Triple Extra Large", family)
	case "4xlarge":
		return fmt.Sprintf("%s Quadruple Extra Large", family)
	case "6xlarge":
		return fmt.Sprintf("%s 6xlarge", family)
	case "8xlarge":
		return fmt.Sprintf("%s Eight Extra Large", family)
	case "9xlarge":
		return fmt.Sprintf("%s 9xlarge", family)
	case "10xlarge":
		return fmt.Sprintf("%s 10xlarge", family)
	case "12xlarge":
		return fmt.Sprintf("%s 12xlarge", family)
	case "16xlarge":
		return fmt.Sprintf("%s 16xlarge", family)
	case "18xlarge":
		return fmt.Sprintf("%s 18xlarge", family)
	case "24xlarge":
		return fmt.Sprintf("%s 24xlarge", family)
	case "32xlarge":
		return fmt.Sprintf("%s 32xlarge", family)
	case "48xlarge":
		return fmt.Sprintf("%s 48xlarge", family)
	case "56xlarge":
		return fmt.Sprintf("%s 56xlarge", family)
	case "96xlarge":
		return fmt.Sprintf("%s 96xlarge", family)
	case "112xlarge":
		return fmt.Sprintf("%s 112xlarge", family)
	case "metal":
		return fmt.Sprintf("%s High Memory", family)
	case "metal-16xl":
		return fmt.Sprintf("%s Metal 16xl", family)
	case "metal-24xl":
		return fmt.Sprintf("%s Metal 24xl", family)
	case "metal-32xl":
		return fmt.Sprintf("%s Metal 32xl", family)
	case "metal-48xl":
		return fmt.Sprintf("%s Metal 48xl", family)
	default:
		// For any other size format
		return fmt.Sprintf("%s %s", family, size)
	}
}

// fetchPricing fetches pricing information for instances
func (s *EC2Scraper) fetchPricing(ctx context.Context, instances map[string]*models.Instance, config *models.ScraperConfig, enabledRegions []string) error {
	// Check cache first if available
	cachedCount := 0
	uncachedInstances := make(map[string]*models.Instance)
	
	if s.instanceCache != nil {
		fmt.Println("Checking cache for pricing data...")
		for instanceType, instance := range instances {
			pricing, hit, err := s.instanceCache.GetPricing("ec2", instanceType)
			if err != nil {
				fmt.Printf("  Cache error for %s: %v\n", instanceType, err)
				uncachedInstances[instanceType] = instance
				continue
			}
			
			if hit && len(pricing) > 0 {
				instance.Pricing = pricing
				cachedCount++
			} else {
				uncachedInstances[instanceType] = instance
			}
		}
		
		fmt.Printf("  Found %d/%d instances with cached pricing\n", cachedCount, len(instances))
		if len(uncachedInstances) > 0 && len(uncachedInstances) <= 10 {
			fmt.Print("  Uncached instances: ")
			for it := range uncachedInstances {
				fmt.Printf("%s ", it)
			}
			fmt.Println()
		}
		
		// If all instances have cached pricing, we're done
		if len(uncachedInstances) == 0 {
			fmt.Println("All pricing data retrieved from cache!")
			return nil
		}
	} else {
		uncachedInstances = instances
	}
	
	// Use us-east-1 for pricing API
	clients, err := s.clientManager.GetClients("us-east-1")
	if err != nil {
		return err
	}

	// Get all EC2 products in one go
	filters := []pricingtypes.Filter{
		{
			Type:  pricingtypes.FilterTypeTermMatch,
			Field: aws.String("ServiceCode"),
			Value: aws.String("AmazonEC2"),
		},
		{
			Type:  pricingtypes.FilterTypeTermMatch,
			Field: aws.String("operatingSystem"),
			Value: aws.String("Linux"),
		},
		{
			Type:  pricingtypes.FilterTypeTermMatch,
			Field: aws.String("preInstalledSw"),
			Value: aws.String("NA"),
		},
		{
			Type:  pricingtypes.FilterTypeTermMatch,
			Field: aws.String("tenancy"),
			Value: aws.String("Shared"),
		},
		{
			Type:  pricingtypes.FilterTypeTermMatch,
			Field: aws.String("capacitystatus"),
			Value: aws.String("Used"),
		},
	}

	// Process all pricing data
	var nextToken *string
	processedCount := 0
	foundPricing := make(map[string]bool)
	lastReportTime := time.Now()
	startTime := time.Now()
	
	if len(uncachedInstances) == 0 {
		fmt.Println("No uncached instances to fetch pricing for")
		return nil
	}
	
	fmt.Printf("Fetching pricing data for %d EC2 instance types (uncached)...\n", len(uncachedInstances))
	fmt.Println("Note: This may take several minutes. The pricing API has no timeout.")
	
	// Track consecutive errors for circuit breaking
	consecutiveErrors := 0
	maxConsecutiveErrors := 10
	
	for {
		input := &pricing.GetProductsInput{
			ServiceCode: aws.String("AmazonEC2"),
			Filters:     filters,
			MaxResults:  aws.Int32(50), // Reduced to avoid throttling
			NextToken:   nextToken,
		}

		// Retry with exponential backoff for throttling
		var output *pricing.GetProductsOutput
		var err error
		for attempt := 0; attempt < 10; attempt++ { // Increased retry attempts
			output, err = clients.PricingClient.GetProducts(ctx, input)
			if err == nil {
				consecutiveErrors = 0 // Reset error counter on success
				break
			}
			
			// Check if it's a throttling error
			if strings.Contains(err.Error(), "ThrottlingException") || strings.Contains(err.Error(), "429") || 
			   strings.Contains(err.Error(), "ServiceUnavailable") || strings.Contains(err.Error(), "RequestTimeout") {
				waitTime := time.Duration(1<<uint(attempt)) * time.Second // 1s, 2s, 4s, 8s, 16s, etc.
				if waitTime > 60*time.Second {
					waitTime = 60 * time.Second // Cap at 60 seconds
				}
				fmt.Printf("  API error, waiting %v before retry (attempt %d/10): %v\n", waitTime, attempt+1, err)
				time.Sleep(waitTime)
				continue
			}
			
			// Non-throttling error - still retry a few times
			consecutiveErrors++
			if consecutiveErrors >= maxConsecutiveErrors {
				return fmt.Errorf("too many consecutive errors (%d), aborting: %v", consecutiveErrors, err)
			}
			
			fmt.Printf("  Error (consecutive: %d/%d): %v\n", consecutiveErrors, maxConsecutiveErrors, err)
			time.Sleep(5 * time.Second)
		}
		
		if err != nil {
			return fmt.Errorf("failed to get products after retries: %v", err)
		}

		// Process each product
		for _, priceItem := range output.PriceList {
			instanceType := s.processPricingItem(priceItem, uncachedInstances)
			if instanceType != "" {
				foundPricing[instanceType] = true
			}
			processedCount++
		}

		// Report progress every 5 seconds
		if time.Since(lastReportTime) > 5*time.Second {
			elapsed := time.Since(startTime)
			fmt.Printf("  Progress: Processed %d products, found pricing for %d/%d uncached instances (%.1f%%) - Elapsed: %v\n", 
				processedCount, len(foundPricing), len(uncachedInstances), 
				float64(len(foundPricing))/float64(len(uncachedInstances))*100,
				elapsed.Round(time.Second))
			lastReportTime = time.Now()
			
			// If we've processed many products and still haven't found all instances, 
			// they might be new/rare types not in pricing API yet
			if processedCount > 10000 && len(foundPricing) < len(uncachedInstances) {
				fmt.Println("  Note: Some instance types may not be in the pricing API yet")
				if len(uncachedInstances) <= 10 {
					fmt.Print("  Still looking for: ")
					for it := range uncachedInstances {
						if !foundPricing[it] {
							fmt.Printf("%s ", it)
						}
					}
					fmt.Println()
				}
			}
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
		
		// If we've found all instances we're looking for, stop early
		if len(foundPricing) == len(uncachedInstances) {
			fmt.Println("  Found pricing for all requested instances, stopping early")
			break
		}
		
		// Safety limit: stop after processing too many products when sampling
		if config.SampleSize > 0 && processedCount > 20000 {
			fmt.Printf("  Reached safety limit of 20000 products in sampling mode. Found %d/%d instances\n", 
				len(foundPricing), len(uncachedInstances))
			break
		}
		
		// Small delay to avoid throttling
		time.Sleep(100 * time.Millisecond)
	}
	
	elapsed := time.Since(startTime)
	fmt.Printf("Completed pricing fetch: found pricing for %d/%d uncached instances after processing %d products in %v\n", 
		len(foundPricing), len(uncachedInstances), processedCount, elapsed.Round(time.Second))
	
	// Save pricing to cache
	if s.instanceCache != nil && len(foundPricing) > 0 {
		fmt.Println("Saving pricing data to cache...")
		for instanceType := range foundPricing {
			if instance, ok := uncachedInstances[instanceType]; ok && len(instance.Pricing) > 0 {
				if err := s.instanceCache.SetPricing("ec2", instanceType, instance.Pricing); err != nil {
					fmt.Printf("  Failed to cache pricing for %s: %v\n", instanceType, err)
				}
			}
		}
	}
	
	// For instances without pricing, mark them as having empty pricing so we don't try again
	if len(foundPricing) < len(uncachedInstances) {
		fmt.Printf("Note: %d instances had no pricing data in the API\n", len(uncachedInstances)-len(foundPricing))
		for instanceType, instance := range uncachedInstances {
			if !foundPricing[instanceType] {
				// Set empty pricing to avoid re-fetching
				instance.Pricing = make(map[string]models.RegionPricing)
				if s.instanceCache != nil {
					// Cache the empty pricing with a shorter TTL (1 hour)
					// Note: This is a simplified approach - in production you might want a separate cache TTL
					s.instanceCache.SetPricing("ec2", instanceType, instance.Pricing)
				}
			}
		}
	}
	
	// Merge pricing from uncached instances back to main instances map
	for instanceType, uncachedInstance := range uncachedInstances {
		if instance, ok := instances[instanceType]; ok {
			instance.Pricing = uncachedInstance.Pricing
		}
	}

	// Only fetch spot pricing if we have instances with pricing data
	hasInstancesWithPricing := false
	for _, inst := range instances {
		if len(inst.Pricing) > 0 {
			hasInstancesWithPricing = true
			break
		}
	}
	
	if hasInstancesWithPricing {
		// Fetch real spot pricing in parallel across regions
		fmt.Println("Fetching spot pricing data...")
		err = s.fetchSpotPricing(s.ctx, instances, enabledRegions)
		if err != nil {
			fmt.Printf("Warning: Failed to fetch spot pricing: %v\n", err)
		}
	} else {
		fmt.Println("Skipping spot pricing fetch - no instances have on-demand pricing")
	}

	return nil
}

// processPricingItem processes a single pricing item and returns the instance type if found
func (s *EC2Scraper) processPricingItem(priceItem string, instances map[string]*models.Instance) string {
	var product map[string]interface{}
	if err := json.Unmarshal([]byte(priceItem), &product); err != nil {
		return ""
	}

	// Extract attributes
	productData, ok := product["product"].(map[string]interface{})
	if !ok {
		return ""
	}
	
	attributes, ok := productData["attributes"].(map[string]interface{})
	if !ok {
		return ""
	}

	instanceType, ok := attributes["instanceType"].(string)
	if !ok {
		return ""
	}

	// Check if we have this instance type
	instance, ok := instances[instanceType]
	if !ok {
		return ""
	}

	location, ok := attributes["location"].(string)
	if !ok {
		return ""
	}

	// Get region from location
	region := s.locationToRegion(location)
	if region == "" {
		return ""
	}

	// Parse pricing terms
	terms, ok := product["terms"].(map[string]interface{})
	if !ok {
		return ""
	}

	// Process OnDemand pricing
	if onDemand, ok := terms["OnDemand"].(map[string]interface{}); ok {
		s.processOnDemandPricing(onDemand, instance, region)
	}

	// Process Reserved pricing
	if reserved, ok := terms["Reserved"].(map[string]interface{}); ok {
		s.processReservedPricing(reserved, instance, region)
	}
	
	return instanceType
}

// processOnDemandPricing processes on-demand pricing
func (s *EC2Scraper) processOnDemandPricing(onDemand map[string]interface{}, instance *models.Instance, region string) {
	for _, termData := range onDemand {
		term, ok := termData.(map[string]interface{})
		if !ok {
			continue
		}

		priceDimensions, ok := term["priceDimensions"].(map[string]interface{})
		if !ok {
			continue
		}

		for _, dimData := range priceDimensions {
			dim, ok := dimData.(map[string]interface{})
			if !ok {
				continue
			}

			// Skip non-instance charges
			if !IsValidPricingDescription(dim) {
				continue
			}

			pricePerUnit, ok := dim["pricePerUnit"].(map[string]interface{})
			if !ok {
				continue
			}

			usdPrice, ok := pricePerUnit["USD"].(string)
			if !ok {
				continue
			}

			price, err := strconv.ParseFloat(usdPrice, 64)
			if err != nil {
				continue
			}

			// Update pricing
			pricing := instance.Pricing[region]
			pricing.OnDemand = price
			instance.Pricing[region] = pricing
			break // Only need the first valid price
		}
	}
}

// processReservedPricing processes reserved pricing
func (s *EC2Scraper) processReservedPricing(reserved map[string]interface{}, instance *models.Instance, region string) {
	for _, termData := range reserved {
		term, ok := termData.(map[string]interface{})
		if !ok {
			continue
		}

		termAttributes, ok := term["termAttributes"].(map[string]interface{})
		if !ok {
			continue
		}

		leaseContractLength, _ := termAttributes["LeaseContractLength"].(string)
		purchaseOption, _ := termAttributes["PurchaseOption"].(string)
		offeringClass, _ := termAttributes["OfferingClass"].(string)

		if offeringClass != "standard" {
			continue // Skip convertible instances for now
		}

		priceDimensions, ok := term["priceDimensions"].(map[string]interface{})
		if !ok {
			continue
		}

		var upfrontCost, hourlyCost float64
		for _, dimData := range priceDimensions {
			dim, ok := dimData.(map[string]interface{})
			if !ok {
				continue
			}

			unit, _ := dim["unit"].(string)
			pricePerUnit, ok := dim["pricePerUnit"].(map[string]interface{})
			if !ok {
				continue
			}

			if usdPrice, ok := pricePerUnit["USD"].(string); ok {
				if price, err := strconv.ParseFloat(usdPrice, 64); err == nil {
					if unit == "Quantity" {
						upfrontCost = price
					} else if unit == "Hrs" {
						hourlyCost = price
					}
				}
			}
		}

		// Calculate effective hourly rate
		var termHours float64
		if leaseContractLength == "1yr" {
			termHours = 365 * 24
		} else if leaseContractLength == "3yr" {
			termHours = 365 * 24 * 3
		}

		effectiveHourly := hourlyCost
		if termHours > 0 {
			effectiveHourly += upfrontCost / termHours
		}

		// Map to pricing keys
		var pricingKey string
		if leaseContractLength == "1yr" && purchaseOption == "All Upfront" {
			pricingKey = "yrTerm1Standard.allUpfront"
		} else if leaseContractLength == "1yr" && purchaseOption == "Partial Upfront" {
			pricingKey = "yrTerm1Standard.partialUpfront"
		} else if leaseContractLength == "1yr" && purchaseOption == "No Upfront" {
			pricingKey = "yrTerm1Standard.noUpfront"
		} else if leaseContractLength == "3yr" && purchaseOption == "All Upfront" {
			pricingKey = "yrTerm3Standard.allUpfront"
		} else if leaseContractLength == "3yr" && purchaseOption == "Partial Upfront" {
			pricingKey = "yrTerm3Standard.partialUpfront"
		} else if leaseContractLength == "3yr" && purchaseOption == "No Upfront" {
			pricingKey = "yrTerm3Standard.noUpfront"
		}

		if pricingKey != "" {
			pricing := instance.Pricing[region]
			if pricing.Reserved == nil {
				pricing.Reserved = make(map[string]float64)
			}
			pricing.Reserved[pricingKey] = effectiveHourly
			instance.Pricing[region] = pricing
		}
	}
}

// fetchSpotPricing fetches real spot pricing data in parallel across regions
func (s *EC2Scraper) fetchSpotPricing(ctx context.Context, instances map[string]*models.Instance, enabledRegions []string) error {
	// Only fetch spot pricing for regions that are both enabled and have pricing data
	regionsToFetch := make(map[string]bool)
	enabledRegionSet := make(map[string]bool)
	
	// Convert enabled regions to a set for fast lookup
	for _, region := range enabledRegions {
		enabledRegionSet[region] = true
	}
	
	// Find regions that have pricing data AND are in our enabled regions
	for _, instance := range instances {
		for region := range instance.Pricing {
			if enabledRegionSet[region] {
				regionsToFetch[region] = true
			}
		}
	}

	if len(regionsToFetch) == 0 {
		fmt.Println("No enabled regions with pricing data found, skipping spot pricing fetch")
		return nil
	}

	fmt.Printf("Fetching spot prices for %d enabled regions in parallel...\n", len(regionsToFetch))
	startTime := time.Now()
	
	// Fetch spot prices in parallel for each region
	g := &errgroup.Group{}
	g.SetLimit(20) // Increased concurrency for spot pricing
	
	var mu sync.Mutex
	spotPrices := make(map[string]map[string]float64) // region -> instanceType -> price
	successCount := 0
	failureCount := 0
	
	for region := range regionsToFetch {
		region := region // capture loop variable
		g.Go(func() error {
			// Retry logic for spot pricing
			var prices map[string]float64
			var err error
			
			for attempt := 0; attempt < 3; attempt++ {
				prices, err = s.fetchSpotPricesForRegion(ctx, region)
				if err == nil {
					break
				}
				
				if strings.Contains(err.Error(), "ThrottlingException") || strings.Contains(err.Error(), "429") {
					waitTime := time.Duration(1<<uint(attempt)) * time.Second
					fmt.Printf("  Throttled for spot prices in %s, waiting %v before retry (attempt %d/3)\n", region, waitTime, attempt+1)
					time.Sleep(waitTime)
					continue
				}
				
				// For other errors, break after first attempt
				break
			}
			
			mu.Lock()
			if err != nil {
				failureCount++
				fmt.Printf("  ✗ Failed to fetch spot prices for %s: %v\n", region, err)
			} else {
				successCount++
				spotPrices[region] = prices
				fmt.Printf("  ✓ Fetched spot prices for %s (%d types)\n", region, len(prices))
			}
			mu.Unlock()
			
			return nil // Don't fail the whole operation
		})
	}
	
	if err := g.Wait(); err != nil {
		return err
	}
	
	elapsed := time.Since(startTime)
	fmt.Printf("Spot pricing fetch completed in %v: %d successful, %d failed regions\n", 
		elapsed.Round(time.Second), successCount, failureCount)
	
	// Apply spot prices to instances
	appliedCount := 0
	missingCount := 0
	for _, instance := range instances {
		for region := range instance.Pricing {
			if regionPrices, ok := spotPrices[region]; ok {
				if spotPrice, ok := regionPrices[instance.InstanceType]; ok {
					pricing := instance.Pricing[region]
					if pricing.Reserved == nil {
						pricing.Reserved = make(map[string]float64)
					}
					pricing.Reserved["spot"] = spotPrice
					instance.Pricing[region] = pricing
					appliedCount++
				} else {
					missingCount++
				}
			}
		}
	}
	
	fmt.Printf("Applied %d spot prices to instances (%d instance-region combinations had no spot price)\n", 
		appliedCount, missingCount)
	
	// Cache spot prices if available
	if s.instanceCache != nil && appliedCount > 0 {
		fmt.Println("Updating cache with spot pricing data...")
		for instanceType, instance := range instances {
			if len(instance.Pricing) > 0 {
				// Update cached pricing with spot prices
				if err := s.instanceCache.SetPricing("ec2", instanceType, instance.Pricing); err != nil {
					fmt.Printf("  Failed to update cache for %s: %v\n", instanceType, err)
				}
			}
		}
	}
	
	return nil
}

// fetchSpotPricesForRegion fetches current spot prices for a specific region
func (s *EC2Scraper) fetchSpotPricesForRegion(ctx context.Context, region string) (map[string]float64, error) {
	clients, err := s.clientManager.GetClients(region)
	if err != nil {
		return nil, err
	}
	
	prices := make(map[string]float64)
	var nextToken *string
	
	for {
		input := &ec2.DescribeSpotPriceHistoryInput{
			StartTime:  aws.Time(time.Now().Add(-1 * time.Hour)), // Last hour
			MaxResults: aws.Int32(1000),
			NextToken:  nextToken,
		}
		
		output, err := clients.EC2Client.DescribeSpotPriceHistory(ctx, input)
		if err != nil {
			return nil, err
		}
		
		// Get the most recent price for each instance type
		for _, spotPrice := range output.SpotPriceHistory {
			if spotPrice.InstanceType != "" && spotPrice.SpotPrice != nil {
				instanceType := string(spotPrice.InstanceType)
				if price, err := strconv.ParseFloat(*spotPrice.SpotPrice, 64); err == nil {
					// Keep the most recent price (first occurrence)
					if _, exists := prices[instanceType]; !exists {
						prices[instanceType] = price
					}
				}
			}
		}
		
		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}
	
	return prices, nil
}

// IsValidPricingDescription checks if a pricing dimension description represents actual instance usage
func IsValidPricingDescription(dim map[string]interface{}) bool {
	desc, ok := dim["description"].(string)
	if !ok {
		return true // If no description, assume it's valid
	}
	
	// Skip unused resource charges
	descLower := strings.ToLower(desc)
	return !strings.Contains(descLower, "unused")
}

// CalculateEstimatedSpotPrice calculates an estimated spot price based on on-demand pricing
func CalculateEstimatedSpotPrice(onDemandPrice float64, instanceType, region string) float64 {
	if onDemandPrice <= 0 {
		return 0
	}
	
	// Spot prices are typically 70-90% off on-demand prices
	// We'll use 80% discount as a baseline with ±5% variance
	spotDiscountBase := 0.80
	spotVariance := 0.05
	
	// Add a small random-like variance based on instance type and region
	hashValue := float64(len(instanceType) + len(region)) / 100.0
	variance := (hashValue - 0.5) * spotVariance * 2 // -5% to +5%
	spotDiscount := spotDiscountBase + variance
	
	// Ensure discount is between 75% and 85%
	if spotDiscount < 0.75 {
		spotDiscount = 0.75
	} else if spotDiscount > 0.85 {
		spotDiscount = 0.85
	}
	
	return onDemandPrice * (1 - spotDiscount)
}

// addSpotPricing adds estimated spot pricing based on on-demand pricing (fallback)
func (s *EC2Scraper) addSpotPricing(instances map[string]*models.Instance) {
	for _, instance := range instances {
		for region, pricing := range instance.Pricing {
			if pricing.OnDemand > 0 {
				spotPrice := CalculateEstimatedSpotPrice(pricing.OnDemand, instance.InstanceType, region)
				
				if pricing.Reserved == nil {
					pricing.Reserved = make(map[string]float64)
				}
				pricing.Reserved["spot"] = spotPrice
				instance.Pricing[region] = pricing
			}
		}
	}
}

// getFamilyName returns human-readable family names to match Vantage
func (s *EC2Scraper) getFamilyName(familyCode string) string {
	// Match Vantage's family naming
	familyMap := map[string]string{
		// General purpose
		"t2": "General purpose", "t3": "General purpose", "t3a": "General purpose", "t4g": "General purpose",
		"m3": "General purpose", "m4": "General purpose", "m5": "General purpose", "m5a": "General purpose",
		"m5ad": "General purpose", "m5d": "General purpose", "m5dn": "General purpose", "m5n": "General purpose",
		"m5zn": "General purpose", "m6a": "General purpose", "m6i": "General purpose", "m6id": "General purpose",
		"m6idn": "General purpose", "m6in": "General purpose", "m7a": "General purpose", "m7g": "General purpose",
		"m7gd": "General purpose", "m7i": "General purpose", "m7i-flex": "General purpose", "m8g": "General purpose",
		"a1": "General purpose", "mac1": "General purpose", "mac2": "General purpose", "mac2-m1": "General purpose",
		"mac2-m2": "General purpose", "mac2-m2pro": "General purpose",
		// Compute optimized
		"c3": "Compute optimized", "c4": "Compute optimized", "c5": "Compute optimized", "c5a": "Compute optimized",
		"c5ad": "Compute optimized", "c5d": "Compute optimized", "c5n": "Compute optimized", "c6a": "Compute optimized",
		"c6g": "Compute optimized", "c6gd": "Compute optimized", "c6gn": "Compute optimized", "c6i": "Compute optimized",
		"c6id": "Compute optimized", "c6in": "Compute optimized", "c7a": "Compute optimized", "c7g": "Compute optimized",
		"c7gd": "Compute optimized", "c7gn": "Compute optimized", "c7i": "Compute optimized", "c7i-flex": "Compute optimized",
		"c8g": "Compute optimized", "hpc6a": "Compute optimized", "hpc6id": "Compute optimized", "hpc7a": "Compute optimized",
		"hpc7g": "Compute optimized",
		// Memory optimized
		"r3": "Memory optimized", "r4": "Memory optimized", "r5": "Memory optimized", "r5a": "Memory optimized",
		"r5ad": "Memory optimized", "r5b": "Memory optimized", "r5d": "Memory optimized", "r5dn": "Memory optimized",
		"r5n": "Memory optimized", "r6a": "Memory optimized", "r6g": "Memory optimized", "r6gd": "Memory optimized",
		"r6i": "Memory optimized", "r6id": "Memory optimized", "r6idn": "Memory optimized", "r6in": "Memory optimized",
		"r7a": "Memory optimized", "r7g": "Memory optimized", "r7gd": "Memory optimized", "r7i": "Memory optimized",
		"r7iz": "Memory optimized", "r8g": "Memory optimized", "u-12tb1": "Memory optimized", "u-18tb1": "Memory optimized",
		"u-24tb1": "Memory optimized", "u-3tb1": "Memory optimized", "u-6tb1": "Memory optimized", "u-9tb1": "Memory optimized",
		"u7i-12tb": "Memory optimized", "u7in-16tb": "Memory optimized", "u7in-24tb": "Memory optimized",
		"u7in-32tb": "Memory optimized", "x1": "Memory optimized", "x1e": "Memory optimized", "x2gd": "Memory optimized",
		"x2idn": "Memory optimized", "x2iedn": "Memory optimized", "x2iezn": "Memory optimized", "x4g": "Memory optimized",
		"z1d": "Memory optimized",
		// Storage optimized
		"d2": "Storage optimized", "d3": "Storage optimized", "d3en": "Storage optimized", "h1": "Storage optimized",
		"i2": "Storage optimized", "i3": "Storage optimized", "i3en": "Storage optimized", "i4g": "Storage optimized",
		"i4i": "Storage optimized", "im4gn": "Storage optimized", "is4gen": "Storage optimized",
		// GPU instances
		"g3": "GPU instance", "g3s": "GPU instance", "g4ad": "GPU instance", "g4dn": "GPU instance",
		"g5": "GPU instance", "g5g": "GPU instance", "g6": "GPU instance", "g6e": "GPU instance",
		"gr6": "GPU instance", "inf1": "GPU instance", "inf2": "GPU instance", "p2": "GPU instance",
		"p3": "GPU instance", "p3dn": "GPU instance", "p4d": "GPU instance", "p4de": "GPU instance",
		"p5": "GPU instance", "p5e": "GPU instance", "p5en": "GPU instance", "trn1": "GPU instance",
		"trn1n": "GPU instance", "trn2": "GPU instance",
		// FPGA
		"f1": "FPGA instance", "vt1": "FPGA instance",
		// Legacy
		"cc2": "Compute optimized", "cr1": "Memory optimized", "hs1": "Storage optimized",
		"cg1": "GPU instance", "hi1": "Storage optimized", "m1": "General purpose", "m2": "Memory optimized",
		"c1": "Compute optimized", "t1": "General purpose",
	}
	
	if family, ok := familyMap[familyCode]; ok {
		return family
	}
	
	// Default based on prefix
	switch {
	case strings.HasPrefix(familyCode, "t") || strings.HasPrefix(familyCode, "m") || strings.HasPrefix(familyCode, "a"):
		return "General purpose"
	case strings.HasPrefix(familyCode, "c") || strings.HasPrefix(familyCode, "hpc"):
		return "Compute optimized"
	case strings.HasPrefix(familyCode, "r") || strings.HasPrefix(familyCode, "x") || strings.HasPrefix(familyCode, "u") || strings.HasPrefix(familyCode, "z"):
		return "Memory optimized"
	case strings.HasPrefix(familyCode, "d") || strings.HasPrefix(familyCode, "h") || strings.HasPrefix(familyCode, "i"):
		return "Storage optimized"
	case strings.HasPrefix(familyCode, "g") || strings.HasPrefix(familyCode, "p") || strings.HasPrefix(familyCode, "inf") || strings.HasPrefix(familyCode, "trn"):
		return "GPU instance"
	case strings.HasPrefix(familyCode, "f") || strings.HasPrefix(familyCode, "vt"):
		return "FPGA instance"
	default:
		return "Unknown"
	}
}

// locationToRegion converts AWS location names to region codes
func (s *EC2Scraper) locationToRegion(location string) string {
	// Comprehensive mapping of all AWS locations to regions
	locationMap := map[string]string{
		"US East (N. Virginia)":     "us-east-1",
		"US East (Ohio)":            "us-east-2",
		"US West (Oregon)":          "us-west-2",
		"US West (N. California)":   "us-west-1",
		"Europe (Ireland)":          "eu-west-1",
		"EU (Ireland)":              "eu-west-1",
		"Europe (Frankfurt)":        "eu-central-1",
		"EU (Frankfurt)":            "eu-central-1",
		"Europe (London)":           "eu-west-2",
		"EU (London)":               "eu-west-2",
		"Europe (Paris)":            "eu-west-3",
		"EU (Paris)":                "eu-west-3",
		"Europe (Stockholm)":        "eu-north-1",
		"EU (Stockholm)":            "eu-north-1",
		"Europe (Milan)":            "eu-south-1",
		"EU (Milan)":                "eu-south-1",
		"Europe (Spain)":            "eu-south-2",
		"EU (Spain)":                "eu-south-2",
		"Europe (Zurich)":           "eu-central-2",
		"EU (Zurich)":               "eu-central-2",
		"Asia Pacific (Tokyo)":      "ap-northeast-1",
		"Asia Pacific (Seoul)":      "ap-northeast-2",
		"Asia Pacific (Singapore)":  "ap-southeast-1",
		"Asia Pacific (Sydney)":     "ap-southeast-2",
		"Asia Pacific (Mumbai)":     "ap-south-1",
		"Asia Pacific (Hong Kong)":  "ap-east-1",
		"Asia Pacific (Jakarta)":    "ap-southeast-3",
		"Asia Pacific (Osaka)":      "ap-northeast-3",
		"Asia Pacific (Hyderabad)":  "ap-south-2",
		"Asia Pacific (Melbourne)":  "ap-southeast-4",
		"South America (Sao Paulo)": "sa-east-1",
		"Canada (Central)":          "ca-central-1",
		"Canada West (Calgary)":     "ca-west-1",
		"Middle East (Bahrain)":     "me-south-1",
		"Middle East (UAE)":         "me-central-1",
		"Africa (Cape Town)":        "af-south-1",
		"Israel (Tel Aviv)":         "il-central-1",
	}

	if region, ok := locationMap[location]; ok {
		return region
	}
	return ""
}

// sampleInstances returns a random sample of instances
func (s *EC2Scraper) sampleInstances(instances []models.Instance, sampleSize int) []models.Instance {
	if sampleSize <= 0 || len(instances) <= sampleSize {
		return instances
	}
	
	// Create a copy to avoid modifying the original
	shuffled := make([]models.Instance, len(instances))
	copy(shuffled, instances)
	
	// Fisher-Yates shuffle
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(shuffled) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	
	// Return first sampleSize elements
	return shuffled[:sampleSize]
}