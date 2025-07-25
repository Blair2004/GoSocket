const WebSocket = require('ws');

// Test with shorter ping intervals
function testBroadcastTiming() {
    const client1 = new WebSocket('ws://localhost:8082/ws');
    const client2 = new WebSocket('ws://localhost:8082/ws');
    
    let client1Ready = false;
    let client2Ready = false;
    
    // Client 1 setup
    client1.on('open', () => {
        console.log('Client 1: Connected');
        client1.send(JSON.stringify({ action: 'join_channel', channel: 'test' }));
        client1Ready = true;
        checkReady();
    });
    
    client1.on('message', (data) => {
        const msg = JSON.parse(data);
        if (msg.event === 'message') {
            console.log('Client 1: Received broadcast at', new Date().toISOString());
        }
    });
    
    // Client 2 setup
    client2.on('open', () => {
        console.log('Client 2: Connected');
        client2.send(JSON.stringify({ action: 'join_channel', channel: 'test' }));
        client2Ready = true;
        checkReady();
    });
    
    client2.on('message', (data) => {
        const msg = JSON.parse(data);
        if (msg.event === 'message') {
            console.log('Client 2: Received broadcast at', new Date().toISOString());
        } else if (msg.event === 'pong') {
            console.log('Client 2: Received pong at', new Date().toISOString());
        }
    });
    
    function checkReady() {
        if (client1Ready && client2Ready) {
            setTimeout(() => {
                console.log('Sending test message at', new Date().toISOString());
                client1.send(JSON.stringify({
                    action: 'send_message',
                    channel: 'test',
                    event: 'message',
                    data: { message: 'Test broadcast timing' }
                }));
                
                // Send a ping to client 2 to trigger processing
                setTimeout(() => {
                    client2.send(JSON.stringify({ action: 'ping' }));
                }, 1000);
                
            }, 2000);
        }
    }
    
    // Close after 10 seconds
    setTimeout(() => {
        client1.close();
        client2.close();
        console.log('Test completed');
    }, 10000);
}

testBroadcastTiming();
