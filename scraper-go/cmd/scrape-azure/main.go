package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloud-instances/scraper/pkg/azure"
	"github.com/cloud-instances/scraper/pkg/models"
	"github.com/spf13/cobra"
)

var (
	outputDir      string
	regions        []string
	maxConcurrency int
	sampleSize     int
	subscriptionID string
)

var rootCmd = &cobra.Command{
	Use:   "scrape-azure",
	Short: "Scrape Azure VM instance data",
	RunE:  run,
}

func init() {
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "../www/azure", "Output directory for JSON files")
	rootCmd.Flags().StringSliceVarP(&regions, "regions", "r", []string{}, "Regions to scrape (empty for all)")
	rootCmd.Flags().IntVarP(&maxConcurrency, "concurrency", "c", 10, "Maximum concurrent operations")
	rootCmd.Flags().IntVar(&sampleSize, "sample", 0, "Sample size for testing (0 = all instances)")
	rootCmd.Flags().StringVarP(&subscriptionID, "subscription", "s", "", "Azure subscription ID (defaults to AZURE_SUBSCRIPTION_ID env var)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Get subscription ID
	if subscriptionID == "" {
		subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
		if subscriptionID == "" {
			return fmt.Errorf("Azure subscription ID required: set AZURE_SUBSCRIPTION_ID or use --subscription flag")
		}
	}

	fmt.Printf("Using Azure subscription: %s\n", subscriptionID)

	ctx := context.Background()
	
	// Create real Azure scraper
	scraper, err := azure.NewRealComputeScraper(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to create Azure scraper: %v", err)
	}

	// If no regions specified, use all Azure regions
	if len(regions) == 0 {
		fmt.Println("No regions specified, using all Azure regions...")
		regions = getAllAzureRegions()
	}

	// Configure scraping
	config := &models.ScraperConfig{
		Regions:        regions,
		MaxConcurrency: maxConcurrency,
		OutputPath:     outputDir,
		Service:        "azure",
		SampleSize:     sampleSize,
	}

	fmt.Printf("Starting Azure VM scraping across %d regions...\n", len(regions))
	instances, err := scraper.Scrape(config)
	if err != nil {
		return fmt.Errorf("failed to scrape Azure instances: %v", err)
	}

	fmt.Printf("\nSuccessfully scraped %d Azure VM types\n", len(instances))
	
	// Convert to frontend format
	fmt.Println("Converting to frontend format...")
	frontendInstances := convertToFrontendFormat(instances)
	
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Save instances JSON with the name expected by compressData.ts
	instancesPath := filepath.Join(outputDir, "instances.json")
	if err := saveJSON(instancesPath, frontendInstances); err != nil {
		return fmt.Errorf("failed to save instances: %v", err)
	}
	fmt.Printf("Saved instances to: %s\n", instancesPath)

	// Save regions metadata to the meta directory
	metaDir := filepath.Join(filepath.Dir(outputDir), "meta")
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		return fmt.Errorf("failed to create meta directory: %v", err)
	}
	
	regionsPath := filepath.Join(metaDir, "regions_azure2.yaml")
	if err := saveRegionsYAML(regionsPath, instances); err != nil {
		return fmt.Errorf("failed to save regions: %v", err)
	}
	fmt.Printf("Saved regions to: %s\n", regionsPath)

	fmt.Println("\nAzure scraping completed successfully!")
	fmt.Println("Next step: Run 'npm run init' in the next directory to generate msgpack files")
	
	return nil
}

// getAllAzureRegions returns all Azure regions
func getAllAzureRegions() []string {
	// From the error message, these are the supported locations
	return []string{
		"eastus", "eastus2", "westus", "centralus", "northcentralus", 
		"southcentralus", "northeurope", "westeurope", "eastasia", "southeastasia", 
		"japaneast", "japanwest", "australiaeast", "australiasoutheast", "australiacentral", 
		"brazilsouth", "southindia", "centralindia", "westindia", "canadacentral", 
		"canadaeast", "westus2", "westcentralus", "uksouth", "ukwest", 
		"koreacentral", "koreasouth", "francecentral", "southafricanorth", "uaenorth", 
		"switzerlandnorth", "germanywestcentral", "norwayeast", "jioindiawest", "westus3", 
		"swedencentral", "qatarcentral", "polandcentral", "italynorth", "israelcentral", 
		"spaincentral", "mexicocentral", "malaysiawest", "newzealandnorth", "indonesiacentral", 
		"chilecentral",
	}
}

