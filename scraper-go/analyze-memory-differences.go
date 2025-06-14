package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"time"
)

// VantageAzureInstance represents the structure from instances.vantage.sh/azure/instances.json
type VantageAzureInstance struct {
	InstanceType string  `json:"instance_type"`
	Memory       float64 `json:"memory"`
	VCPU         interface{} `json:"vcpu"`
	Category     string  `json:"category"`
	PrettyName   string  `json:"pretty_name"`
}

type MemoryDifference struct {
	InstanceType   string
	VantageMemory  float64
	OurMemory      float64
	Difference     float64
	DiffPercent    float64
}

func fetchVantageAzureData() ([]VantageAzureInstance, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	
	resp, err := client.Get("https://instances.vantage.sh/azure/instances.json")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Vantage Azure data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Vantage API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var instances []VantageAzureInstance
	if err := json.Unmarshal(body, &instances); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return instances, nil
}

func main() {
	// Load Vantage data
	vantageInstances, err := fetchVantageAzureData()
	if err != nil {
		fmt.Printf("Error fetching Vantage data: %v\n", err)
		os.Exit(1)
	}

	// Load our data
	file, err := os.Open("azure-instances-real.json")
	if err != nil {
		fmt.Printf("Error opening our data file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	body, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading our data: %v\n", err)
		os.Exit(1)
	}

	var ourInstances []VantageAzureInstance
	if err := json.Unmarshal(body, &ourInstances); err != nil {
		fmt.Printf("Error unmarshaling our data: %v\n", err)
		os.Exit(1)
	}

	// Create maps for easier lookup
	vantageMap := make(map[string]VantageAzureInstance)
	for _, inst := range vantageInstances {
		vantageMap[inst.InstanceType] = inst
	}

	ourMap := make(map[string]VantageAzureInstance)
	for _, inst := range ourInstances {
		ourMap[inst.InstanceType] = inst
	}

	// Find memory differences
	var memoryDiffs []MemoryDifference
	
	for instanceType, ourInst := range ourMap {
		if vantageInst, exists := vantageMap[instanceType]; exists {
			if vantageInst.Memory != ourInst.Memory && vantageInst.Memory > 0 && ourInst.Memory > 0 {
				diff := ourInst.Memory - vantageInst.Memory
				diffPercent := (diff / vantageInst.Memory) * 100
				
				memoryDiffs = append(memoryDiffs, MemoryDifference{
					InstanceType:   instanceType,
					VantageMemory:  vantageInst.Memory,
					OurMemory:      ourInst.Memory,
					Difference:     diff,
					DiffPercent:    diffPercent,
				})
			}
		}
	}

	// Sort by absolute difference
	sort.Slice(memoryDiffs, func(i, j int) bool {
		absI := memoryDiffs[i].Difference
		if absI < 0 {
			absI = -absI
		}
		absJ := memoryDiffs[j].Difference
		if absJ < 0 {
			absJ = -absJ
		}
		return absI > absJ
	})

	fmt.Printf("=== MEMORY DIFFERENCES ANALYSIS ===\n")
	fmt.Printf("Total matching instances: %d\n", len(ourMap))
	fmt.Printf("Instances with memory differences: %d\n", len(memoryDiffs))
	fmt.Printf("Memory accuracy: %.1f%%\n\n", float64(len(ourMap)-len(memoryDiffs))/float64(len(ourMap))*100)

	if len(memoryDiffs) == 0 {
		fmt.Println("No memory differences found!")
		return
	}

	fmt.Printf("%-15s | %10s | %10s | %10s | %8s\n", "Instance", "Vantage", "Ours", "Diff (GB)", "Diff %")
	fmt.Printf("%-15s-|-%10s-|-%10s-|-%10s-|-%8s\n", "---------------", "----------", "----------", "----------", "--------")

	for _, diff := range memoryDiffs {
		fmt.Printf("%-15s | %10.1f | %10.1f | %+10.1f | %+7.1f%%\n",
			diff.InstanceType,
			diff.VantageMemory,
			diff.OurMemory,
			diff.Difference,
			diff.DiffPercent)
	}

	// Analysis
	fmt.Printf("\n=== ANALYSIS ===\n")
	
	// Group by pattern
	patterns := make(map[string]int)
	for _, diff := range memoryDiffs {
		if diff.Difference > 0 {
			patterns["Our memory > Vantage"]++
		} else {
			patterns["Our memory < Vantage"]++
		}
	}
	
	fmt.Printf("Patterns:\n")
	for pattern, count := range patterns {
		fmt.Printf("- %s: %d instances\n", pattern, count)
	}

	// Check for common differences
	fmt.Printf("\nCommon differences:\n")
	diffCounts := make(map[float64]int)
	for _, diff := range memoryDiffs {
		// Round to nearest GB for grouping
		rounded := float64(int(diff.Difference + 0.5))
		diffCounts[rounded]++
	}
	
	for diffVal, count := range diffCounts {
		if count > 1 {
			fmt.Printf("- %+.0fGB difference: %d instances\n", diffVal, count)
		}
	}
}