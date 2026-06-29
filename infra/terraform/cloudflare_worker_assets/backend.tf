terraform {
  backend "s3" {
    key          = "cloudflare_worker_assets.terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true
    encrypt      = true
  }
}
