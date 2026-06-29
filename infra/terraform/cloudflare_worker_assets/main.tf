# Creates the Cloudflare storage the worker serves from: the R2 buckets that hold
# the built static assets uploaded by deployment/index.ts, and the Workers KV
# namespaces. Bucket names must match wrangler.jsonc and the *_CF_BUCKET CI
# secrets; the KV namespace ids (see outputs) go into wrangler.jsonc.
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

resource "cloudflare_workers_kv_namespace" "assets" {
  for_each = var.kv_namespace_titles

  account_id = var.cloudflare_account_id
  title      = each.value
}
