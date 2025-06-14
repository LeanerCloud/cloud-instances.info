package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cloud-instances/scraper/pkg/models"
	"gopkg.in/yaml.v3"
)

type GCPFrontendInstance struct {
	InstanceType       string                           `json:"instance_type"`
	PrettyName         string                           `json:"pretty_name"`
	VCPU               int                              `json:"vcpu"`
	Memory             float64                          `json:"memory"`
	GPU                *GPU                             `json:"gpu,omitempty"`
	Family             string                           `json:"family"`
	Arch               []string                         `json:"arch"`
	Storage            *Storage                         `json:"storage,omitempty"`
	NetworkPerformance string                           `json:"network_performance"`
	Pricing            map[string]map[string]PricingOS `json:"pricing"`
}

type GPU struct {
	Count int    `json:"count"`
	Type  string `json:"type"`
}

type Storage struct {
	LocalSSD       int    `json:"local_ssd,omitempty"`
	PersistentDisk int    `json:"persistent_disk,omitempty"`
	Type           string `json:"type,omitempty"`
}

type PricingOS struct {
	OnDemand    float64            `json:"ondemand"`
	Preemptible float64            `json:"preemptible,omitempty"`
	Committed   map[string]float64 `json:"committed,omitempty"`
}

type Region struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Zones       []string `json:"zones"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: generate-gcp-data <output-dir>")
	}

	outputDir := os.Args[1]
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Generate sample data without requiring GCP authentication
	fmt.Println("Generating sample GCP Compute Engine instances...")
	instances := generateSampleGCPInstances()

	fmt.Printf("Generated %d sample GCP instances\n", len(instances))

	// Convert to frontend format
	frontendInstances := make([]GCPFrontendInstance, 0, len(instances))
	for _, instance := range instances {
		frontendInstance := convertToFrontendFormat(instance)
		frontendInstances = append(frontendInstances, frontendInstance)
	}

	// Generate instances JSON
	instancesPath := filepath.Join(outputDir, "gcp-instances.json")
	if err := writeJSON(instancesPath, frontendInstances); err != nil {
		log.Fatalf("Failed to write instances data: %v", err)
	}
	fmt.Printf("Generated %s with %d instances\n", instancesPath, len(frontendInstances))

	// Generate regions YAML
	regions := generateRegionsData()
	regionsPath := filepath.Join(outputDir, "gcp-regions.yaml")
	if err := writeYAML(regionsPath, regions); err != nil {
		log.Fatalf("Failed to write regions data: %v", err)
	}
	fmt.Printf("Generated %s with %d regions\n", regionsPath, len(regions))

	fmt.Println("✅ GCP data generation completed successfully!")
}

func generateSampleGCPInstances() []models.Instance {
	instances := []models.Instance{
		{
			InstanceType: "e2-micro",
			PrettyName:   "E2 Micro",
			Memory:       "1.0",
			VCPU:         "0.25",
			Family:       "General purpose",
			Pricing: map[string]models.RegionPricing{
				"us-central1": {OnDemand: 0.0056, Reserved: map[string]float64{"1yr": 0.0039, "3yr": 0.0028}},
				"us-east1":    {OnDemand: 0.0056, Reserved: map[string]float64{"1yr": 0.0039, "3yr": 0.0028}},
				"europe-west1": {OnDemand: 0.0061, Reserved: map[string]float64{"1yr": 0.0043, "3yr": 0.0030}},
			},
			Regions: map[string]string{"us-central1": "us-central1", "us-east1": "us-east1", "europe-west1": "europe-west1"},
		},
		{
			InstanceType: "e2-small",
			PrettyName:   "E2 Small",
			Memory:       "2.0",
			VCPU:         "1",
			Family:       "General purpose",
			Pricing: map[string]models.RegionPricing{
				"us-central1": {OnDemand: 0.0113, Reserved: map[string]float64{"1yr": 0.0079, "3yr": 0.0056}},
				"us-east1":    {OnDemand: 0.0113, Reserved: map[string]float64{"1yr": 0.0079, "3yr": 0.0056}},
				"europe-west1": {OnDemand: 0.0123, Reserved: map[string]float64{"1yr": 0.0086, "3yr": 0.0061}},
			},
			Regions: map[string]string{"us-central1": "us-central1", "us-east1": "us-east1", "europe-west1": "europe-west1"},
		},
		{
			InstanceType: "e2-medium",
			PrettyName:   "E2 Medium",
			Memory:       "4.0",
			VCPU:         "1",
			Family:       "General purpose",
			Pricing: map[string]models.RegionPricing{
				"us-central1": {OnDemand: 0.0226, Reserved: map[string]float64{"1yr": 0.0158, "3yr": 0.0113}},
				"us-east1":    {OnDemand: 0.0226, Reserved: map[string]float64{"1yr": 0.0158, "3yr": 0.0113}},
				"europe-west1": {OnDemand: 0.0246, Reserved: map[string]float64{"1yr": 0.0172, "3yr": 0.0123}},
			},
			Regions: map[string]string{"us-central1": "us-central1", "us-east1": "us-east1", "europe-west1": "europe-west1"},
		},
		{
			InstanceType: "n1-standard-1",
			PrettyName:   "N1 Standard",
			Memory:       "3.75",
			VCPU:         "1",
			Family:       "General purpose",
			Pricing: map[string]models.RegionPricing{
				"us-central1": {OnDemand: 0.0475, Reserved: map[string]float64{"1yr": 0.0333, "3yr": 0.0238}},
				"us-east1":    {OnDemand: 0.0475, Reserved: map[string]float64{"1yr": 0.0333, "3yr": 0.0238}},
				"europe-west1": {OnDemand: 0.0518, Reserved: map[string]float64{"1yr": 0.0363, "3yr": 0.0259}},
			},
			Regions: map[string]string{"us-central1": "us-central1", "us-east1": "us-east1", "europe-west1": "europe-west1"},
		},
		{
			InstanceType: "c2-standard-4",
			PrettyName:   "C2 Standard",
			Memory:       "16.0",
			VCPU:         "4",
			Family:       "Compute optimized",
			Pricing: map[string]models.RegionPricing{
				"us-central1": {OnDemand: 0.1992, Reserved: map[string]float64{"1yr": 0.1394, "3yr": 0.0996}},
				"us-east1":    {OnDemand: 0.1992, Reserved: map[string]float64{"1yr": 0.1394, "3yr": 0.0996}},
				"europe-west1": {OnDemand: 0.2171, Reserved: map[string]float64{"1yr": 0.1520, "3yr": 0.1086}},
			},
			Regions: map[string]string{"us-central1": "us-central1", "us-east1": "us-east1", "europe-west1": "europe-west1"},
		},
	}
	return instances
}

func convertToFrontendFormat(instance models.Instance) GCPFrontendInstance {
	// Parse memory as float
	var memory float64
	if _, err := fmt.Sscanf(instance.Memory, "%f", &memory); err != nil {
		memory = 0
	}

	// Parse vCPU count
	var vcpu int
	if _, err := fmt.Sscanf(string(instance.VCPU), "%d", &vcpu); err != nil {
		vcpu = 0
	}

	// Convert pricing
	pricing := make(map[string]map[string]PricingOS)
	for region, regionPricing := range instance.Pricing {
		pricing[region] = make(map[string]PricingOS)
		
		// Linux pricing - convert from RegionPricing to PricingOS
		linuxPricing := PricingOS{
			OnDemand:    regionPricing.OnDemand,
			Preemptible: regionPricing.OnDemand * 0.2, // Preemptible is ~20% of on-demand
			Committed:   make(map[string]float64),
		}
		
		// Add committed use discounts
		for term, price := range regionPricing.Reserved {
			linuxPricing.Committed[term] = price
		}
		
		pricing[region]["linux"] = linuxPricing
		
		// Windows pricing (typically 2x Linux pricing)
		windowsPricing := PricingOS{
			OnDemand:  regionPricing.OnDemand * 2.0,
			Committed: make(map[string]float64),
		}
		
		for term, price := range regionPricing.Reserved {
			windowsPricing.Committed[term] = price * 2.0
		}
		
		pricing[region]["windows"] = windowsPricing
	}

	return GCPFrontendInstance{
		InstanceType:       instance.InstanceType,
		PrettyName:         instance.PrettyName,
		VCPU:               vcpu,
		Memory:             memory,
		Family:             instance.Family,
		Arch:               []string{"x64"}, // Most GCP instances are x64
		NetworkPerformance: instance.NetworkPerformance,
		Pricing:            pricing,
	}
}

func generateRegionsData() []Region {
	return []Region{
		{
			Name:        "us-central1",
			DisplayName: "US Central (Iowa)",
			Zones:       []string{"us-central1-a", "us-central1-b", "us-central1-c", "us-central1-f"},
		},
		{
			Name:        "us-east1",
			DisplayName: "US East (South Carolina)",
			Zones:       []string{"us-east1-b", "us-east1-c", "us-east1-d"},
		},
		{
			Name:        "us-east4",
			DisplayName: "US East (Northern Virginia)",
			Zones:       []string{"us-east4-a", "us-east4-b", "us-east4-c"},
		},
		{
			Name:        "us-west1",
			DisplayName: "US West (Oregon)",
			Zones:       []string{"us-west1-a", "us-west1-b", "us-west1-c"},
		},
		{
			Name:        "us-west2",
			DisplayName: "US West (Los Angeles)",
			Zones:       []string{"us-west2-a", "us-west2-b", "us-west2-c"},
		},
		{
			Name:        "us-west3",
			DisplayName: "US West (Salt Lake City)",
			Zones:       []string{"us-west3-a", "us-west3-b", "us-west3-c"},
		},
		{
			Name:        "us-west4",
			DisplayName: "US West (Las Vegas)",
			Zones:       []string{"us-west4-a", "us-west4-b", "us-west4-c"},
		},
		{
			Name:        "europe-west1",
			DisplayName: "Europe West (Belgium)",
			Zones:       []string{"europe-west1-b", "europe-west1-c", "europe-west1-d"},
		},
		{
			Name:        "europe-west2",
			DisplayName: "Europe West (London)",
			Zones:       []string{"europe-west2-a", "europe-west2-b", "europe-west2-c"},
		},
		{
			Name:        "europe-west3",
			DisplayName: "Europe West (Frankfurt)",
			Zones:       []string{"europe-west3-a", "europe-west3-b", "europe-west3-c"},
		},
		{
			Name:        "asia-east1",
			DisplayName: "Asia East (Taiwan)",
			Zones:       []string{"asia-east1-a", "asia-east1-b", "asia-east1-c"},
		},
		{
			Name:        "asia-southeast1",
			DisplayName: "Asia Southeast (Singapore)",
			Zones:       []string{"asia-southeast1-a", "asia-southeast1-b", "asia-southeast1-c"},
		},
	}
}

func writeJSON(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func writeYAML(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()
	return encoder.Encode(data)
}