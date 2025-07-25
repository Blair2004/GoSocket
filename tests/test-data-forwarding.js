#!/usr/bin/env node

const WebSocket = require('ws');

// Test script to verify optional data forwarding to Laravel
class DataForwardingTest {
    constructor() {
        this.ws = null;
        this.testResults = [];
    }

    async runTests() {
        console.log('🧪 Testing optional data forwarding for join/leave operations...\n');
        
        try {
            await this.connect();
            await this.testJoinWithData();
            await this.testJoinWithoutData();
            await this.testLeaveWithData();
            await this.testLeaveWithoutData();
            
            this.printResults();
        } catch (error) {
            console.error('❌ Test failed:', error.message);
        } finally {
            if (this.ws) {
                this.ws.close();
            }
        }
    }

    async connect() {
        return new Promise((resolve, reject) => {
            this.ws = new WebSocket('ws://localhost:8080/ws');
            
            this.ws.on('open', () => {
                console.log('✅ Connected to WebSocket server');
                resolve();
            });
            
            this.ws.on('message', (data) => {
                const message = JSON.parse(data);
                console.log('📨 Received:', JSON.stringify(message, null, 2));
            });
            
            this.ws.on('error', (error) => {
                reject(error);
            });
            
            setTimeout(() => {
                if (this.ws.readyState !== WebSocket.OPEN) {
                    reject(new Error('Connection timeout'));
                }
            }, 5000);
        });
    }

    async testJoinWithData() {
        console.log('\n🔍 Test 1: Join channel WITH data');
        const customData = {
            userType: 'premium',
            preferences: {
                notifications: true,
                theme: 'dark'
            },
            metadata: 'test-join-data'
        };
        
        this.sendMessage({
            action: 'join_channel',
            channel: 'test-channel-with-data',
            data: customData
        });
        
        await this.wait(1000);
        this.testResults.push('✅ Join with data sent successfully');
    }

    async testJoinWithoutData() {
        console.log('\n🔍 Test 2: Join channel WITHOUT data');
        
        this.sendMessage({
            action: 'join_channel',
            channel: 'test-channel-no-data'
            // No data field - should be handled as null/undefined
        });
        
        await this.wait(1000);
        this.testResults.push('✅ Join without data sent successfully');
    }

    async testLeaveWithData() {
        console.log('\n🔍 Test 3: Leave channel WITH data');
        const customData = {
            reason: 'user_requested',
            sessionDuration: 1200,
            activityCount: 5
        };
        
        this.sendMessage({
            action: 'leave_channel',
            channel: 'test-channel-with-data',
            data: customData
        });
        
        await this.wait(1000);
        this.testResults.push('✅ Leave with data sent successfully');
    }

    async testLeaveWithoutData() {
        console.log('\n🔍 Test 4: Leave channel WITHOUT data');
        
        this.sendMessage({
            action: 'leave_channel',
            channel: 'test-channel-no-data'
            // No data field - should use default system data
        });
        
        await this.wait(1000);
        this.testResults.push('✅ Leave without data sent successfully');
    }

    sendMessage(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            console.log('📤 Sending:', JSON.stringify(message, null, 2));
            this.ws.send(JSON.stringify(message));
        } else {
            console.error('❌ WebSocket not connected');
        }
    }

    wait(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    printResults() {
        console.log('\n📊 Test Results:');
        console.log('================');
        this.testResults.forEach(result => console.log(result));
        console.log('\n💡 Check your Laravel logs to verify:');
        console.log('   - Join with data: Custom data was forwarded');
        console.log('   - Join without data: Data field is null/empty');
        console.log('   - Leave with data: Custom data was forwarded');
        console.log('   - Leave without data: Default system data was used');
    }
}

// Run the test
const test = new DataForwardingTest();
test.runTests().catch(console.error);
