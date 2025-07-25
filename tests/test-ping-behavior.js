const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:8082/ws');

let connected = false;
let startTime = Date.now();

ws.on('open', function open() {
    console.log('✅ Connected to server');
    connected = true;
    
    // Send a ping immediately to test the new behavior
    console.log('📤 Sending ping to test new behavior...');
    ws.send(JSON.stringify({ action: 'ping' }));
});

ws.on('message', function message(data) {
    const msg = JSON.parse(data);
    const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
    console.log(`[${elapsed}s] 📥 Received:`, JSON.stringify(msg, null, 2));
    
    if (msg.event === 'pong') {
        console.log(`[${elapsed}s] 🏓 PONG received from server - ping handled internally!`);
    }
});

ws.on('close', function close(code, reason) {
    const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
    console.log(`[${elapsed}s] ❌ Connection closed: code=${code}, reason=${reason}`);
    connected = false;
});

ws.on('error', function error(err) {
    const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
    console.log(`[${elapsed}s] ❌ Error:`, err.message);
});

// Send a ping every 5 seconds to test
setInterval(() => {
    if (connected) {
        const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
        console.log(`[${elapsed}s] 📤 Sending ping...`);
        ws.send(JSON.stringify({ action: 'ping' }));
    }
}, 5000);

// Close after 30 seconds
setTimeout(() => {
    console.log('🔴 Test complete, closing connection...');
    ws.close();
    process.exit(0);
}, 30000);
