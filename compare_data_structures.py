#!/usr/bin/env python3
import json
import sys
from collections import defaultdict
from typing import Dict, List, Any, Set

def load_json(filepath):
    with open(filepath, 'r') as f:
        return json.load(f)

def get_data_structure(data: List[Dict], name: str) -> Dict[str, Any]:
    """Analyze the data structure of instances"""
    structure = {
        'name': name,
        'total_instances': len(data),
        'fields': set(),
        'field_types': defaultdict(set),
        'pricing_structure': {
            'regions': set(),
            'os_types': set(),
            'pricing_fields': set(),
            'reserved_keys': set()
        },
        'sample_instance': None,
        'sample_pricing': None,
        'instances_with_pricing': 0,
        'instances_with_actual_pricing_data': 0
    }
    
    # Analyze all instances
    for inst in data:
        # Collect all fields
        structure['fields'].update(inst.keys())
        
        # Track field types
        for field, value in inst.items():
            structure['field_types'][field].add(type(value).__name__)
        
        # Analyze pricing structure
        if 'pricing' in inst and inst['pricing']:
            structure['instances_with_pricing'] += 1
            
            # Check if there's actual pricing data (not just empty dicts)
            has_actual_data = False
            
            for region, region_data in inst['pricing'].items():
                structure['pricing_structure']['regions'].add(region)
                
                if isinstance(region_data, dict) and region_data:
                    has_actual_data = True
                    
                    for os_type, os_data in region_data.items():
                        structure['pricing_structure']['os_types'].add(os_type)
                        
                        if isinstance(os_data, dict):
                            structure['pricing_structure']['pricing_fields'].update(os_data.keys())
                            
                            # Check for reserved pricing structure
                            if 'reserved' in os_data and isinstance(os_data['reserved'], dict):
                                structure['pricing_structure']['reserved_keys'].update(os_data['reserved'].keys())
            
            if has_actual_data:
                structure['instances_with_actual_pricing_data'] += 1
    
    # Get a sample instance with pricing
    for inst in data:
        if inst.get('pricing'):
            for region, region_data in inst['pricing'].items():
                if region_data and any(region_data.values()):
                    structure['sample_instance'] = inst['instance_type']
                    structure['sample_pricing'] = {
                        'region': region,
                        'data': region_data
                    }
                    break
            if structure['sample_pricing']:
                break
    
    return structure

def compare_structures(go_struct: Dict, py_struct: Dict) -> Dict[str, Any]:
    """Compare two data structures"""
    comparison = {
        'field_differences': {
            'go_only': list(go_struct['fields'] - py_struct['fields']),
            'python_only': list(py_struct['fields'] - go_struct['fields']),
            'common': list(go_struct['fields'] & py_struct['fields'])
        },
        'pricing_differences': {
            'regions': {
                'go_only': list(go_struct['pricing_structure']['regions'] - py_struct['pricing_structure']['regions']),
                'python_only': list(py_struct['pricing_structure']['regions'] - go_struct['pricing_structure']['regions']),
                'common': len(go_struct['pricing_structure']['regions'] & py_struct['pricing_structure']['regions'])
            },
            'os_types': {
                'go': list(go_struct['pricing_structure']['os_types']),
                'python': list(py_struct['pricing_structure']['os_types'])
            },
            'pricing_fields': {
                'go': list(go_struct['pricing_structure']['pricing_fields']),
                'python': list(py_struct['pricing_structure']['pricing_fields'])
            },
            'reserved_keys': {
                'go': list(go_struct['pricing_structure']['reserved_keys']),
                'python': list(py_struct['pricing_structure']['reserved_keys'])
            }
        }
    }
    
    return comparison

