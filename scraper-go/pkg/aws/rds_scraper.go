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

// RDSScraper implements scraping for RDS instances
type RDSScraper struct {
	clientManager *ClientManager
	ctx           context.Context
}

// NewRDSScraper creates a new RDS scraper
func NewRDSScraper(ctx context.Context) (*RDSScraper, error) {
	cm, err := NewClientManager(ctx)
	if err != nil {
		return nil, err
	}
	
	return &RDSScraper{
		clientManager: cm,
		ctx:           ctx,
	}, nil
}

// GetServiceName returns the service name
func (s *RDSScraper) GetServiceName() string {
	return "RDS"
}

// Scrape fetches RDS instance data from AWS
func (s *RDSScraper) Scrape(config *models.ScraperConfig) ([]models.Instance, error) {
	// Get pricing data from us-east-1
	clients, err := s.clientManager.GetClients("us-east-1")
	if err != nil {
		return nil, fmt.Errorf("failed to get clients: %v", err)
	}

	// RDS instances are prefixed with "db."
	instanceTypes := make(map[string]*models.Instance)
	var mu sync.Mutex
	
	// Get all RDS products from pricing API
	filters := []pricingtypes.Filter{
		{
			Type:  pricingtypes.FilterTypeTermMatch,
			Field: aws.String("ServiceCode"),
			Value: aws.String("AmazonRDS"),
		},
		{
			Type:  pricingtypes.FilterTypeTermMatch,
			Field: aws.String("productFamily"),
			Value: aws.String("Database Instance"),
		},
	}

	var nextToken *string
	for {
		input := &pricing.GetProductsInput{
			ServiceCode: aws.String("AmazonRDS"),
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

	// Get EBS info from EC2 instances
	s.addEBSInfo(instanceTypes)

	// Convert map to slice
	instances := make([]models.Instance, 0, len(instanceTypes))
	for _, instance := range instanceTypes {
		instances = append(instances, *instance)
	}

	return instances, nil
}

// parseProduct parses a pricing API product into an Instance
func (s *RDSScraper) parseProduct(priceItem string) (*models.Instance, error) {
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

	// Skip if not a database instance
	instanceType, ok := attributes["instanceType"].(string)
	if !ok || instanceType == "" {
		return nil, fmt.Errorf("missing instance type")
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

	if clockSpeed, ok := attributes["clockSpeedInGhz"].(string); ok {
		instance.ClockSpeed = fmt.Sprintf("%s GHz", clockSpeed)
	}

	if networkPerformance, ok := attributes["networkPerformance"].(string); ok {
		instance.NetworkPerformance = networkPerformance
	}

	// Family from instance type (e.g., "db.m5.large" -> "m5")
	parts := strings.Split(instanceType, ".")
	if len(parts) >= 2 {
		instance.Family = s.getRDSFamilyName(parts[1])
	}

	// Storage
	if storageMedia, ok := attributes["storageMedia"].(string); ok {
		storage := map[string]interface{}{
			"ssd":     storageMedia == "SSD",
			"devices": 0,
			"size":    0,
		}
		storageJSON, _ := json.Marshal(storage)
		instance.Storage = json.RawMessage(storageJSON)
	}

	// Engine info
	if engine, ok := attributes["databaseEngine"].(string); ok {
		instance.Engine = engine
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

					// Skip non-instance charges
					if desc, ok := dim["description"].(string); ok {
						descLower := strings.ToLower(desc)
						if strings.Contains(descLower, "storage") || 
						   strings.Contains(descLower, "i/o") ||
						   strings.Contains(descLower, "backup") {
							continue
						}
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
	}

	// Pretty name
	instance.PrettyName = s.generatePrettyName(instance)

	return instance, nil
}

// getRDSFamilyName returns human-readable family names for RDS
func (s *RDSScraper) getRDSFamilyName(familyCode string) string {
	familyMap := map[string]string{
		"t2": "T2 General Purpose", "t3": "T3 General Purpose", "t4g": "T4G General Purpose",
		"m3": "M3 General Purpose", "m4": "M4 General Purpose", "m5": "M5 General Purpose",
		"m5d": "M5D General Purpose", "m6g": "M6G General Purpose", "m6gd": "M6GD General Purpose",
		"m6i": "M6I General Purpose", "m6id": "M6ID General Purpose", "m7g": "M7G General Purpose",
		"r3": "R3 Memory Optimized", "r4": "R4 Memory Optimized", "r5": "R5 Memory Optimized",
		"r5b": "R5B Memory Optimized", "r5d": "R5D Memory Optimized", "r6g": "R6G Memory Optimized",
		"r6gd": "R6GD Memory Optimized", "r6i": "R6I Memory Optimized", "r7g": "R7G Memory Optimized",
		"x1": "X1 Memory Optimized", "x1e": "X1E Memory Optimized", "x2g": "X2G Memory Optimized",
		"x2idn": "X2IDN Memory Optimized", "x2iedn": "X2IEDN Memory Optimized", "x2iezn": "X2IEZN Memory Optimized",
		"z1d": "Z1D Memory Optimized",
		"i3": "I3 High I/O",
	}
	
	if family, ok := familyMap[familyCode]; ok {
		return family
	}
	
	return strings.ToUpper(familyCode) + " General Purpose"
}

// generatePrettyName generates a human-readable name for RDS instances
func (s *RDSScraper) generatePrettyName(instance *models.Instance) string {
	// RDS instance types look like "db.r4.large"
	parts := strings.Split(instance.InstanceType, ".")
	if len(parts) != 3 {
		return instance.InstanceType
	}

	family := strings.ToUpper(parts[1])
	size := parts[2]

	// Size mapping similar to EC2
	sizeMap := map[string]string{
		"micro": "Micro", "small": "Small", "medium": "Medium", "large": "Large",
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

// addEBSInfo adds EBS information from EC2 instances
func (s *RDSScraper) addEBSInfo(instances map[string]*models.Instance) {
	// This would require fetching EC2 instance data and matching
	// For now, set default values
	for _, instance := range instances {
		// These fields are added by the Python scraper
		// We'll set them to 0 for now
		instance.Storage = json.RawMessage(`{"ssd":false,"devices":0,"size":0}`)
	}
}

// locationToRegion converts AWS location names to region codes
func (s *RDSScraper) locationToRegion(location string) string {
	// Reuse the mapping from EC2 scraper
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
		"Middle East (Bahrain)":   "me-south-1",
		"Africa (Cape Town)":      "af-south-1",
		"Asia Pacific (Hong Kong)": "ap-east-1",
		"Asia Pacific (Jakarta)":  "ap-southeast-3",
		"Asia Pacific (Osaka)":    "ap-northeast-3",
		"Europe (Milan)":          "eu-south-1",
		"Europe (Spain)":          "eu-south-2",
		"Europe (Zurich)":         "eu-central-2",
		"Middle East (UAE)":       "me-central-1",
		"Asia Pacific (Hyderabad)": "ap-south-2",
		"Asia Pacific (Melbourne)": "ap-southeast-4",
		"Israel (Tel Aviv)":       "il-central-1",
		"Canada West (Calgary)":   "ca-west-1",
	}

	if region, ok := locationMap[location]; ok {
		return region
	}
	return ""
}