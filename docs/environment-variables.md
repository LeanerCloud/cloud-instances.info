# Environment Variables and Cloud Resources

This document describes all environment variables and cloud resources required to run cloud-instances.info, comparing the original Vantage setup with the LeanerCloud fork.

## Summary of Changes

| Category | Vantage (Original) | LeanerCloud (Fork) |
|----------|-------------------|-------------------|
| Scraper credentials | Same | Same |
| Marketing/Analytics | Extensive | Minimal (optional GTM only) |
| Database | PostgreSQL for diff-writer | Not required |
| Deployment | CloudFlare R2 + Workers | CloudFlare R2 + Workers |

---

## Required Environment Variables

### Cloud Provider Credentials (Scrapers)

These are **required** for the scrapers to fetch pricing data.

#### AWS

| Variable | Description | Required |
|----------|-------------|----------|
| `AWS_ACCESS_KEY_ID` | AWS access key for pricing API | Yes |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | Yes |

**Required IAM Permissions:**
- `ec2:DescribeInstanceTypes`
- `ec2:DescribeSpotPriceHistory`
- `elasticache:DescribeEngineDefaultParameters`
- `pricing:GetProducts`
- `rds:DescribeDBEngineVersions`
- `rds:DescribeOrderableDBInstanceOptions`

#### Azure

| Variable | Description | Required |
|----------|-------------|----------|
| `AZURE_TENANT_ID` | Azure AD tenant ID | Yes |
| `AZURE_CLIENT_ID` | Service principal client ID | Yes |
| `AZURE_CLIENT_SECRET` | Service principal secret | Yes |
| `AZURE_SUBSCRIPTION_ID` | Azure subscription ID | Yes |

**Required Azure Permissions:**
- Reader role on the subscription
- Access to Azure Retail Prices API (public, no special permissions)

See [docs/setting-up-azure.md](./setting-up-azure.md) for setup instructions.

#### GCP

| Variable | Description | Required |
|----------|-------------|----------|
| `GCP_API_KEY` | Google Cloud API key | Yes |

**Required GCP APIs:**
- Cloud Billing API (`cloudbilling.googleapis.com`)

---

### Next.js Build Configuration

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `NEXT_PUBLIC_URL` | Public URL of the site (e.g., `https://cloud-instances.info/`) | Yes | - |
| `DENY_ROBOTS_TXT` | Set to `"1"` to block search engine indexing (for staging) | No | Not set |
| `NEXT_PUBLIC_REMOVE_ADVERTS` | Set to `"1"` to hide sponsor banner | No | Not set |

---

### CloudFlare Deployment

#### Production

| Variable | Description | Required |
|----------|-------------|----------|
| `PRODUCTION_HOSTNAME` | Production domain (e.g., `cloud-instances.info`) | Yes |
| `PRODUCTION_CF_BUCKET` | R2 bucket name for production | Yes |
| `PRODUCTION_CF_AWS_ACCESS_KEY_ID` | R2 access key for production bucket | Yes |
| `PRODUCTION_CF_AWS_SECRET_ACCESS_KEY` | R2 secret key for production bucket | Yes |
| `PRODUCTION_CF_NAMESPACE` | KV namespace ID for production | Yes |
| `PRODUCTION_CF_ZONE_ID` | CloudFlare zone ID for cache purging | Yes |
| `CLOUDFLARE_ACCOUNT_ID` | CloudFlare account ID | Yes |
| `CLOUDFLARE_API_TOKEN` | CloudFlare API token | Yes |

#### Staging

| Variable | Description | Required |
|----------|-------------|----------|
| `STAGING_HOSTNAME` | Staging domain | Yes |
| `STAGING_CF_BUCKET` | R2 bucket name for staging | Yes |
| `STAGING_CF_AWS_ACCESS_KEY_ID` | R2 access key for staging bucket | Yes |
| `STAGING_CF_AWS_SECRET_ACCESS_KEY` | R2 secret key for staging bucket | Yes |
| `STAGING_CF_NAMESPACE` | KV namespace ID for staging | Yes |
| `STAGING_CF_ZONE_ID` | CloudFlare zone ID for staging | Yes |
| `DEPLOYMENT_CF_ACCOUNT_ID` | CloudFlare account ID for deployments | Yes |

