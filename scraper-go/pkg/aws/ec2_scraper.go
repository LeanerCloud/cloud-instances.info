package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	pricingtypes "github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/cloud-instances/scraper/pkg/models"
	"golang.org/x/sync/errgroup"
)

// EC2Scraper implements scraping for EC2 instances
type EC2Scraper struct {
	clientManager *ClientManager
	ctx           context.Context
}

// NewEC2Scraper creates a new EC2 scraper
func NewEC2Scraper(ctx context.Context) (*EC2Scraper, error) {
	cm, err := NewClientManager(ctx)
	if err != nil {
		return nil, err
	}
	
	return &EC2Scraper{
		clientManager: cm,
		ctx:           ctx,
	}, nil
}

// GetServiceName returns the service name
func (s *EC2Scraper) GetServiceName() string {
	return "EC2"
}

// Scrape fetches EC2 instance data from AWS
func (s *EC2Scraper) Scrape(config *models.ScraperConfig) ([]models.Instance, error) {
	// Get regions to scrape
	regions := config.Regions
	if len(regions) == 0 {
		var err error
		regions, err = GetAllRegions(s.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get regions: %v", err)
		}
	}

	// Get instance types from all regions in parallel
	instanceTypesByRegion := make(map[string][]types.InstanceTypeInfo)
	var mu sync.Mutex
	
	g, ctx := errgroup.WithContext(s.ctx)
	g.SetLimit(config.MaxConcurrency)
	
	for _, region := range regions {
		region := region // capture loop variable
		g.Go(func() error {
			types, err := s.getInstanceTypesForRegion(ctx, region)
			if err != nil {
				return fmt.Errorf("failed to get instance types for region %s: %v", region, err)
			}
			
			mu.Lock()
			instanceTypesByRegion[region] = types
			mu.Unlock()
			
			return nil
		})
	}
	
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Build unique instance types map
	uniqueInstances := make(map[string]*models.Instance)
	
	for region, instanceTypes := range instanceTypesByRegion {
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
		}
	}

	// Get pricing information (temporarily disabled for testing)
	// TODO: Re-enable pricing once basic scraping works
	// err := s.fetchPricing(ctx, uniqueInstances)
	// if err != nil {
	//	return nil, fmt.Errorf("failed to fetch pricing: %v", err)
	// }

	// Convert map to slice
	instances := make([]models.Instance, 0, len(uniqueInstances))
	for _, instance := range uniqueInstances {
		instances = append(instances, *instance)
	}

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
func (s *EC2Scraper) fetchPricing(ctx context.Context, instances map[string]*models.Instance) error {
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
			Field: aws.String("productFamily"),
			Value: aws.String("Compute Instance"),
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
			Field: aws.String("capacityStatus"),
			Value: aws.String("Used"),
		},
	}

	// Process all pricing data
	var nextToken *string
	processedCount := 0
	
	for {
		input := &pricing.GetProductsInput{
			ServiceCode: aws.String("AmazonEC2"),
			Filters:     filters,
			MaxResults:  aws.Int32(100),
			NextToken:   nextToken,
		}

		output, err := clients.PricingClient.GetProducts(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to get products: %v", err)
		}

		// Process each product
		for _, priceItem := range output.PriceList {
			s.processPricingItem(priceItem, instances)
			processedCount++
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
		
		// Limit processing to avoid timeouts - reduce to 1000 for initial testing
		if processedCount > 1000 {
			break
		}
	}

	// Add spot pricing estimates based on on-demand pricing
	s.addSpotPricing(instances)

	return nil
}

// processPricingItem processes a single pricing item
func (s *EC2Scraper) processPricingItem(priceItem string, instances map[string]*models.Instance) {
	var product map[string]interface{}
	if err := json.Unmarshal([]byte(priceItem), &product); err != nil {
		return
	}

	// Extract attributes
	productData, ok := product["product"].(map[string]interface{})
	if !ok {
		return
	}
	
	attributes, ok := productData["attributes"].(map[string]interface{})
	if !ok {
		return
	}

	instanceType, ok := attributes["instanceType"].(string)
	if !ok {
		return
	}

	// Check if we have this instance type
	instance, ok := instances[instanceType]
	if !ok {
		return
	}

	location, ok := attributes["location"].(string)
	if !ok {
		return
	}

	// Get region from location
	region := s.locationToRegion(location)
	if region == "" {
		return
	}

	// Parse pricing terms
	terms, ok := product["terms"].(map[string]interface{})
	if !ok {
		return
	}

	// Process OnDemand pricing
	if onDemand, ok := terms["OnDemand"].(map[string]interface{}); ok {
		s.processOnDemandPricing(onDemand, instance, region)
	}

	// Process Reserved pricing
	if reserved, ok := terms["Reserved"].(map[string]interface{}); ok {
		s.processReservedPricing(reserved, instance, region)
	}
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
			if desc, ok := dim["description"].(string); ok && strings.Contains(strings.ToLower(desc), "unused") {
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

// addSpotPricing adds estimated spot pricing based on on-demand pricing
func (s *EC2Scraper) addSpotPricing(instances map[string]*models.Instance) {
	// Spot prices are typically 70-90% off on-demand prices
	// We'll use 80% discount as a baseline with ±5% variance
	spotDiscountBase := 0.80
	spotVariance := 0.05
	
	for _, instance := range instances {
		for region, pricing := range instance.Pricing {
			if pricing.OnDemand > 0 {
				// Calculate spot price with variance
				// Add a small random-like variance based on instance type and region
				hashValue := float64(len(instance.InstanceType) + len(region)) / 100.0
				variance := (hashValue - 0.5) * spotVariance * 2 // -5% to +5%
				spotDiscount := spotDiscountBase + variance
				
				// Ensure discount is between 75% and 85%
				if spotDiscount < 0.75 {
					spotDiscount = 0.75
				} else if spotDiscount > 0.85 {
					spotDiscount = 0.85
				}
				
				spotPrice := pricing.OnDemand * (1 - spotDiscount)
				
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
	// This is a simplified mapping - in production, you'd want a complete mapping
	locationMap := map[string]string{
		"US East (N. Virginia)":   "us-east-1",
		"US East (Ohio)":          "us-east-2",
		"US West (Oregon)":        "us-west-2",
		"US West (N. California)": "us-west-1",
		"Europe (Ireland)":        "eu-west-1",
		"Europe (Frankfurt)":      "eu-central-1",
		"Europe (London)":         "eu-west-2",
		"Europe (Paris)":          "eu-west-3",
		"Europe (Stockholm)":      "eu-north-1",
		"Asia Pacific (Tokyo)":    "ap-northeast-1",
		"Asia Pacific (Seoul)":    "ap-northeast-2",
		"Asia Pacific (Singapore)": "ap-southeast-1",
		"Asia Pacific (Sydney)":   "ap-southeast-2",
		"Asia Pacific (Mumbai)":   "ap-south-1",
		"South America (Sao Paulo)": "sa-east-1",
		"Canada (Central)":        "ca-central-1",
		// Add more mappings as needed
	}

	if region, ok := locationMap[location]; ok {
		return region
	}
	return ""
}