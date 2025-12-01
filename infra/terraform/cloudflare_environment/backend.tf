terraform {
  backend "s3" {
    key          = "cloudflare_env.terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true
    encrypt      = true
  }
}

