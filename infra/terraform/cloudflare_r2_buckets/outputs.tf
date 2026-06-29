output "bucket_names" {
  description = "Names of the created R2 buckets"
  value       = [for b in cloudflare_r2_bucket.assets : b.name]
}
