const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:8082/ws');

ws.on('open', function open() {
    console.log('✅ Connected to server');
    
    // Join a channel first
    console.log('📤 Joining channel...');
    ws.send(JSON.stringify({
        action: 'join_channel',
        channel: 'test-channel'
    }));
    
    // Send a message after a short delay
    setTimeout(() => {
        console.log('📤 Sending message...');
        ws.send(JSON.stringify({
            action: 'send_message',
            channel: 'test-channel',
            event: 'test-event',
            data: { message: 'Hello from test client!' }
        }));
        
        // Send ping
        setTimeout(() => {
            console.log('📤 Sending ping...');
            ws.send(JSON.stringify({
                action: 'ping'
            }));
            
            // Close after another delay
            setTimeout(() => {
                console.log('🔴 Closing connection');
                ws.close();
            }, 1000);
        }, 20000 );
    }, 1000);
});

ws.on('message', function message(data) {
    const msg = JSON.parse(data);
    console.log('📥 Received:', JSON.stringify(msg, null, 2));
});

ws.on('close', function close() {
    console.log('❌ Disconnected');
    process.exit(0);
});

ws.on('error', function error(err) {
    console.log('❌ Error:', err.message);
    process.exit(1);
});
