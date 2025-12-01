terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = ">= 5.1.0, <= 5.4.0" # Allow a range that includes both versions
    }
    external = {
      source = "hashicorp/external"
    }
  }
  required_version = ">= 1.0.0"
}

# Main provider configuration (version 5.4.0)
provider "cloudflare" {}

# Additional provider configuration for version 5.1.0 with alias
provider "cloudflare" {
  alias     = "v5_1_0"
}
