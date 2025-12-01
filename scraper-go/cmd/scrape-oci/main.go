package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/core"
	"github.com/oracle/oci-go-sdk/v65/identity"
	"gopkg.in/yaml.v2"
)

type OCIRegion struct {
	Name                string   `yaml:"name"`
	DisplayName         string   `yaml:"displayname"`
	AvailabilityDomains []string `yaml:"availabilitydomains"`
}

type OCIInstance struct {
	InstanceType              string                 `json:"instance_type"`
	PrettyName                string                 `json:"pretty_name"`
	VCPU                      int                    `json:"vcpu"`
	Memory                    float64                `json:"memory"`
	GPU                       *GPU                   `json:"gpu,omitempty"`
	Family                    string                 `json:"family"`
	Arch                      []string               `json:"arch"`
	Storage                   *Storage               `json:"storage,omitempty"`
	NetworkPerformance        string                 `json:"network_performance"`
	ShapeType                 string                 `json:"shape_type"`
	Processor                 string                 `json:"processor,omitempty"`
	BaselineOCPUUtilization   string                 `json:"baseline_ocpu_utilization,omitempty"`
	MaxBandwidth              string                 `json:"max_bandwidth,omitempty"`
	AvailabilityDomain        string                 `json:"availability_domain,omitempty"`
	Pricing                   map[string]interface{} `json:"pricing"`
}

type GPU struct {
	Count int    `json:"count"`
	Type  string `json:"type"`
}

type Storage struct {
	BootVolume    int    `json:"boot_volume,omitempty"`
	BlockVolume   int    `json:"block_volume,omitempty"`
	LocalStorage  int    `json:"local_storage,omitempty"`
	Type          string `json:"type,omitempty"`
}

func main() {
	ctx := context.Background()

	// Create OCI clients
	configProvider := common.DefaultConfigProvider()
	
	identityClient, err := identity.NewIdentityClientWithConfigurationProvider(configProvider)
	if err != nil {
		log.Fatalf("Failed to create identity client: %v", err)
	}

	// Get tenancy OCID
	tenancyOCID, err := configProvider.TenancyOCID()
	if err != nil {
		log.Fatalf("Failed to get tenancy OCID: %v", err)
	}

	// Get all regions
	regions, err := getRegions(ctx, identityClient, tenancyOCID)
	if err != nil {
		log.Fatalf("Failed to get regions: %v", err)
	}

	// Get all shapes across regions
	instances, err := getShapes(ctx, configProvider, regions)
	if err != nil {
		log.Fatalf("Failed to get shapes: %v", err)
	}

	// Sort instances by name
	sort.Slice(instances, func(i, j int) bool {
		return instances[i].InstanceType < instances[j].InstanceType
	})

	// Save instances
	instancesJSON, err := json.MarshalIndent(instances, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal instances: %v", err)
	}

	outputDir := "../../www/oci"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	if err := ioutil.WriteFile(filepath.Join(outputDir, "instances.json"), instancesJSON, 0644); err != nil {
		log.Fatalf("Failed to write instances.json: %v", err)
	}

	// Save regions
	regionsYAML, err := yaml.Marshal(regions)
	if err != nil {
		log.Fatalf("Failed to marshal regions: %v", err)
	}

	if err := ioutil.WriteFile("../../next/public/data/oci-regions.yaml", regionsYAML, 0644); err != nil {
		log.Fatalf("Failed to write regions.yaml: %v", err)
	}

	fmt.Printf("Successfully scraped %d OCI instance types across %d regions\n", len(instances), len(regions))
}

func getRegions(ctx context.Context, client identity.IdentityClient, tenancyOCID string) ([]OCIRegion, error) {
	var regions []OCIRegion

	// List all subscribed regions
	request := identity.ListRegionSubscriptionsRequest{
		TenancyId: &tenancyOCID,
	}

	response, err := client.ListRegionSubscriptions(ctx, request)
	if err != nil {
		// If we can't list subscriptions, fall back to known regions
		return getKnownRegions(), nil
	}

	for _, subscription := range response.Items {
		region := OCIRegion{
			Name:        *subscription.RegionName,
			DisplayName: getRegionDisplayName(*subscription.RegionName),
		}

		// Get availability domains for this region
		regionClient, err := identity.NewIdentityClientWithConfigurationProvider(
			common.DefaultConfigProvider(),
		)
		if err == nil {
			regionClient.SetRegion(*subscription.RegionName)
			adRequest := identity.ListAvailabilityDomainsRequest{
				CompartmentId: &tenancyOCID,
			}
			adResponse, err := regionClient.ListAvailabilityDomains(ctx, adRequest)
			if err == nil {
				for _, ad := range adResponse.Items {
					region.AvailabilityDomains = append(region.AvailabilityDomains, *ad.Name)
				}
			}
		}

		regions = append(regions, region)
	}

	// Sort regions by name
	sort.Slice(regions, func(i, j int) bool {
		return regions[i].Name < regions[j].Name
	})

	return regions, nil
}

