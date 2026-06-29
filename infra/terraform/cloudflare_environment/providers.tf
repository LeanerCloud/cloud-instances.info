terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.21"
    }
    external = {
      source = "hashicorp/external"
    }
  }
  required_version = ">= 1.0.0"
}

provider "cloudflare" {}
