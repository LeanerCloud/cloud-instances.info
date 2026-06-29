output "bucket_names" {
  description = "Names of the created R2 buckets"
  value       = [for b in cloudflare_r2_bucket.assets : b.name]
}

output "kv_namespace_ids" {
  description = "Map of KV namespace title => id. Paste these ids into wrangler.jsonc (kv_namespaces[].id)."
  value       = { for title, ns in cloudflare_workers_kv_namespace.assets : title => ns.id }
}
