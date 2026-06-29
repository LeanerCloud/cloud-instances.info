# cloudflare_worker_assets

Creates the Cloudflare storage the worker serves from for each environment:

- **R2 buckets** holding the built static assets. Names must stay in sync with
  `wrangler.jsonc` (`r2_buckets[].bucket_name`) and the `PRODUCTION_CF_BUCKET` /
  `STAGING_CF_BUCKET` GitHub Actions secrets.
- **Workers KV namespaces**. Their generated ids must be pasted into
  `wrangler.jsonc` (`kv_namespaces[].id`, replacing the `REPLACE_WITH_*_KV_ID`
  placeholders).

By default it creates, for both `production` and `staging`:

- R2 buckets `cloud-instances-assets-production` / `cloud-instances-assets-staging`
- KV namespaces titled the same

## Usage

```sh
export CLOUDFLARE_API_TOKEN=...        # token with R2 + Workers KV edit on the account
export TF_STATE_BUCKET=...             # S3 bucket for remote state (see aws_prerequisites)

terraform init -backend-config="bucket=$TF_STATE_BUCKET"
terraform apply -var="cloudflare_account_id=<account-id>"

# Then copy the printed kv_namespace_ids into wrangler.jsonc:
terraform output kv_namespace_ids
```

Optionally pin a region with `-var="bucket_location=weur"` (apac, eeur, enam,
weur, wnam, oc); omit to let Cloudflare choose.
