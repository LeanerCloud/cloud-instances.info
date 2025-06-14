package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cloud-instances/scraper/pkg/aws"
	"github.com/cloud-instances/scraper/pkg/models"
)

func main() {
	ctx := context.Background()
	
	fmt.Println("Starting EC2 scraper with full pricing test...")
	startTime := time.Now()
	
	scraper, err := aws.NewEC2Scraper(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create scraper: %v\n", err)
		os.Exit(1)
	}
	
	config := &models.ScraperConfig{
		Regions:        []string{"us-east-1"},
		MaxConcurrency: 10,
		Service:        "ec2",
	}
	
	instances, err := scraper.Scrape(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to scrape: %v\n", err)
		os.Exit(1)
	}
	
	// Save to file
	file, err := os.Create("/tmp/ec2-pricing-test.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(instances); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
	
	duration := time.Since(startTime)
	fmt.Printf("\nCompleted in %v. Found %d instances.\n", duration, len(instances))
	
	// Count pricing data
	pricedCount := 0
	spotCount := 0
	for _, inst := range instances {
		if len(inst.Pricing) > 0 {
			pricedCount++
			for _, pricing := range inst.Pricing {
				if pricing.Reserved != nil && pricing.Reserved["spot"] > 0 {
					spotCount++
					break
				}
			}
		}
	}
	
	fmt.Printf("Instances with pricing: %d\n", pricedCount)
	fmt.Printf("Instances with spot pricing: %d\n", spotCount)
}