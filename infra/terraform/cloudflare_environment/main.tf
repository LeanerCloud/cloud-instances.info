# Get the Cloudflare zone for the domain
data "cloudflare_zone" "domain_zone" {
  zone_id = var.cloudflare_zone_id
}

# The build script creates a data source containing the script source code encoded as base64
data "external" "worker_script" {
  program = ["${path.module}/worker/build.sh", "${path.module}/worker"]
}

# Decode the worker script
locals {
  worker_script_content = base64decode(data.external.worker_script.result.content)
}

# Create Cloudflare R2 bucket with version 5.1.0 provider
resource "cloudflare_r2_bucket" "website_bucket" {
  provider      = cloudflare.v5_1_0 # Use the aliased provider
  account_id    = var.cloudflare_account_id
  name          = var.bucket_name
  location      = var.bucket_location
  storage_class = "Standard"
  jurisdiction  = "default"
}

# Create Cloudflare Worker script
resource "cloudflare_workers_script" "r2_website_worker" {
  script_name        = "${var.bucket_name}-worker"
  account_id         = var.cloudflare_account_id
  content            = local.worker_script_content
  compatibility_date = var.worker_compatibility_date
  main_module        = "worker.js"
}

# Create a Worker route to associate the worker with a hostname pattern
resource "cloudflare_workers_route" "r2_website_route" {
  zone_id = data.cloudflare_zone.domain_zone.zone_id
  pattern = var.subdomain == "@" ? "${var.domain_name}/*" : "${var.subdomain}.${var.domain_name}/*"
  script  = cloudflare_workers_script.r2_website_worker.script_name
}

# The custom domain exposes our R2 bucket through the CloudFlare CDN
resource "cloudflare_r2_custom_domain" "custom_domain" {
  account_id  = var.cloudflare_account_id
  bucket_name = cloudflare_r2_bucket.website_bucket.name
  domain      = var.subdomain == "@" ? "${var.domain_name}/*" : "${var.subdomain}.${var.domain_name}"
  enabled     = true
  zone_id     = data.cloudflare_zone.domain_zone.zone_id
  min_tls     = "1.0"
}