def print_structure_analysis(go_struct: Dict, py_struct: Dict, comparison: Dict):
    """Print detailed analysis of data structures"""
    print("=== Data Structure Comparison ===\n")
    
    print(f"Go Scraper:")
    print(f"  Total instances: {go_struct['total_instances']}")
    print(f"  Instances with pricing key: {go_struct['instances_with_pricing']}")
    print(f"  Instances with actual pricing data: {go_struct['instances_with_actual_pricing_data']}")
    print(f"  Total fields: {len(go_struct['fields'])}")
    print(f"  Pricing regions: {len(go_struct['pricing_structure']['regions'])}")
    
    print(f"\nPython Scraper:")
    print(f"  Total instances: {py_struct['total_instances']}")
    print(f"  Instances with pricing key: {py_struct['instances_with_pricing']}")
    print(f"  Instances with actual pricing data: {py_struct['instances_with_actual_pricing_data']}")
    print(f"  Total fields: {len(py_struct['fields'])}")
    print(f"  Pricing regions: {len(py_struct['pricing_structure']['regions'])}")
    
    print("\n=== Field Differences ===")
    print(f"Common fields ({len(comparison['field_differences']['common'])}): {sorted(comparison['field_differences']['common'])}")
    
    if comparison['field_differences']['go_only']:
        print(f"\nGo-only fields: {sorted(comparison['field_differences']['go_only'])}")
    
    if comparison['field_differences']['python_only']:
        print(f"\nPython-only fields: {sorted(comparison['field_differences']['python_only'])}")
    
    print("\n=== Pricing Structure ===")
    print(f"OS Types:")
    print(f"  Go: {comparison['pricing_differences']['os_types']['go']}")
    print(f"  Python: {comparison['pricing_differences']['os_types']['python']}")
    
    print(f"\nPricing Fields:")
    print(f"  Go: {sorted(comparison['pricing_differences']['pricing_fields']['go'])}")
    print(f"  Python: {sorted(comparison['pricing_differences']['pricing_fields']['python'])}")
    
    print(f"\nReserved Pricing Keys:")
    print(f"  Go: {sorted(comparison['pricing_differences']['reserved_keys']['go'])}")
    print(f"  Python: {sorted(comparison['pricing_differences']['reserved_keys']['python'])}")
    
    print(f"\nRegion Coverage:")
    print(f"  Common regions: {comparison['pricing_differences']['regions']['common']}")
    if comparison['pricing_differences']['regions']['go_only']:
        print(f"  Go-only regions: {comparison['pricing_differences']['regions']['go_only'][:5]}...")
    if comparison['pricing_differences']['regions']['python_only']:
        print(f"  Python-only regions: {comparison['pricing_differences']['regions']['python_only'][:5]}...")
    
    print("\n=== Sample Pricing Structure ===")
    
    if go_struct['sample_pricing']:
        print(f"\nGo Scraper - Instance: {go_struct['sample_instance']}, Region: {go_struct['sample_pricing']['region']}")
        print(json.dumps(go_struct['sample_pricing']['data'], indent=2))
    else:
        print("\nGo Scraper - No pricing data found")
    
    if py_struct['sample_pricing']:
        print(f"\nPython Scraper - Instance: {py_struct['sample_instance']}, Region: {py_struct['sample_pricing']['region']}")
        print(json.dumps(py_struct['sample_pricing']['data'], indent=2))
    else:
        print("\nPython Scraper - No pricing data found")

def check_specific_instances(go_data: List[Dict], py_data: List[Dict], instance_types: List[str]):
    """Check specific instances in both datasets"""
    print("\n=== Specific Instance Comparison ===")
    
    go_by_type = {inst['instance_type']: inst for inst in go_data}
    py_by_type = {inst['instance_type']: inst for inst in py_data}
    
    for inst_type in instance_types:
        print(f"\n{inst_type}:")
        
        if inst_type in go_by_type:
            go_inst = go_by_type[inst_type]
            go_pricing = go_inst.get('pricing', {})
            go_regions_with_data = sum(1 for r_data in go_pricing.values() if r_data)
            print(f"  Go: Found, pricing for {go_regions_with_data} regions")
            
            # Show sample pricing if available
            for region in ['eastus', 'westus2', 'westeurope']:
                if region in go_pricing and go_pricing[region]:
                    print(f"    {region}: {list(go_pricing[region].keys())}")
                    break
        else:
            print(f"  Go: Not found")
        
        if inst_type in py_by_type:
            py_inst = py_by_type[inst_type]
            py_pricing = py_inst.get('pricing', {})
            py_regions_with_data = sum(1 for r_data in py_pricing.values() if r_data)
            print(f"  Python: Found, pricing for {py_regions_with_data} regions")
            
            # Show sample pricing if available
            for region in ['eastus', 'westus2', 'westeurope']:
                if region in py_pricing and py_pricing[region]:
                    print(f"    {region}: {list(py_pricing[region].keys())}")
                    break
        else:
            print(f"  Python: Not found")

def main():
    print("Loading data files...")
    go_data = load_json('/Users/cristi/Dropbox (Maestral)/devel/cloud-instances.info/www/azure/instances.json')
    py_data = load_json('/tmp/azure-vantage.json')
    
    print("Analyzing data structures...")
    go_struct = get_data_structure(go_data, 'Go Scraper')
    py_struct = get_data_structure(py_data, 'Python Scraper')
    
    comparison = compare_structures(go_struct, py_struct)
    
    print_structure_analysis(go_struct, py_struct, comparison)
    
    # Check some specific instances
    check_specific_instances(go_data, py_data, [
        'Standard_D4s_v3', 'Standard_B1s', 'Standard_E2s_v3', 
        'Standard_F2s_v2', 'Standard_NC6', 'Standard_M128s'
    ])

if __name__ == '__main__':
    main()