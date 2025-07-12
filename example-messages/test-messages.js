#!/usr/bin/env node

/**
 * Socket Server Message Testing Script
 * 
 * This script demonstrates how to send various messages to the socket server
 * using the example messages in this directory.
 */

const WebSocket = require('ws');
const fs = require('fs');
const path = require('path');

class SocketTester {
    constructor(url = 'ws://localhost:8080/ws') {
        this.url = url;
        this.ws = null;
        this.connected = false;
        this.authenticated = false;
        this.joinedChannels = new Set();
    }

    async connect() {
        return new Promise((resolve, reject) => {
            this.ws = new WebSocket(this.url);
            
            this.ws.on('open', () => {
                console.log('✅ Connected to socket server');
                this.connected = true;
                resolve();
            });
            
            this.ws.on('message', (data) => {
                const message = JSON.parse(data);
                this.handleMessage(message);
            });
            
            this.ws.on('error', (error) => {
                console.error('❌ WebSocket error:', error);
                reject(error);
            });
            
            this.ws.on('close', () => {
                console.log('🔌 Disconnected from socket server');
                this.connected = false;
            });
        });
    }

    handleMessage(message) {
        console.log('📨 Received:', JSON.stringify(message, null, 2));
        
        switch (message.event) {
            case 'connected':
                console.log('🎉 Welcome message received');
                break;
            case 'joined_channel':
                this.joinedChannels.add(message.data.channel);
                console.log(`🏠 Joined channel: ${message.data.channel}`);
                break;
            case 'left_channel':
                this.joinedChannels.delete(message.data.channel);
                console.log(`🚪 Left channel: ${message.data.channel}`);
                break;
            case 'error':
                console.error('⚠️ Server error:', message.data.error);
                break;
            case 'pong':
                console.log('🏓 Pong received');
                break;
            case 'message':
                console.log(`💬 Message in ${message.channel}: ${JSON.stringify(message.data)}`);
                break;
        }
    }

    loadExampleMessage(filename) {
        const filePath = path.join(__dirname, filename);
        const content = fs.readFileSync(filePath, 'utf8');
        return JSON.parse(content);
    }

    sendMessage(message) {
        if (!this.connected) {
            console.error('❌ Not connected to server');
            return;
        }
        
        console.log('📤 Sending:', JSON.stringify(message, null, 2));
        this.ws.send(JSON.stringify(message));
    }

    async wait(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    async runDemo() {
        try {
            console.log('🚀 Starting Socket Server Demo...\n');
            
            // Connect to server
            await this.connect();
            await this.wait(1000);
            
            // Test authentication
            console.log('\n1. Testing Authentication...');
            const authMessage = this.loadExampleMessage('authenticate.json');
            this.sendMessage(authMessage);
            await this.wait(2000);
            
            // Test joining channels
            console.log('\n2. Testing Channel Joining...');
            const joinGeneral = this.loadExampleMessage('join_channel.json');
            this.sendMessage(joinGeneral);
            await this.wait(1000);
            
            const joinOrders = this.loadExampleMessage('join_channel_orders.json');
            this.sendMessage(joinOrders);
            await this.wait(1000);
            
            // Test sending messages
            console.log('\n3. Testing Message Sending...');
            const textMessage = this.loadExampleMessage('send_message.json');
            this.sendMessage(textMessage);
            await this.wait(1000);
            
            const orderUpdate = this.loadExampleMessage('send_message_order_update.json');
            this.sendMessage(orderUpdate);
            await this.wait(1000);
            
            const notification = this.loadExampleMessage('send_message_notification.json');
            this.sendMessage(notification);
            await this.wait(1000);
            
            // Test ping/pong
            console.log('\n4. Testing Ping/Pong...');
            const ping = this.loadExampleMessage('ping.json');
            this.sendMessage(ping);
            await this.wait(1000);
            
            // Test leaving channel
            console.log('\n5. Testing Channel Leaving...');
            const leaveChannel = this.loadExampleMessage('leave_channel.json');
            this.sendMessage(leaveChannel);
            await this.wait(1000);
            
            console.log('\n✅ Demo completed successfully!');
            console.log(`📊 Joined channels: ${Array.from(this.joinedChannels).join(', ')}`);
            
        } catch (error) {
            console.error('❌ Demo failed:', error);
        } finally {
            if (this.ws) {
                this.ws.close();
            }
        }
    }
}

// Run the demo if this script is executed directly
if (require.main === module) {
    const tester = new SocketTester();
    tester.runDemo();
}

module.exports = SocketTester;
