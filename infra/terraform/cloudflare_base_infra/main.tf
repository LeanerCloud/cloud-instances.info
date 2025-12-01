# main.tf
terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.4.0"
    }
    namecheap = {
      source  = "namecheap/namecheap"
      version = "~> 2.2.0"
    }
  }
  required_version = ">= 1.0.0"
}

# Provider configurations
provider "cloudflare" {}

provider "namecheap" {
  user_name   = var.namecheap_username
  api_user    = var.namecheap_api_user
  api_key     = var.namecheap_api_key
  client_ip   = var.namecheap_client_ip
  use_sandbox = false
}

# Parent zone
resource "cloudflare_zone" "domain_zone" {
  name = var.domain_name
  account = {
    id = var.cloudflare_account_id
  }
  type = "full"
}


# Configure Namecheap DNS to point to Cloudflare nameservers
resource "namecheap_domain_records" "domain_records" {
  domain = var.domain_name
  mode   = "OVERWRITE"

  nameservers = cloudflare_zone.domain_zone.name_servers

  depends_on = [cloudflare_zone.domain_zone]
}
