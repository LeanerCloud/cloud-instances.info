

output "cloudflare_name_servers" {
  description = "Cloudflare nameservers for the domain"
  value       = cloudflare_zone.domain_zone.name_servers
}

