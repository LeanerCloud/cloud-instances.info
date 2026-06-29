terraform {
  backend "s3" {
    key          = "cloudflare_r2_buckets.terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true
    encrypt      = true
  }
}
