package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// RealClientManager manages real Azure API clients
type RealClientManager struct {
	computeClient *armcompute.VirtualMachineSizesClient
	ctx           context.Context
	subscriptionID string
}

// NewRealClientManager creates a new Azure client manager with real Azure credentials
func NewRealClientManager(ctx context.Context, subscriptionID string) (*RealClientManager, error) {
	// Use DefaultAzureCredential which will try multiple authentication methods:
	// 1. Environment variables
	// 2. Managed identity
	// 3. Azure CLI
	// 4. Azure PowerShell
	// 5. Interactive browser
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %v", err)
	}

	// Create compute client
	computeClient, err := armcompute.NewVirtualMachineSizesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %v", err)
	}

	return &RealClientManager{
		computeClient:  computeClient,
		ctx:           ctx,
		subscriptionID: subscriptionID,
	}, nil
}

// GetVMSizes fetches VM sizes for a specific location using real Azure APIs
func (cm *RealClientManager) GetVMSizes(location string) ([]*armcompute.VirtualMachineSize, error) {
	pager := cm.computeClient.NewListPager(location, nil)
	
	var allSizes []*armcompute.VirtualMachineSize
	for pager.More() {
		page, err := pager.NextPage(cm.ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get VM sizes for location %s: %v", location, err)
		}
		
		allSizes = append(allSizes, page.Value...)
	}
	
	return allSizes, nil
}

// GetSubscriptionID returns the subscription ID
func (cm *RealClientManager) GetSubscriptionID() string {
	return cm.subscriptionID
}