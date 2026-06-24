#!/bin/bash
#
# Simulate CI/CD execution locally to test dependencies and configuration
# This script mimics what the GitHub Actions workflows do
#
# Usage:
#   ./scripts/simulate-ci.sh [environment]
#
# Environments:
#   staging (default) - Simulates staging-release.yml
#   production        - Simulates production-release.yml
#   pr                - Simulates ephemeral-environment-for-pull-request.yaml
#   test              - Simulates tests.yml (no cloud credentials needed)
#

set -e

ENVIRONMENT="${1:-staging}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo_step() {
    echo -e "${BLUE}==>${NC} $1"
}

echo_success() {
    echo -e "${GREEN}✓${NC} $1"
}

echo_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

echo_error() {
    echo -e "${RED}✗${NC} $1"
}

check_env_var() {
    local var_name="$1"
    local required="$2"

    if [ -z "${!var_name}" ]; then
        if [ "$required" = "required" ]; then
            echo_error "$var_name is not set (REQUIRED)"
            return 1
        else
            echo_warning "$var_name is not set (optional)"
            return 0
        fi
    else
        echo_success "$var_name is set"
        return 0
    fi
}

check_command() {
    local cmd="$1"
    if command -v "$cmd" &> /dev/null; then
        echo_success "$cmd is installed ($(command -v $cmd))"
        return 0
    else
        echo_error "$cmd is not installed"
        return 1
    fi
}

#
# Check prerequisites
#
check_prerequisites() {
    echo_step "Checking prerequisites..."

    local errors=0

    check_command "go" || ((errors++))
    check_command "node" || ((errors++))
    check_command "npm" || ((errors++))
    check_command "docker" || ((errors++))

    if [ "$ENVIRONMENT" = "pr" ]; then
        check_command "terraform" || ((errors++))
        check_command "aws" || ((errors++))
    fi

    if [ $errors -gt 0 ]; then
        echo_error "Missing $errors required tools"
        return 1
    fi

    echo ""
    return 0
}

#
# Check cloud provider credentials
#
check_cloud_credentials() {
    echo_step "Checking cloud provider credentials..."

    local errors=0

    # AWS
    echo ""
    echo "AWS:"
    check_env_var "AWS_ACCESS_KEY_ID" "required" || ((errors++))
    check_env_var "AWS_SECRET_ACCESS_KEY" "required" || ((errors++))

    # Azure
    echo ""
    echo "Azure:"
    check_env_var "AZURE_TENANT_ID" "required" || ((errors++))
    check_env_var "AZURE_CLIENT_ID" "required" || ((errors++))
    check_env_var "AZURE_CLIENT_SECRET" "required" || ((errors++))
    check_env_var "AZURE_SUBSCRIPTION_ID" "required" || ((errors++))

    # GCP
    echo ""
    echo "GCP:"
    check_env_var "GCP_API_KEY" "required" || ((errors++))

    echo ""
    if [ $errors -gt 0 ]; then
        echo_error "Missing $errors required cloud credentials"
        echo ""
        echo "Set these environment variables or create a .env file:"
        echo "  export AWS_ACCESS_KEY_ID=..."
        echo "  export AWS_SECRET_ACCESS_KEY=..."
        echo "  export AZURE_TENANT_ID=..."
        echo "  export AZURE_CLIENT_ID=..."
        echo "  export AZURE_CLIENT_SECRET=..."
        echo "  export AZURE_SUBSCRIPTION_ID=..."
        echo "  export GCP_API_KEY=..."
        return 1
    fi

    return 0
}

