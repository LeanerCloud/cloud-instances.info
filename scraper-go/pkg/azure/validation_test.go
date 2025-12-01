package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/cloud-instances/scraper/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getAzureSubscriptionID tries to get subscription ID from environment or Azure CLI
func getAzureSubscriptionID() string {
	// First try environment variable
	if subID := os.Getenv("AZURE_SUBSCRIPTION_ID"); subID != "" {
		return subID
	}

	// Try Azure CLI
	cmd := exec.Command("az", "account", "show", "--query", "id", "-o", "tsv")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	subscriptionID := strings.TrimSpace(string(output))
	if subscriptionID != "" && subscriptionID != "null" {
		return subscriptionID
	}

	return ""
}

// TestValidationAgainstVantageData tests our Azure scraper output against Vantage production data
func TestValidationAgainstVantageData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping validation test in short mode")
	}

	// Try to get subscription ID from Azure CLI or environment
	subscriptionID := getAzureSubscriptionID()
	if subscriptionID == "" {
		fmt.Println("No Azure credentials found, using cached data for comparison")
		// Fall back to using existing real scraper output file
		referenceData, ourData, err := loadComparisonData()
		if err != nil {
			t.Skipf("Could not load comparison data: %v", err)
			return
		}
		
		compareAzureData(t, referenceData, ourData)
		return
	}

	fmt.Printf("Found Azure subscription ID: %s\n", subscriptionID[:8]+"...")

	// Fetch Vantage Azure data
	azureInstances, err := fetchAzureData()
	if err != nil {
		t.Skipf("Could not fetch Vantage Azure data: %v", err)
		return
	}

	if len(azureInstances) == 0 {
		t.Skip("No Vantage Azure instances found")
		return
	}

	fmt.Printf("Fetched %d Azure instances from Vantage for validation\n", len(azureInstances))

	// Run our scraper to get real data
	ctx := context.Background()
	
	scraper, err := NewRealComputeScraper(ctx, subscriptionID)
	if err != nil {
		t.Skipf("Could not create Azure scraper: %v", err)
		return
	}

	// Get our instances with limited set for memory analysis
	config := &models.ScraperConfig{
		Regions:        []string{"eastus", "westus2"},
		SampleSize:     100, // Focus on 100 instances for detailed analysis
		MaxConcurrency: 10,
	}
	
	ourInstances, err := scraper.Scrape(config)
	if err != nil {
		t.Skipf("Could not scrape Azure data: %v", err)
		return
	}

	fmt.Printf("Scraped %d Azure instances from our implementation\n", len(ourInstances))

	// Compare data
	compareAzureData(t, azureInstances, ourInstances)
	
	// Detailed memory analysis
	analyzeMemoryDifferences(t, azureInstances, ourInstances)
}

// fetchAzureData fetches Azure instance data from reference API
func fetchAzureData() ([]AzureInstance, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	
	resp, err := client.Get("https://instances.reference.sh/azure/instances.json")
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

	var instances []AzureInstance
	if err := json.Unmarshal(body, &instances); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return instances, nil
}

// loadComparisonData loads both Vantage and our scraper data for comparison
func loadComparisonData() ([]AzureInstance, []AzureInstance, error) {
	// Load Vantage data
	azureInstances, err := fetchAzureData()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch Vantage data: %v", err)
	}

	// Load our data from the generated file
	file, err := os.Open("../../azure-instances-real.json")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open our Azure data file: %v", err)
	}
	defer file.Close()

	body, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read our Azure data: %v", err)
	}

	var ourInstances []AzureInstance
	if err := json.Unmarshal(body, &ourInstances); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal our Azure data: %v", err)
	}

	return azureInstances, ourInstances, nil
}