func getKnownRegions() []OCIRegion {
	// Fallback list of known OCI regions
	return []OCIRegion{
		{Name: "us-ashburn-1", DisplayName: "US East (Ashburn)", AvailabilityDomains: []string{"Uocm:US-ASHBURN-AD-1", "Uocm:US-ASHBURN-AD-2", "Uocm:US-ASHBURN-AD-3"}},
		{Name: "us-phoenix-1", DisplayName: "US West (Phoenix)", AvailabilityDomains: []string{"Uocm:PHX-AD-1", "Uocm:PHX-AD-2", "Uocm:PHX-AD-3"}},
		{Name: "ca-toronto-1", DisplayName: "Canada Southeast (Toronto)", AvailabilityDomains: []string{"Uocm:CA-TORONTO-1-AD-1"}},
		{Name: "ca-montreal-1", DisplayName: "Canada Southeast (Montreal)", AvailabilityDomains: []string{"Uocm:CA-MONTREAL-1-AD-1"}},
		{Name: "sa-saopaulo-1", DisplayName: "Brazil East (Sao Paulo)", AvailabilityDomains: []string{"Uocm:SA-SAOPAULO-1-AD-1"}},
		{Name: "sa-vinhedo-1", DisplayName: "Brazil Southeast (Vinhedo)", AvailabilityDomains: []string{"Uocm:SA-VINHEDO-1-AD-1"}},
		{Name: "uk-london-1", DisplayName: "UK South (London)", AvailabilityDomains: []string{"Uocm:UK-LONDON-1-AD-1", "Uocm:UK-LONDON-1-AD-2", "Uocm:UK-LONDON-1-AD-3"}},
		{Name: "uk-cardiff-1", DisplayName: "UK West (Cardiff)", AvailabilityDomains: []string{"Uocm:UK-CARDIFF-1-AD-1"}},
		{Name: "eu-frankfurt-1", DisplayName: "Germany Central (Frankfurt)", AvailabilityDomains: []string{"Uocm:EU-FRANKFURT-1-AD-1", "Uocm:EU-FRANKFURT-1-AD-2", "Uocm:EU-FRANKFURT-1-AD-3"}},
		{Name: "eu-amsterdam-1", DisplayName: "Netherlands Northwest (Amsterdam)", AvailabilityDomains: []string{"Uocm:EU-AMSTERDAM-1-AD-1"}},
		{Name: "eu-zurich-1", DisplayName: "Switzerland North (Zurich)", AvailabilityDomains: []string{"Uocm:EU-ZURICH-1-AD-1"}},
		{Name: "ap-mumbai-1", DisplayName: "India West (Mumbai)", AvailabilityDomains: []string{"Uocm:AP-MUMBAI-1-AD-1"}},
		{Name: "ap-hyderabad-1", DisplayName: "India South (Hyderabad)", AvailabilityDomains: []string{"Uocm:AP-HYDERABAD-1-AD-1", "Uocm:AP-HYDERABAD-1-AD-2", "Uocm:AP-HYDERABAD-1-AD-3"}},
		{Name: "ap-seoul-1", DisplayName: "South Korea Central (Seoul)", AvailabilityDomains: []string{"Uocm:AP-SEOUL-1-AD-1"}},
		{Name: "ap-tokyo-1", DisplayName: "Japan East (Tokyo)", AvailabilityDomains: []string{"Uocm:AP-TOKYO-1-AD-1"}},
		{Name: "ap-osaka-1", DisplayName: "Japan Central (Osaka)", AvailabilityDomains: []string{"Uocm:AP-OSAKA-1-AD-1"}},
		{Name: "ap-sydney-1", DisplayName: "Australia East (Sydney)", AvailabilityDomains: []string{"Uocm:AP-SYDNEY-1-AD-1"}},
		{Name: "ap-melbourne-1", DisplayName: "Australia Southeast (Melbourne)", AvailabilityDomains: []string{"Uocm:AP-MELBOURNE-1-AD-1"}},
		{Name: "me-jeddah-1", DisplayName: "Saudi Arabia West (Jeddah)", AvailabilityDomains: []string{"Uocm:ME-JEDDAH-1-AD-1"}},
		{Name: "me-dubai-1", DisplayName: "UAE East (Dubai)", AvailabilityDomains: []string{"Uocm:ME-DUBAI-1-AD-1"}},
		{Name: "il-jerusalem-1", DisplayName: "Israel Central (Jerusalem)", AvailabilityDomains: []string{"Uocm:IL-JERUSALEM-1-AD-1"}},
		{Name: "ap-singapore-1", DisplayName: "Singapore", AvailabilityDomains: []string{"Uocm:AP-SINGAPORE-1-AD-1"}},
	}
}

