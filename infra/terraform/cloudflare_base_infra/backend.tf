terraform {
  backend "s3" {

    key          = "cloudflare_base_infra.terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true
    encrypt      = true
  }
}

