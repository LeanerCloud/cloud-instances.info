package azure

// VantageAzureInstance represents the structure from instances.vantage.sh/azure/instances.json
type VantageAzureInstance struct {
	ACU                  int                               `json:"ACU"`
	GPU                  string                            `json:"GPU"`
	AcceleratedNetworking bool                             `json:"accelerated_networking"`
	Arch                 []string                          `json:"arch"`
	AvailabilityZones    map[string]interface{}            `json:"availability_zones"`
	CachedDisk           int                               `json:"cached_disk"`
	CapacitySupport      bool                              `json:"capacity_support"`
	Category             string                            `json:"category"`
	Confidential         interface{}                       `json:"confidential"`
	Encryption           interface{}                       `json:"encryption"`
	Family               string                            `json:"family"`
	Hibernation          interface{}                       `json:"hibernation"`
	HypervGenerations    string                            `json:"hyperv_generations"`
	InstanceType         string                            `json:"instance_type"`
	IOPS                 interface{}                       `json:"iops"`
	LowPriority          bool                              `json:"low_priority"`
	Memory               float64                           `json:"memory"`
	MemoryMaintenance    bool                              `json:"memory_maintenance"`
	PremiumIO            bool                              `json:"premium_io"`
	PrettyName           string                            `json:"pretty_name"`
	PrettyNameAzure      string                            `json:"pretty_name_azure"`
	Pricing              map[string]VantageAzurePricing    `json:"pricing"`
	RDMA                 bool                              `json:"rdma,omitempty"`
	ReadIO               interface{}                       `json:"read_io"`
	Size                 interface{}                       `json:"size"`
	Storage              interface{}                       `json:"storage"`
	TrustedLaunch        bool                              `json:"trusted_launch,omitempty"`
	UltraSSD             bool                              `json:"ultra_ssd,omitempty"`
	UncachedDisk         int                               `json:"uncached_disk"`
	UncachedDiskIO       interface{}                       `json:"uncached_disk_io"`
	VCPU                 interface{}                       `json:"vcpu"`
	VCPUsAvailable       interface{}                       `json:"vcpus_available"`
	VCPUsPerCore         interface{}                       `json:"vcpus_percore"`
	VMDeployment         interface{}                       `json:"vm_deployment"`
	WriteIO              interface{}                       `json:"write_io"`
}

// VantageAzurePricing represents pricing structure
type VantageAzurePricing struct {
	Linux   VantageAzureOSPricing `json:"linux,omitempty"`
	Windows VantageAzureOSPricing `json:"windows,omitempty"`
}

// VantageAzureOSPricing represents OS-specific pricing
type VantageAzureOSPricing struct {
	Basic          float64            `json:"basic,omitempty"`
	BasicSpot      float64            `json:"basic-spot,omitempty"`
	OnDemand       float64            `json:"ondemand,omitempty"`
	LowPriority    float64            `json:"lowpriority,omitempty"`
	HybridBenefit  float64            `json:"hybridbenefit,omitempty"`
	Reserved       map[string]float64 `json:"reserved,omitempty"`
	SpotMin        float64            `json:"spot_min,omitempty"`
}