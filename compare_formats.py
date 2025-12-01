#!/usr/bin/env python3
import json

def load_json(filepath):
    with open(filepath, 'r') as f:
        return json.load(f)

def compare_data_formats():
    go_data = load_json('/Users/cristi/Dropbox (Maestral)/devel/cloud-instances.info/www/azure/instances.json')
    py_data = load_json('/tmp/azure-vantage.json')
    
    # Find a common instance with pricing in both
    go_by_type = {inst['instance_type']: inst for inst in go_data}
    py_by_type = {inst['instance_type']: inst for inst in py_data}
    
    common_types = set(go_by_type.keys()) & set(py_by_type.keys())
    
    # Find instances with pricing in both
    for inst_type in sorted(common_types):
        go_inst = go_by_type[inst_type]
        py_inst = py_by_type[inst_type]
        
        if go_inst.get('pricing') and py_inst.get('pricing'):
            # Find a common region
            go_regions = set(go_inst['pricing'].keys())
            py_regions = set(py_inst['pricing'].keys())
            
            # Try to find a matching region (accounting for naming differences)
            for go_region in go_regions:
                if go_region == 'eastus' and 'us-east' in py_regions:
                    print(f"=== Comparing {inst_type} in {go_region} (us-east) ===")
                    print("\nGo format:")
                    print(json.dumps(go_inst['pricing'][go_region], indent=2))
                    print("\nPython format:")
                    print(json.dumps(py_inst['pricing']['us-east'], indent=2))
                    
                    # Check field differences
                    print("\n=== Field Analysis ===")
                    go_linux = go_inst['pricing'][go_region].get('linux', {})
                    py_linux = py_inst['pricing']['us-east'].get('linux', {})
                    
                    go_fields = set(go_linux.keys())
                    py_fields = set(py_linux.keys())
                    
                    print(f"Go Linux fields: {sorted(go_fields)}")
                    print(f"Python Linux fields: {sorted(py_fields)}")
                    print(f"Missing in Go: {sorted(py_fields - go_fields)}")
                    
                    # Check reserved pricing structure
                    if 'reserved' in go_linux and 'reserved' in py_linux:
                        go_reserved = set(go_linux['reserved'].keys())
                        py_reserved = set(py_linux['reserved'].keys())
                        print(f"\nGo reserved keys: {sorted(go_reserved)}")
                        print(f"Python reserved keys: {sorted(py_reserved)}")
                        print(f"Missing in Go: {sorted(py_reserved - go_reserved)}")
                    
                    # Check Windows
                    go_windows = go_inst['pricing'][go_region].get('windows', {})
                    py_windows = py_inst['pricing']['us-east'].get('windows', {})
                    
                    go_win_fields = set(go_windows.keys())
                    py_win_fields = set(py_windows.keys())
                    
                    print(f"\nGo Windows fields: {sorted(go_win_fields)}")
                    print(f"Python Windows fields: {sorted(py_win_fields)}")
                    print(f"Missing in Go: {sorted(py_win_fields - go_win_fields)}")
                    
                    return inst_type, go_region
    
    return None, None

def check_basic_instances():
    """Check which instances have 'basic' pricing"""
    py_data = load_json('/tmp/azure-vantage.json')
    
    print("\n=== Instances with 'basic' pricing ===")
    basic_instances = []
    
    for inst in py_data:
        if inst.get('pricing'):
            for region, region_data in inst['pricing'].items():
                if region_data.get('linux', {}).get('basic'):
                    basic_instances.append(inst['instance_type'])
                    break
    
    print(f"Found {len(basic_instances)} instances with basic pricing:")
    print(basic_instances[:10])

if __name__ == '__main__':
    inst_type, region = compare_data_formats()
    check_basic_instances()