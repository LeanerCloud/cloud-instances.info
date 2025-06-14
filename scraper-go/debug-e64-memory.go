package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

func main() {
	// Get Azure credentials
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		fmt.Printf("Failed to get Azure credentials: %v\n", err)
		os.Exit(1)
	}

	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		fmt.Println("AZURE_SUBSCRIPTION_ID environment variable not set")
		os.Exit(1)
	}

	// Create VM sizes client
	client, err := armcompute.NewVirtualMachineSizesClient(subscriptionID, cred, nil)
	if err != nil {
		fmt.Printf("Failed to create VM sizes client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	region := "eastus"

	fmt.Printf("Fetching VM sizes for region: %s\n", region)
	
	pager := client.NewListPager(region, nil)
	
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			fmt.Printf("Failed to get VM sizes: %v\n", err)
			os.Exit(1)
		}

		for _, size := range page.Value {
			if size.Name != nil && (*size.Name == "Standard_E64a_v4" || *size.Name == "Standard_E64as_v4") {
				fmt.Printf("\n=== %s ===\n", *size.Name)
				fmt.Printf("Name: %s\n", *size.Name)
				
				if size.MemoryInMB != nil {
					fmt.Printf("Memory (Azure API): %d MB\n", *size.MemoryInMB)
					fmt.Printf("Memory (Our conversion): %.1f GB\n", float64(*size.MemoryInMB)/1024)
				}
				
				if size.NumberOfCores != nil {
					fmt.Printf("vCPUs: %d\n", *size.NumberOfCores)
				}
				
				if size.ResourceDiskSizeInMB != nil {
					fmt.Printf("Resource Disk: %d MB (%.1f GB)\n", *size.ResourceDiskSizeInMB, float64(*size.ResourceDiskSizeInMB)/1024)
				}
				
				if size.MaxDataDiskCount != nil {
					fmt.Printf("Max Data Disks: %d\n", *size.MaxDataDiskCount)
				}
				
				if size.OSDiskSizeInMB != nil {
					fmt.Printf("OS Disk: %d MB (%.1f GB)\n", *size.OSDiskSizeInMB, float64(*size.OSDiskSizeInMB)/1024)
				}
			}
		}
	}
}