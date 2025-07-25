const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:8082/ws');

let connected = false;
let startTime = Date.now();

ws.on('open', function open() {
    console.log('✅ Connected to server');
    connected = true;
    
    // Join a channel first
    console.log('📤 Joining channel...');
    ws.send(JSON.stringify({
        action: 'join_channel',
        channel: 'test-channel'
    }));
});

ws.on('message', function message(data) {
    const msg = JSON.parse(data);
    const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
    console.log(`[${elapsed}s] 📥 Received:`, JSON.stringify(msg, null, 2));
    
    // Log ping frames specifically
    if (msg.event === 'pong') {
        console.log(`[${elapsed}s] 🏓 PONG received from server`);
    }
});

ws.on('ping', function ping(data) {
    const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
    console.log(`[${elapsed}s] 🏓 PING received from server (WebSocket ping frame)`);
    // WebSocket automatically sends pong response
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

// Keep the connection alive for observation
console.log('🔍 Monitoring connection for automatic server pings (should happen every 30s)...');
console.log('Press Ctrl+C to exit');

// Send a ping every 10 seconds to keep connection alive and test responsiveness
setInterval(() => {
    if (connected) {
        const elapsed = ((Date.now() - startTime) / 1000).toFixed(1);
        console.log(`[${elapsed}s] 📤 Sending ping to server...`);
        ws.send(JSON.stringify({ action: 'ping' }));
    }
}, 10000);

// Graceful shutdown
process.on('SIGINT', () => {
    console.log('\\n🔴 Shutting down...');
    if (connected) {
        ws.close(1000, 'Manual shutdown');
    }
    process.exit(0);
});
