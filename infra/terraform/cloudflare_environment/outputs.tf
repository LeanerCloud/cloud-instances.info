
output "bucket_name" {
  description = "Name of the created R2 bucket"
  value       = cloudflare_r2_bucket.website_bucket.name
}

output "website_url" {
  description = "URL for the website"
  value       = var.subdomain == "@" ? "https://${var.domain_name}" : "https://${var.subdomain}.${var.domain_name}"
}