// compareAzureData performs detailed comparison between Vantage and our Azure data
func compareAzureData(t *testing.T, azureInstances, ourInstances []AzureInstance) {
	fmt.Printf("\n=== AZURE DATA COMPARISON ===\n")
	fmt.Printf("Vantage instances: %d\n", len(azureInstances))
	fmt.Printf("Our instances: %d\n", len(ourInstances))

	// Create maps for easier lookup
	azureMap := make(map[string]AzureInstance)
	for _, inst := range azureInstances {
		azureMap[inst.InstanceType] = inst
	}

	ourMap := make(map[string]AzureInstance)
	for _, inst := range ourInstances {
		ourMap[inst.InstanceType] = inst
	}

	// Check instances that exist in both
	matchingInstances := 0
	fieldMatches := make(map[string]int)
	fieldMismatches := make(map[string]int)

	fmt.Printf("\n=== MATCHING INSTANCES ===\n")
	for instanceType, ourInst := range ourMap {
		if referenceInst, exists := azureMap[instanceType]; exists {
			matchingInstances++
			fmt.Printf("\n--- %s ---\n", instanceType)
			compareInstanceFields(t, referenceInst, ourInst, fieldMatches, fieldMismatches)
		}
	}

	// Check instances only in Vantage
	onlyInVantage := 0
	fmt.Printf("\n=== INSTANCES ONLY IN VANTAGE ===\n")
	for instanceType := range azureMap {
		if _, exists := ourMap[instanceType]; !exists {
			onlyInVantage++
			if onlyInVantage <= 10 { // Show first 10
				fmt.Printf("- %s\n", instanceType)
			}
		}
	}
	if onlyInVantage > 10 {
		fmt.Printf("... and %d more\n", onlyInVantage-10)
	}

	// Check instances only in our data
	onlyInOurs := 0
	fmt.Printf("\n=== INSTANCES ONLY IN OUR DATA ===\n")
	for instanceType := range ourMap {
		if _, exists := azureMap[instanceType]; !exists {
			onlyInOurs++
			fmt.Printf("- %s\n", instanceType)
		}
	}

	// Summary
	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("Matching instances: %d\n", matchingInstances)
	fmt.Printf("Only in Vantage: %d\n", onlyInVantage)
	fmt.Printf("Only in our data: %d\n", onlyInOurs)

	fmt.Printf("\n=== FIELD COMPARISON ===\n")
	for field, matches := range fieldMatches {
		mismatches := fieldMismatches[field]
		total := matches + mismatches
		if total > 0 {
			accuracy := float64(matches) / float64(total) * 100
			fmt.Printf("%-20s: %3d/%3d matches (%.1f%%)\n", field, matches, total, accuracy)
		}
	}

	// Assertions for the test
	assert.Greater(t, matchingInstances, 0, "Should have at least some matching instances")
	assert.Greater(t, len(ourInstances), 0, "Should have generated some instances")
}

// compareInstanceFields compares fields between two instances
func compareInstanceFields(t *testing.T, reference, ours AzureInstance, matches, mismatches map[string]int) {
	// Helper to increment counters
	recordResult := func(field string, match bool) {
		if match {
			matches[field]++
		} else {
			mismatches[field]++
		}
	}

	// Compare basic fields
	recordResult("InstanceType", reference.InstanceType == ours.InstanceType)
	recordResult("PrettyName", reference.PrettyName == ours.PrettyName)
	recordResult("Memory", reference.Memory == ours.Memory)
	recordResult("Category", reference.Category == ours.Category)
	recordResult("Family", reference.Family == ours.Family)
	recordResult("ACU", reference.ACU == ours.ACU)
	recordResult("GPU", reference.GPU == ours.GPU)
	recordResult("AcceleratedNetworking", reference.AcceleratedNetworking == ours.AcceleratedNetworking)
	recordResult("PremiumIO", reference.PremiumIO == ours.PremiumIO)
	recordResult("HypervGenerations", reference.HypervGenerations == ours.HypervGenerations)

	// Compare vCPU (handle different types)
	vCPUMatch := false
	if reference.VCPU != nil && ours.VCPU != nil {
		vCPUMatch = reflect.DeepEqual(reference.VCPU, ours.VCPU)
	} else {
		vCPUMatch = (reference.VCPU == nil && ours.VCPU == nil)
	}
	recordResult("VCPU", vCPUMatch)

	// Compare arrays
	archMatch := reflect.DeepEqual(reference.Arch, ours.Arch)
	recordResult("Arch", archMatch)

	// Show specific differences for this instance
	if reference.Memory != ours.Memory {
		fmt.Printf("  Memory: Vantage=%.1fGB, Ours=%.1fGB (diff=%.1fGB)\n", 
			reference.Memory, ours.Memory, ours.Memory-reference.Memory)
	}
	if reference.Category != ours.Category {
		fmt.Printf("  Category: Vantage='%s', Ours='%s'\n", reference.Category, ours.Category)
	}
	if reference.ACU != ours.ACU {
		fmt.Printf("  ACU: Vantage=%d, Ours=%d\n", reference.ACU, ours.ACU)
	}
	if !vCPUMatch {
		fmt.Printf("  VCPU: Vantage=%v, Ours=%v\n", reference.VCPU, ours.VCPU)
	}
}

