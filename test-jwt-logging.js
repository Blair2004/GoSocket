const WebSocket = require('ws');
const jwt = require('jsonwebtoken');

console.log('Testing JWT authentication logging...\n');

// Create a test JWT with a WRONG secret (to trigger "signature invalid" error)
const testToken = jwt.sign(
    { user_id: '123', channel: 'test-channel' },
    'WRONG_SECRET_KEY', // Using wrong secret intentionally
    { algorithm: 'HS256' }
);

console.log('Test JWT Token (signed with WRONG secret):');
console.log(testToken);
console.log('\n');

// Connect to WebSocket server
const ws = new WebSocket('ws://localhost:8003/ws');

ws.on('open', function open() {
    console.log('Connected to WebSocket server');
    
    // Send authentication message with the invalid token
    const authMessage = {
        event: 'authenticate',
        token: testToken
    };
    
    console.log('Sending authentication message...');
    ws.send(JSON.stringify(authMessage));
});

ws.on('message', function incoming(data) {
    console.log('Received from server:', data.toString());
});

ws.on('error', function error(err) {
    console.error('WebSocket error:', err.message);
});

ws.on('close', function close() {
    console.log('Connection closed');
    process.exit(0);
});

// Close after 3 seconds
setTimeout(() => {
    console.log('\nClosing connection...');
    ws.close();
}, 3000);
