


variable "cloudflare_account_id" {
  description = "Cloudflare account ID"
  type        = string
}

variable "domain_name" {
  description = "Domain name for the website"
  type        = string
}

variable "namecheap_username" {
  description = "Namecheap username"
  type        = string
  sensitive   = true
}

variable "namecheap_api_user" {
  description = "Namecheap API user (same as username if you're the account owner)"
  type        = string
  sensitive   = true
}

variable "namecheap_api_key" {
  description = "Namecheap API key"
  type        = string
  sensitive   = true
}

variable "namecheap_client_ip" {
  description = "Client IP address for Namecheap API access"
  type        = string
}

