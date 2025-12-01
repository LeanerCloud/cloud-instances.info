package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cloud-instances/scraper/pkg/gcp"
	"github.com/cloud-instances/scraper/pkg/models"
)

func main() {
	var (
		outputDir  = flag.String("output", "../www/gcp", "Output directory for scraped data")
		regions    = flag.String("regions", "", "Comma-separated list of regions to scrape (empty for all)")
		sampleSize = flag.Int("sample", 0, "Number of instances to sample (0 for all)")
		projectID  = flag.String("project", "", "GCP project ID (uses ADC default if not specified)")
	)
	flag.Parse()

	ctx := context.Background()

	// Get project ID
	if *projectID == "" {
		*projectID = os.Getenv("GCP_PROJECT_ID")
		if *projectID == "" {
			// Try to get from gcloud config
			*projectID = "cloud-pricing-project" // Default project for public pricing
		}
	}

	fmt.Printf("Starting GCP VM scraping for project: %s\n", *projectID)

	// Parse regions
	var regionList []string
	if *regions != "" {
		for _, r := range models.ParseCommaSeparatedList(*regions) {
			regionList = append(regionList, r)
		}
		fmt.Printf("Regions specified: %v\n", regionList)
	} else {
		fmt.Println("No regions specified, will fetch all available regions...")
	}

	// Create scraper
	scraper, err := gcp.NewComputeScraperV2(ctx, *projectID)
	if err != nil {
		log.Fatalf("Failed to create scraper: %v", err)
	}

	// Create scraper config
	config := &models.ScraperConfig{
		Regions:        regionList,
		SampleSize:     *sampleSize,
		MaxConcurrency: 10,
	}

	// Scrape instances
	instances, err := scraper.ScrapeToGCPFormat(config)
	if err != nil {
		log.Fatalf("Failed to scrape instances: %v", err)
	}

	// Create output directory
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Save instances
	outputPath := filepath.Join(*outputDir, "instances.json")
	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(instances); err != nil {
		log.Fatalf("Failed to encode instances: %v", err)
	}

	fmt.Printf("\nSuccessfully scraped %d GCP instance types\n", len(instances))
	fmt.Printf("Output saved to: %s\n", outputPath)
	
	// Print sample instance for verification
	if len(instances) > 0 {
		fmt.Println("\nSample instance structure:")
		sampleData, _ := json.MarshalIndent(instances[0], "", "  ")
		fmt.Println(string(sampleData))
	}
}