func getRegionDisplayName(regionName string) string {
	displayNames := map[string]string{
		"us-ashburn-1":    "US East (Ashburn)",
		"us-phoenix-1":    "US West (Phoenix)",
		"ca-toronto-1":    "Canada Southeast (Toronto)",
		"ca-montreal-1":   "Canada Southeast (Montreal)",
		"sa-saopaulo-1":   "Brazil East (Sao Paulo)",
		"sa-vinhedo-1":    "Brazil Southeast (Vinhedo)",
		"uk-london-1":     "UK South (London)",
		"uk-cardiff-1":    "UK West (Cardiff)",
		"eu-frankfurt-1":  "Germany Central (Frankfurt)",
		"eu-amsterdam-1":  "Netherlands Northwest (Amsterdam)",
		"eu-zurich-1":     "Switzerland North (Zurich)",
		"ap-mumbai-1":     "India West (Mumbai)",
		"ap-hyderabad-1":  "India South (Hyderabad)",
		"ap-seoul-1":      "South Korea Central (Seoul)",
		"ap-tokyo-1":      "Japan East (Tokyo)",
		"ap-osaka-1":      "Japan Central (Osaka)",
		"ap-sydney-1":     "Australia East (Sydney)",
		"ap-melbourne-1":  "Australia Southeast (Melbourne)",
		"me-jeddah-1":     "Saudi Arabia West (Jeddah)",
		"me-dubai-1":      "UAE East (Dubai)",
		"il-jerusalem-1":  "Israel Central (Jerusalem)",
		"ap-singapore-1":  "Singapore",
	}

	if displayName, ok := displayNames[regionName]; ok {
		return displayName
	}
	return regionName
}

func getShapes(ctx context.Context, configProvider common.ConfigurationProvider, regions []OCIRegion) ([]OCIInstance, error) {
	instanceMap := make(map[string]*OCIInstance)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Process each region
	for _, region := range regions {
		wg.Add(1)
		go func(regionName string, ads []string) {
			defer wg.Done()

			// Create compute client for this region
			computeClient, err := core.NewComputeClientWithConfigurationProvider(configProvider)
			if err != nil {
				log.Printf("Failed to create compute client for region %s: %v", regionName, err)
				return
			}
			computeClient.SetRegion(regionName)

			// Get compartment ID
			compartmentID, err := configProvider.TenancyOCID()
			if err != nil {
				log.Printf("Failed to get compartment ID: %v", err)
				return
			}

			// List shapes in each availability domain
			for _, ad := range ads {
				request := core.ListShapesRequest{
					CompartmentId:      &compartmentID,
					AvailabilityDomain: &ad,
				}

				response, err := computeClient.ListShapes(ctx, request)
				if err != nil {
					// Try without availability domain
					request.AvailabilityDomain = nil
					response, err = computeClient.ListShapes(ctx, request)
					if err != nil {
						log.Printf("Failed to list shapes in region %s: %v", regionName, err)
						continue
					}
				}

				for _, shape := range response.Items {
					mu.Lock()
					if _, exists := instanceMap[*shape.Shape]; !exists {
						instance := convertShape(shape, regionName)
						instanceMap[*shape.Shape] = instance
					} else {
						// Add pricing for this region
						addOCIRegionPricing(instanceMap[*shape.Shape], regionName)
					}
					mu.Unlock()
				}
				break // Only need to query one AD as shapes are same across ADs
			}
		}(region.Name, region.AvailabilityDomains)
	}

	wg.Wait()

	// Convert map to slice
	var instances []OCIInstance
	for _, instance := range instanceMap {
		instances = append(instances, *instance)
	}

	return instances, nil
}