#### Ephemeral PR Environments

| Variable | Description | Required |
|----------|-------------|----------|
| `TF_STATE_BUCKET` | S3 bucket for Terraform state | Yes |
| `R2_ACCESS_KEY_ID` | R2 access key for PR environments | Yes |
| `R2_SECRET_ACCESS_KEY` | R2 secret key for PR environments | Yes |
| `PR_CF_NAMESPACE` | KV namespace ID for PR environments | Yes |

---

### Optional: Notifications

| Variable | Description | Required |
|----------|-------------|----------|
| `SLACK_WEBHOOK_URL` | Slack webhook for deployment notifications | No |
| `SCRAPER_SLACK_WEBHOOK_URL` | Slack webhook for scraper error alerts | No |

---

### Optional: Analytics

| Variable | Description | Required |
|----------|-------------|----------|
| `NEXT_PUBLIC_GOOGLE_TAG_MANAGER_ID` | Google Tag Manager container ID | No |

---

## Removed Variables (Vantage-specific)

These variables were used by Vantage but are **not needed** in the LeanerCloud fork:

| Variable | Original Purpose | Status |
|----------|-----------------|--------|
| `NEXT_PUBLIC_SENTRY_DSN` | Sentry error tracking | Removed (can be re-added if needed) |
| `SENTRY_ORG` | Sentry organization | Removed |
| `SENTRY_PROJECT` | Sentry project name | Removed |
| `SENTRY_AUTH_TOKEN` | Sentry auth for source maps | Removed |
| `NEXT_PUBLIC_ENABLE_VANTAGE_SCRIPT_TAG` | Vantage tracking script | Removed |
| `NEXT_PUBLIC_UNIFY_TAG_ID` | Unify marketing tag | Removed |
| `NEXT_PUBLIC_UNIFY_API_KEY` | Unify API key | Removed |
| `DB_CONNECTION_STRING` | PostgreSQL for diff-writer | Removed |
| `NEXT_PUBLIC_INSTANCESKV_URL` | Vantage KV service URL | Removed |

---

## Cloud Resources Required

### AWS

| Resource | Purpose |
|----------|---------|
| IAM User/Role | Scraper access to pricing APIs |
| S3 Bucket | Terraform state storage (for ephemeral environments) |

### Azure

| Resource | Purpose |
|----------|---------|
| App Registration (Service Principal) | Scraper access to Azure pricing |
| Subscription | Required for VM SKU enumeration |

### GCP

| Resource | Purpose |
|----------|---------|
| API Key | Access to Cloud Billing API |
| Project | Project with Billing API enabled |

### CloudFlare

| Resource | Purpose |
|----------|---------|
| R2 Buckets | Static site hosting (production, staging, PR environments) |
| Workers | Request routing and redirects |
| KV Namespaces | Instance data caching |
| DNS Zone | Domain management |

---

## Workflow Triggers

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `production-release.yml` | Push to `main` + every 8 hours | Deploy to production |
| `staging-release.yml` | Push to `develop` | Deploy to staging |
| `tests.yml` | All pushes | Run tests, linting, type checks |
| `ephemeral-environment-for-pull-request.yaml` | PR open/sync/close | Create/destroy PR preview environments |

---

## Migration Checklist

When migrating from Vantage to LeanerCloud fork:

- [ ] Set up AWS IAM user with required permissions
- [ ] Create Azure service principal
- [ ] Create GCP API key with Billing API access
- [ ] Create CloudFlare R2 buckets (production, staging)
- [ ] Create CloudFlare KV namespaces
- [ ] Set up CloudFlare Workers
- [ ] Configure GitHub repository secrets
- [ ] (Optional) Set up Slack webhooks for notifications
- [ ] (Optional) Set up Google Tag Manager
