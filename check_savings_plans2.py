#!/usr/bin/env python3
import requests
import json
import urllib.parse

def check_azure_apis():
    """Check different Azure API queries to find Savings Plans"""
    
    base_url = "https://prices.azure.com/api/retail/prices"
    
    # Test different service names
    test_queries = [
        # Try Savings Plans as a service
        "serviceName eq 'Savings Plan for Compute' and armRegionName eq 'eastus' and currencyCode eq 'USD'",
        
        # Try with productName containing savings
        "productName eq 'Savings Plan' and armRegionName eq 'eastus' and currencyCode eq 'USD'",
        
        # Try Azure Compute Savings Plan
        "serviceName eq 'Azure Compute Savings Plan' and armRegionName eq 'eastus' and currencyCode eq 'USD'",
        
        # Try searching for any savings-related items
        "(productName eq 'Virtual Machines DSv3 Series' or productName eq 'Compute Savings Plan') and armSkuName eq 'Standard_D4s_v3' and armRegionName eq 'eastus' and currencyCode eq 'USD'",
    ]
    
    for i, query in enumerate(test_queries, 1):
        print(f"\n=== Test {i} ===")
        print(f"Query: {query}")
        
        url = f"{base_url}?$filter={urllib.parse.quote(query)}"
        
        try:
            response = requests.get(url)
            if response.status_code == 200:
                data = response.json()
                print(f"Found {len(data['Items'])} items")
                
                if data['Items']:
                    # Show first few items
                    for j, item in enumerate(data['Items'][:3]):
                        print(f"  Item {j+1}:")
                        print(f"    Service: {item.get('serviceName', 'N/A')}")
                        print(f"    Product: {item.get('productName', 'N/A')}")
                        print(f"    SKU: {item.get('skuName', 'N/A')}")
                        print(f"    Type: {item.get('type', 'N/A')}")
                        print(f"    Price: {item.get('unitPrice', 'N/A')}")
            else:
                print(f"Error: {response.status_code}")
                if response.text:
                    print(f"Response: {response.text[:200]}")
        except Exception as e:
            print(f"Exception: {e}")

if __name__ == '__main__':
    check_azure_apis()