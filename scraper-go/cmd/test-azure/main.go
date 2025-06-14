package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloud-instances/scraper/pkg/azure"
	"github.com/cloud-instances/scraper/pkg/models"
)

func main() {
	// Get subscription ID from Azure CLI or environment variable
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	if subscriptionID == "" {
		fmt.Println("Getting subscription ID from Azure CLI...")
		// Try to get from Azure CLI context
		// This should match the subscription from 'az account show'
		subscriptionID = "189a26dd-5eff-4eb1-8577-49787e750745" // From your Azure CLI output
	}

	fmt.Printf("Using Azure subscription: %s\n", subscriptionID)

	ctx := context.Background()
	
	// Create real Azure scraper
	scraper, err := azure.NewRealComputeScraper(ctx, subscriptionID)
	if err != nil {
		fmt.Printf("Failed to create Azure scraper: %v\n", err)
		os.Exit(1)
	}

	// Configure scraping with a small sample
	config := &models.ScraperConfig{
		Regions:        []string{"eastus", "westus2"}, // Limit to 2 regions for testing
		MaxConcurrency: 2,
		OutputPath:     "/tmp/azure-test",
		Service:        "azure",
		SampleSize:     10, // Small sample for testing
	}

	fmt.Println("Starting Azure VM size scraping...")
	instances, err := scraper.Scrape(config)
	if err != nil {
		fmt.Printf("Failed to scrape Azure instances: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSuccessfully scraped %d Azure VM types\n", len(instances))
	
	// Display first few instances
	fmt.Println("\nFirst 5 Azure VM types:")
	for i, instance := range instances {
		if i >= 5 {
			break
		}
		
		regionsStr := ""
		if len(instance.AvailabilityZones) > 0 {
			regions := make([]string, 0, len(instance.AvailabilityZones))
			for region := range instance.AvailabilityZones {
				regions = append(regions, region)
			}
			regionsStr = fmt.Sprintf(" (regions: %v)", regions)
		}
		
		vcpuStr := "unknown"
		if instance.VCPU != nil {
			vcpuStr = fmt.Sprintf("%v", instance.VCPU)
		}
		
		fmt.Printf("  %s: %s vCPU, %.1f GB RAM, %s%s\n", 
			instance.InstanceType, 
			vcpuStr, 
			instance.Memory, 
			instance.Category,
			regionsStr)
	}

	// Save to JSON file for inspection
	outputFile := "azure-instances-real.json"
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Failed to create output file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(instances); err != nil {
		fmt.Printf("Failed to write JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nFull results saved to: %s\n", outputFile)
	fmt.Println("Azure scraping completed successfully!")
}