variable "cloudflare_account_id" {
  description = "Cloudflare account ID"
  type        = string
}

variable "bucket_names" {
  description = "R2 bucket names to create (one per environment). Must match wrangler.jsonc and the *_CF_BUCKET CI secrets."
  type        = set(string)
  default = [
    "cloud-instances-assets-production",
    "cloud-instances-assets-staging",
  ]
}

variable "bucket_location" {
  description = "Optional R2 location hint (apac, eeur, enam, weur, wnam, oc). Leave null to let Cloudflare choose automatically."
  type        = string
  default     = null
}
