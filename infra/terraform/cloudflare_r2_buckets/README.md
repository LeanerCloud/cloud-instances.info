# cloudflare_r2_buckets

Creates the Cloudflare R2 buckets that store the built static assets for each
environment. The bucket names here must stay in sync with `wrangler.jsonc`
(`r2_buckets[].bucket_name`) and the `PRODUCTION_CF_BUCKET` / `STAGING_CF_BUCKET`
GitHub Actions secrets.

By default it creates:

- `cloud-instances-assets-production`
- `cloud-instances-assets-staging`

## Usage

```sh
export CLOUDFLARE_API_TOKEN=...        # token with R2 edit on the account
export TF_STATE_BUCKET=...             # S3 bucket for remote state (see aws_prerequisites)

terraform init -backend-config="bucket=$TF_STATE_BUCKET"
terraform apply -var="cloudflare_account_id=<account-id>"
```

Optionally pin a region with `-var="bucket_location=weur"` (apac, eeur, enam,
weur, wnam, oc); omit to let Cloudflare choose.

> KV namespaces (the other `wrangler.jsonc` prerequisite) are not created here.
> They can be added with `cloudflare_workers_kv_namespace` resources following the
> same pattern.
