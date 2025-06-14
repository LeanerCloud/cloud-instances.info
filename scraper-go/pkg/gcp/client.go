package gcp

import (
	"context"
	"fmt"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// ClientManager manages GCP clients
type ClientManager struct {
	computeService *compute.Service
	ctx            context.Context
}

// NewClientManager creates a new GCP client manager
func NewClientManager(ctx context.Context) (*ClientManager, error) {
	// Create compute service
	computeService, err := compute.NewService(ctx, option.WithScopes(compute.ComputeReadonlyScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create compute service: %v", err)
	}

	return &ClientManager{
		computeService: computeService,
		ctx:            ctx,
	}, nil
}

// GetComputeService returns the compute service
func (cm *ClientManager) GetComputeService() *compute.Service {
	return cm.computeService
}

// GetAvailableZones returns all available zones across all regions
func (cm *ClientManager) GetAvailableZones(projectID string) ([]string, error) {
	zones := []string{}
	
	// List all zones
	call := cm.computeService.Zones.List(projectID)
	if err := call.Pages(cm.ctx, func(page *compute.ZoneList) error {
		for _, zone := range page.Items {
			if zone.Status == "UP" {
				zones = append(zones, zone.Name)
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to list zones: %v", err)
	}

	return zones, nil
}

// GetAvailableRegions returns all available regions
func (cm *ClientManager) GetAvailableRegions(projectID string) ([]string, error) {
	regions := []string{}
	
	// List all regions
	call := cm.computeService.Regions.List(projectID)
	if err := call.Pages(cm.ctx, func(page *compute.RegionList) error {
		for _, region := range page.Items {
			if region.Status == "UP" {
				regions = append(regions, region.Name)
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to list regions: %v", err)
	}

	return regions, nil
}