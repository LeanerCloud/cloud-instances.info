#!/usr/bin/env python3
import json
import sys
from collections import defaultdict

def load_json(filepath):
    with open(filepath, 'r') as f:
        return json.load(f)

def normalize_instance_type(instance_type):
    """Normalize instance type for comparison"""
    return instance_type.lower().replace('_', '').replace('-', '')

def compare_pricing(go_data, python_data):
    # Create mapping of normalized instance types
    go_instances = {normalize_instance_type(i['instance_type']): i for i in go_data}
    python_instances = {normalize_instance_type(i['instance_type']): i for i in python_data}
    
    # Also create mapping by original names for debugging
    go_by_name = {i['instance_type']: i for i in go_data}
    python_by_name = {i['instance_type']: i for i in python_data}
    
    # Statistics
    stats = {
        'total_go': len(go_instances),
        'total_python': len(python_instances),
        'common_instances': 0,
        'go_only': [],
        'python_only': [],
        'pricing_differences': defaultdict(int),
        'missing_pricing': defaultdict(int),
        'price_comparison': [],
        'exact_matches': 0,
        'pricing_coverage': {'go': 0, 'python': 0}
    }
    
    # Find instances only in one dataset
    go_types = set(go_instances.keys())
    python_types = set(python_instances.keys())
    
    stats['go_only'] = list(go_types - python_types)
    stats['python_only'] = list(python_types - go_types)
    stats['common_instances'] = len(go_types & python_types)
    
    # Count instances with pricing
    for inst in go_data:
        if inst.get('pricing') and any(inst['pricing'].values()):
            stats['pricing_coverage']['go'] += 1
    
    for inst in python_data:
        if inst.get('pricing') and any(inst['pricing'].values()):
            stats['pricing_coverage']['python'] += 1
    
    # Compare pricing for common instances
    exact_price_matches = 0
    instances_compared = 0
    
    for norm_type in go_types & python_types:
        go_inst = go_instances[norm_type]
        python_inst = python_instances[norm_type]
        
        # Sample a few regions for comparison
        regions_to_check = ['eastus', 'westus2', 'westeurope']
        
        instance_has_match = True
        
        for region in regions_to_check:
            go_pricing = go_inst.get('pricing', {}).get(region, {})
            python_pricing = python_inst.get('pricing', {}).get(region, {})
            
            if not go_pricing and not python_pricing:
                continue
                
            # Compare Linux pricing
            if 'linux' in go_pricing or 'linux' in python_pricing:
                go_linux = go_pricing.get('linux', {})
                python_linux = python_pricing.get('linux', {})
                
                # Check on-demand pricing
                if 'ondemand' in go_linux and 'ondemand' in python_linux:
                    go_price = go_linux['ondemand']
                    python_price = python_linux['ondemand']
                    if abs(go_price - python_price) > 0.001:  # Allow small rounding differences
                        instance_has_match = False
                        stats['price_comparison'].append({
                            'instance': go_inst['instance_type'],
                            'region': region,
                            'type': 'linux_ondemand',
                            'go_price': go_price,
                            'python_price': python_price,
                            'diff': go_price - python_price,
                            'diff_pct': ((go_price - python_price) / python_price * 100) if python_price > 0 else 0
                        })
                elif 'ondemand' in go_linux and 'ondemand' not in python_linux:
                    stats['missing_pricing']['python_missing_ondemand'] += 1
                elif 'ondemand' not in go_linux and 'ondemand' in python_linux:
                    stats['missing_pricing']['go_missing_ondemand'] += 1
                
                # Check reserved pricing
                go_reserved = go_linux.get('reserved', {})
                python_reserved = python_linux.get('reserved', {})
                
                # Check 1-year reserved
                go_1yr = go_reserved.get('yrTerm1Standard.allUpfront', 0)
                python_1yr = python_reserved.get('yrTerm1Standard.allUpfront', 0)
                
                if go_1yr > 0 and python_1yr > 0:
                    if abs(go_1yr - python_1yr) > 0.001:
                        stats['price_comparison'].append({
                            'instance': go_inst['instance_type'],
                            'region': region,
                            'type': 'linux_reserved_1yr',
                            'go_price': go_1yr,
                            'python_price': python_1yr,
                            'diff': go_1yr - python_1yr,
                            'diff_pct': ((go_1yr - python_1yr) / python_1yr * 100) if python_1yr > 0 else 0
                        })
                elif go_1yr > 0 and python_1yr == 0:
                    stats['missing_pricing']['python_missing_reserved'] += 1
                elif go_1yr == 0 and python_1yr > 0:
                    stats['missing_pricing']['go_missing_reserved'] += 1
        
        if instance_has_match and go_inst.get('pricing') and python_inst.get('pricing'):
            stats['exact_matches'] += 1
    
    return stats

