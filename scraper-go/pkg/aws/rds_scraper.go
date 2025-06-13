package aws

import (
	"context"
	"fmt"

	"github.com/cloud-instances/scraper/pkg/models"
)

// RDSScraper implements scraping for RDS instances
type RDSScraper struct {
	clientManager *ClientManager
	ctx           context.Context
}

// NewRDSScraper creates a new RDS scraper
func NewRDSScraper(ctx context.Context) (*RDSScraper, error) {
	cm, err := NewClientManager(ctx)
	if err != nil {
		return nil, err
	}
	
	return &RDSScraper{
		clientManager: cm,
		ctx:           ctx,
	}, nil
}

// GetServiceName returns the service name
func (s *RDSScraper) GetServiceName() string {
	return "RDS"
}

// Scrape fetches RDS instance data from AWS
func (s *RDSScraper) Scrape(config *models.ScraperConfig) ([]models.Instance, error) {
	// TODO: Implement RDS scraping
	return nil, fmt.Errorf("RDS scraper not yet implemented")
}