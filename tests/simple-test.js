const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:8082/ws');

ws.on('open', () => {
    console.log('✅ Connected');
});

ws.on('message', (data) => {
    const msg = JSON.parse(data);
    console.log('📥 Received:', msg);
    
    if (msg.event === 'connected') {
        console.log('🎉 Welcome received, sending ping...');
        ws.send(JSON.stringify({ action: 'ping' }));
    }
});

ws.on('close', (code, reason) => {
    console.log('❌ Connection closed:', code, reason.toString());
    process.exit(0);
});

ws.on('error', (error) => {
    console.log('❌ Error:', error.message);
});

setTimeout(() => {
    console.log('🔴 Test complete');
    ws.close();
}, 10000);
