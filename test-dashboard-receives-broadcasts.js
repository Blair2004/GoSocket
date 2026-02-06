#!/usr/bin/env node

const WebSocket = require('ws');

// Configuration
const WS_URL = 'ws://localhost:8003/ws';
const JWT_TOKEN = 'YOUR_JWT_TOKEN_HERE'; // Replace with actual JWT
const CHANNEL_NAME = 'test-channel-' + Date.now();

console.log('='.repeat(60));
console.log('WebSocket Dashboard Broadcast Test');
console.log('='.repeat(60));
console.log('');
console.log('This test will:');
console.log('1. Connect Client A (simulated regular user)');
console.log('2. Client A joins channel:', CHANNEL_NAME);
console.log('3. Client A sends 3 test messages to the channel');
console.log('4. Dashboard should receive all these messages if connected');
console.log('');
console.log('Instructions:');
console.log('- Open the dashboard in your browser');
console.log('- Make sure JWT token is configured in dashboard settings');
console.log('- Join the channel:', CHANNEL_NAME);
console.log('- Watch the message log for incoming messages');
console.log('');
console.log('='.repeat(60));
console.log('');

// Wait for user to press enter
const readline = require('readline');
const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

rl.question('Press ENTER when dashboard is ready and connected to channel...', () => {
    rl.close();
    startTest();
});

function startTest() {
    console.log('\n🚀 Starting test...\n');
    
    const clientA = new WebSocket(WS_URL);
    
    clientA.on('open', () => {
        console.log('✅ Client A connected');
        
        // Authenticate if JWT provided
        if (JWT_TOKEN && JWT_TOKEN !== 'YOUR_JWT_TOKEN_HERE') {
            console.log('🔐 Client A authenticating...');
            clientA.send(JSON.stringify({
                action: 'authenticate',
                token: JWT_TOKEN
            }));
        }
        
        // Wait a bit then join channel
        setTimeout(() => {
            console.log(`📺 Client A joining channel: ${CHANNEL_NAME}`);
            clientA.send(JSON.stringify({
                action: 'join_channel',
                channel: CHANNEL_NAME,
                data: {},
                private: false
            }));
            
            // Wait then send messages
            setTimeout(() => {
                sendTestMessages(clientA);
            }, 1000);
        }, 1000);
    });
    
    clientA.on('message', (data) => {
        try {
            const msg = JSON.parse(data.toString());
            console.log('📨 Client A received:', msg.event || msg.action, JSON.stringify(msg.data || {}).substring(0, 50));
        } catch (e) {
            console.log('📨 Client A received:', data.toString().substring(0, 100));
        }
    });
    
    clientA.on('error', (error) => {
        console.error('❌ Client A error:', error.message);
    });
    
    clientA.on('close', () => {
        console.log('🔌 Client A disconnected');
    });
}

function sendTestMessages(client) {
    const messages = [
        {
            title: 'Test Message 1',
            content: 'Hello from Client A!',
            timestamp: new Date().toISOString()
        },
        {
            title: 'Test Message 2',
            content: 'This is the second test message',
            counter: 2
        },
        {
            title: 'Test Message 3',
            content: 'Final test message - dashboard should see all of these!',
            counter: 3,
            final: true
        }
    ];
    
    let index = 0;
    const interval = setInterval(() => {
        if (index >= messages.length) {
            clearInterval(interval);
            console.log('\n✅ All test messages sent!');
            console.log('📋 Check the dashboard message log - you should see all 3 messages');
            console.log('\n⏰ Closing client in 5 seconds...');
            setTimeout(() => {
                client.close();
                process.exit(0);
            }, 5000);
            return;
        }
        
        const message = {
            action: 'send_message',
            channel: CHANNEL_NAME,
            event: 'test_message',
            data: messages[index]
        };
        
        console.log(`\n💬 Client A sending message ${index + 1}:`, messages[index].title);
        client.send(JSON.stringify(message));
        index++;
    }, 2000); // Send every 2 seconds
}