// Frontend format structures matching Vantage exactly
type AzureFrontendInstance struct {
	ACU                  int                      `json:"ACU"`
	GPU                  string                   `json:"GPU"`
	AcceleratedNetworking bool                    `json:"accelerated_networking"`
	Arch                 []string                 `json:"arch"`
	AvailabilityZones    map[string]interface{}   `json:"availability_zones"`
	CachedDisk           int                      `json:"cached_disk"`
	CapacitySupport      bool                     `json:"capacity_support"`
	Category             string                   `json:"category"`
	Confidential         interface{}              `json:"confidential"`
	Encryption           interface{}              `json:"encryption"`
	Family               string                   `json:"family"`
	Hibernation          interface{}              `json:"hibernation"`
	HypervGenerations    string                   `json:"hyperv_generations"`
	InstanceType         string                   `json:"instance_type"`
	IOPS                 interface{}              `json:"iops"`
	LowPriority          bool                     `json:"low_priority"`
	Memory               float64                  `json:"memory"`
	MemoryMaintenance    bool                     `json:"memory_maintenance"`
	PremiumIO            bool                     `json:"premium_io"`
	PrettyName           string                   `json:"pretty_name"`
	PrettyNameAzure      string                   `json:"pretty_name_azure"`
	Pricing              map[string]AzurePricing  `json:"pricing"`
	RDMA                 bool                     `json:"rdma"`
	ReadIO               int                      `json:"read_io"`
	Size                 interface{}              `json:"size"`
	Storage              interface{}              `json:"storage"`
	TrustedLaunch        interface{}              `json:"trusted_launch"`
	UltraSSD             bool                     `json:"ultra_ssd"`
	UncachedDisk         int                      `json:"uncached_disk"`
	UncachedDiskIO       interface{}              `json:"uncached_disk_io"`
	VCPU                 int                      `json:"vcpu"`
	VCPUsAvailable       interface{}              `json:"vcpus_available"`
	VCPUsPerCore         interface{}              `json:"vcpus_percore"`
	VMDeployment         interface{}              `json:"vm_deployment"`
	WriteIO              int                      `json:"write_io"`
}

type AzurePricing struct {
	Linux   AzureOSPricing `json:"linux"`
	Windows AzureOSPricing `json:"windows"`
}

type AzureOSPricing struct {
	Basic         float64            `json:"basic,omitempty"`
	BasicSpot     float64            `json:"basic-spot,omitempty"`
	OnDemand      float64            `json:"ondemand,omitempty"`
	LowPriority   float64            `json:"lowpriority,omitempty"`
	HybridBenefit float64            `json:"hybridbenefit,omitempty"`
	Reserved      map[string]float64 `json:"reserved,omitempty"`
	SpotMin       float64            `json:"spot_min,omitempty"`
}

func convertToFrontendFormat(instances []azure.AzureInstance) []AzureFrontendInstance {
	result := make([]AzureFrontendInstance, 0, len(instances))
	
	for _, inst := range instances {
		// Convert VCPU (handle both int and *int)
		vcpu := 0
		if v, ok := inst.VCPU.(int); ok {
			vcpu = v
		} else if v, ok := inst.VCPU.(*int); ok && v != nil {
			vcpu = *v
		}
		
		// Convert pricing to frontend format
		pricing := make(map[string]AzurePricing)
		for region, azurePricing := range inst.Pricing {
			pricing[region] = AzurePricing{
				Linux: AzureOSPricing{
					Basic:      azurePricing.Linux.Basic,
					BasicSpot:  azurePricing.Linux.BasicSpot,
					OnDemand:   azurePricing.Linux.OnDemand,
					LowPriority: azurePricing.Linux.LowPriority,
					Reserved:   azurePricing.Linux.Reserved,
					SpotMin:    azurePricing.Linux.SpotMin,
				},
				Windows: AzureOSPricing{
					Basic:         azurePricing.Windows.Basic,
					BasicSpot:     azurePricing.Windows.BasicSpot,
					OnDemand:      azurePricing.Windows.OnDemand,
					LowPriority:   azurePricing.Windows.LowPriority,
					HybridBenefit: azurePricing.Windows.HybridBenefit,
					Reserved:      azurePricing.Windows.Reserved,
					SpotMin:       azurePricing.Windows.SpotMin,
				},
			}
		}
		
		
		// Convert architecture
		arch := inst.Arch
		if len(arch) == 0 {
			arch = []string{"x64"}
		}
		
		frontend := AzureFrontendInstance{
			ACU:                  inst.ACU,
			GPU:                  inst.GPU, // Keep as string to match Vantage
			AcceleratedNetworking: inst.AcceleratedNetworking,
			Arch:                 arch,
			AvailabilityZones:    inst.AvailabilityZones, // Use from original data
			CachedDisk:           inst.CachedDisk,
			CapacitySupport:      inst.CapacitySupport,
			Category:             inst.Category,
			Confidential:         inst.Confidential,
			Encryption:           inst.Encryption,
			Family:               inst.Family,
			Hibernation:          inst.Hibernation,
			HypervGenerations:    inst.HypervGenerations,
			InstanceType:         inst.InstanceType,
			IOPS:                 inst.IOPS,
			LowPriority:          inst.LowPriority,
			Memory:               inst.Memory,
			MemoryMaintenance:    inst.MemoryMaintenance,
			PremiumIO:            inst.PremiumIO,
			PrettyName:           inst.PrettyName,
			PrettyNameAzure:      inst.PrettyNameAzure,
			Pricing:              pricing,
			RDMA:                 inst.RDMA,
			ReadIO:               0, // Default from Vantage reference
			Size:                 inst.Size,
			Storage:              inst.Storage,
			TrustedLaunch:        inst.TrustedLaunch,
			UltraSSD:             inst.UltraSSD,
			UncachedDisk:         inst.UncachedDisk,
			UncachedDiskIO:       inst.UncachedDiskIO,
			VCPU:                 vcpu,
			VCPUsAvailable:       inst.VCPUsAvailable,
			VCPUsPerCore:         inst.VCPUsPerCore,
			VMDeployment:         inst.VMDeployment,
			WriteIO:              0, // Default from Vantage reference
		}
		
		result = append(result, frontend)
	}
	
	return result
}

