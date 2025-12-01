#!/usr/bin/env python3
import requests
import json
import urllib.parse

def search_savings_plans():
    """Search for Savings Plans in Azure Pricing API"""
    
    base_url = "https://prices.azure.com/api/retail/prices"
    
    # Search for anything with "savings" in various fields
    queries = [
        # Search in skuName
        "contains(skuName, 'Savings') and currencyCode eq 'USD'",
        
        # Search in productName
        "contains(productName, 'Savings') and currencyCode eq 'USD'",
        
        # Search in meterName
        "contains(meterName, 'Savings') and currencyCode eq 'USD'",
        
        # Look for Compute Savings Plan specifically
        "contains(productName, 'Compute Savings') and currencyCode eq 'USD'",
    ]
    
    for query in queries:
        print(f"\n=== Query: {query} ===")
        url = f"{base_url}?$filter={urllib.parse.quote(query)}"
        
        try:
            response = requests.get(url)
            if response.status_code == 200:
                data = response.json()
                print(f"Found {len(data.get('Items', []))} items")
                
                if data.get('Items'):
                    # Group by unique combinations
                    unique_combos = set()
                    for item in data['Items']:
                        combo = (
                            item.get('serviceName', ''),
                            item.get('productName', ''),
                            item.get('skuName', ''),
                            item.get('type', '')
                        )
                        unique_combos.add(combo)
                    
                    print("\nUnique combinations found:")
                    for service, product, sku, price_type in sorted(unique_combos)[:10]:
                        print(f"  Service: {service}")
                        print(f"  Product: {product}")
                        print(f"  SKU: {sku}")
                        print(f"  Type: {price_type}")
                        print("  ---")
                        
                    # Check if any relate to VMs
                    vm_related = [item for item in data['Items'] if 'virtual machine' in item.get('productName', '').lower() or 'compute' in item.get('productName', '').lower()]
                    if vm_related:
                        print(f"\nFound {len(vm_related)} VM/Compute related items")
                        print("Sample:")
                        item = vm_related[0]
                        print(f"  Service: {item.get('serviceName')}")
                        print(f"  Product: {item.get('productName')}")
                        print(f"  SKU: {item.get('skuName')}")
                        print(f"  Meter: {item.get('meterName')}")
                        print(f"  Type: {item.get('type')}")
                        print(f"  Term: {item.get('reservationTerm', 'N/A')}")
            else:
                print(f"Error: {response.status_code}")
        except Exception as e:
            print(f"Exception: {e}")
            
    # Also check what the nextPageLink looks like
    print("\n=== Checking a regular VM query to understand structure ===")
    vm_query = "serviceName eq 'Virtual Machines' and armSkuName eq 'Standard_D4s_v3' and armRegionName eq 'eastus' and currencyCode eq 'USD'"
    url = f"{base_url}?$filter={urllib.parse.quote(vm_query)}"
    
    response = requests.get(url)
    if response.status_code == 200:
        data = response.json()
        print(f"Response keys: {list(data.keys())}")
        if 'Items' in data and data['Items']:
            item = data['Items'][0]
            print(f"\nSample item keys: {list(item.keys())}")

if __name__ == '__main__':
    search_savings_plans()