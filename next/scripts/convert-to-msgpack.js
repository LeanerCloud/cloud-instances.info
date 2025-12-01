const fs = require('fs');
const msgpack = require('@msgpack/msgpack');
const { execSync } = require('child_process');

// Convert JSON to msgpack.xz
function convertToMsgpack(provider) {
  const jsonFile = `data/${provider}-instances.json`;
  const msgpackFile = `data/${provider}-instances.msgpack`;
  const xzFile = `data/${provider}-instances.msgpack.xz`;
  
  // Read JSON
  const jsonData = JSON.parse(fs.readFileSync(jsonFile, 'utf8'));
  
  // Convert to msgpack
  const packed = msgpack.encode(jsonData);
  fs.writeFileSync(msgpackFile, Buffer.from(packed));
  
  // Compress with xz
  execSync(`xz -f ${msgpackFile}`);
  
  console.log(`✅ Created ${xzFile}`);
}

// Convert GCP and OCI
convertToMsgpack('gcp');
convertToMsgpack('oci');