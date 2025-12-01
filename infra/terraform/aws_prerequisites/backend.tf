terraform {
  backend "s3" {
    key          = "aws_prerequisites.terraform.tfstate"
    region       = "us-east-1"
    use_lockfile = true
    encrypt      = true
  }
}
