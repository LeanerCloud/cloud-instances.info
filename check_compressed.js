const fs = require('fs');
const msgpack = require('msgpack-lite');

// Read the first 100 instances msgpack file
const data = msgpack.decode(fs.readFileSync('next/public/first-100-azure-instances.msgpack'));

console.log(`Loaded ${data.length} items from msgpack`);

// First item should be the rainbow table
const rainbowTable = data[0];
console.log(`\nRainbow table has ${rainbowTable.length} entries`);

// Check instances for pricing
let foundPricing = 0;
for (let i = 1; i < Math.min(11, data.length); i++) {
    const inst = data[i];
    if (inst.pricing && Object.keys(inst.pricing).length > 0) {
        console.log(`\nInstance ${i}: ${inst.instance_type || 'unknown'}`);
        console.log(`Pricing regions: ${Object.keys(inst.pricing).slice(0, 5).join(', ')}...`);
        
        // Show first region's pricing
        const firstRegion = Object.keys(inst.pricing)[0];
        console.log(`Region ${firstRegion}:`, JSON.stringify(inst.pricing[firstRegion], null, 2));
        foundPricing++;
    }
}

if (foundPricing === 0) {
    console.log('\nNo pricing found in first 10 instances');
    
    // Search all instances
    for (let i = 1; i < data.length; i++) {
        const inst = data[i];
        if (inst.pricing && Object.keys(inst.pricing).length > 0) {
            console.log(`\nFirst instance with pricing at index ${i}: ${inst.instance_type}`);
            console.log(`Pricing regions: ${Object.keys(inst.pricing).slice(0, 5).join(', ')}`);
            break;
        }
    }
}

// Also check the original JSON
const original = JSON.parse(fs.readFileSync('www/azure/instances.json', 'utf8'));
for (const inst of original.slice(0, 100)) {
    if (inst.pricing && Object.keys(inst.pricing).length > 0) {
        console.log(`\n\nOriginal JSON - Instance: ${inst.instance_type}`);
        console.log(`Pricing regions: ${Object.keys(inst.pricing).slice(0, 5).join(', ')}`);
        const firstRegion = Object.keys(inst.pricing)[0];
        console.log(`Region ${firstRegion}:`, JSON.stringify(inst.pricing[firstRegion], null, 2));
        break;
    }
}