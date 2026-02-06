#!/usr/bin/env node

const WebSocket = require('ws');
const http = require('http');

// Configuration
const WS_URL = 'ws://localhost:8003/ws';
const SERVER_HOST = 'localhost';
const SERVER_PORT = 8003;
const SERVER_TOKEN = '2WARMMVuwLmr9oRAjf9vZiLUbytz9j';
const CHANNEL_NAME = 'course.ab366aad-e7d0-41a3-9ff0-cbe0331d72cf.lesson.f85ea88a-5a7d-4ffb-b435-cfec6549bdcf';

console.log('='.repeat(70));
console.log('🔍 Debugging Broadcast Message Reception');
console.log('='.repeat(70));
console.log('\nStep 1: Connecting client to channel:', CHANNEL_NAME);
console.log('');

// Create WebSocket client
const ws = new WebSocket(WS_URL);

ws.on('open', () => {
    console.log('✅ Client connected to WebSocket server');
    console.log('📤 Joining channel:', CHANNEL_NAME);
    
    ws.send(JSON.stringify({
        action: 'join_channel',
        channel: CHANNEL_NAME
    }));
    
    // Wait for join to complete, then send broadcast
    setTimeout(() => {
        console.log('\n' + '='.repeat(70));
        console.log('Step 2: Sending HTTP broadcast to channel');
        console.log('='.repeat(70) + '\n');
        
        sendBroadcast();
    }, 1000);
});

ws.on('message', (data) => {
    const message = JSON.parse(data.toString());
    console.log('\n' + '🎉'.repeat(35));
    console.log('📥 CLIENT RECEIVED MESSAGE:');
    console.log('🎉'.repeat(35));
    console.log(JSON.stringify(message, null, 2));
    console.log('🎉'.repeat(35) + '\n');
    
    // Close after receiving
    setTimeout(() => {
        console.log('✅ Test completed successfully! Client received the broadcast.');
        ws.close();
        process.exit(0);
    }, 500);
});

ws.on('close', () => {
    console.log('\n❌ Client disconnected from server');
});

ws.on('error', (error) => {
    console.error('❌ WebSocket Error:', error.message);
});

function sendBroadcast() {
    const payload = {
        broadcast_type: 'channel',
        channel: CHANNEL_NAME,
        event: 'test_broadcast_event',
        data: {
            message: 'This is a test broadcast message',
            timestamp: new Date().toISOString(),
            test_id: 'test-' + Math.random().toString(36).substring(7),
            content: 'Hello from the broadcast!',
            details: {
                lesson_id: 'f85ea88a-5a7d-4ffb-b435-cfec6549bdcf',
                course_id: 'ab366aad-e7d0-41a3-9ff0-cbe0331d72cf',
                action: 'lesson_update'
            }
        }
    };

    const postData = JSON.stringify(payload);

    const options = {
        hostname: SERVER_HOST,
        port: SERVER_PORT,
        path: '/broadcast',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(postData),
            'Authorization': `Bearer ${SERVER_TOKEN}`
        }
    };

    console.log('📡 Broadcasting message:');
    console.log(JSON.stringify(payload, null, 2));
    console.log('');

    const req = http.request(options, (res) => {
        let data = '';

        res.on('data', (chunk) => {
            data += chunk;
        });

        res.on('end', () => {
            console.log('📬 HTTP Response Status:', res.statusCode);
            console.log('📬 HTTP Response:', data);
            console.log('');
            
            if (res.statusCode === 200) {
                console.log('✅ Broadcast API call successful');
                console.log('⏳ Waiting for client to receive message...\n');
                
                // If no message received after 3 seconds, something is wrong
                setTimeout(() => {
                    console.log('\n⚠️  WARNING: Client did not receive message after 3 seconds!');
                    console.log('This indicates the broadcast is not reaching the client.');
                    ws.close();
                    process.exit(1);
                }, 3000);
            } else {
                console.log('❌ Broadcast API call failed!');
                ws.close();
                process.exit(1);
            }
        });
    });

    req.on('error', (error) => {
        console.error('❌ HTTP Error:', error.message);
        ws.close();
        process.exit(1);
    });

    req.write(postData);
    req.end();
}