// analyzeMemoryDifferences performs detailed memory difference analysis
func analyzeMemoryDifferences(t *testing.T, azureInstances, ourInstances []AzureInstance) {
	fmt.Printf("\n=== MEMORY DIFFERENCES ANALYSIS ===\n")
	
	// Create maps for easier lookup
	azureMap := make(map[string]AzureInstance)
	for _, inst := range azureInstances {
		azureMap[inst.InstanceType] = inst
	}

	ourMap := make(map[string]AzureInstance)
	for _, inst := range ourInstances {
		ourMap[inst.InstanceType] = inst
	}

	type MemoryDiff struct {
		InstanceType   string
		VantageMemory  float64
		OurMemory      float64
		Difference     float64
		DiffPercent    float64
	}

	var memoryDiffs []MemoryDiff
	totalMatches := 0
	
	for instanceType, ourInst := range ourMap {
		if referenceInst, exists := azureMap[instanceType]; exists {
			totalMatches++
			if referenceInst.Memory != ourInst.Memory && referenceInst.Memory > 0 && ourInst.Memory > 0 {
				diff := ourInst.Memory - referenceInst.Memory
				diffPercent := (diff / referenceInst.Memory) * 100
				
				memoryDiffs = append(memoryDiffs, MemoryDiff{
					InstanceType:   instanceType,
					VantageMemory:  referenceInst.Memory,
					OurMemory:      ourInst.Memory,
					Difference:     diff,
					DiffPercent:    diffPercent,
				})
			}
		}
	}

	fmt.Printf("Total matching instances: %d\n", totalMatches)
	fmt.Printf("Instances with memory differences: %d\n", len(memoryDiffs))
	if totalMatches > 0 {
		fmt.Printf("Memory accuracy: %.1f%%\n", float64(totalMatches-len(memoryDiffs))/float64(totalMatches)*100)
	}

	if len(memoryDiffs) == 0 {
		fmt.Printf("✅ Perfect memory accuracy! No differences found.\n")
		return
	}

	// Sort by absolute difference (largest first)
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

	fmt.Printf("\n%-15s | %10s | %10s | %10s | %8s\n", "Instance", "Vantage", "Ours", "Diff (GB)", "Diff %")
	fmt.Printf("%-15s-|-%10s-|-%10s-|-%10s-|-%8s\n", "---------------", "----------", "----------", "----------", "--------")

	// Show top 10 differences
	maxShow := 10
	if len(memoryDiffs) < maxShow {
		maxShow = len(memoryDiffs)
	}

	for i := 0; i < maxShow; i++ {
		diff := memoryDiffs[i]
		fmt.Printf("%-15s | %10.1f | %10.1f | %+10.1f | %+7.1f%%\n",
			diff.InstanceType,
			diff.VantageMemory,
			diff.OurMemory,
			diff.Difference,
			diff.DiffPercent)
	}

	if len(memoryDiffs) > maxShow {
		fmt.Printf("... and %d more differences\n", len(memoryDiffs)-maxShow)
	}

	// Analysis patterns
	fmt.Printf("\n=== MEMORY DIFFERENCE PATTERNS ===\n")
	
	ourHigher := 0
	ourLower := 0
	for _, diff := range memoryDiffs {
		if diff.Difference > 0 {
			ourHigher++
		} else {
			ourLower++
		}
	}
	
	fmt.Printf("Our memory > Vantage: %d instances\n", ourHigher)
	fmt.Printf("Our memory < Vantage: %d instances\n", ourLower)

	// Check for rounding differences (common in memory specifications)
	roundingDiffs := 0
	for _, diff := range memoryDiffs {
		absDiff := diff.Difference
		if absDiff < 0 {
			absDiff = -absDiff
		}
		// Check if difference is likely due to rounding (< 5% and round numbers)
		if absDiff < diff.VantageMemory*0.05 && (int(absDiff)%4 == 0 || int(absDiff)%8 == 0) {
			roundingDiffs++
		}
	}
	
	if roundingDiffs > 0 {
		fmt.Printf("Likely rounding differences: %d instances\n", roundingDiffs)
		fmt.Printf("💡 This suggests different memory calculation methods (e.g., decimal vs binary units)\n")
	}
}

