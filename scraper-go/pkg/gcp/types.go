package gcp

// GCPInstance represents the structure for GCP instances
type GCPInstance struct {
	InstanceType         string                 `json:"instance_type"`
	PrettyName           string                 `json:"pretty_name"`
	Pricing              map[string]GCPPricing  `json:"pricing"`
	VCPU                 int                    `json:"vcpu"`
	Memory               float64                `json:"memory"`
	GPU                  *GPUInfo               `json:"gpu,omitempty"`
	Family               string                 `json:"family"`
	Arch                 []string               `json:"arch"`
	Storage              *GCPStorage            `json:"storage,omitempty"`
	NetworkPerformance   string                 `json:"network_performance"`
	SustainedUseDiscount bool                   `json:"sustained_use_discount"`
	Preemptible          bool                   `json:"preemptible"`
	SoleTenant           bool                   `json:"sole_tenant,omitempty"`
	CustomSizing         bool                   `json:"custom_sizing,omitempty"`
	MachineType          string                 `json:"machine_type,omitempty"`
	Generation           string                 `json:"generation,omitempty"`
}

// GCPPricing represents pricing structure for a region
type GCPPricing struct {
	Linux   *OSPricing `json:"linux,omitempty"`
	Windows *OSPricing `json:"windows,omitempty"`
}

// OSPricing represents OS-specific pricing
type OSPricing struct {
	OnDemand     float64            `json:"ondemand"`
	Preemptible  float64            `json:"preemptible,omitempty"`
	Spot         float64            `json:"spot,omitempty"`
	Committed    map[string]float64 `json:"committed,omitempty"`
}

// GCPStorage represents storage configuration
type GCPStorage struct {
	LocalSSD       int    `json:"local_ssd,omitempty"`
	PersistentDisk int    `json:"persistent_disk,omitempty"`
	Type           string `json:"type,omitempty"`
}

// GPUInfo represents GPU configuration
type GPUInfo struct {
	Count int    `json:"count"`
	Type  string `json:"type"`
}

// CloudBillingPricingResponse represents response from Cloud Billing API
type CloudBillingPricingResponse struct {
	SKUs         []SKU  `json:"skus"`
	NextPageToken string `json:"nextPageToken"`
}

// SKU represents a pricing SKU from Cloud Billing API
type SKU struct {
	Name        string      `json:"name"`
	SKUID       string      `json:"skuId"`
	Description string      `json:"description"`
	Category    Category    `json:"category"`
	ServiceRegions []string `json:"serviceRegions"`
	PricingInfo []PricingInfo `json:"pricingInfo"`
}

// Category represents SKU category
type Category struct {
	ServiceDisplayName string `json:"serviceDisplayName"`
	ResourceFamily     string `json:"resourceFamily"`
	ResourceGroup      string `json:"resourceGroup"`
	UsageType          string `json:"usageType"`
}

// PricingInfo represents pricing information
type PricingInfo struct {
	Summary                string          `json:"summary"`
	PricingExpression      PricingExpression `json:"pricingExpression"`
	CurrencyConversionRate float64         `json:"currencyConversionRate"`
	EffectiveTime          string          `json:"effectiveTime"`
}

// PricingExpression represents pricing formula
type PricingExpression struct {
	UsageUnit            string         `json:"usageUnit"`
	UsageUnitDescription string         `json:"usageUnitDescription"`
	BaseUnit             string         `json:"baseUnit"`
	BaseUnitDescription  string         `json:"baseUnitDescription"`
	BaseUnitConversionFactor float64    `json:"baseUnitConversionFactor"`
	DisplayQuantity      float64        `json:"displayQuantity"`
	TieredRates          []TieredRate   `json:"tieredRates"`
}

// TieredRate represents tiered pricing
type TieredRate struct {
	StartUsageAmount float64      `json:"startUsageAmount"`
	UnitPrice        *Money       `json:"unitPrice"`
}

// Money represents currency amount
type Money struct {
	CurrencyCode string `json:"currencyCode"`
	Units        string `json:"units"`
	Nanos        int32  `json:"nanos"`
}