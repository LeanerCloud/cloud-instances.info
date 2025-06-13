#!/usr/bin/env python3
"""Compare Go scraper outputs with instances.vantage.sh data"""
import json
import os
import subprocess
import sys
import tempfile
import requests
from pathlib import Path

def download_vantage_data(output_dir):
    """Download JSON files from instances.vantage.sh"""
    print("Downloading data from instances.vantage.sh...")
    
    endpoints = {
        "instances.json": "https://instances.vantage.sh/instances.json",
        "rds/instances.json": "https://instances.vantage.sh/rds/instances.json",
        "cache/instances.json": "https://instances.vantage.sh/cache/instances.json",
        "redshift/instances.json": "https://instances.vantage.sh/redshift/instances.json",
        "opensearch/instances.json": "https://instances.vantage.sh/opensearch/instances.json",
    }
    
    for file_path, url in endpoints.items():
        output_file = os.path.join(output_dir, file_path)
        os.makedirs(os.path.dirname(output_file), exist_ok=True)
        
        try:
            response = requests.get(url, timeout=30)
            response.raise_for_status()
            
            with open(output_file, 'w') as f:
                json.dump(response.json(), f, indent=2)
            
            print(f"  ✓ Downloaded {file_path}")
        except Exception as e:
            print(f"  ✗ Failed to download {file_path}: {e}")
            return False
    
    return True

def run_go_scrapers(output_dir, use_cache=False):
    """Run the Go scrapers"""
    print("Running Go scrapers...")
    
    cmd = [
        "go", "run", "cmd/scraper/main.go",
        "-o", output_dir,
        "-c", "20",  # Higher concurrency
        "-s", "ec2",  # Only EC2 for now
    ]
    
    if use_cache:
        cmd.extend(["--use-cache"])
    
    env = os.environ.copy()
    env['AWS_PROFILE'] = 'framed'
    
    result = subprocess.run(cmd, cwd="scraper-go", env=env, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Go scraper failed: {result.stderr}")
        return False
    print(result.stdout)
    return True

def compare_json_files(vantage_file, go_file, service_name):
    """Compare Vantage and Go JSON outputs"""
    print(f"\n--- Comparing {service_name} ---")
    
    if not os.path.exists(vantage_file):
        print(f"  ❌ Vantage file missing: {vantage_file}")
        return
    
    if not os.path.exists(go_file):
        print(f"  ❌ Go file missing: {go_file}")
        return
    
    # Load files
    with open(vantage_file, 'r') as f:
        vantage_data = json.load(f)
    with open(go_file, 'r') as f:
        go_data = json.load(f)
    
    print(f"  Vantage instances: {len(vantage_data)}")
    print(f"  Go instances: {len(go_data)}")
    
    # Create maps
    vantage_map = {inst.get('instance_type', ''): inst for inst in vantage_data}
    go_map = {inst.get('instance_type', ''): inst for inst in go_data}
    
    # Find differences
    vantage_types = set(vantage_map.keys())
    go_types = set(go_map.keys())
    
    missing_in_go = vantage_types - go_types
    missing_in_vantage = go_types - vantage_types
    common = vantage_types & go_types
    
    if missing_in_go:
        print(f"  Missing in Go ({len(missing_in_go)}): {list(missing_in_go)[:5]}...")
    if missing_in_vantage:
        print(f"  Missing in Vantage ({len(missing_in_vantage)}): {list(missing_in_vantage)[:5]}...")
    
    print(f"  Common instances: {len(common)}")
    
    # Compare a few sample instances
    if common:
        samples = list(common)[:3]  # Check first 3
        for sample in samples:
            print(f"\n  Comparing instance: {sample}")
            
            vantage_inst = vantage_map[sample]
            go_inst = go_map[sample]
            
            # Compare key fields
            fields = ['memory', 'vCPU', 'family', 'network_performance', 'generation', 'pretty_name']
            mismatches = []
            
            for field in fields:
                vantage_val = vantage_inst.get(field)
                go_val = go_inst.get(field)
                
                # Normalize vCPU for comparison
                if field == 'vCPU':
                    vantage_val = str(vantage_val) if vantage_val is not None else None
                    go_val = str(go_val) if go_val is not None else None
                
                if vantage_val != go_val:
                    mismatches.append(f"    ❌ {field}: Vantage={vantage_val}, Go={go_val}")
                else:
                    if field in ['memory', 'vCPU', 'family']:  # Show key fields
                        print(f"    ✅ {field}: {vantage_val}")
            
            for mismatch in mismatches:
                print(mismatch)
            
            # Check pricing structure
            if 'pricing' in vantage_inst and 'pricing' in go_inst:
                vantage_regions = set(vantage_inst['pricing'].keys())
                go_regions = set(go_inst['pricing'].keys())
                if len(vantage_regions) > 0 and len(go_regions) == 0:
                    print(f"    ⚠️  Go scraper missing pricing data (Vantage has {len(vantage_regions)} regions)")

def main():
    """Main comparison function"""
    with tempfile.TemporaryDirectory() as vantage_dir:
        with tempfile.TemporaryDirectory() as go_dir:
            # Download Vantage data
            if not download_vantage_data(vantage_dir):
                print("Failed to download Vantage data")
                return 1
            
            # Run Go scrapers
            if not run_go_scrapers(go_dir):
                print("Go scrapers failed")
                return 1
            
            # Compare outputs
            services = [
                ("instances.json", "EC2"),
                ("rds/instances.json", "RDS"),
                ("cache/instances.json", "ElastiCache"),
                ("redshift/instances.json", "Redshift"),
                ("opensearch/instances.json", "OpenSearch"),
            ]
            
            for file_path, service_name in services:
                vantage_file = os.path.join(vantage_dir, file_path)
                go_file = os.path.join(go_dir, file_path)
                compare_json_files(vantage_file, go_file, service_name)
            
            print("\n" + "="*50)
            print("Comparison complete!")
            
            # Test caching
            response = input("\nTest caching? (y/n): ")
            if response.lower() == 'y':
                print("\nRunning Go scrapers with cache...")
                if run_go_scrapers(go_dir, use_cache=True):
                    print("✓ Cache test passed")

if __name__ == "__main__":
    sys.exit(main())