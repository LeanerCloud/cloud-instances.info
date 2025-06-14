package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
	pricingtypes "github.com/aws/aws-sdk-go-v2/service/pricing/types"
	"github.com/cloud-instances/scraper/pkg/models"
	"golang.org/x/sync/errgroup"
)

// OpenSearchScraper implements scraping for OpenSearch instances
type OpenSearchScraper struct {
	clientManager *ClientManager
	ctx           context.Context
}

// NewOpenSearchScraper creates a new OpenSearch scraper
func NewOpenSearchScraper(ctx context.Context) (*OpenSearchScraper, error) {
	cm, err := NewClientManager(ctx)
	if err != nil {
		return nil, err
	}
	
	return &OpenSearchScraper{
		clientManager: cm,
		ctx:           ctx,
	}, nil
}

// GetServiceName returns the service name
func (s *OpenSearchScraper) GetServiceName() string {
	return "OpenSearch"
}

// Scrape fetches OpenSearch instance data from AWS
func (s *OpenSearchScraper) Scrape(config *models.ScraperConfig) ([]models.Instance, error) {
	// Get pricing data from us-east-1
	clients, err := s.clientManager.GetClients("us-east-1")
	if err != nil {
		return nil, fmt.Errorf("failed to get clients: %v", err)
	}

	// OpenSearch instances use instance types like "t2.small.search", "m5.large.search"
	instanceTypes := make(map[string]*models.Instance)
	var mu sync.Mutex
	
	// Get all OpenSearch products from pricing API
	// Note: OpenSearch still uses "AmazonES" service code for backwards compatibility
	filters := []pricingtypes.Filter{
		{
			Type:  pricingtypes.FilterTypeTermMatch,
			Field: aws.String("ServiceCode"),
			Value: aws.String("AmazonES"),
		},
	}

	var nextToken *string
	for {
		input := &pricing.GetProductsInput{
			ServiceCode: aws.String("AmazonES"),
			Filters:     filters,
			MaxResults:  aws.Int32(100),
			NextToken:   nextToken,
		}

		output, err := clients.PricingClient.GetProducts(s.ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to get products: %v", err)
		}
		

		// Process products in parallel
		g, _ := errgroup.WithContext(s.ctx)
		g.SetLimit(config.MaxConcurrency)
		
		for _, priceItem := range output.PriceList {
			priceItem := priceItem // capture loop variable
			g.Go(func() error {
				instance, err := s.parseProduct(priceItem)
				if err != nil {
					return nil // Skip errors for individual products
				}
				
				if instance != nil {
					mu.Lock()
					if existing, ok := instanceTypes[instance.InstanceType]; ok {
						// Merge regions and pricing
						for region, desc := range instance.Regions {
							existing.Regions[region] = desc
						}
						for region, pricing := range instance.Pricing {
							if _, ok := existing.Pricing[region]; !ok {
								existing.Pricing[region] = pricing
							}
						}
					} else {
						instanceTypes[instance.InstanceType] = instance
					}
					mu.Unlock()
				}
				
				return nil
			})
		}
		
		if err := g.Wait(); err != nil {
			return nil, err
		}

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	// Convert map to slice
	instances := make([]models.Instance, 0, len(instanceTypes))
	for _, instance := range instanceTypes {
		instances = append(instances, *instance)
	}

	return instances, nil
}

// parseProduct parses a pricing API product into an Instance
func (s *OpenSearchScraper) parseProduct(priceItem string) (*models.Instance, error) {
	var product map[string]interface{}
	if err := json.Unmarshal([]byte(priceItem), &product); err != nil {
		return nil, err
	}

	// Extract attributes
	productData, ok := product["product"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing product data")
	}
	
	attributes, ok := productData["attributes"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing attributes")
	}

	// Get instance type
	instanceType, ok := attributes["instanceType"].(string)
	if !ok || instanceType == "" {
		return nil, fmt.Errorf("missing instance type")
	}

	// Skip if not a valid OpenSearch instance type
	if !s.isValidOpenSearchInstanceType(instanceType) {
		return nil, fmt.Errorf("invalid opensearch instance type: %s", instanceType)
	}

	// Get location and convert to region
	location, ok := attributes["location"].(string)
	if !ok {
		return nil, fmt.Errorf("missing location")
	}
	
	region := s.locationToRegion(location)
	if region == "" {
		return nil, fmt.Errorf("unknown location: %s", location)
	}

	// Create instance
	instance := &models.Instance{
		InstanceType: instanceType,
		Pricing:      make(map[string]models.RegionPricing),
		Regions:      make(map[string]string),
	}

	// Basic attributes
	if memory, ok := attributes["memory"].(string); ok {
		// Extract numeric value from strings like "32 GiB"
		memStr := strings.TrimSpace(strings.Replace(memory, "GiB", "", 1))
		memStr = strings.TrimSpace(strings.Replace(memStr, "GB", "", 1))
		instance.Memory = memStr
	}

	if vcpu, ok := attributes["vcpu"].(string); ok {
		instance.VCPU = json.Number(vcpu)
	}

	if physicalProcessor, ok := attributes["physicalProcessor"].(string); ok {
		instance.PhysicalProcessor = physicalProcessor
	}

	if clockSpeed, ok := attributes["clockSpeed"].(string); ok {
		instance.ClockSpeed = clockSpeed
	}

	if networkPerformance, ok := attributes["networkPerformance"].(string); ok {
		instance.NetworkPerformance = networkPerformance
	}

	// Family from instance type (e.g., "m5.large.search" -> "m5")
	parts := strings.Split(instanceType, ".")
	if len(parts) >= 2 {
		instance.Family = s.getOpenSearchFamilyName(parts[0])
	}

	// Storage info - OpenSearch instances typically have EBS storage
	if storageType, ok := attributes["storageType"].(string); ok {
		storage := map[string]interface{}{
			"ssd":     storageType == "EBS" || storageType == "SSD",
			"devices": 0,
			"size":    0,
		}
		storageJSON, _ := json.Marshal(storage)
		instance.Storage = json.RawMessage(storageJSON)
	}

	// Region
	instance.Regions[region] = region

	// Pricing
	instance.Pricing[region] = models.RegionPricing{
		OnDemand: 0,
		Reserved: make(map[string]float64),
	}

	// Parse pricing from terms
	if terms, ok := product["terms"].(map[string]interface{}); ok {
		// OnDemand pricing
		if onDemand, ok := terms["OnDemand"].(map[string]interface{}); ok {
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

					pricePerUnit, ok := dim["pricePerUnit"].(map[string]interface{})
					if !ok {
						continue
					}

					if usdPrice, ok := pricePerUnit["USD"].(string); ok {
						if price, err := strconv.ParseFloat(usdPrice, 64); err == nil {
							pricing := instance.Pricing[region]
							pricing.OnDemand = price
							instance.Pricing[region] = pricing
						}
					}
				}
			}
		}

		// Reserved pricing
		if reserved, ok := terms["Reserved"].(map[string]interface{}); ok {
			s.parseReservedPricing(reserved, instance, region)
		}
	}

	// Pretty name
	instance.PrettyName = s.generatePrettyName(instance)

	return instance, nil
}