func convertShape(shape core.Shape, region string) *OCIInstance {
	instance := &OCIInstance{
		InstanceType:       *shape.Shape,
		PrettyName:         prettifyOCIName(*shape.Shape),
		VCPU:               int(*shape.Ocpus), // OCI OCPUs are equivalent to vCPUs
		Memory:             float64(*shape.MemoryInGBs),
		Family:             getOCIFamily(*shape.Shape),
		Arch:               []string{"x64"}, // Default, will be updated for ARM shapes
		NetworkPerformance: fmt.Sprintf("%.0f Gbps", *shape.NetworkingBandwidthInGbps),
		ShapeType:          getShapeType(*shape.Shape),
		Processor:          getProcessor(*shape.Shape),
		Pricing:            make(map[string]interface{}),
	}

	// Handle ARM instances
	if strings.Contains(*shape.Shape, ".A1") {
		instance.Arch = []string{"arm64"}
	}

	// Add GPU info if present
	if shape.Gpus != nil && *shape.Gpus > 0 {
		instance.GPU = &GPU{
			Count: *shape.Gpus,
			Type:  getGPUType(*shape.Shape),
		}
	}

	// Add storage info if present
	if shape.LocalDisksTotalSizeInGBs != nil && *shape.LocalDisksTotalSizeInGBs > 0 {
		instance.Storage = &Storage{
			LocalStorage: int(*shape.LocalDisksTotalSizeInGBs),
			Type:         "NVMe SSD",
		}
	}

	// Add bandwidth info
	if shape.NetworkingBandwidthInGbps != nil {
		instance.MaxBandwidth = fmt.Sprintf("%.0f Gbps", *shape.NetworkingBandwidthInGbps)
	}

	// Add initial pricing
	addOCIRegionPricing(instance, region)

	return instance
}

func addOCIRegionPricing(instance *OCIInstance, region string) {
	// Simplified OCI pricing structure
	// In production, you would fetch from OCI pricing API
	basePricePerOCPU := 0.0340  // Example base price per OCPU hour
	
	onDemandPrice := float64(instance.VCPU) * basePricePerOCPU
	
	regionPricing := map[string]interface{}{
		"linux": map[string]interface{}{
			"ondemand": onDemandPrice,
			"preemptible": onDemandPrice * 0.5, // Preemptible is roughly 50% cheaper
			"reserved": map[string]float64{
				"1yr": onDemandPrice * 0.7,  // 1 year reserved ~30% discount
				"3yr": onDemandPrice * 0.55, // 3 year reserved ~45% discount
			},
		},
		"windows": map[string]interface{}{
			"ondemand": onDemandPrice + (float64(instance.VCPU) * 0.015), // Windows license cost
			"reserved": map[string]float64{
				"1yr": (onDemandPrice + (float64(instance.VCPU) * 0.015)) * 0.7,
				"3yr": (onDemandPrice + (float64(instance.VCPU) * 0.015)) * 0.55,
			},
		},
	}

	instance.Pricing[region] = regionPricing
}

func getOCIFamily(shape string) string {
	switch {
	case strings.HasPrefix(shape, "VM.Standard.E"):
		return "Standard AMD"
	case strings.HasPrefix(shape, "VM.Standard3"):
		return "Standard Intel"
	case strings.HasPrefix(shape, "VM.Standard2"):
		return "Standard X7"
	case strings.HasPrefix(shape, "VM.Standard.B"):
		return "Standard X7 Burstable"
	case strings.HasPrefix(shape, "VM.Standard.A"):
		return "Standard Ampere A1"
	case strings.HasPrefix(shape, "VM.DenseIO"):
		return "Dense I/O"
	case strings.HasPrefix(shape, "VM.GPU"):
		return "GPU"
	case strings.HasPrefix(shape, "VM.HPC"):
		return "HPC"
	case strings.HasPrefix(shape, "VM.Optimized"):
		return "Optimized"
	case strings.HasPrefix(shape, "BM."):
		return "Bare Metal"
	default:
		return "Standard"
	}
}

func getProcessor(shape string) string {
	switch {
	case strings.Contains(shape, ".E4") || strings.Contains(shape, ".E3"):
		return "AMD EPYC"
	case strings.Contains(shape, ".A1"):
		return "Ampere Altra"
	case strings.Contains(shape, "GPU"):
		return "NVIDIA GPU"
	case strings.Contains(shape, "HPC"):
		return "Intel Xeon"
	default:
		return "Intel Xeon"
	}
}

func prettifyOCIName(shape string) string {
	// Convert shape name to prettier display name
	parts := strings.Split(shape, ".")
	if len(parts) < 3 {
		return shape
	}

	// VM.Standard.E4.Flex -> Standard E4 Flex
	return strings.Join(parts[1:], " ")
}

func getShapeType(shape string) string {
	switch {
	case strings.HasPrefix(shape, "VM."):
		return "Virtual Machine"
	case strings.HasPrefix(shape, "BM."):
		return "Bare Metal"
	default:
		return "Virtual Machine"
	}
}

func getGPUType(shape string) string {
	switch {
	case strings.Contains(shape, "GPU3"):
		return "NVIDIA V100"
	case strings.Contains(shape, "GPU2"):
		return "NVIDIA P100"
	case strings.Contains(shape, "GPU4"):
		return "NVIDIA A100"
	default:
		return "NVIDIA GPU"
	}
}