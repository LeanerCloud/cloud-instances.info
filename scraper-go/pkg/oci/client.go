package oci

import (
	"context"
	"fmt"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
)

// ClientManager manages OCI clients
type ClientManager struct {
	computeClient core.ComputeClient
	ctx           context.Context
}

// NewClientManager creates a new OCI client manager
func NewClientManager(ctx context.Context) (*ClientManager, error) {
	// Create configuration provider
	configProvider := common.DefaultConfigProvider()
	
	// Create compute client
	computeClient, err := core.NewComputeClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %v", err)
	}

	return &ClientManager{
		computeClient: computeClient,
		ctx:           ctx,
	}, nil
}

// GetComputeClient returns the compute client
func (cm *ClientManager) GetComputeClient() core.ComputeClient {
	return cm.computeClient
}


// GetRegions returns all available regions
func GetAvailableRegions() []string {
	// OCI regions as of 2024 - this could be made dynamic with the Identity service
	return []string{
		"us-ashburn-1",     // US East (Ashburn)
		"us-phoenix-1",     // US West (Phoenix)
		"ca-toronto-1",     // Canada Central (Toronto)
		"ca-montreal-1",    // Canada Southeast (Montreal)
		"sa-saopaulo-1",    // Brazil East (Sao Paulo)
		"eu-frankfurt-1",   // Germany Central (Frankfurt)
		"eu-zurich-1",      // Switzerland North (Zurich)
		"eu-amsterdam-1",   // Netherlands Northwest (Amsterdam)
		"uk-london-1",      // UK South (London)
		"ap-mumbai-1",      // India West (Mumbai)
		"ap-hyderabad-1",   // India South (Hyderabad)
		"ap-seoul-1",       // South Korea Central (Seoul)
		"ap-chuncheon-1",   // South Korea North (Chuncheon)
		"ap-tokyo-1",       // Japan East (Tokyo)
		"ap-osaka-1",       // Japan Central (Osaka)
		"ap-sydney-1",      // Australia East (Sydney)
		"ap-melbourne-1",   // Australia Southeast (Melbourne)
		"me-jeddah-1",      // Saudi Arabia West (Jeddah)
		"me-dubai-1",       // UAE East (Dubai)
		"af-johannesburg-1", // South Africa North (Johannesburg)
		"il-jerusalem-1",   // Israel Central (Jerusalem)
	}
}