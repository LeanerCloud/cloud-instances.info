# terraform

[![uses terraform](https://img.shields.io/badge/uses-terraform-blueviolet.svg)](https://www.terraform.io/)

Terraform resources to use to bootstrap the infrastructure for this repo

---

## In this directory

### [tf_state_s3_bucket.tf](./tf_state_s3_bucket.tf)

Creates an S3 bucket used for Terraform state storage, for all the other Terraform modules we have.

### [iam.tf](./iam.tf)

Creates the minimum policy required (and an IAM user + binding):

- to run the EC2/RDS pricing scraper at build time.
- to use the S3 bucket for Terraform state storage at deploy time.

This user's access key is created manually, then configured in the github action secrets and referenced in the GIthub actions.
