package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

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
func (s *RealComputeScraper) Scrape(config *models.ScraperConfig) ([]AzureInstance, error) {
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
	uniqueInstances := make(map[string]*AzureInstance)
	totalSizes := 0
	
	for region, sizes := range sizesByRegion {
		totalSizes += len(sizes)
		for _, size := range sizes {
			if size.Name == nil || *size.Name == "" {
				continue
			}
			
			instanceType := s.convertToNaming(*size.Name)
			
			if _, exists := uniqueInstances[instanceType]; !exists {
				instance := s.convertRealVMSizeToInstance(size)
				uniqueInstances[instanceType] = instance
			}
			
			// Add region availability to AvailabilityZones
			// Using normalized region names to match Python scraper format
			if uniqueInstances[instanceType].AvailabilityZones == nil {
				uniqueInstances[instanceType].AvailabilityZones = make(map[string]interface{})
			}
			normalizedRegion := s.normalizeRegionName(region)
			uniqueInstances[instanceType].AvailabilityZones[normalizedRegion] = true
			
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
	var instancesToPrice map[string]*AzureInstance
	if config.SampleSize > 0 && len(uniqueInstances) > config.SampleSize {
		fmt.Printf("[Azure] Sampling %d random instances from %d total instances\n", 
			config.SampleSize, len(uniqueInstances))
		
		instancesToPrice = s.sampleInstances(uniqueInstances, config.SampleSize)
	} else {
		instancesToPrice = uniqueInstances
	}

	fmt.Printf("[Azure] Fetching pricing for %d VM types...\n", len(instancesToPrice))
	
	// Fetch pricing from Azure APIs
	err := s.fetchPricing(instancesToPrice, regions)
	if err != nil {
		fmt.Printf("[Azure] Warning: Failed to fetch pricing: %v. Skipping pricing data.\n", err)
	}

	// Convert map to slice
	instances := make([]AzureInstance, 0, len(instancesToPrice))
	for _, instance := range instancesToPrice {
		instances = append(instances, *instance)
	}

	fmt.Printf("[Azure] Azure API scrape completed: returning %d instances\n", len(instances))
	return instances, nil
}

// convertRealVMSizeToInstance converts Azure VM size from real API to standard format
func (s *RealComputeScraper) convertRealVMSizeToInstance(size *armcompute.VirtualMachineSize) *AzureInstance {
	// Convert Standard_D4s_v3 to d4sv3 format
	instanceType := s.convertToNaming(*size.Name)
	
	instance := &AzureInstance{
		InstanceType:         instanceType,
		PrettyName:          s.generatePrettyName(*size.Name),
		PrettyNameAzure:     *size.Name,
		Pricing:             make(map[string]AzurePricing),
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

	// Storage information from Azure data
	instance.Storage = s.getStorageInfo(size)

	// Family and category based on Azure naming
	instance.Family = s.getFamilyCode(*size.Name)
	instance.Category = s.getCategory(*size.Name)

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

// fetchPricing fetches real pricing data from Azure Retail Pricing API
func (s *RealComputeScraper) fetchPricing(instances map[string]*AzureInstance, regions []string) error {
	fmt.Printf("[Azure] Fetching real pricing for %d instances across %d regions\n", len(instances), len(regions))
	
	// Use Azure Retail Pricing API to get pricing data
	err := s.fetchAzureRetailPricing(instances, regions)
	if err != nil {
		fmt.Printf("[Azure] Warning: Failed to fetch pricing from Azure Retail Pricing API: %v\n", err)
	}
	
	return nil
}

// AzureRetailPricingResponse represents the response from Azure Retail Pricing API
type AzureRetailPricingResponse struct {
	Items    []AzureRetailPricingItem `json:"Items"`
	NextPage string                   `json:"NextPageLink"`
	Count    int                      `json:"Count"`
}

// AzureRetailPricingItem represents a pricing item from Azure Retail Pricing API
type AzureRetailPricingItem struct {
	CurrencyCode     string  `json:"currencyCode"`
	TierMinimumUnits float64 `json:"tierMinimumUnits"`
	RetailPrice      float64 `json:"retailPrice"`
	UnitPrice        float64 `json:"unitPrice"`
	ArmRegionName    string  `json:"armRegionName"`
	Location         string  `json:"location"`
	EffectiveDate    string  `json:"effectiveDate"`
	MeterID          string  `json:"meterId"`
	MeterName        string  `json:"meterName"`
	ProductID        string  `json:"productId"`
	SkuID            string  `json:"skuId"`
	ProductName      string  `json:"productName"`
	SkuName          string  `json:"skuName"`
	ServiceName      string  `json:"serviceName"`
	ServiceID        string  `json:"serviceId"`
	ServiceFamily    string  `json:"serviceFamily"`
	UnitOfMeasure    string  `json:"unitOfMeasure"`
	Type             string  `json:"type"`
	IsPrimaryRegion  bool    `json:"isPrimaryMeterRegion"`
	ArmSkuName       string  `json:"armSkuName"`
	ReservationTerm  string  `json:"reservationTerm"` // Added field for reservation term
}

// fetchAzureRetailPricing fetches pricing from Azure Retail Pricing API
func (s *RealComputeScraper) fetchAzureRetailPricing(instances map[string]*AzureInstance, regions []string) error {
	fmt.Println("[Azure] Fetching pricing from Azure Retail Pricing API...")
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	baseURL := "https://prices.azure.com/api/retail/prices"
	
	// Create work items batched by instance type
	type workItem struct {
		instanceType string
		instance     *AzureInstance
		regions      []string
	}
	
	var workItems []workItem
	for instanceType, instance := range instances {
		// Fetch pricing for ALL regions, not just where instance is available
		// This ensures consistent pricing data across all regions
		workItems = append(workItems, workItem{
			instanceType: instanceType,
			instance:     instance,
			regions:      regions,
		})
	}
	
	// Count total combinations
	totalCombinations := 0
	for _, item := range workItems {
		totalCombinations += len(item.regions)
	}
	fmt.Printf("[Azure] Fetching pricing for %d instances across %d region combinations\n", len(workItems), totalCombinations)
	
	// Process work items in parallel with limited concurrency
	var mu sync.Mutex
	applied := 0
	
	g, ctx := errgroup.WithContext(s.ctx)
	g.SetLimit(20) // Limit concurrent requests to avoid rate limiting
	
	for _, item := range workItems {
		item := item // capture loop variable
		g.Go(func() error {
			// Exponential backoff retry logic
			maxRetries := 3
			baseDelay := time.Second
			
			for attempt := 0; attempt < maxRetries; attempt++ {
				if attempt > 0 {
					// Exponential backoff: 1s, 2s, 4s
					delay := baseDelay * time.Duration(1<<(attempt-1))
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(delay):
					}
				}
				
				// Build filter for this specific VM size and multiple regions
				azureSKU := item.instance.PrettyNameAzure
				allSuccess := true
				
				// Batch regions to avoid exceeding URL length limits
				const maxRegionsPerBatch = 10
				for i := 0; i < len(item.regions); i += maxRegionsPerBatch {
					end := i + maxRegionsPerBatch
					if end > len(item.regions) {
						end = len(item.regions)
					}
					
					batchRegions := item.regions[i:end]
					
					// Build region filter with OR conditions
					regionFilters := make([]string, len(batchRegions))
					for j, region := range batchRegions {
						regionFilters[j] = fmt.Sprintf("armRegionName eq '%s'", region)
					}
					regionFilter := "(" + strings.Join(regionFilters, " or ") + ")"
					
					filter := fmt.Sprintf("serviceName eq 'Virtual Machines' and armSkuName eq '%s' and %s and currencyCode eq 'USD'", azureSKU, regionFilter)
					
					requestURL := fmt.Sprintf("%s?$filter=%s", baseURL, url.QueryEscape(filter))
					
					resp, err := client.Get(requestURL)
					if err != nil {
						if attempt == maxRetries-1 {
							fmt.Printf("[Azure] Warning: Failed to fetch pricing for %s (batch %d-%d) after %d attempts: %v\n", item.instanceType, i, end, maxRetries, err)
						}
						allSuccess = false
						continue
					}
					
					if resp.StatusCode == 429 {
						// Rate limited, retry with backoff
						resp.Body.Close()
						allSuccess = false
						break // Break out of region batch loop to retry
					}
					
					if resp.StatusCode != 200 {
						resp.Body.Close()
						if attempt == maxRetries-1 {
							fmt.Printf("[Azure] Warning: Pricing API returned status %d for %s (batch %d-%d)\n", resp.StatusCode, item.instanceType, i, end)
						}
						allSuccess = false
						continue
					}
					
					var pricingResp AzureRetailPricingResponse
					if err := json.NewDecoder(resp.Body).Decode(&pricingResp); err != nil {
						resp.Body.Close()
						if attempt == maxRetries-1 {
							fmt.Printf("[Azure] Warning: Failed to decode pricing for %s (batch %d-%d): %v\n", item.instanceType, i, end, err)
						}
						allSuccess = false
						continue
					}
					resp.Body.Close()
					
					// Group pricing items by region
					pricingByRegion := make(map[string][]AzureRetailPricingItem)
					for _, pItem := range pricingResp.Items {
						normalizedRegion := s.normalizeRegionName(pItem.ArmRegionName)
						if normalizedRegion != "" {
							pricingByRegion[normalizedRegion] = append(pricingByRegion[normalizedRegion], pItem)
						}
					}
					
					// Process pricing items for each region
					mu.Lock()
					for region, regionItems := range pricingByRegion {
						pricing := s.processPricingItems(regionItems, azureSKU)
						if pricing != nil {
							item.instance.Pricing[region] = *pricing
							applied++
							if applied%100 == 0 {
								fmt.Printf("[Azure] Progress: Applied %d pricing entries...\n", applied)
							}
						}
					}
					mu.Unlock()
				}
				
				// If all batches succeeded, break out of retry loop
				if allSuccess {
					break
				}
			}
			
			return nil
		})
	}
	
	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return fmt.Errorf("error fetching pricing: %v", err)
	}
	
	fmt.Printf("[Azure] Applied real pricing to %d instance/region combinations\n", applied)
	
	return nil
}

// extractVMSizeFromPricing extracts VM size from pricing item
func (s *RealComputeScraper) extractVMSizeFromPricing(item AzureRetailPricingItem) string {
	// Try ARM SKU name first
	if item.ArmSkuName != "" {
		return item.ArmSkuName
	}
	
	// Try extracting from SKU name
	if strings.Contains(item.SkuName, " ") {
		parts := strings.Fields(item.SkuName)
		for _, part := range parts {
			if strings.Contains(strings.ToLower(part), "standard_") ||
			   strings.Contains(strings.ToLower(part), "basic_") {
				return part
			}
		}
	}
	
	return ""
}

// normalizeRegionName converts ARM region name to our standard format
func (s *RealComputeScraper) normalizeRegionName(armRegion string) string {
	// Map ARM region names to our standardized hyphenated format
	// This matches the Python scraper's format
	regionMap := map[string]string{
		"eastus":             "us-east",
		"eastus2":            "us-east-2",
		"westus":             "us-west",
		"westus2":            "us-west-2",
		"westus3":            "us-west-3",
		"centralus":          "us-central",
		"northcentralus":     "us-north-central",
		"southcentralus":     "us-south-central",
		"westcentralus":      "us-west-central",
		"northeurope":        "europe-north",
		"westeurope":         "europe-west",
		"eastasia":           "asia-pacific-east",
		"southeastasia":      "asia-pacific-southeast",
		"japaneast":          "japan-east",
		"japanwest":          "japan-west",
		"koreacentral":       "korea-central",
		"koreasouth":         "korea-south",
		"australiaeast":      "australia-east",
		"australiasoutheast": "australia-southeast",
		"australiacentral":   "australia-central",
		"brazilsouth":        "brazil-south",
		"southindia":         "south-india",
		"centralindia":       "india-central",
		"westindia":          "india-west",
		"canadacentral":      "canada-central",
		"canadaeast":         "canada-east",
		"uksouth":            "uk-south",
		"ukwest":             "uk-west",
		"francecentral":      "france-central",
		"germanywestcentral": "germany-west-central",
		"norwayeast":         "norway-east",
		"switzerlandnorth":   "switzerland-north",
		"swedencentral":      "sweden-central",
		"uaenorth":           "uae-north",
		"southafricanorth":   "south-africa-north",
	}
	
	if mapped, exists := regionMap[strings.ToLower(armRegion)]; exists {
		return mapped
	}
	
	// Default: try to convert camelCase to hyphenated format
	// e.g., "southeastasia" -> "southeast-asia"
	return strings.ToLower(armRegion)
}

// roundToDecimals rounds a float64 to the specified number of decimal places
func roundToDecimals(value float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(value*multiplier) / multiplier
}

// processPricingItems processes Azure pricing items and returns structured pricing
func (s *RealComputeScraper) processPricingItems(items []AzureRetailPricingItem, azureSKU string) *AzurePricing {
	if len(items) == 0 {
		return nil
	}
	
	// Initialize pricing structures
	linuxPricing := AzureOSPricing{
		Reserved: make(map[string]float64),
	}
	windowsPricing := AzureOSPricing{
		Reserved: make(map[string]float64),
	}
	
	// Debug: log first few items to understand structure
	if len(items) > 0 && (azureSKU == "Standard_D4s_v3" || azureSKU == "Standard_A1_v2") { // Debug specific SKUs
		fmt.Printf("[Azure] Debug pricing for %s in region %s - found %d items\n", azureSKU, items[0].ArmRegionName, len(items))
		for i, item := range items {
			if i < 10 { // Show first 10 items
				fmt.Printf("  - MeterName: %s, SkuName: %s, ProductName: %s, PriceType: %s, ReservationTerm: %s, UnitOfMeasure: %s, Price: %f\n", 
					item.MeterName, item.SkuName, item.ProductName, item.Type, item.ReservationTerm, item.UnitOfMeasure, item.UnitPrice)
			}
		}
	}
	
	// Process each pricing item
	for _, item := range items {
		// Skip if not hourly pricing (except for reservations which might be yearly)
		if item.UnitOfMeasure != "1 Hour" && item.Type != "Reservation" {
			continue
		}
		
		price := item.UnitPrice
		
		// Handle reservation pricing conversion
		// The Azure Retail Pricing API returns upfront total cost for reservations
		if item.Type == "Reservation" {
			// Debug output to understand the pricing
			if azureSKU == "Standard_D4s_v3" || azureSKU == "Standard_B1s" {
				fmt.Printf("[Azure] Raw reservation price for %s: %.4f (UnitOfMeasure: %s, ReservationTerm: %s)\n", 
					azureSKU, price, item.UnitOfMeasure, item.ReservationTerm)
			}
			
			// For reservations, the API returns the total upfront cost
			// We need to convert this to hourly rate
			// Azure Retail Pricing API always returns reservation prices as total upfront cost
			if item.ReservationTerm == "3 Years" || 
			   strings.Contains(strings.ToLower(item.ProductName), "3 year") ||
			   strings.Contains(strings.ToLower(item.MeterName), "3 year") {
				// 3 year term: 26,280 hours (3 * 365 * 24)
				price = price / 26280.0
			} else {
				// 1 year term: 8,760 hours (365 * 24)
				price = price / 8760.0
			}
		}
		
		price = roundToDecimals(price, 4)
		
		// Determine OS and pricing type from MeterName and SkuName
		meterName := strings.ToLower(item.MeterName)
		skuName := strings.ToLower(item.SkuName)
		productName := strings.ToLower(item.ProductName)
		
		isWindows := strings.Contains(meterName, "windows") || strings.Contains(skuName, "windows") || strings.Contains(productName, "windows")
		isLowPriority := strings.Contains(meterName, "low priority") || strings.Contains(skuName, "low priority")
		isSpot := strings.Contains(meterName, "spot") || strings.Contains(skuName, "spot")
		
		// Check for reserved instances and savings plans
		isReserved := false
		var reservedKey string
		
		// Check if this is a reservation price type (matching Python logic)
		if item.Type == "Reservation" {
			isReserved = true
			
			// Determine if it's a Savings Plan or Standard Reservation
			isSavingsPlan := strings.Contains(meterName, "savings plan") || 
			                 strings.Contains(skuName, "savings plan") ||
			                 strings.Contains(productName, "savings plan")
			
			// Use reservationTerm field if available, otherwise parse from product info
			if item.ReservationTerm == "1 Year" {
				if isSavingsPlan {
					reservedKey = "yrTerm1Savings.allUpfront"
				} else {
					reservedKey = "yrTerm1Standard.allUpfront"
				}
			} else if item.ReservationTerm == "3 Years" {
				if isSavingsPlan {
					reservedKey = "yrTerm3Savings.allUpfront"
				} else {
					reservedKey = "yrTerm3Standard.allUpfront"
				}
			} else {
				// Fallback: try to determine from product/meter/sku names
				if strings.Contains(strings.ToLower(item.ProductName), "3 year") || 
				   strings.Contains(strings.ToLower(item.MeterName), "3 year") ||
				   strings.Contains(strings.ToLower(item.SkuName), "3 year") {
					if isSavingsPlan {
						reservedKey = "yrTerm3Savings.allUpfront"
					} else {
						reservedKey = "yrTerm3Standard.allUpfront"
					}
				} else {
					// Default to 1 year if not explicitly 3 year
					if isSavingsPlan {
						reservedKey = "yrTerm1Savings.allUpfront"
					} else {
						reservedKey = "yrTerm1Standard.allUpfront"
					}
				}
			}
		} else if strings.Contains(meterName, "1 year reserved") || strings.Contains(skuName, "1 year reserved") || 
		   strings.Contains(productName, "1 year reserved") || strings.Contains(meterName, "1-year reserved") {
			isReserved = true
			if strings.Contains(meterName, "azure hybrid benefit") || strings.Contains(skuName, "azure hybrid benefit") ||
			   strings.Contains(productName, "azure hybrid benefit") {
				reservedKey = "yrTerm1Standard.hybridbenefit"
			} else {
				reservedKey = "yrTerm1Standard.allUpfront"
			}
		} else if strings.Contains(meterName, "3 year reserved") || strings.Contains(skuName, "3 year reserved") || 
		          strings.Contains(productName, "3 year reserved") || strings.Contains(meterName, "3-year reserved") {
			isReserved = true
			if strings.Contains(meterName, "azure hybrid benefit") || strings.Contains(skuName, "azure hybrid benefit") ||
			   strings.Contains(productName, "azure hybrid benefit") {
				reservedKey = "yrTerm3Standard.hybridbenefit"
			} else {
				reservedKey = "yrTerm3Standard.allUpfront"
			}
		} else if strings.Contains(meterName, "1 year savings plan") || strings.Contains(skuName, "1 year savings plan") ||
		          strings.Contains(productName, "1 year savings plan") || strings.Contains(meterName, "1-year savings plan") {
			isReserved = true
			if strings.Contains(meterName, "azure hybrid benefit") || strings.Contains(skuName, "azure hybrid benefit") ||
			   strings.Contains(productName, "azure hybrid benefit") {
				reservedKey = "yrTerm1Savings.hybridbenefit"
			} else {
				reservedKey = "yrTerm1Savings.allUpfront"
			}
		} else if strings.Contains(meterName, "3 year savings plan") || strings.Contains(skuName, "3 year savings plan") ||
		          strings.Contains(productName, "3 year savings plan") || strings.Contains(meterName, "3-year savings plan") {
			isReserved = true
			if strings.Contains(meterName, "azure hybrid benefit") || strings.Contains(skuName, "azure hybrid benefit") ||
			   strings.Contains(productName, "azure hybrid benefit") {
				reservedKey = "yrTerm3Savings.hybridbenefit"
			} else {
				reservedKey = "yrTerm3Savings.allUpfront"
			}
		}
		
		// Check for basic pricing (A-series instances)
		isBasic := strings.Contains(meterName, "basic") || strings.Contains(skuName, "basic")
		
		// Check for Azure Hybrid Benefit
		isHybridBenefit := strings.Contains(meterName, "azure hybrid benefit") || 
		                   strings.Contains(skuName, "azure hybrid benefit") ||
		                   strings.Contains(productName, "azure hybrid benefit")
		
		// Apply pricing based on type (matching Python split_pricing logic)
		if item.Type == "Consumption" {
			if isBasic && isSpot {
				// Basic spot pricing
				if isWindows {
					windowsPricing.BasicSpot = price
				} else {
					linuxPricing.BasicSpot = price
				}
			} else if isBasic {
				// Basic pricing
				if isWindows {
					windowsPricing.Basic = price
				} else {
					linuxPricing.Basic = price
				}
			} else if isSpot {
				// Spot pricing
				if isWindows {
					windowsPricing.SpotMin = price
				} else {
					linuxPricing.SpotMin = price
				}
			} else if isLowPriority {
				// Low priority pricing
				if isWindows {
					windowsPricing.LowPriority = price
				} else {
					linuxPricing.LowPriority = price
				}
			} else if isHybridBenefit && isWindows {
				// Azure Hybrid Benefit (Windows only)
				windowsPricing.HybridBenefit = price
			} else {
				// Regular on-demand pricing
				if isWindows {
					windowsPricing.OnDemand = price
				} else {
					linuxPricing.OnDemand = price
				}
			}
		} else if item.Type == "Reservation" && isReserved && reservedKey != "" {
			// Reserved instance pricing
			if isWindows {
				windowsPricing.Reserved[reservedKey] = price
			} else {
				linuxPricing.Reserved[reservedKey] = price
			}
		} else if item.Type == "DevTestConsumption" {
			// Dev/Test pricing - we don't currently use this
			// Python stores it as "devtest" but we'll skip for now
		}
	}
	
	// Debug: log what we collected
	if azureSKU == "Standard_D4s_v3" || azureSKU == "Standard_A1_v2" {
		fmt.Printf("[Azure] Collected pricing - Linux: ondemand=%f, lowpri=%f, spot=%f, reserved=%d items\n", 
			linuxPricing.OnDemand, linuxPricing.LowPriority, linuxPricing.SpotMin, len(linuxPricing.Reserved))
		fmt.Printf("[Azure] Collected pricing - Windows: ondemand=%f, lowpri=%f, spot=%f, reserved=%d items\n", 
			windowsPricing.OnDemand, windowsPricing.LowPriority, windowsPricing.SpotMin, len(windowsPricing.Reserved))
		if len(linuxPricing.Reserved) > 0 {
			fmt.Printf("[Azure] Linux reserved keys: %v\n", linuxPricing.Reserved)
		}
		if len(windowsPricing.Reserved) > 0 {
			fmt.Printf("[Azure] Windows reserved keys: %v\n", windowsPricing.Reserved)
		}
	}
	
	// Add Windows Hybrid Benefit pricing (same as Linux on-demand)
	// This represents Windows pricing when you bring your own license
	if linuxPricing.OnDemand > 0 && windowsPricing.OnDemand > 0 {
		windowsPricing.HybridBenefit = linuxPricing.OnDemand
	}
	
	// Return pricing structure even if we only have partial data
	return &AzurePricing{
		Linux:   linuxPricing,
		Windows: windowsPricing,
	}
}