// parseReservedPricing parses reserved pricing information
func (s *OpenSearchScraper) parseReservedPricing(reserved map[string]interface{}, instance *models.Instance, region string) {
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
			pricingKey = "yrTerm1.allUpfront"
		} else if leaseContractLength == "1yr" && purchaseOption == "Partial Upfront" {
			pricingKey = "yrTerm1.partialUpfront"
		} else if leaseContractLength == "1yr" && purchaseOption == "No Upfront" {
			pricingKey = "yrTerm1.noUpfront"
		} else if leaseContractLength == "3yr" && purchaseOption == "All Upfront" {
			pricingKey = "yrTerm3.allUpfront"
		} else if leaseContractLength == "3yr" && purchaseOption == "Partial Upfront" {
			pricingKey = "yrTerm3.partialUpfront"
		} else if leaseContractLength == "3yr" && purchaseOption == "No Upfront" {
			pricingKey = "yrTerm3.noUpfront"
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

// isValidOpenSearchInstanceType checks if the instance type is a valid OpenSearch instance type
func (s *OpenSearchScraper) isValidOpenSearchInstanceType(instanceType string) bool {
	// OpenSearch instance types typically end with ".search" 
	return strings.HasSuffix(instanceType, ".search")
}

// getOpenSearchFamilyName returns human-readable family names for OpenSearch
func (s *OpenSearchScraper) getOpenSearchFamilyName(familyCode string) string {
	familyMap := map[string]string{
		"t2": "T2 General Purpose",
		"t3": "T3 General Purpose",
		"m3": "M3 General Purpose",
		"m4": "M4 General Purpose",
		"m5": "M5 General Purpose",
		"m6g": "M6G General Purpose",
		"c4": "C4 Compute Optimized",
		"c5": "C5 Compute Optimized",
		"c6g": "C6G Compute Optimized",
		"r3": "R3 Memory Optimized",
		"r4": "R4 Memory Optimized",
		"r5": "R5 Memory Optimized",
		"r6g": "R6G Memory Optimized",
		"r6gd": "R6GD Memory Optimized",
		"i2": "I2 Storage Optimized",
		"i3": "I3 Storage Optimized",
		"d2": "D2 Storage Optimized",
		"ultrawarm1": "UltraWarm",
	}
	
	if family, ok := familyMap[familyCode]; ok {
		return family
	}
	
	return strings.ToUpper(familyCode) + " General Purpose"
}

// generatePrettyName generates a human-readable name for OpenSearch instances
func (s *OpenSearchScraper) generatePrettyName(instance *models.Instance) string {
	// OpenSearch instance types look like "m5.large.search"
	parts := strings.Split(instance.InstanceType, ".")
	if len(parts) < 2 {
		return instance.InstanceType
	}

	family := strings.ToUpper(parts[0])
	size := parts[1]

	// Size mapping
	sizeMap := map[string]string{
		"small": "Small", "medium": "Medium", "large": "Large",
		"xlarge": "Extra Large", "2xlarge": "Double Extra Large",
		"4xlarge": "Quadruple Extra Large", "8xlarge": "Eight Extra Large",
		"12xlarge": "12xlarge", "16xlarge": "16xlarge", "24xlarge": "24xlarge",
	}

	sizeName, ok := sizeMap[size]
	if !ok {
		sizeName = size
	}

	return fmt.Sprintf("%s %s", family, sizeName)
}

// locationToRegion converts AWS location names to region codes
func (s *OpenSearchScraper) locationToRegion(location string) string {
	// Comprehensive location mapping including all variations seen in pricing API
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
		"Asia Pacific (Thailand)":   "ap-southeast-2", // Map to closest
		"Asia Pacific (Malaysia)":   "ap-southeast-1", // Map to closest
		"Asia Pacific (Taipei)":     "ap-northeast-1", // Map to closest
		"South America (Sao Paulo)": "sa-east-1",
		"Canada (Central)":          "ca-central-1",
		"Canada West (Calgary)":     "ca-west-1",
		"Middle East (Bahrain)":     "me-south-1",
		"Middle East (UAE)":         "me-central-1",
		"Africa (Cape Town)":        "af-south-1",
		"Israel (Tel Aviv)":         "il-central-1",
		"Mexico (Central)":          "us-east-1", // Map to closest
		// Skip gov cloud and china regions for now
	}

	if region, ok := locationMap[location]; ok {
		return region
	}
	// Skip unknown regions instead of erroring
	return ""
}