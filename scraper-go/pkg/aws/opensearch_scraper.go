package aws

import (
	"context"
	"fmt"

	"github.com/cloud-instances/scraper/pkg/models"
)

// OpenSearchScraper implements scraping for OpenSearch instances
type OpenSearchScraper struct {
	clientManager *ClientManager
	ctx           context.Context
}

// NewOpenSearchScraper creates a new OpenSearch scraper
func NewOpenSearchScraper(ctx context.Context) (*OpenSearchScraper, error) {
	cm, err := NewClientManager(ctx)
	if err != nil {
		return nil, err
	}
	
	return &OpenSearchScraper{
		clientManager: cm,
		ctx:           ctx,
	}, nil
}

// GetServiceName returns the service name
func (s *OpenSearchScraper) GetServiceName() string {
	return "OpenSearch"
}

// Scrape fetches OpenSearch instance data from AWS
func (s *OpenSearchScraper) Scrape(config *models.ScraperConfig) ([]models.Instance, error) {
	// TODO: Implement OpenSearch scraping
	return nil, fmt.Errorf("OpenSearch scraper not yet implemented")
}