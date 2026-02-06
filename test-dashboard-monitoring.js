const WebSocket = require('ws');

console.log('🚀 Testing dashboard message monitoring...\n');

const ws = new WebSocket('ws://localhost:8003/ws');

ws.on('open', () => {
    console.log('✅ Connected to server\n');
    
    // Join test channel
    console.log('📤 Joining channel "test-channel"...');
    ws.send(JSON.stringify({
        action: 'join_channel',
        channel: 'test-channel'
    }));
    
    // Wait a bit, then send some test messages
    setTimeout(() => {
        console.log('📤 Sending test messages...\n');
        
        // Message 1
        ws.send(JSON.stringify({
            action: 'send_message',
            channel: 'test-channel',
            event: 'user_action',
            data: { message: 'User clicked button', user: 'John', action: 'click' }
        }));
        
        setTimeout(() => {
            // Message 2
            ws.send(JSON.stringify({
                action: 'send_message',
                channel: 'test-channel',
                event: 'notification',
                data: { title: 'New Order', body: 'Order #12345 received', priority: 'high' }
            }));
        }, 1000);
        
        setTimeout(() => {
            // Message 3
            ws.send(JSON.stringify({
                action: 'send_message',
                channel: 'test-channel',
                event: 'status_update',
                data: { status: 'processing', progress: 50, task: 'data-import' }
            }));
        }, 2000);
        
        setTimeout(() => {
            console.log('\n✅ Test messages sent!');
            console.log('💡 Check your dashboard to see these messages appear in the Message Monitor\n');
            ws.close();
        }, 3000);
    }, 500);
});

ws.on('message', (data) => {
    const msg = JSON.parse(data);
    console.log('📥 Received:', JSON.stringify(msg, null, 2));
});

ws.on('close', () => {
    console.log('\n👋 Disconnected');
    process.exit(0);
});

ws.on('error', (error) => {
    console.error('❌ Error:', error.message);
});
