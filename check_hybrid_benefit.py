#!/usr/bin/env python3
import requests
import json
import urllib.parse

def check_hybrid_benefit():
    """Check if Azure API returns Hybrid Benefit pricing"""
    
    base_url = "https://prices.azure.com/api/retail/prices"
    
    # Query for A1 v2 which should have hybrid benefit
    query = "serviceName eq 'Virtual Machines' and armSkuName eq 'Standard_A1_v2' and armRegionName eq 'eastus' and currencyCode eq 'USD'"
    
    url = f"{base_url}?$filter={urllib.parse.quote(query)}"
    
    print(f"Checking A1 v2 pricing for Hybrid Benefit...")
    
    response = requests.get(url)
    if response.status_code == 200:
        data = response.json()
        
        print(f"\nFound {len(data['Items'])} items for A1 v2")
        
        # Look for Windows items
        windows_items = [item for item in data['Items'] if 'windows' in item.get('productName', '').lower()]
        
        print(f"\nWindows items: {len(windows_items)}")
        for item in windows_items:
            print(f"\n  Product: {item['productName']}")
            print(f"  SKU: {item['skuName']}")
            print(f"  Meter: {item['meterName']}")
            print(f"  Type: {item['type']}")
            print(f"  Price: {item['unitPrice']}")
            
            # Check all fields for hybrid benefit
            for key, value in item.items():
                if 'hybrid' in str(value).lower():
                    print(f"  *** Found 'hybrid' in {key}: {value}")
        
        # Check if hybrid benefit might be a separate SKU
        print("\n\nAll unique meter names:")
        unique_meters = set(item['meterName'] for item in data['Items'])
        for meter in sorted(unique_meters):
            print(f"  - {meter}")

if __name__ == '__main__':
    check_hybrid_benefit()