def print_summary(stats):
    print("=== Azure Pricing Comparison Summary ===\n")
    
    print(f"Total instances in Go scraper: {stats['total_go']}")
    print(f"Total instances in Python scraper: {stats['total_python']}")
    print(f"Common instances: {stats['common_instances']}")
    print(f"Go-only instances: {len(stats['go_only'])}")
    print(f"Python-only instances: {len(stats['python_only'])}")
    
    print(f"\nInstances with pricing data:")
    print(f"  Go scraper: {stats['pricing_coverage']['go']} ({stats['pricing_coverage']['go']/stats['total_go']*100:.1f}%)")
    print(f"  Python scraper: {stats['pricing_coverage']['python']} ({stats['pricing_coverage']['python']/stats['total_python']*100:.1f}%)")
    
    print(f"\nExact price matches among common instances: {stats['exact_matches']}")
    
    print("\n=== Missing Pricing Data ===")
    for key, count in stats['missing_pricing'].items():
        print(f"{key}: {count}")
    
    print("\n=== Price Differences ===")
    if stats['price_comparison']:
        # Sort by absolute difference
        differences = sorted(stats['price_comparison'], key=lambda x: abs(x['diff']), reverse=True)
        
        print(f"\nFound {len(differences)} pricing differences")
        print("\nTop 10 largest differences:")
        print(f"{'Instance':<20} {'Region':<15} {'Type':<20} {'Go Price':>10} {'Py Price':>10} {'Diff':>10} {'Diff %':>8}")
        print("-" * 103)
        
        for diff in differences[:10]:
            print(f"{diff['instance']:<20} {diff['region']:<15} {diff['type']:<20} "
                  f"{diff['go_price']:>10.4f} {diff['python_price']:>10.4f} "
                  f"{diff['diff']:>10.4f} {diff['diff_pct']:>7.1f}%")
    else:
        print("No significant pricing differences found!")
    
    # Show some examples of unique instances
    if stats['go_only']:
        print(f"\n=== Examples of Go-only instances ({len(stats['go_only'])} total) ===")
        for inst in stats['go_only'][:5]:
            print(f"  - {inst}")
    
    if stats['python_only']:
        print(f"\n=== Examples of Python-only instances ({len(stats['python_only'])} total) ===")
        for inst in stats['python_only'][:5]:
            print(f"  - {inst}")

def compare_specific_instance(go_data, python_data, instance_type):
    """Compare pricing for a specific instance type"""
    go_inst = None
    python_inst = None
    
    for inst in go_data:
        if inst['instance_type'].lower() == instance_type.lower():
            go_inst = inst
            break
    
    for inst in python_data:
        if inst['instance_type'].lower() == instance_type.lower():
            python_inst = inst
            break
    
    print(f"\n=== Detailed comparison for {instance_type} ===")
    
    if not go_inst:
        print(f"Instance {instance_type} not found in Go scraper data")
    if not python_inst:
        print(f"Instance {instance_type} not found in Python scraper data")
    
    if go_inst and python_inst:
        # Compare eastus pricing
        region = 'eastus'
        go_pricing = go_inst.get('pricing', {}).get(region, {})
        python_pricing = python_inst.get('pricing', {}).get(region, {})
        
        print(f"\nRegion: {region}")
        print("\nGo scraper pricing:")
        print(json.dumps(go_pricing, indent=2))
        print("\nPython scraper pricing:")
        print(json.dumps(python_pricing, indent=2))

def main():
    go_data = load_json('/Users/cristi/Dropbox (Maestral)/devel/cloud-instances.info/www/azure/instances.json')
    python_data = load_json('/tmp/azure-vantage.json')
    
    stats = compare_pricing(go_data, python_data)
    print_summary(stats)
    
    # Compare a few specific instances
    compare_specific_instance(go_data, python_data, 'd4sv3')
    compare_specific_instance(go_data, python_data, 'b1s')

if __name__ == '__main__':
    main()