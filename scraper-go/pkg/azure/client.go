package azure

import (
	"context"
	"net/http"
	"time"
)

// ClientManager manages Azure API clients
type ClientManager struct {
	httpClient *http.Client
	ctx        context.Context
}

// NewClientManager creates a new Azure client manager
func NewClientManager(ctx context.Context) (*ClientManager, error) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	return &ClientManager{
		httpClient: httpClient,
		ctx:        ctx,
	}, nil
}

// GetHTTPClient returns the HTTP client
func (cm *ClientManager) GetHTTPClient() *http.Client {
	return cm.httpClient
}

// GetAvailableRegions returns all available Azure regions
func GetAvailableRegions() []string {
	// Azure regions as of 2024
	return []string{
		"eastus",           // East US
		"eastus2",          // East US 2
		"southcentralus",   // South Central US
		"westus2",          // West US 2
		"westus3",          // West US 3
		"australiaeast",    // Australia East
		"southeastasia",    // Southeast Asia
		"northeurope",      // North Europe
		"swedencentral",    // Sweden Central
		"uksouth",          // UK South
		"westeurope",       // West Europe
		"centralus",        // Central US
		"southafricanorth", // South Africa North
		"centralindia",     // Central India
		"eastasia",         // East Asia
		"japaneast",        // Japan East
		"koreacentral",     // Korea Central
		"canadacentral",    // Canada Central
		"francecentral",    // France Central
		"germanywestcentral", // Germany West Central
		"norwayeast",       // Norway East
		"switzerlandnorth", // Switzerland North
		"uaenorth",         // UAE North
		"brazilsouth",      // Brazil South
		"qatarcentral",     // Qatar Central
		"centraluseuap",    // Central US EUAP
		"eastus2euap",      // East US 2 EUAP
	}
}