#!/usr/bin/env python3
import json

def final_comparison():
    """Final comparison of Go vs Python scraper outputs"""
    
    go_data = json.load(open('/Users/cristi/Dropbox (Maestral)/devel/cloud-instances.info/www/azure/instances.json'))
    py_data = json.load(open('/tmp/azure-vantage.json'))
    
    # Create mappings
    go_by_type = {inst['instance_type']: inst for inst in go_data}
    py_by_type = {inst['instance_type']: inst for inst in py_data}
    
    # Find common instances with pricing
    common_with_pricing = []
    for inst_type in sorted(set(go_by_type.keys()) & set(py_by_type.keys())):
        go_inst = go_by_type[inst_type]
        py_inst = py_by_type[inst_type]
        
        if go_inst.get('pricing') and py_inst.get('pricing'):
            # Check eastus/us-east
            if 'eastus' in go_inst['pricing'] and 'us-east' in py_inst['pricing']:
                common_with_pricing.append({
                    'type': inst_type,
                    'go': go_inst['pricing']['eastus'],
                    'py': py_inst['pricing']['us-east']
                })
    
    print(f"=== Final Comparison Summary ===")
    print(f"Total common instances with pricing: {len(common_with_pricing)}")
    
    # Analyze field differences
    field_stats = {
        'linux': {
            'ondemand': {'match': 0, 'diff': 0, 'missing_go': 0, 'missing_py': 0},
            'lowpriority': {'match': 0, 'diff': 0, 'missing_go': 0, 'missing_py': 0},
            'spot_min': {'match': 0, 'diff': 0, 'missing_go': 0, 'missing_py': 0},
            'basic': {'match': 0, 'diff': 0, 'missing_go': 0, 'missing_py': 0},
            'hybridbenefit': {'match': 0, 'diff': 0, 'missing_go': 0, 'missing_py': 0},
            'reserved': {'match': 0, 'diff': 0, 'missing_go': 0, 'missing_py': 0}
        },
        'windows': {
            'ondemand': {'match': 0, 'diff': 0, 'missing_go': 0, 'missing_py': 0},
            'lowpriority': {'match': 0, 'diff': 0, 'missing_go': 0, 'missing_py': 0},
            'spot_min': {'match': 0, 'diff': 0, 'missing_go': 0, 'missing_py': 0},
            'hybridbenefit': {'match': 0, 'diff': 0, 'missing_go': 0, 'missing_py': 0},
            'reserved': {'match': 0, 'diff': 0, 'missing_go': 0, 'missing_py': 0}
        }
    }
    
    for item in common_with_pricing:
        for os in ['linux', 'windows']:
            go_os = item['go'].get(os, {})
            py_os = item['py'].get(os, {})
            
            # Check each field
            for field in ['ondemand', 'lowpriority', 'spot_min', 'basic', 'hybridbenefit']:
                go_val = go_os.get(field)
                py_val = py_os.get(field)
                
                if go_val is not None and py_val is not None:
                    if abs(go_val - py_val) < 0.001:
                        field_stats[os][field]['match'] += 1
                    else:
                        field_stats[os][field]['diff'] += 1
                elif go_val is None and py_val is not None:
                    field_stats[os][field]['missing_go'] += 1
                elif go_val is not None and py_val is None:
                    field_stats[os][field]['missing_py'] += 1
            
            # Check reserved
            go_res = go_os.get('reserved', {})
            py_res = py_os.get('reserved', {})
            
            if go_res and py_res:
                # Check standard reserved pricing only (Go doesn't have Savings)
                if 'yrTerm1Standard.allUpfront' in go_res and 'yrTerm1Standard.allUpfront' in py_res:
                    if abs(go_res['yrTerm1Standard.allUpfront'] - py_res['yrTerm1Standard.allUpfront']) < 0.001:
                        field_stats[os]['reserved']['match'] += 1
                    else:
                        field_stats[os]['reserved']['diff'] += 1
                else:
                    field_stats[os]['reserved']['missing_go'] += 1
            elif not go_res and py_res:
                field_stats[os]['reserved']['missing_go'] += 1
            elif go_res and not py_res:
                field_stats[os]['reserved']['missing_py'] += 1
    
    # Print results
    print("\n=== Field Comparison Results ===")
    for os in ['linux', 'windows']:
        print(f"\n{os.upper()}:")
        for field, stats in field_stats[os].items():
            total = sum(stats.values())
            if total > 0:
                match_pct = (stats['match'] / total * 100) if total > 0 else 0
                print(f"  {field:12} - Matches: {stats['match']:3d}, Diffs: {stats['diff']:3d}, "
                      f"Missing in Go: {stats['missing_go']:3d}, Missing in Py: {stats['missing_py']:3d} "
                      f"(Match rate: {match_pct:.1f}%)")
    
    # Check specific differences
    print("\n=== Key Differences to Address ===")
    
    # Check hybridbenefit
    hybrid_examples = []
    for item in common_with_pricing[:100]:
        if 'windows' in item['py'] and 'hybridbenefit' in item['py']['windows']:
            if 'windows' not in item['go'] or 'hybridbenefit' not in item['go']['windows']:
                hybrid_examples.append(item['type'])
    
    if hybrid_examples:
        print(f"\n1. Windows HybridBenefit missing in Go ({len(hybrid_examples)} instances):")
        print(f"   Examples: {hybrid_examples[:5]}")
    
    # Check Windows reserved
    win_reserved_examples = []
    for item in common_with_pricing[:100]:
        if 'windows' in item['py'] and 'reserved' in item['py']['windows']:
            if 'windows' not in item['go'] or 'reserved' not in item['go']['windows']:
                win_reserved_examples.append(item['type'])
    
    if win_reserved_examples:
        print(f"\n2. Windows Reserved pricing missing in Go ({len(win_reserved_examples)} instances):")
        print(f"   Examples: {win_reserved_examples[:5]}")
    
    # Basic pricing
    basic_examples = []
    for item in common_with_pricing:
        if 'linux' in item['py'] and 'basic' in item['py']['linux']:
            if 'linux' not in item['go'] or 'basic' not in item['go']['linux']:
                basic_examples.append(item['type'])
    
    if basic_examples:
        print(f"\n3. Basic pricing missing in Go ({len(basic_examples)} instances):")
        print(f"   Examples: {basic_examples}")

if __name__ == '__main__':
    final_comparison()