#
# Check deployment credentials
#
check_deployment_credentials() {
    echo_step "Checking deployment credentials for $ENVIRONMENT..."

    local errors=0

    echo ""
    echo "CloudFlare:"
    check_env_var "CLOUDFLARE_ACCOUNT_ID" "required" || ((errors++))
    check_env_var "CLOUDFLARE_API_TOKEN" "required" || ((errors++))

    if [ "$ENVIRONMENT" = "production" ]; then
        echo ""
        echo "Production:"
        check_env_var "PRODUCTION_HOSTNAME" "required" || ((errors++))
        check_env_var "PRODUCTION_CF_BUCKET" "required" || ((errors++))
        check_env_var "PRODUCTION_CF_AWS_ACCESS_KEY_ID" "required" || ((errors++))
        check_env_var "PRODUCTION_CF_AWS_SECRET_ACCESS_KEY" "required" || ((errors++))
        check_env_var "PRODUCTION_CF_NAMESPACE" "required" || ((errors++))
        check_env_var "PRODUCTION_CF_ZONE_ID" "required" || ((errors++))
    elif [ "$ENVIRONMENT" = "staging" ]; then
        echo ""
        echo "Staging:"
        check_env_var "STAGING_HOSTNAME" "required" || ((errors++))
        check_env_var "STAGING_CF_BUCKET" "required" || ((errors++))
        check_env_var "STAGING_CF_AWS_ACCESS_KEY_ID" "required" || ((errors++))
        check_env_var "STAGING_CF_AWS_SECRET_ACCESS_KEY" "required" || ((errors++))
        check_env_var "STAGING_CF_NAMESPACE" "required" || ((errors++))
        check_env_var "STAGING_CF_ZONE_ID" "required" || ((errors++))
        check_env_var "DEPLOYMENT_CF_ACCOUNT_ID" "required" || ((errors++))
    elif [ "$ENVIRONMENT" = "pr" ]; then
        echo ""
        echo "PR Environment:"
        check_env_var "TF_STATE_BUCKET" "required" || ((errors++))
        check_env_var "R2_ACCESS_KEY_ID" "required" || ((errors++))
        check_env_var "R2_SECRET_ACCESS_KEY" "required" || ((errors++))
        check_env_var "PR_CF_NAMESPACE" "required" || ((errors++))
        check_env_var "CLOUDFLARE_ZONE_ID" "required" || ((errors++))
    fi

    echo ""
    echo "Optional:"
    check_env_var "SLACK_WEBHOOK_URL" "optional"
    check_env_var "SCRAPER_SLACK_WEBHOOK_URL" "optional"
    check_env_var "NEXT_PUBLIC_GOOGLE_TAG_MANAGER_ID" "optional"

    echo ""
    if [ $errors -gt 0 ]; then
        echo_error "Missing $errors required deployment credentials"
        return 1
    fi

    return 0
}

#
# Run tests (simulates tests.yml)
#
run_tests() {
    echo_step "Running tests..."

    cd "$PROJECT_ROOT"

    # gofmt check
    echo ""
    echo "Checking Go formatting..."
    cd scraper
    if [ -z "$(gofmt -l .)" ]; then
        echo_success "Go code is properly formatted"
    else
        echo_error "Go code needs formatting:"
        gofmt -l .
        echo "Run: cd scraper && gofmt -w ."
    fi
    cd "$PROJECT_ROOT"

    # Prettier check
    echo ""
    echo "Checking Prettier formatting..."
    if command -v prettier &> /dev/null; then
        if prettier --check . 2>/dev/null; then
            echo_success "Code is properly formatted"
        else
            echo_warning "Some files need formatting. Run: prettier --write ."
        fi
    else
        echo_warning "Prettier not installed globally. Run: npm install -g prettier"
    fi

    # TypeScript and tests
    echo ""
    echo "Running Next.js tests..."
    cd next
    if [ ! -d "node_modules" ]; then
        echo "Installing dependencies..."
        npm ci
    fi

    echo "Checking types..."
    npm run check-types || echo_warning "Type check failed"

    echo "Running unit tests..."
    npm run test -- --coverage || echo_warning "Some tests failed"

    cd "$PROJECT_ROOT"
}

#
# Build site (simulates make all)
#
build_site() {
    echo_step "Building site..."

    cd "$PROJECT_ROOT"

    # Set Next.js URL based on environment
    if [ "$ENVIRONMENT" = "production" ]; then
        export NEXT_PUBLIC_URL="https://${PRODUCTION_HOSTNAME:-cloud-instances.info}/"
    elif [ "$ENVIRONMENT" = "staging" ]; then
        export NEXT_PUBLIC_URL="https://${STAGING_HOSTNAME:-staging.cloud-instances.info}/"
        export DENY_ROBOTS_TXT="1"
    elif [ "$ENVIRONMENT" = "pr" ]; then
        export NEXT_PUBLIC_URL="https://pull-request-test.cloud-instances.info/"
        export DENY_ROBOTS_TXT="1"
    fi

    echo "NEXT_PUBLIC_URL=$NEXT_PUBLIC_URL"

    # Run make all
    echo ""
    echo "Running 'make all' (this runs the scrapers and builds Next.js)..."
    echo "This may take several minutes..."
    echo ""

    make all

    echo ""
    echo_success "Build completed!"
    echo "Output is in: $PROJECT_ROOT/www/"
}

