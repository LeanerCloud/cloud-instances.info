#!/usr/bin/env python3
import msgpack
import json

# Load the first 100 instances
with open('next/public/first-100-azure-instances.msgpack', 'rb') as f:
    data = msgpack.unpackb(f.read(), raw=False)

print(f"Loaded {len(data)} items from msgpack")

# First item should be the rainbow table
rainbow_table = data[0]
print(f"\nRainbow table has {len(rainbow_table)} entries")

# Check first few instances
print("\nFirst few instances with pricing:")
found_pricing = 0
for i, inst in enumerate(data[1:11]):  # Check first 10 instances after rainbow table
    if isinstance(inst, dict) and 'pricing' in inst and inst['pricing']:
        print(f"\nInstance {i}: {inst.get('instance_type', 'unknown')}")
        print(f"Pricing regions: {list(inst['pricing'].keys())[:5]}...")
        # Show one region's pricing
        for region, pricing in list(inst['pricing'].items())[:1]:
            print(f"Region {region}: {json.dumps(pricing, indent=2)}")
        found_pricing += 1
        
if found_pricing == 0:
    print("\nNo instances with pricing found in first 10 instances")
    # Check all instances
    for i, inst in enumerate(data[1:]):
        if isinstance(inst, dict) and 'pricing' in inst and inst['pricing']:
            print(f"\nFirst instance with pricing at index {i}: {inst.get('instance_type', 'unknown')}")
            print(f"Pricing regions: {list(inst['pricing'].keys())[:5]}...")
            break

# Also load the original JSON to compare
with open('www/azure/instances.json', 'r') as f:
    original = json.load(f)

# Find same instance
for orig_inst in original[:100]:
    if orig_inst.get('pricing') and any(orig_inst['pricing'].values()):
        print(f"\n\nOriginal JSON - Instance: {orig_inst['instance_type']}")
        print(f"Pricing regions: {list(orig_inst['pricing'].keys())[:5]}...")
        for region, pricing in list(orig_inst['pricing'].items())[:1]:
            print(f"Region {region}: {json.dumps(pricing, indent=2)}")
        break