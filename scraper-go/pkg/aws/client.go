package aws

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/pricing"
)

// RegionClients holds AWS service clients for a specific region
type RegionClients struct {
	Region        string
	Config        aws.Config
	EC2Client     *ec2.Client
	PricingClient *pricing.Client
}

// ClientManager manages AWS clients across multiple regions
type ClientManager struct {
	clients map[string]*RegionClients
	mu      sync.RWMutex
	ctx     context.Context
}

// NewClientManager creates a new AWS client manager
func NewClientManager(ctx context.Context) (*ClientManager, error) {
	return &ClientManager{
		clients: make(map[string]*RegionClients),
		ctx:     ctx,
	}, nil
}

// GetClients returns AWS clients for a specific region
func (cm *ClientManager) GetClients(region string) (*RegionClients, error) {
	cm.mu.RLock()
	if clients, ok := cm.clients[region]; ok {
		cm.mu.RUnlock()
		return clients, nil
	}
	cm.mu.RUnlock()

	// Create new clients for the region
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Double-check after acquiring write lock
	if clients, ok := cm.clients[region]; ok {
		return clients, nil
	}

	cfg, err := config.LoadDefaultConfig(cm.ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config for region %s: %v", region, err)
	}

	clients := &RegionClients{
		Region:        region,
		Config:        cfg,
		EC2Client:     ec2.NewFromConfig(cfg),
		PricingClient: pricing.NewFromConfig(cfg),
	}

	cm.clients[region] = clients
	return clients, nil
}

// GetAllRegions fetches all available AWS regions
func GetAllRegions(ctx context.Context) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	client := ec2.NewFromConfig(cfg)
	
	result, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(false), // Only get opted-in regions
	})
	if err != nil {
		return nil, fmt.Errorf("unable to describe regions: %v", err)
	}

	regions := make([]string, 0, len(result.Regions))
	for _, region := range result.Regions {
		if region.RegionName != nil {
			regions = append(regions, *region.RegionName)
		}
	}

	return regions, nil
}

// GetRegionDescriptions returns a map of region names to their descriptions
func GetRegionDescriptions(ctx context.Context) (map[string]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	client := ec2.NewFromConfig(cfg)
	
	result, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to describe regions: %v", err)
	}

	descriptions := make(map[string]string)
	for _, region := range result.Regions {
		if region.RegionName != nil {
			descriptions[*region.RegionName] = *region.RegionName
		}
	}

	return descriptions, nil
}