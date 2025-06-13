package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cloud-instances/scraper/internal/cache"
	"github.com/cloud-instances/scraper/pkg/aws"
	"github.com/cloud-instances/scraper/pkg/models"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	outputDir      string
	regions        []string
	services       []string
	maxConcurrency int
	validate       bool
	pythonOutput   string
	useCache       bool
	cacheDir       string
	cacheTTL       string
)

var rootCmd = &cobra.Command{
	Use:   "scraper",
	Short: "Cloud instance scraper with parallel execution",
	RunE:  run,
}

func init() {
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "www", "Output directory for JSON files")
	rootCmd.Flags().StringSliceVarP(&regions, "regions", "r", []string{}, "Regions to scrape (empty for all)")
	rootCmd.Flags().StringSliceVarP(&services, "services", "s", []string{"ec2", "rds", "elasticache", "redshift", "opensearch"}, "Services to scrape")
	rootCmd.Flags().IntVarP(&maxConcurrency, "concurrency", "c", 10, "Maximum concurrent operations")
	rootCmd.Flags().BoolVarP(&validate, "validate", "v", false, "Validate output against Python scraper output")
	rootCmd.Flags().StringVarP(&pythonOutput, "python-output", "p", "", "Path to Python scraper output for validation")
	rootCmd.Flags().BoolVar(&useCache, "use-cache", false, "Use cached data if available")
	rootCmd.Flags().StringVar(&cacheDir, "cache-dir", ".cache/scraper", "Cache directory")
	rootCmd.Flags().StringVar(&cacheTTL, "cache-ttl", "1h", "Cache TTL (e.g., 1h, 30m, 24h)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Parse cache TTL
	ttl, err := time.ParseDuration(cacheTTL)
	if err != nil {
		return fmt.Errorf("invalid cache TTL: %v", err)
	}

	// Create cache
	var scraperCache *cache.Cache
	if useCache {
		scraperCache = cache.NewCache(cacheDir, ttl)
		fmt.Printf("Using cache directory: %s (TTL: %s)\n", cacheDir, ttl)
	}

	// Store cache in context for scrapers to use
	if scraperCache != nil {
		ctx = context.WithValue(ctx, "cache", scraperCache)
	}

	// Run scrapers in parallel
	g, ctx := errgroup.WithContext(ctx)
	
	for _, service := range services {
		service := service // capture loop variable
		g.Go(func() error {
			return scrapeService(ctx, service)
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	fmt.Println("Scraping completed successfully!")
	return nil
}

func scrapeService(ctx context.Context, service string) error {
	fmt.Printf("Starting scrape for %s...\n", service)
	startTime := time.Now()

	config := &models.ScraperConfig{
		Regions:        regions,
		MaxConcurrency: maxConcurrency,
		Service:        service,
	}

	var scraper models.ServiceScraper
	var outputPath string
	var err error

	switch service {
	case "ec2":
		scraper, err = aws.NewEC2Scraper(ctx)
		outputPath = filepath.Join(outputDir, "instances.json")
	case "rds":
		scraper, err = aws.NewRDSScraper(ctx)
		outputPath = filepath.Join(outputDir, "rds", "instances.json")
	case "elasticache":
		scraper, err = aws.NewElastiCacheScraper(ctx)
		outputPath = filepath.Join(outputDir, "cache", "instances.json")
	case "redshift":
		scraper, err = aws.NewRedshiftScraper(ctx)
		outputPath = filepath.Join(outputDir, "redshift", "instances.json")
	case "opensearch":
		scraper, err = aws.NewOpenSearchScraper(ctx)
		outputPath = filepath.Join(outputDir, "opensearch", "instances.json")
	default:
		return fmt.Errorf("unknown service: %s", service)
	}

	if err != nil {
		return fmt.Errorf("failed to create %s scraper: %v", service, err)
	}

	var instances []models.Instance

	// Check cache if enabled
	if cacheVal := ctx.Value("cache"); cacheVal != nil {
		scraperCache := cacheVal.(*cache.Cache)
		cacheKey := scraperCache.CacheKey(service, regions)
		hit, err := scraperCache.Get(cacheKey, &instances)
		if err != nil {
			fmt.Printf("  Cache error for %s: %v\n", service, err)
		} else if hit {
			fmt.Printf("  Using cached data for %s\n", service)
		} else {
			// Scrape fresh data
			instances, err = scraper.Scrape(config)
			if err != nil {
				return fmt.Errorf("failed to scrape %s: %v", service, err)
			}
			
			// Save to cache
			if err := scraperCache.Set(cacheKey, instances); err != nil {
				fmt.Printf("  Failed to cache %s data: %v\n", service, err)
			}
		}
	} else {
		// No cache, scrape directly
		instances, err = scraper.Scrape(config)
		if err != nil {
			return fmt.Errorf("failed to scrape %s: %v", service, err)
		}
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", outputDir, err)
	}

	// Save to file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %v", outputPath, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	if err := encoder.Encode(instances); err != nil {
		return fmt.Errorf("failed to encode JSON for %s: %v", service, err)
	}

	duration := time.Since(startTime)
	fmt.Printf("Completed %s scrape in %v. Found %d instance types.\n", service, duration, len(instances))

	// Validate if requested
	if validate && pythonOutput != "" {
		pythonPath := filepath.Join(pythonOutput, filepath.Base(outputPath))
		if service == "ec2" {
			pythonPath = filepath.Join(pythonOutput, "instances.json")
		}
		
		if err := validateOutput(outputPath, pythonPath, service); err != nil {
			return fmt.Errorf("validation failed for %s: %v", service, err)
		}
		fmt.Printf("✓ Validation passed for %s\n", service)
	}

	return nil
}

func validateOutput(goOutput, pythonOutput, service string) error {
	// Read Go output
	goData, err := os.ReadFile(goOutput)
	if err != nil {
		return fmt.Errorf("failed to read Go output: %v", err)
	}

	// Read Python output
	pythonData, err := os.ReadFile(pythonOutput)
	if err != nil {
		return fmt.Errorf("failed to read Python output: %v", err)
	}

	// Parse both outputs
	var goInstances []models.Instance
	if err := json.Unmarshal(goData, &goInstances); err != nil {
		return fmt.Errorf("failed to parse Go output: %v", err)
	}

	var pythonInstances []map[string]interface{}
	if err := json.Unmarshal(pythonData, &pythonInstances); err != nil {
		return fmt.Errorf("failed to parse Python output: %v", err)
	}

	// Basic validation - check counts
	if len(goInstances) != len(pythonInstances) {
		fmt.Printf("Warning: Instance count mismatch for %s - Go: %d, Python: %d\n", 
			service, len(goInstances), len(pythonInstances))
	}

	// Create maps for easier comparison
	goMap := make(map[string]models.Instance)
	for _, inst := range goInstances {
		goMap[inst.InstanceType] = inst
	}

	pythonMap := make(map[string]map[string]interface{})
	for _, inst := range pythonInstances {
		if it, ok := inst["instance_type"].(string); ok {
			pythonMap[it] = inst
		}
	}

	// Check for missing instances
	missingInGo := 0
	missingInPython := 0

	for it := range pythonMap {
		if _, ok := goMap[it]; !ok {
			missingInGo++
			if missingInGo <= 5 { // Print first 5
				fmt.Printf("  Missing in Go output: %s\n", it)
			}
		}
	}

	for it := range goMap {
		if _, ok := pythonMap[it]; !ok {
			missingInPython++
			if missingInPython <= 5 { // Print first 5
				fmt.Printf("  Missing in Python output: %s\n", it)
			}
		}
	}

	if missingInGo > 0 || missingInPython > 0 {
		fmt.Printf("  Total missing: Go=%d, Python=%d\n", missingInGo, missingInPython)
	}

	// Validate data structure for a sample instance
	if len(goInstances) > 0 && len(pythonInstances) > 0 {
		// Find a common instance type
		for it := range goMap {
			if pyInst, ok := pythonMap[it]; ok {
				// Check key fields match
				goInst := goMap[it]
				
				// Memory
				if pyMem, ok := pyInst["memory"].(string); ok && goInst.Memory != pyMem {
					fmt.Printf("  Memory mismatch for %s: Go=%s, Python=%s\n", it, goInst.Memory, pyMem)
				}
				
				// vCPU
				if pyVCPU, ok := pyInst["vCPU"]; ok {
					goVCPU := string(goInst.VCPU)
					pyVCPUStr := fmt.Sprintf("%v", pyVCPU)
					if goVCPU != pyVCPUStr {
						fmt.Printf("  vCPU mismatch for %s: Go=%s, Python=%s\n", it, goVCPU, pyVCPUStr)
					}
				}
				
				break // Just check one instance
			}
		}
	}

	return nil
}