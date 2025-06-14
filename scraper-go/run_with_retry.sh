#!/bin/bash

# Script to run the scraper with retry logic for timeouts

set -e

# Configuration
MAX_ATTEMPTS=3
CURRENT_ATTEMPT=1
AWS_PROFILE="${AWS_PROFILE:-framed}"
CACHE_ENABLED=true
OUTPUT_DIR="www"
SERVICES="ec2,rds,elasticache,redshift,opensearch"
REGIONS=""  # Empty means all regions

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --no-cache)
            CACHE_ENABLED=false
            shift
            ;;
        --services)
            SERVICES="$2"
            shift 2
            ;;
        --regions)
            REGIONS="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --max-attempts)
            MAX_ATTEMPTS="$2"
            shift 2
            ;;
        *)
            echo "Unknown argument: $1"
            echo "Usage: $0 [--no-cache] [--services services] [--regions regions] [--output dir] [--max-attempts N]"
            exit 1
            ;;
    esac
done

# Build the scraper
echo "Building scraper..."
make build

# Function to run the scraper
run_scraper() {
    local cmd="AWS_PROFILE=$AWS_PROFILE ./bin/scraper"
    
    if [ "$CACHE_ENABLED" = true ]; then
        cmd="$cmd --use-cache"
    fi
    
    cmd="$cmd -s $SERVICES -o $OUTPUT_DIR"
    
    if [ -n "$REGIONS" ]; then
        cmd="$cmd -r $REGIONS"
    fi
    
    echo "Running: $cmd"
    eval $cmd
}

# Main retry loop
while [ $CURRENT_ATTEMPT -le $MAX_ATTEMPTS ]; do
    echo "========================================="
    echo "Attempt $CURRENT_ATTEMPT of $MAX_ATTEMPTS"
    echo "========================================="
    
    start_time=$(date +%s)
    
    if run_scraper; then
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        echo "✅ Scraper completed successfully in $duration seconds"
        exit 0
    else
        exit_code=$?
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        
        echo "❌ Scraper failed with exit code $exit_code after $duration seconds"
        
        if [ $CURRENT_ATTEMPT -lt $MAX_ATTEMPTS ]; then
            echo "Retrying in 30 seconds..."
            sleep 30
        else
            echo "Maximum attempts reached. Exiting."
            exit $exit_code
        fi
    fi
    
    CURRENT_ATTEMPT=$((CURRENT_ATTEMPT + 1))
done