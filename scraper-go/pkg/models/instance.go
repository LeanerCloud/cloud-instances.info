package models

import (
	"encoding/json"
	"time"
)

// Instance represents a cloud instance with its specifications and pricing
type Instance struct {
	InstanceType          string                       `json:"instance_type"`
	PrettyName           string                       `json:"pretty_name"`
	Memory               string                       `json:"memory"`
	VCPU                 json.Number                  `json:"vCPU"`
	PhysicalProcessor    string                       `json:"physical_processor,omitempty"`
	ClockSpeed           string                       `json:"clock_speed,omitempty"`
	Storage              json.RawMessage              `json:"storage"`
	NetworkPerformance   string                       `json:"network_performance,omitempty"`
	Family               string                       `json:"family"`
	EnhancedNetworking   interface{}                  `json:"enhanced_networking,omitempty"`
	VPCOnly              bool                         `json:"vpc_only,omitempty"`
	IPv6Support          bool                         `json:"ipv6_support,omitempty"`
	PlacementGroupSupport bool                        `json:"placement_group_support,omitempty"`
	LinuxVirtualization  []string                     `json:"linux_virtualization,omitempty"`
	Generation           string                       `json:"generation,omitempty"`
	GPUModel             string                       `json:"gpu_model,omitempty"`
	GPUMemory            string                       `json:"gpu_memory,omitempty"`
	GPUCount             json.Number                  `json:"gpu,omitempty"`
	Pricing              map[string]RegionPricing     `json:"pricing"`
	Regions              map[string]string            `json:"regions"`
	
	// Service-specific fields
	Engine               string                       `json:"engine,omitempty"`               // RDS
	MaxConnections       int                          `json:"max_connections,omitempty"`      // RDS
	MaxIops              int                          `json:"max_iops,omitempty"`             // RDS
	MinEBS               string                       `json:"min_ebs,omitempty"`              // OpenSearch
	MaxEBS               string                       `json:"max_ebs_gp2,omitempty"`          // OpenSearch
	MaxEBSGP3            string                       `json:"max_ebs_gp3,omitempty"`          // OpenSearch
	MaxHTTPPayload       string                       `json:"max_http_payload,omitempty"`     // OpenSearch
	NodeType             string                       `json:"node_type,omitempty"`            // Redshift
	Nodes                string                       `json:"nodes,omitempty"`                // Redshift
}

// RegionPricing contains pricing information for a specific region
type RegionPricing struct {
	OnDemand float64                   `json:"ondemand"`
	Reserved map[string]float64        `json:"reserved,omitempty"`
}

// PricingFilter represents filters for AWS Pricing API
type PricingFilter struct {
	Field string
	Value string
	Type  string // TERM_MATCH or others
}

// ScraperConfig holds configuration for scrapers
type ScraperConfig struct {
	Regions         []string
	MaxConcurrency  int
	OutputPath      string
	Service         string
	CacheDuration   time.Duration
	SampleSize      int  // For testing: limit to N random instances (0 = all)
}

// ServiceScraper interface that all service scrapers must implement
type ServiceScraper interface {
	Scrape(config *ScraperConfig) ([]Instance, error)
	GetServiceName() string
}