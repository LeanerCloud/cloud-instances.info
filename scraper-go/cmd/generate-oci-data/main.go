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

type OCIFrontendInstance struct {
	InstanceType       string                           `json:"instance_type"`
	PrettyName         string                           `json:"pretty_name"`
	VCPU               int                              `json:"vcpu"`
	Memory             float64                          `json:"memory"`
	GPU                *GPU                             `json:"gpu,omitempty"`
	Family             string                           `json:"family"`
	Arch               []string                         `json:"arch"`
	Storage            *Storage                         `json:"storage,omitempty"`
	NetworkPerformance string                           `json:"network_performance"`
	ShapeType          string                           `json:"shape_type"`
	Processor          string                           `json:"processor,omitempty"`
	Pricing            map[string]map[string]PricingOS `json:"pricing"`
}

type GPU struct {
	Count int    `json:"count"`
	Type  string `json:"type"`
}

type Storage struct {
	BootVolume   int    `json:"boot_volume,omitempty"`
	BlockVolume  int    `json:"block_volume,omitempty"`
	LocalStorage int    `json:"local_storage,omitempty"`
	Type         string `json:"type,omitempty"`
}

type PricingOS struct {
	OnDemand float64            `json:"ondemand"`
	Reserved map[string]float64 `json:"reserved,omitempty"`
}

