#!/usr/bin/env python3
import json
from collections import defaultdict

def normalize_region_name(region):
    """Convert between Azure internal names and hyphenated names"""
    # Common mappings
    mappings = {
        'eastus': 'us-east',
        'eastus2': 'us-east-2',
        'westus': 'us-west',
        'westus2': 'us-west-2',
        'westus3': 'us-west-3',
        'centralus': 'us-central',
        'northcentralus': 'us-north-central',
        'southcentralus': 'us-south-central',
        'westcentralus': 'us-west-central',
        'westeurope': 'europe-west',
        'northeurope': 'europe-north',
        'australiaeast': 'australia-east',
        'australiasoutheast': 'australia-southeast',
        'australiacentral': 'australia-central',
        'canadacentral': 'canada-central',
        'canadaeast': 'canada-east',
        'japaneast': 'japan-east',
        'japanwest': 'japan-west',
        'koreacentral': 'korea-central',
        'koreasouth': 'korea-south',
        'southeastasia': 'asia-pacific-southeast',
        'eastasia': 'asia-pacific-east',
        'brazilsouth': 'brazil-south',
        'uksouth': 'uk-south',
        'ukwest': 'uk-west',
        'francecentral': 'france-central',
        'southafricanorth': 'south-africa-north',
        'centralindia': 'india-central',
        'southindia': 'south-india',
        'westindia': 'india-west',
        'uaenorth': 'uae-north',
        'switzerlandnorth': 'switzerland-north',
        'germanywestcentral': 'germany-west-central',
        'norwayeast': 'norway-east'
    }
    
    # If it's already a hyphenated name, return it
    if '-' in region:
        return region
    
    # Convert Azure internal name to hyphenated
    return mappings.get(region, region)

def load_json(filepath):
    with open(filepath, 'r') as f:
        return json.load(f)

def compare_pricing_normalized(go_data, python_data):
    """Compare pricing with normalized region names"""
    # Create instance mappings
    go_by_type = {inst['instance_type']: inst for inst in go_data}
    py_by_type = {inst['instance_type']: inst for inst in python_data}
    
    # Find common instances
    common_types = set(go_by_type.keys()) & set(py_by_type.keys())
    
    print(f"Total instances - Go: {len(go_by_type)}, Python: {len(py_by_type)}")
    print(f"Common instance types: {len(common_types)}")
    
    # Compare pricing for common instances
    pricing_matches = 0
    pricing_differences = []
    instances_with_both_pricing = 0
    
    # Sample detailed comparisons
    sample_comparisons = []
    
    for inst_type in sorted(common_types):
        go_inst = go_by_type[inst_type]
        py_inst = py_by_type[inst_type]
        
        go_pricing = go_inst.get('pricing', {})
        py_pricing = py_inst.get('pricing', {})
        
        if not go_pricing or not py_pricing:
            continue
        
        instances_with_both_pricing += 1
        
        # Try to find matching regions
        matched_any = False
        for go_region, go_region_data in go_pricing.items():
            normalized_region = normalize_region_name(go_region)
            
            # Look for this region in Python data
            py_region_data = None
            if normalized_region in py_pricing:
                py_region_data = py_pricing[normalized_region]
            elif go_region in py_pricing:
                py_region_data = py_pricing[go_region]
            
            if py_region_data and go_region_data:
                matched_any = True
                
                # Compare Linux on-demand pricing
                go_linux = go_region_data.get('linux', {})
                py_linux = py_region_data.get('linux', {})
                
                if 'ondemand' in go_linux and 'ondemand' in py_linux:
                    go_price = go_linux['ondemand']
                    py_price = py_linux['ondemand']
                    
                    if abs(go_price - py_price) < 0.001:
                        pricing_matches += 1
                    else:
                        pricing_differences.append({
                            'instance': inst_type,
                            'region': go_region,
                            'normalized_region': normalized_region,
                            'go_price': go_price,
                            'py_price': py_price,
                            'diff': go_price - py_price,
                            'diff_pct': ((go_price - py_price) / py_price * 100) if py_price > 0 else 0
                        })
                
                # Add to sample if we haven't collected enough
                if len(sample_comparisons) < 5 and go_linux and py_linux:
                    sample_comparisons.append({
                        'instance': inst_type,
                        'region': go_region,
                        'go_data': go_region_data,
                        'py_data': py_region_data
                    })
                
                break  # Found a match, move to next instance
        
        if matched_any:
            pass  # Already counted above
    
    print(f"\nInstances with pricing in both datasets: {instances_with_both_pricing}")
    print(f"Exact price matches: {pricing_matches}")
    print(f"Price differences: {len(pricing_differences)}")
    
    # Show some differences
    if pricing_differences:
        print("\n=== Top Price Differences ===")
        sorted_diffs = sorted(pricing_differences, key=lambda x: abs(x['diff']), reverse=True)
        print(f"{'Instance':<20} {'Region':<20} {'Go Price':>10} {'Py Price':>10} {'Diff':>10} {'Diff %':>8}")
        print("-" * 88)
        for diff in sorted_diffs[:10]:
            print(f"{diff['instance']:<20} {diff['region']:<20} "
                  f"{diff['go_price']:>10.4f} {diff['py_price']:>10.4f} "
                  f"{diff['diff']:>10.4f} {diff['diff_pct']:>7.1f}%")
    
    # Show sample comparisons
    print("\n=== Sample Detailed Comparisons ===")
    for sample in sample_comparisons[:3]:
        print(f"\nInstance: {sample['instance']}, Region: {sample['region']}")
        print("Go data:")
        print(json.dumps(sample['go_data'], indent=2))
        print("Python data:")
        print(json.dumps(sample['py_data'], indent=2))
    
    # Check reserved pricing format differences
    print("\n=== Reserved Pricing Key Analysis ===")
    go_reserved_keys = set()
    py_reserved_keys = set()
    
    for inst in go_data:
        if inst.get('pricing'):
            for region_data in inst['pricing'].values():
                for os_data in region_data.values():
                    if isinstance(os_data, dict) and 'reserved' in os_data:
                        go_reserved_keys.update(os_data['reserved'].keys())
    
    for inst in python_data:
        if inst.get('pricing'):
            for region_data in inst['pricing'].values():
                for os_data in region_data.values():
                    if isinstance(os_data, dict) and 'reserved' in os_data:
                        py_reserved_keys.update(os_data['reserved'].keys())
    
    print(f"Go reserved keys: {sorted(go_reserved_keys)}")
    print(f"Python reserved keys: {sorted(py_reserved_keys)}")
    
    return pricing_matches, pricing_differences

def main():
    go_data = load_json('/Users/cristi/Dropbox (Maestral)/devel/cloud-instances.info/www/azure/instances.json')
    python_data = load_json('/tmp/azure-vantage.json')
    
    print("=== Normalized Azure Pricing Comparison ===\n")
    
    matches, differences = compare_pricing_normalized(go_data, python_data)
    
    # Check instance coverage
    print("\n=== Instance Coverage ===")
    go_types = {inst['instance_type'] for inst in go_data}
    py_types = {inst['instance_type'] for inst in python_data}
    
    print(f"Instances only in Go scraper: {len(go_types - py_types)}")
    print("Examples:", sorted(list(go_types - py_types))[:10])
    
    print(f"\nInstances only in Python scraper: {len(py_types - go_types)}")
    print("Examples:", sorted(list(py_types - go_types))[:10])

if __name__ == '__main__':
    main()