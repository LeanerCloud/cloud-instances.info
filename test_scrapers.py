#!/usr/bin/env python3
"""Test script to compare Python and Go scraper outputs"""
import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path

def run_go_ec2_scraper(output_dir):
    """Run the Go EC2 scraper"""
    print("Running Go EC2 scraper...")
    cmd = [
        "go", "run", "cmd/scraper/main.go",
        "-s", "ec2",
        "-c", "10",
        "-r", "us-east-1,us-west-2",
        "-o", output_dir
    ]
    env = os.environ.copy()
    env['AWS_PROFILE'] = 'framed'
    
    result = subprocess.run(cmd, cwd="scraper-go", env=env, capture_output=True, text=True)
    if result.returncode != 0:
        print(f"Go scraper failed: {result.stderr}")
        return False
    print(result.stdout)
    return True

def run_python_ec2_scraper(output_file):
    """Run the Python EC2 scraper (limited regions for speed)"""
    print("Running Python EC2 scraper...")
    
    # Import and run directly
    os.environ['AWS_PROFILE'] = 'framed'
    import ec2
    
    # Temporarily limit regions for faster testing
    original_get_regions = ec2.get_regions
    ec2.get_regions = lambda: ["us-east-1", "us-west-2"]
    
    try:
        from scrape import scrape
        scrape(output_file)
        return True
    except Exception as e:
        print(f"Python scraper failed: {e}")
        return False
    finally:
        ec2.get_regions = original_get_regions

def compare_outputs(python_file, go_file):
    """Compare Python and Go outputs"""
    print("\n=== Comparing EC2 outputs ===")
    
    # Load files
    with open(python_file, 'r') as f:
        python_data = json.load(f)
    with open(go_file, 'r') as f:
        go_data = json.load(f)
    
    print(f"Python instances: {len(python_data)}")
    print(f"Go instances: {len(go_data)}")
    
    # Create maps
    python_map = {inst['instance_type']: inst for inst in python_data}
    go_map = {inst['instance_type']: inst for inst in go_data}
    
    # Find differences
    python_types = set(python_map.keys())
    go_types = set(go_map.keys())
    
    missing_in_go = python_types - go_types
    missing_in_python = go_types - python_types
    common = python_types & go_types
    
    if missing_in_go:
        print(f"\nMissing in Go ({len(missing_in_go)}): {list(missing_in_go)[:5]}...")
    if missing_in_python:
        print(f"\nMissing in Python ({len(missing_in_python)}): {list(missing_in_python)[:5]}...")
    
    print(f"\nCommon instances: {len(common)}")
    
    # Compare a sample instance
    if common:
        sample = list(common)[0]
        print(f"\nComparing instance: {sample}")
        
        py_inst = python_map[sample]
        go_inst = go_map[sample]
        
        # Compare key fields
        fields = ['memory', 'vCPU', 'family', 'network_performance', 'generation']
        for field in fields:
            py_val = py_inst.get(field)
            go_val = go_inst.get(field)
            
            # Special handling for vCPU
            if field == 'vCPU':
                py_val = str(py_val) if py_val is not None else None
                go_val = str(go_val) if go_val is not None else None
            
            if py_val != go_val:
                print(f"  ❌ {field}: Python={py_val}, Go={go_val}")
            else:
                print(f"  ✅ {field}: {py_val}")

def main():
    """Main test function"""
    # Create temp directories
    with tempfile.TemporaryDirectory() as python_dir:
        with tempfile.TemporaryDirectory() as go_dir:
            python_file = os.path.join(python_dir, "instances.json")
            go_file = os.path.join(go_dir, "instances.json")
            
            # Run scrapers
            if not run_python_ec2_scraper(python_file):
                print("Python scraper failed")
                return 1
            
            if not run_go_ec2_scraper(go_dir):
                print("Go scraper failed")
                return 1
            
            # Compare outputs
            compare_outputs(python_file, go_file)
            
            # Ask user
            response = input("\nOutputs look good? (y/n): ")
            if response.lower() == 'y':
                print("Great! The Go scraper is working correctly.")
                return 0
            else:
                print("Please fix the differences and try again.")
                return 1

if __name__ == "__main__":
    sys.exit(main())