type Region struct {
	Name             string   `json:"name"`
	DisplayName      string   `json:"display_name"`
	AvailabilityDomains []string `json:"availability_domains"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: generate-oci-data <output-dir>")
	}

	outputDir := os.Args[1]
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Generate sample data without requiring OCI authentication
	fmt.Println("Generating sample OCI Compute instances...")
	instances := generateSampleOCIInstances()

	fmt.Printf("Generated %d sample OCI instances\n", len(instances))

	// Convert to frontend format
	frontendInstances := make([]OCIFrontendInstance, 0, len(instances))
	for _, instance := range instances {
		frontendInstance := convertToFrontendFormat(instance)
		frontendInstances = append(frontendInstances, frontendInstance)
	}

	// Generate instances JSON
	instancesPath := filepath.Join(outputDir, "oci-instances.json")
	if err := writeJSON(instancesPath, frontendInstances); err != nil {
		log.Fatalf("Failed to write instances data: %v", err)
	}
	fmt.Printf("Generated %s with %d instances\n", instancesPath, len(frontendInstances))

	// Generate regions YAML
	regions := generateRegionsData()
	regionsPath := filepath.Join(outputDir, "oci-regions.yaml")
	if err := writeYAML(regionsPath, regions); err != nil {
		log.Fatalf("Failed to write regions data: %v", err)
	}
	fmt.Printf("Generated %s with %d regions\n", regionsPath, len(regions))

	fmt.Println("✅ OCI data generation completed successfully!")
}

func generateSampleOCIInstances() []models.Instance {
	instances := []models.Instance{
		{
			InstanceType: "VM.Standard2.1",
			PrettyName:   "VM Standard2.1",
			Memory:       "15.0",
			VCPU:         "1",
			Family:       "General purpose",
			Pricing: map[string]models.RegionPricing{
				"us-ashburn-1":   {OnDemand: 0.0464, Reserved: map[string]float64{"1yr": 0.0325, "3yr": 0.0232}},
				"us-phoenix-1":   {OnDemand: 0.0464, Reserved: map[string]float64{"1yr": 0.0325, "3yr": 0.0232}},
				"eu-frankfurt-1": {OnDemand: 0.0510, Reserved: map[string]float64{"1yr": 0.0357, "3yr": 0.0255}},
			},
			Regions: map[string]string{"us-ashburn-1": "us-ashburn-1", "us-phoenix-1": "us-phoenix-1", "eu-frankfurt-1": "eu-frankfurt-1"},
		},
		{
			InstanceType: "VM.Standard2.2",
			PrettyName:   "VM Standard2.2",
			Memory:       "30.0",
			VCPU:         "2",
			Family:       "General purpose",
			Pricing: map[string]models.RegionPricing{
				"us-ashburn-1":   {OnDemand: 0.0928, Reserved: map[string]float64{"1yr": 0.0650, "3yr": 0.0464}},
				"us-phoenix-1":   {OnDemand: 0.0928, Reserved: map[string]float64{"1yr": 0.0650, "3yr": 0.0464}},
				"eu-frankfurt-1": {OnDemand: 0.1021, Reserved: map[string]float64{"1yr": 0.0715, "3yr": 0.0511}},
			},
			Regions: map[string]string{"us-ashburn-1": "us-ashburn-1", "us-phoenix-1": "us-phoenix-1", "eu-frankfurt-1": "eu-frankfurt-1"},
		},
		{
			InstanceType: "VM.Standard.E3.Flex",
			PrettyName:   "VM Standard E3 Flex",
			Memory:       "16.0",
			VCPU:         "1",
			Family:       "General purpose",
			Pricing: map[string]models.RegionPricing{
				"us-ashburn-1":   {OnDemand: 0.0255, Reserved: map[string]float64{"1yr": 0.0179, "3yr": 0.0128}},
				"us-phoenix-1":   {OnDemand: 0.0255, Reserved: map[string]float64{"1yr": 0.0179, "3yr": 0.0128}},
				"eu-frankfurt-1": {OnDemand: 0.0281, Reserved: map[string]float64{"1yr": 0.0197, "3yr": 0.0141}},
			},
			Regions: map[string]string{"us-ashburn-1": "us-ashburn-1", "us-phoenix-1": "us-phoenix-1", "eu-frankfurt-1": "eu-frankfurt-1"},
		},
		{
			InstanceType: "BM.Standard2.52",
			PrettyName:   "BM Standard2.52",
			Memory:       "768.0",
			VCPU:         "52",
			Family:       "Bare Metal",
			Pricing: map[string]models.RegionPricing{
				"us-ashburn-1":   {OnDemand: 2.4128, Reserved: map[string]float64{"1yr": 1.6890, "3yr": 1.2064}},
				"us-phoenix-1":   {OnDemand: 2.4128, Reserved: map[string]float64{"1yr": 1.6890, "3yr": 1.2064}},
				"eu-frankfurt-1": {OnDemand: 2.6541, Reserved: map[string]float64{"1yr": 1.8579, "3yr": 1.3270}},
			},
			Regions: map[string]string{"us-ashburn-1": "us-ashburn-1", "us-phoenix-1": "us-phoenix-1", "eu-frankfurt-1": "eu-frankfurt-1"},
		},
		{
			InstanceType: "VM.Standard.E4.Flex",
			PrettyName:   "VM Standard E4 Flex",
			Memory:       "64.0",
			VCPU:         "4",
			Family:       "General purpose",
			Pricing: map[string]models.RegionPricing{
				"us-ashburn-1":   {OnDemand: 0.1020, Reserved: map[string]float64{"1yr": 0.0714, "3yr": 0.0510}},
				"us-phoenix-1":   {OnDemand: 0.1020, Reserved: map[string]float64{"1yr": 0.0714, "3yr": 0.0510}},
				"eu-frankfurt-1": {OnDemand: 0.1122, Reserved: map[string]float64{"1yr": 0.0785, "3yr": 0.0561}},
			},
			Regions: map[string]string{"us-ashburn-1": "us-ashburn-1", "us-phoenix-1": "us-phoenix-1", "eu-frankfurt-1": "eu-frankfurt-1"},
		},
	}
	return instances
}

func convertToFrontendFormat(instance models.Instance) OCIFrontendInstance {
	// Parse memory as float
	var memory float64
	if _, err := fmt.Sscanf(instance.Memory, "%f", &memory); err != nil {
		memory = 0
	}

	// Parse vCPU count (OCI uses OCPUs)
	var vcpu int
	if _, err := fmt.Sscanf(string(instance.VCPU), "%d", &vcpu); err != nil {
		vcpu = 0
	}

	// Determine shape type from instance type
	shapeType := "VM.Standard"
	if len(instance.InstanceType) > 0 {
		if instance.InstanceType[0:2] == "BM" {
			shapeType = "BM.Standard"
		} else if instance.InstanceType[0:2] == "VM" {
			shapeType = "VM.Standard"
		}
	}

	// Convert pricing
	pricing := make(map[string]map[string]PricingOS)
	for region, regionPricing := range instance.Pricing {
		pricing[region] = make(map[string]PricingOS)
		
		// Linux pricing - convert from RegionPricing to PricingOS
		linuxPricing := PricingOS{
			OnDemand: regionPricing.OnDemand,
			Reserved: make(map[string]float64),
		}
		
		// Add reserved pricing
		for term, price := range regionPricing.Reserved {
			linuxPricing.Reserved[term] = price
		}
		
		pricing[region]["linux"] = linuxPricing
		
		// Windows pricing (typically higher than Linux)
		windowsPricing := PricingOS{
			OnDemand: regionPricing.OnDemand * 1.5,
			Reserved: make(map[string]float64),
		}
		
		for term, price := range regionPricing.Reserved {
			windowsPricing.Reserved[term] = price * 1.5
		}
		
		pricing[region]["windows"] = windowsPricing
	}

	return OCIFrontendInstance{
		InstanceType:       instance.InstanceType,
		PrettyName:         instance.PrettyName,
		VCPU:               vcpu,
		Memory:             memory,
		Family:             instance.Family,
		Arch:               []string{"x64"}, // Most OCI instances are x64
		NetworkPerformance: instance.NetworkPerformance,
		ShapeType:          shapeType,
		Processor:          "Intel Xeon", // Default processor type
		Pricing:            pricing,
	}
}

func generateRegionsData() []Region {
	return []Region{
		{
			Name:        "us-ashburn-1",
			DisplayName: "US East (Ashburn)",
			AvailabilityDomains: []string{"Uocm:US-ASHBURN-AD-1", "Uocm:US-ASHBURN-AD-2", "Uocm:US-ASHBURN-AD-3"},
		},
		{
			Name:        "us-phoenix-1",
			DisplayName: "US West (Phoenix)",
			AvailabilityDomains: []string{"Uocm:PHX-AD-1", "Uocm:PHX-AD-2", "Uocm:PHX-AD-3"},
		},
		{
			Name:        "ca-central-1",
			DisplayName: "Canada Central (Toronto)",
			AvailabilityDomains: []string{"Uocm:CA-CENTRAL-1-AD-1"},
		},
		{
			Name:        "eu-frankfurt-1",
			DisplayName: "Germany Central (Frankfurt)",
			AvailabilityDomains: []string{"Uocm:EU-FRANKFURT-1-AD-1", "Uocm:EU-FRANKFURT-1-AD-2", "Uocm:EU-FRANKFURT-1-AD-3"},
		},
		{
			Name:        "eu-zurich-1",
			DisplayName: "Switzerland North (Zurich)",
			AvailabilityDomains: []string{"Uocm:EU-ZURICH-1-AD-1"},
		},
		{
			Name:        "uk-london-1",
			DisplayName: "UK South (London)",
			AvailabilityDomains: []string{"Uocm:UK-LONDON-1-AD-1", "Uocm:UK-LONDON-1-AD-2", "Uocm:UK-LONDON-1-AD-3"},
		},
		{
			Name:        "ap-tokyo-1",
			DisplayName: "Japan East (Tokyo)",
			AvailabilityDomains: []string{"Uocm:AP-TOKYO-1-AD-1"},
		},
		{
			Name:        "ap-osaka-1",
			DisplayName: "Japan Central (Osaka)",
			AvailabilityDomains: []string{"Uocm:AP-OSAKA-1-AD-1"},
		},
		{
			Name:        "ap-mumbai-1",
			DisplayName: "India West (Mumbai)",
			AvailabilityDomains: []string{"Uocm:AP-MUMBAI-1-AD-1"},
		},
		{
			Name:        "ap-seoul-1",
			DisplayName: "South Korea Central (Seoul)",
			AvailabilityDomains: []string{"Uocm:AP-SEOUL-1-AD-1"},
		},
		{
			Name:        "ap-sydney-1",
			DisplayName: "Australia East (Sydney)",
			AvailabilityDomains: []string{"Uocm:AP-SYDNEY-1-AD-1"},
		},
		{
			Name:        "sa-saopaulo-1",
			DisplayName: "Brazil East (Sao Paulo)",
			AvailabilityDomains: []string{"Uocm:SA-SAOPAULO-1-AD-1"},
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