const fs = require('fs');

// Load the original Azure data
const data = JSON.parse(fs.readFileSync('www/azure/instances.json', 'utf8'));

// Count instances with pricing
const withPricing = data.filter(inst => inst.pricing && Object.keys(inst.pricing).length > 0);
console.log(`Total instances: ${data.length}`);
console.log(`Instances with pricing: ${withPricing.length}`);

// Check specific regions
const regionsWithPricing = new Set();
for (const inst of withPricing) {
    Object.keys(inst.pricing).forEach(region => regionsWithPricing.add(region));
}

console.log(`\nUnique regions with pricing: ${regionsWithPricing.size}`);
console.log('Regions:', Array.from(regionsWithPricing).sort().slice(0, 10).join(', '), '...');

// Check if 'eastus' has pricing
const eastusInstances = withPricing.filter(inst => inst.pricing.eastus);
console.log(`\nInstances with 'eastus' pricing: ${eastusInstances.length}`);

// Show a sample instance with eastus pricing
if (eastusInstances.length > 0) {
    const sample = eastusInstances[0];
    console.log(`\nSample instance with eastus pricing:`);
    console.log(`Instance: ${sample.instance_type}`);
    console.log(`eastus pricing:`, JSON.stringify(sample.pricing.eastus, null, 2));
}

// Check the first 100 instances specifically
const first100 = data.slice(0, 100);
const first100WithPricing = first100.filter(inst => inst.pricing && inst.pricing.eastus);
console.log(`\nFirst 100 instances with eastus pricing: ${first100WithPricing.length}`);

// List instance types in first 100
console.log('\nFirst 10 instance types:');
for (let i = 0; i < 10; i++) {
    const inst = data[i];
    const hasPricing = inst.pricing && Object.keys(inst.pricing).length > 0;
    console.log(`  ${inst.instance_type}: ${hasPricing ? 'has pricing' : 'NO PRICING'}`);
}