#
# Deploy (simulates deployment step)
#
deploy() {
    echo_step "Deploying to $ENVIRONMENT..."

    cd "$PROJECT_ROOT/deployment"

    if [ ! -d "node_modules" ]; then
        echo "Installing deployment dependencies..."
        npm ci
    fi

    # Set deployment variables based on environment
    if [ "$ENVIRONMENT" = "production" ]; then
        export DEPLOYMENT_CF_BUCKET="$PRODUCTION_CF_BUCKET"
        export DEPLOYMENT_CF_ACCOUNT_ID="$CLOUDFLARE_ACCOUNT_ID"
        export DEPLOYMENT_CF_AWS_ACCESS_KEY_ID="$PRODUCTION_CF_AWS_ACCESS_KEY_ID"
        export DEPLOYMENT_CF_AWS_SECRET_ACCESS_KEY="$PRODUCTION_CF_AWS_SECRET_ACCESS_KEY"
        export DEPLOYMENT_CF_API_KEY="$CLOUDFLARE_API_TOKEN"
        export DEPLOYMENT_CF_NAMESPACE="$PRODUCTION_CF_NAMESPACE"
    elif [ "$ENVIRONMENT" = "staging" ]; then
        export DEPLOYMENT_CF_BUCKET="$STAGING_CF_BUCKET"
        export DEPLOYMENT_CF_ACCOUNT_ID="$DEPLOYMENT_CF_ACCOUNT_ID"
        export DEPLOYMENT_CF_AWS_ACCESS_KEY_ID="$STAGING_CF_AWS_ACCESS_KEY_ID"
        export DEPLOYMENT_CF_AWS_SECRET_ACCESS_KEY="$STAGING_CF_AWS_SECRET_ACCESS_KEY"
        export DEPLOYMENT_CF_API_KEY="$CLOUDFLARE_API_TOKEN"
        export DEPLOYMENT_CF_NAMESPACE="$STAGING_CF_NAMESPACE"
    fi

    echo "Deploying..."
    npm run start

    echo ""
    echo_success "Deployment completed!"
}

#
# Main
#
main() {
    echo "========================================"
    echo "  CI/CD Simulation - $ENVIRONMENT"
    echo "========================================"
    echo ""

    # Load .env file if it exists
    if [ -f "$PROJECT_ROOT/.env" ]; then
        echo_step "Loading .env file..."
        set -a
        source "$PROJECT_ROOT/.env"
        set +a
        echo ""
    fi

    case "$ENVIRONMENT" in
        test)
            check_prerequisites || exit 1
            run_tests
            ;;
        staging|production)
            check_prerequisites || exit 1
            check_cloud_credentials || exit 1
            check_deployment_credentials || exit 1

            echo ""
            echo "All checks passed! Ready to build and deploy."
            echo ""
            read -p "Continue with build? (y/n) " -n 1 -r
            echo ""
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                build_site

                echo ""
                read -p "Continue with deployment? (y/n) " -n 1 -r
                echo ""
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    deploy
                fi
            fi
            ;;
        pr)
            check_prerequisites || exit 1
            check_cloud_credentials || exit 1
            check_deployment_credentials || exit 1
            echo ""
            echo_warning "PR environment simulation requires Terraform setup."
            echo "See: infra/terraform/cloudflare_environment/"
            ;;
        check)
            check_prerequisites
            echo ""
            check_cloud_credentials
            echo ""
            check_deployment_credentials
            ;;
        *)
            echo "Unknown environment: $ENVIRONMENT"
            echo ""
            echo "Usage: $0 [environment]"
            echo ""
            echo "Environments:"
            echo "  test       - Run tests only (no cloud credentials needed)"
            echo "  staging    - Simulate staging deployment"
            echo "  production - Simulate production deployment"
            echo "  pr         - Simulate PR environment"
            echo "  check      - Only check credentials (no build/deploy)"
            exit 1
            ;;
    esac

    echo ""
    echo "========================================"
    echo "  Done!"
    echo "========================================"
}

main "$@"