// TestVantageDataStructure tests that we can parse Vantage Azure data correctly
func TestVantageDataStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Vantage data structure test in short mode")
	}

	azureInstances, err := fetchAzureData()
	if err != nil {
		t.Skipf("Could not fetch Vantage Azure data: %v", err)
		return
	}

	require.Greater(t, len(azureInstances), 0, "Should have some instances")

	// Test structure of first instance
	first := azureInstances[0]
	
	assert.NotEmpty(t, first.InstanceType, "Instance type should not be empty")
	assert.NotEmpty(t, first.PrettyName, "Pretty name should not be empty")
	assert.Greater(t, first.Memory, 0.0, "Memory should be positive")
	assert.NotEmpty(t, first.Category, "Category should not be empty")
	
	// Test pricing structure
	assert.NotEmpty(t, first.Pricing, "Should have pricing data")
	
	// Find a region with pricing data
	var sampleRegion string
	var samplePricing AzurePricing
	for region, pricing := range first.Pricing {
		sampleRegion = region
		samplePricing = pricing
		break
	}
	
	assert.NotEmpty(t, sampleRegion, "Should have at least one region with pricing")
	
	// Test that either Linux or Windows pricing exists
	hasLinux := samplePricing.Linux.OnDemand > 0 || samplePricing.Linux.Basic > 0
	hasWindows := samplePricing.Windows.OnDemand > 0 || samplePricing.Windows.Basic > 0
	assert.True(t, hasLinux || hasWindows, "Should have pricing for at least one OS")

	fmt.Printf("Vantage data validation passed for %d Azure instances\n", len(azureInstances))
}

// TestOurDataConsistency tests consistency within our own Azure scraper data
func TestOurDataConsistency(t *testing.T) {
	// Load our generated data
	file, err := os.Open("../../azure-instances-real.json")
	if err != nil {
		t.Skipf("Could not open our Azure data file: %v", err)
		return
	}
	defer file.Close()

	body, err := io.ReadAll(file)
	if err != nil {
		t.Skipf("Could not read our Azure data: %v", err)
		return
	}

	var instances []AzureInstance
	if err := json.Unmarshal(body, &instances); err != nil {
		t.Skipf("Could not unmarshal our Azure data: %v", err)
		return
	}

	require.Greater(t, len(instances), 0, "Should have some instances")

	for _, instance := range instances {
		t.Run(fmt.Sprintf("Consistency_%s", instance.InstanceType), func(t *testing.T) {
			// Basic validations
			assert.NotEmpty(t, instance.InstanceType)
			assert.NotEmpty(t, instance.PrettyName)
			assert.NotEmpty(t, instance.Family)
			assert.NotEmpty(t, instance.Category)
			
			// Memory should be reasonable
			assert.Greater(t, instance.Memory, 0.0)
			assert.Less(t, instance.Memory, 10000.0) // Reasonable upper bound
			
			// vCPU should be reasonable
			if instance.VCPU != nil {
				switch v := instance.VCPU.(type) {
				case int:
					assert.Greater(t, v, 0)
					assert.Less(t, v, 1000)
				case float64:
					assert.Greater(t, v, 0.0)
					assert.Less(t, v, 1000.0)
				}
			}
			
			// Validate known categories
			knownCategories := []string{
				"generalpurpose", "memoryoptimized", "computeoptimized", 
				"storageoptimized", "acceleratedcompute",
			}
			assert.Contains(t, knownCategories, instance.Category, 
				"Category should be one of known categories")
		})
	}
}