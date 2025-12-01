const fs = require('fs');
const msgpack = require('@msgpack/msgpack');

// Load and check the first msgpack file
const data = msgpack.decode(fs.readFileSync('next/public/first-100-azure-instances.msgpack'));

console.log('Total items in msgpack:', data.length);
console.log('First item is rainbow table?', Array.isArray(data[0]) && typeof data[0][0] === 'string');

// Check a few instances
for (let i = 1; i < Math.min(6, data.length); i++) {
    const inst = data[i];
    console.log(`\nInstance ${i}:`);
    console.log('  Type:', inst.instance_type);
    console.log('  Pricing type:', typeof inst.pricing);
    console.log('  Pricing is array?', Array.isArray(inst.pricing));
    
    if (inst.pricing) {
        if (Array.isArray(inst.pricing)) {
            console.log('  Pricing array length:', inst.pricing.length);
            if (inst.pricing.length > 0) {
                console.log('  First pricing item:', JSON.stringify(inst.pricing[0]).substring(0, 100) + '...');
            }
        } else {
            console.log('  Pricing keys:', Object.keys(inst.pricing).slice(0, 5));
        }
    }
}