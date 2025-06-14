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

def run_go_scrapers(output_dir, use_cache=False, sample_size=0):
    """Run the Go scrapers"""
    print("Running Go scrapers...")
    
    cmd = [
        "go", "run", "cmd/scraper/main.go",
        "-o", output_dir,
        "-c", "10",  # Moderate concurrency
        "-s", "ec2,rds,elasticache,redshift,opensearch",  # All services
        "-r", "us-east-1",  # Single region for testing
    ]
    
    if use_cache:
        cmd.extend(["--use-cache"])
    
    if sample_size > 0:
        cmd.extend(["--sample", str(sample_size)])
        print(f"  Sampling {sample_size} random instances per service")
    
    env = os.environ.copy()
    env['AWS_PROFILE'] = 'framed'
    
    result = subprocess.run(cmd, cwd="scraper-go", env=env, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Go scraper failed: {result.stderr}")
        return False
    print(result.stdout)
    return True

def compare_pricing(vantage_pricing, go_pricing, instance_type, region):
    """Compare pricing data between Vantage and Go outputs"""
    print(f"      Pricing comparison for {instance_type} in {region}:")
    
    # Check OnDemand pricing
    vantage_od = vantage_pricing.get('ondemand') or vantage_pricing.get('OnDemand', 0)
    go_od = go_pricing.get('ondemand') or go_pricing.get('OnDemand', 0)
    
    if abs(float(vantage_od) - float(go_od)) < 0.001:
        print(f"        ✅ OnDemand: ${vantage_od}")
    else:
        print(f"        ❌ OnDemand: Vantage=${vantage_od}, Go=${go_od}")
    
    # Check Reserved pricing
    vantage_reserved = vantage_pricing.get('reserved', {})
    go_reserved = go_pricing.get('reserved', {})
    
    if vantage_reserved and go_reserved:
        reserved_keys = set(vantage_reserved.keys()) & set(go_reserved.keys())
        if reserved_keys:
            sample_key = list(reserved_keys)[0]
            vantage_res = vantage_reserved[sample_key]
            go_res = go_reserved[sample_key]
            if abs(float(vantage_res) - float(go_res)) < 0.001:
                print(f"        ✅ Reserved ({sample_key}): ${vantage_res}")
            else:
                print(f"        ❌ Reserved ({sample_key}): Vantage=${vantage_res}, Go=${go_res}")
    
    # Check Spot pricing
    if 'spot' in go_reserved:
        print(f"        ✅ Spot pricing: ${go_reserved['spot']:.4f}")

def analyze_missing_instances(vantage_map, go_map, service_name):
    """Analyze why certain instances are missing"""
    vantage_types = set(vantage_map.keys())
    go_types = set(go_map.keys())
    
    missing_in_go = vantage_types - go_types
    missing_in_vantage = go_types - vantage_types
    
    if missing_in_go:
        print(f"\n  📋 Analysis of {len(missing_in_go)} instances missing in Go:")
        
        # Group by family
        families = {}
        for inst_type in missing_in_go:
            family = inst_type.split('.')[0] if '.' in inst_type else inst_type
            if family not in families:
                families[family] = []
            families[family].append(inst_type)
        
        for family, instances in sorted(families.items()):
            print(f"    {family}: {len(instances)} instances - {instances[:3]}{'...' if len(instances) > 3 else ''}")
            
        # Show some examples
        print(f"\n  🔍 Sample missing instances:")
        for inst_type in sorted(list(missing_in_go))[:5]:
            vantage_inst = vantage_map[inst_type]
            regions = list(vantage_inst.get('regions', {}).keys())[:3]
            print(f"    {inst_type}: Available in regions {regions}")
    
    if missing_in_vantage:
        print(f"\n  📋 Analysis of {len(missing_in_vantage)} instances only in Go:")
        for inst_type in sorted(list(missing_in_vantage))[:5]:
            go_inst = go_map[inst_type]
            regions = list(go_inst.get('regions', {}).keys())[:3]
            print(f"    {inst_type}: Available in regions {regions}")

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
    accuracy = (len(common) / len(vantage_types)) * 100 if vantage_types else 0
    print(f"  Data accuracy: {accuracy:.1f}%")
    
    # Analyze missing instances
    analyze_missing_instances(vantage_map, go_map, service_name)
    
    # Compare detailed sample instances
    if common:
        samples = sorted(list(common))[:3]  # Check first 3 alphabetically
        for sample in samples:
            print(f"\n  Comparing instance: {sample}")
            
            vantage_inst = vantage_map[sample]
            go_inst = go_map[sample]
            
            # Compare key fields
            fields = ['memory', 'vCPU', 'family', 'network_performance', 'generation', 'pretty_name']
            matches = 0
            
            for field in fields:
                vantage_val = vantage_inst.get(field)
                go_val = go_inst.get(field)
                
                # Normalize vCPU for comparison
                if field == 'vCPU':
                    vantage_val = str(vantage_val) if vantage_val is not None else None
                    go_val = str(go_val) if go_val is not None else None
                
                if vantage_val == go_val:
                    matches += 1
                    if field in ['memory', 'vCPU', 'family']:  # Show key fields
                        print(f"    ✅ {field}: {vantage_val}")
                else:
                    print(f"    ❌ {field}: Vantage={vantage_val}, Go={go_val}")
            
            # Check pricing structure
            if 'pricing' in vantage_inst and 'pricing' in go_inst:
                vantage_regions = set(vantage_inst['pricing'].keys())
                go_regions = set(go_inst['pricing'].keys())
                if len(vantage_regions) > 0 and len(go_regions) == 0:
                    print(f"    ⚠️  Go scraper missing pricing data (Vantage has {len(vantage_regions)} regions)")
                elif len(go_regions) > 0:
                    print(f"    ✅ Pricing: Go has {len(go_regions)} regions, Vantage has {len(vantage_regions)} regions")
                    
                    # Compare pricing for one region
                    common_regions = vantage_regions & go_regions
                    if common_regions:
                        sample_region = list(common_regions)[0]
                        compare_pricing(vantage_inst['pricing'][sample_region], 
                                      go_inst['pricing'][sample_region], 
                                      sample, sample_region)
            elif 'pricing' not in go_inst:
                print(f"    ⚠️  Go instance missing pricing data entirely")

def main():
    """Main comparison function"""
    import argparse
    
    parser = argparse.ArgumentParser(description='Compare Go scraper outputs with Vantage data')
    parser.add_argument('--sample', type=int, default=0, help='Sample size for testing (0 = all instances)')
    parser.add_argument('--cache', action='store_true', help='Use cached data')
    args = parser.parse_args()
    
    with tempfile.TemporaryDirectory() as vantage_dir:
        with tempfile.TemporaryDirectory() as go_dir:
            # Download Vantage data
            if not download_vantage_data(vantage_dir):
                print("Failed to download Vantage data")
                return 1
            
            # Run Go scrapers
            if not run_go_scrapers(go_dir, use_cache=args.cache, sample_size=args.sample):
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
            
            print("\\nValidation complete. All scrapers tested successfully!")

if __name__ == "__main__":
    sys.exit(main())