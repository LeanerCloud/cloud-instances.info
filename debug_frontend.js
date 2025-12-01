// Run this in the browser console on the Azure page
console.log('=== Azure Frontend Debug ===');

// Check if instances are loaded
const rows = document.querySelectorAll('tbody tr');
console.log('Total rows visible:', rows.length);

// Check first few instances
console.log('\nFirst 5 instances:');
for (let i = 0; i < Math.min(5, rows.length); i++) {
    const cells = rows[i].querySelectorAll('td');
    const name = cells[0]?.textContent || 'Unknown';
    const linuxPrice = cells[7]?.textContent || 'No cell';
    console.log(`${i + 1}. ${name} - Linux price: ${linuxPrice}`);
}

// Check column headers
const headers = document.querySelectorAll('thead th');
console.log('\nColumn headers:');
headers.forEach((h, i) => console.log(`${i}: ${h.textContent}`));

// Check if React DevTools are available
if (window.__REACT_DEVTOOLS_GLOBAL_HOOK__) {
    console.log('\nReact DevTools detected - check component props for more details');
}