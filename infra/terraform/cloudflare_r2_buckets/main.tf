# Creates the Cloudflare R2 buckets that hold the built static assets uploaded by
# deployment/index.ts and served by the worker. Bucket names must match
# wrangler.jsonc and the *_CF_BUCKET CI secrets.
terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 5.21"
    }
  }
  required_version = ">= 1.0.0"
}

# Reads the API token from the CLOUDFLARE_API_TOKEN environment variable.
provider "cloudflare" {}

resource "cloudflare_r2_bucket" "assets" {
  for_each = var.bucket_names

  account_id = var.cloudflare_account_id
  name       = each.value
  location   = var.bucket_location
}
