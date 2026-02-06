#!/usr/bin/env node

const http = require('http');

// Configuration
const SERVER_HOST = 'localhost';
const SERVER_PORT = 8003;
const SERVER_TOKEN = '2WARMMVuwLmr9oRAjf9vZiLUbytz9j';
const CHANNEL_NAME = 'course.ab366aad-e7d0-41a3-9ff0-cbe0331d72cf.lesson.f85ea88a-5a7d-4ffb-b435-cfec6549bdcf';

// Create test message payload
const payload = {
    broadcast_type: 'channel',
    channel: CHANNEL_NAME,
    event: 'test_broadcast',
    data: {
        message: 'This is a test broadcast from Node.js',
        timestamp: new Date().toISOString(),
        test_id: Math.random().toString(36).substring(7)
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

console.log('Sending broadcast to channel:', CHANNEL_NAME);
console.log('Payload:', JSON.stringify(payload, null, 2));

const req = http.request(options, (res) => {
    let data = '';

    res.on('data', (chunk) => {
        data += chunk;
    });

    res.on('end', () => {
        console.log('\nResponse Status:', res.statusCode);
        console.log('Response:', data);
        
        if (res.statusCode === 200) {
            console.log('\n✅ Broadcast sent successfully!');
            console.log('The dashboard should receive this message if it is joined to the channel.');
        } else {
            console.log('\n❌ Broadcast failed!');
        }
    });
});

req.on('error', (error) => {
    console.error('Error:', error.message);
});

req.write(postData);
req.end();
