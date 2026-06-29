# Deployment

This fork tracks Vantage's `ec2instances.info` upstream and deploys two environments to
Cloudflare using Vantage's deployment method (build via `make all` -> upload assets to a
Cloudflare R2 bucket via `deployment/index.ts` -> deploy the worker via `wrangler` -> purge cache).

| Environment | URL                          | Branch         | Worker env   |
| ----------- | ---------------------------- | -------------- | ------------ |
| Staging     | `staging-cloud-instances.uk` | `staging-auto` | `staging`    |
| Production  | `cloud-instances.info`       | `master`       | `production` |

## Pipeline

```
development (human-owned)
     │  every 8h (cron) or manual dispatch  [auto-rebase-deploy.yml]
     ▼
rebase development onto upstream/main
     │  conflict ─► open/update GitHub issue + Slack, STOP (no deploy)
     ▼  success (deploys ALWAYS run — each build re-scrapes fresh pricing)
force-push staging-auto ─► build + deploy staging ─► staging-cloud-instances.uk
     │  staging green
     ▼
master = validated commit ─► build + deploy production ─► cloud-instances.info
```

- `development` stays human-owned and is never force-pushed by CI.
- `staging-auto` and `master` are **bot-managed** branches (force-updated each run to the rebased /
  validated commit). Treat them as pointers, not hand-curated history.
- Deploys run every 8 hours regardless of code changes, because pricing data is re-scraped on every
  build.

## Workflows

- `.github/workflows/deploy.yml` — reusable build+deploy (called by the others; not triggered directly).
- `.github/workflows/auto-rebase-deploy.yml` — the 8-hourly orchestrator (rebase -> staging ->
  promote -> production). **Must exist on the default branch (`master`) for the schedule to fire.**
- `.github/workflows/staging-release.yml` — deploys staging on a direct push to `staging-auto` or via
  manual dispatch.
- `.github/workflows/production-release.yml` — deploys production on a push to `master` or via manual
  dispatch.

## One-time setup (manual)

### Domains

- `staging-cloud-instances.uk` — registered; ensure it is an **active zone** in this fork's Cloudflare
  account (point nameservers at Cloudflare if registered elsewhere).
- `cloud-instances.info` — already in Cloudflare.
- DNS records + TLS certs are created automatically by the worker `custom_domain` routes in
  `wrangler.jsonc`. No manual DNS, no Terraform.

### Cloudflare resources

Apply the `infra/terraform/cloudflare_worker_assets` module to create the R2 buckets and KV
namespaces (see its README), then wire the results into `wrangler.jsonc` and the secrets below:

- R2 buckets: `cloud-instances-assets-production`, `cloud-instances-assets-staging`.
- KV namespaces — run `terraform output kv_namespace_ids` and paste each id into `wrangler.jsonc`
  (replacing `REPLACE_WITH_PRODUCTION_KV_ID`, `REPLACE_WITH_STAGING_KV_ID`).

### GitHub

- Create the `staging-auto` branch (or let the first orchestrator run create it).
- Ensure branch protection on `staging-auto` and `master` allows the Actions bot to force-update them
  (or wire a GitHub App / PAT if they are protected against it).

### Secrets (repo Settings -> Secrets and variables -> Actions)

Shared (build): `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`,
`AZURE_TENANT_ID`, `AZURE_SUBSCRIPTION_ID`, `GCP_API_KEY`, `GCP_PROJECT_ID`, `GCP_CLIENT_EMAIL`,
`GCP_PRIVATE_KEY`, `SENTRY_DSN`, `SENTRY_ORG`, `SENTRY_PROJECT`, `SENTRY_AUTH_TOKEN`,
`GOOGLE_TAG_MANAGER_ID`, `UMAMI_WEBSITE_ID`, `UMAMI_SCRIPT_URL`, `SLACK_WEBHOOK_URL`,
`SCRAPER_SLACK_WEBHOOK_URL`.

Shared (Cloudflare): `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_ACCOUNT_ID`, `DEPLOYMENT_CF_ACCOUNT_ID`
(same value as `CLOUDFLARE_ACCOUNT_ID`).

Production: `PRODUCTION_CF_BUCKET`, `PRODUCTION_CF_NAMESPACE`, `PRODUCTION_CF_ZONE_ID`,
`PRODUCTION_CF_AWS_ACCESS_KEY_ID`, `PRODUCTION_CF_AWS_SECRET_ACCESS_KEY`.

Staging: `STAGING_CF_BUCKET`, `STAGING_CF_NAMESPACE`, `STAGING_CF_ZONE_ID`,
`STAGING_CF_AWS_ACCESS_KEY_ID`, `STAGING_CF_AWS_SECRET_ACCESS_KEY`.

> Hostnames are hard-coded in the workflows (public, non-secret), so `STAGING_HOSTNAME` /
> `PRODUCTION_HOSTNAME` secrets are not required by CI; `.env.example` still lists them for local use.

## On a rebase conflict

The orchestrator aborts, opens (or comments on) an issue labelled `upstream-rebase-conflict` listing
the conflicting files, and alerts Slack. Nothing deploys. Resolve by rebasing `development` locally
onto `upstream/main`, pushing it, then re-running `auto-rebase-deploy.yml`.
