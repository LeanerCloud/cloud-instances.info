variable "cloudflare_account_id" {
  description = "Cloudflare account ID"
  type        = string
}

variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID"
  type        = string
}

variable "bucket_name" {
  description = "Name of the R2 bucket"
  type        = string
}

variable "bucket_location" {
  description = "Location hint for the R2 bucket (e.g., weur for Western Europe)"
  type        = string
  default     = "weur"
}

variable "domain_name" {
  description = "Domain name for the website"
  type        = string
}

variable "subdomain" {
  description = "Subdomain for the website (use '@' for apex domain)"
  type        = string
  default     = "@"
}

variable "worker_compatibility_date" {
  description = "Compatibility date for the Cloudflare Worker"
  type        = string
  default     = "2025-05-09"
}

variable "worker_environment" {
  description = "Environment for the Cloudflare Worker (e.g., production, development)"
  type        = string
  default     = "production"
}

