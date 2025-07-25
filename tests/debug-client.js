const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:8083/ws');
let pingCount = 0;
let pongCount = 0;

ws.on('open', () => {
    console.log('✅ Connected to server');
    
    // Send a ping immediately
    console.log('📤 Sending ping #1...');
    ws.send(JSON.stringify({ action: 'ping' }));
    pingCount++;
    
    // Send another ping after 2 seconds
    setTimeout(() => {
        console.log('📤 Sending ping #2...');
        ws.send(JSON.stringify({ action: 'ping' }));
        pingCount++;
    }, 2000);
    
    // Send a ping with data to test the logic
    setTimeout(() => {
        console.log('📤 Sending ping #3 with data...');
        ws.send(JSON.stringify({ action: 'ping', data: { test: 'data' } }));
        pingCount++;
    }, 4000);
    
    // Send a ping with channel to test the logic
    setTimeout(() => {
        console.log('📤 Sending ping #4 with channel...');
        ws.send(JSON.stringify({ action: 'ping', channel: 'test-channel' }));
        pingCount++;
    }, 6000);
});

ws.on('message', (data) => {
    try {
        const msg = JSON.parse(data);
        console.log('📥 Received:', msg);
        
        if (msg.event === 'pong') {
            pongCount++;
            console.log(`🏓 Pong received! Count: ${pongCount}/${pingCount}`);
        }
    } catch (error) {
        console.log('❌ Error parsing message:', error.message);
    }
});

ws.on('close', (code, reason) => {
    console.log('❌ Connection closed:', { code, reason: reason.toString() });
    console.log(`📊 Final stats: Sent ${pingCount} pings, received ${pongCount} pongs`);
});

ws.on('error', (error) => {
    console.log('❌ WebSocket error:', error.message);
});

// Close after 10 seconds
setTimeout(() => {
    console.log('🔴 Closing connection...');
    ws.close();
}, 10000);
