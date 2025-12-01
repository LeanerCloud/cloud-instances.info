#!/usr/bin/env python3
import json

def load_json(filepath):
    with open(filepath, 'r') as f:
        return json.load(f)

def analyze_pricing_differences():
    go_data = load_json('/Users/cristi/Dropbox (Maestral)/devel/cloud-instances.info/www/azure/instances.json')
    py_data = load_json('/tmp/azure-vantage.json')
    
    # Create instance mappings
    go_by_type = {inst['instance_type']: inst for inst in go_data}
    py_by_type = {inst['instance_type']: inst for inst in py_data}
    
    # Analyze coverage
    print("=== Instance Coverage Analysis ===")
    print(f"Go scraper total: {len(go_by_type)}")
    print(f"Python scraper total: {len(py_by_type)}")
    
    # Count instances with pricing
    go_with_pricing = sum(1 for inst in go_data if inst.get('pricing') and any(inst['pricing'].values()))
    py_with_pricing = sum(1 for inst in py_data if inst.get('pricing') and any(inst['pricing'].values()))
    
    print(f"\nInstances with pricing:")
    print(f"  Go: {go_with_pricing}/{len(go_data)} ({go_with_pricing/len(go_data)*100:.1f}%)")
    print(f"  Python: {py_with_pricing}/{len(py_data)} ({py_with_pricing/len(py_data)*100:.1f}%)")
    
    # Analyze instance type differences
    go_types = set(go_by_type.keys())
    py_types = set(py_by_type.keys())
    
    go_only = go_types - py_types
    py_only = py_types - go_types
    
    print(f"\nInstance type differences:")
    print(f"  Go-only: {len(go_only)}")
    print(f"  Python-only: {len(py_only)}")
    
    # Analyze why Go has fewer instances with pricing
    print("\n=== Analyzing Go instances without pricing ===")
    go_no_pricing = [inst for inst in go_data if not inst.get('pricing') or not any(inst['pricing'].values())]
    
    # Group by patterns
    patterns = {}
    for inst in go_no_pricing[:50]:  # Check first 50
        family = inst['instance_type'][0]  # First letter usually indicates family
        if 'promo' in inst['instance_type']:
            pattern = 'promo'
        elif 'v1' in inst['instance_type'] and 'v1' not in inst['instance_type'].replace('v10', '').replace('v11', '').replace('v12', '').replace('v13', '').replace('v14', '').replace('v15', ''):
            pattern = 'v1_series'
        elif any(x in inst['instance_type'] for x in ['cc', 'dc', 'ec']):
            pattern = 'confidential'
        else:
            pattern = family + '_family'
        
        if pattern not in patterns:
            patterns[pattern] = []
        patterns[pattern].append(inst['instance_type'])
    
    print("Sample instances without pricing (grouped by pattern):")
    for pattern, instances in sorted(patterns.items()):
        print(f"  {pattern}: {instances[:5]}")
    
    # Check specific instances that Python has but Go doesn't
    print("\n=== Checking specific Python-only instances ===")
    for inst_type in sorted(list(py_only))[:10]:
        py_inst = py_by_type[inst_type]
        has_pricing = bool(py_inst.get('pricing') and any(py_inst['pricing'].values()))
        print(f"  {inst_type}: has_pricing={has_pricing}")
    
    # Analyze pricing key differences
    print("\n=== Pricing Key Analysis ===")
    go_pricing_keys = set()
    py_pricing_keys = set()
    
    for inst in go_data:
        if inst.get('pricing'):
            for region_data in inst['pricing'].values():
                for os_data in region_data.values():
                    if isinstance(os_data, dict):
                        go_pricing_keys.update(os_data.keys())
    
    for inst in py_data:
        if inst.get('pricing'):
            for region_data in inst['pricing'].values():
                for os_data in region_data.values():
                    if isinstance(os_data, dict):
                        py_pricing_keys.update(os_data.keys())
    
    print(f"Go pricing keys: {sorted(go_pricing_keys)}")
    print(f"Python pricing keys: {sorted(py_pricing_keys)}")
    print(f"Missing in Go: {sorted(py_pricing_keys - go_pricing_keys)}")

if __name__ == '__main__':
    analyze_pricing_differences()