func saveJSON(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func saveRegionsYAML(path string, instances []azure.AzureInstance) error {
	// Build regions map from instances
	regionsMap := make(map[string]string)
	for _, inst := range instances {
		for region := range inst.AvailabilityZones {
			if _, exists := regionsMap[region]; !exists {
				// Convert region code to display name
				regionsMap[region] = regionCodeToDisplayName(region)
			}
		}
	}
	
	// Write as simple YAML format
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	
	for region, displayName := range regionsMap {
		if _, err := fmt.Fprintf(file, "%s: %s\n", region, displayName); err != nil {
			return err
		}
	}
	
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func regionCodeToDisplayName(region string) string {
	displayNames := map[string]string{
		"eastus":             "East US",
		"eastus2":            "East US 2",
		"westus":             "West US",
		"westus2":            "West US 2",
		"westus3":            "West US 3",
		"centralus":          "Central US",
		"northcentralus":     "North Central US",
		"southcentralus":     "South Central US",
		"westcentralus":      "West Central US",
		"canadacentral":      "Canada Central",
		"canadaeast":         "Canada East",
		"brazilsouth":        "Brazil South",
		"northeurope":        "North Europe",
		"westeurope":         "West Europe",
		"uksouth":            "UK South",
		"ukwest":             "UK West",
		"francecentral":      "France Central",
		"germanywestcentral": "Germany West Central",
		"norwayeast":         "Norway East",
		"switzerlandnorth":   "Switzerland North",
		"swedencentral":      "Sweden Central",
		"polandcentral":      "Poland Central",
		"eastasia":           "East Asia",
		"southeastasia":      "Southeast Asia",
		"japaneast":          "Japan East",
		"japanwest":          "Japan West",
		"koreacentral":       "Korea Central",
		"koreasouth":         "Korea South",
		"centralindia":       "Central India",
		"southindia":         "South India",
		"westindia":          "West India",
		"jioindiawest":       "Jio India West",
		"australiaeast":      "Australia East",
		"australiasoutheast": "Australia Southeast",
		"australiacentral":   "Australia Central",
		"uaenorth":           "UAE North",
		"southafricanorth":   "South Africa North",
		"qatarcentral":       "Qatar Central",
		"israelcentral":      "Israel Central",
		"italynorth":         "Italy North",
		"spaincentral":       "Spain Central",
		"mexicocentral":      "Mexico Central",
		"malaysiawest":       "Malaysia West",
		"newzealandnorth":    "New Zealand North",
		"indonesiacentral":   "Indonesia Central",
		"chilecentral":       "Chile Central",
	}
	
	if displayName, ok := displayNames[region]; ok {
		return displayName
	}
	
	// Default: capitalize first letter of each word
	parts := strings.Split(region, "")
	if len(parts) > 0 {
		parts[0] = strings.ToUpper(parts[0])
	}
	return strings.Join(parts, "")
}