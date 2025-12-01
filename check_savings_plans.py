#!/usr/bin/env python3
import requests
import json

def check_savings_plan_api():
    """Check if Azure Retail Pricing API returns Savings Plans"""
    
    # Test with a common instance type
    base_url = "https://prices.azure.com/api/retail/prices"
    
    # Query for D4s v3 in eastus - check all price types
    filter_query = "serviceName eq 'Virtual Machines' and armSkuName eq 'Standard_D4s_v3' and armRegionName eq 'eastus' and currencyCode eq 'USD'"
    
    url = f"{base_url}?$filter={filter_query}"
    
    print(f"Fetching pricing from: {url}")
    
    response = requests.get(url)
    if response.status_code == 200:
        data = response.json()
        
        print(f"\nFound {len(data['Items'])} pricing items")
        
        # Group by type
        by_type = {}
        for item in data['Items']:
            price_type = item.get('type', 'Unknown')
            if price_type not in by_type:
                by_type[price_type] = []
            by_type[price_type].append(item)
        
        print("\nPricing types found:")
        for price_type, items in by_type.items():
            print(f"  {price_type}: {len(items)} items")
        
        # Check for Savings Plans
        print("\nChecking for Savings Plans:")
        found_savings = False
        for item in data['Items']:
            if 'savings' in item.get('skuName', '').lower() or \
               'savings' in item.get('productName', '').lower() or \
               'savings' in item.get('meterName', '').lower():
                found_savings = True
                print(f"  Found: {item['meterName']} - {item['productName']} - Type: {item['type']}")
        
        if not found_savings:
            print("  No Savings Plans found in the response")
        
        # Check reservation details
        print("\nReservation items:")
        for item in data['Items']:
            if item['type'] == 'Reservation':
                print(f"  {item['meterName']} - {item['productName']} - Term: {item.get('reservationTerm', 'N/A')}")
                
        # Check for Windows reserved
        print("\nWindows reserved pricing:")
        for item in data['Items']:
            if item['type'] == 'Reservation' and 'windows' in item.get('productName', '').lower():
                print(f"  Found: {item['meterName']} - {item['productName']}")
    else:
        print(f"Error: {response.status_code}")

if __name__ == '__main__':
    check_savings_plan_api()