const WebSocket = require('ws');

class BroadcastTimingTest {
    constructor() {
        this.startTime = Date.now();
        this.client1 = null;
        this.client2 = null;
        this.messagesSent = 0;
        this.messagesReceived = 0;
        this.lastPongTimes = { client1: 0, client2: 0 };
        this.messageReceiveTimes = [];
    }

    log(message, ...args) {
        const elapsed = ((Date.now() - this.startTime) / 1000).toFixed(3);
        const timestamp = new Date().toISOString().substr(11, 12);
        console.log(`[${timestamp}] [${elapsed}s] ${message}`, ...args);
    }

    async runTest() {
        this.log('🚀 Starting broadcast timing test...');
        
        try {
            await this.createClients();
            await this.waitForConnections();
            await this.authenticateClients();
            await this.joinChannel();
            await this.runMessageTests();
            this.showResults();
        } catch (error) {
            this.log('❌ Test error:', error.message);
        } finally {
            this.cleanup();
        }
    }

    async createClients() {
        this.log('Creating Client 1 (sender)...');
        this.client1 = new WebSocket('ws://localhost:8082/ws');
        this.setupClient1Handlers();

        this.log('Creating Client 2 (receiver)...');
        this.client2 = new WebSocket('ws://localhost:8082/ws');
        this.setupClient2Handlers();
    }

    setupClient1Handlers() {
        this.client1.on('open', () => {
            this.log('Client 1: ✅ Connected');
        });

        this.client1.on('message', (data) => {
            try {
                const msg = JSON.parse(data);
                this.log('Client 1: 📥 Received:', msg.event, msg.data ? JSON.stringify(msg.data) : '');
                
                if (msg.event === 'pong') {
                    this.lastPongTimes.client1 = Date.now();
                    this.log('Client 1: 🏓 Pong received');
                }
            } catch (error) {
                this.log('Client 1: ❌ Parse error:', error.message);
            }
        });

        this.client1.on('close', (code, reason) => {
            this.log('Client 1: ❌ Disconnected:', code, reason.toString());
        });

        this.client1.on('error', (error) => {
            this.log('Client 1: ❌ Error:', error.message);
        });
    }

    setupClient2Handlers() {
        this.client2.on('open', () => {
            this.log('Client 2: ✅ Connected');
        });

        this.client2.on('message', (data) => {
            try {
                const msg = JSON.parse(data);
                const receiveTime = Date.now();
                
                if (msg.event === 'pong') {
                    this.lastPongTimes.client2 = receiveTime;
                    this.log('Client 2: 🏓 Pong received');
                } else if (msg.event === 'message' || msg.channel === 'test-broadcast') {
                    this.messagesReceived++;
                    const timeSinceLastPong = receiveTime - this.lastPongTimes.client2;
                    
                    this.log(`Client 2: 📥 Received message (${timeSinceLastPong}ms after last pong):`, 
                        msg.data ? JSON.stringify(msg.data) : msg.event);
                    
                    this.messageReceiveTimes.push({
                        message: msg.data || msg.event,
                        receiveTime: receiveTime,
                        timeSinceLastPong: timeSinceLastPong
                    });
                } else {
                    this.log('Client 2: 📥 Received:', msg.event, msg.data ? JSON.stringify(msg.data) : '');
                }
            } catch (error) {
                this.log('Client 2: ❌ Parse error:', error.message);
            }
        });

        this.client2.on('close', (code, reason) => {
            this.log('Client 2: ❌ Disconnected:', code, reason.toString());
        });

        this.client2.on('error', (error) => {
            this.log('Client 2: ❌ Error:', error.message);
        });
    }

    async waitForConnections() {
        this.log('Waiting for connections...');
        
        await this.waitForCondition(() => 
            this.client1.readyState === WebSocket.OPEN && 
            this.client2.readyState === WebSocket.OPEN, 
            10000, 
            'Connection timeout'
        );
        
        this.log('Both clients connected');
    }

    async authenticateClients() {
        this.log('Authenticating clients...');
        
        // Send authentication for both clients
        this.client1.send(JSON.stringify({
            action: 'authenticate',
            token: 'your-signin-keycccyour-signin-key'
        }));
        
        this.client2.send(JSON.stringify({
            action: 'authenticate',
            token: 'your-signin-keycccyour-signin-key'
        }));
        
        await this.sleep(2000);
        this.log('Authentication sent');
    }

    async joinChannel() {
        this.log('Joining test channel...');
        
        this.client1.send(JSON.stringify({
            action: 'join_channel',
            channel: 'test-broadcast'
        }));
        
        this.client2.send(JSON.stringify({
            action: 'join_channel',
            channel: 'test-broadcast'
        }));
        
        await this.sleep(2000);
        this.log('Both clients joined channel');
    }

    async runMessageTests() {
        this.log('🧪 Starting message timing tests...');
        
        // Test 1: Send ping to client 2, then immediately send a message
        this.log('Test 1: Triggering pong for client 2, then sending message...');
        
        this.client2.send(JSON.stringify({ action: 'ping' }));
        await this.sleep(100); // Wait for pong response
        
        this.sendTestMessage('Test message 1 - right after client 2 pong');
        await this.sleep(3000);
        
        // Test 2: Send message at random time
        this.log('Test 2: Sending message at random time...');
        this.sendTestMessage('Test message 2 - random timing');
        await this.sleep(3000);
        
        // Test 3: Send ping to client 2, wait, then send message
        this.log('Test 3: Triggering pong for client 2, waiting, then sending message...');
        
        this.client2.send(JSON.stringify({ action: 'ping' }));
        await this.sleep(2000); // Wait longer after pong
        
        this.sendTestMessage('Test message 3 - 2 seconds after client 2 pong');
        await this.sleep(3000);
        
        // Test 4: Multiple messages in sequence
        this.log('Test 4: Sending multiple messages in sequence...');
        for (let i = 4; i <= 6; i++) {
            this.sendTestMessage(`Test message ${i} - sequence`);
            await this.sleep(500);
        }
        
        await this.sleep(5000);
    }

    sendTestMessage(message) {
        this.messagesSent++;
        const sentTime = Date.now();
        this.log(`Client 1: 📤 Sending: "${message}"`);
        
        this.client1.send(JSON.stringify({
            action: 'send_message',
            channel: 'test-broadcast',
            event: 'message',
            data: { message: message }
        }));
    }

    showResults() {
        this.log('📊 Test Results:');
        this.log(`   Messages sent: ${this.messagesSent}`);
        this.log(`   Messages received by client 2: ${this.messagesReceived}`);
        
        this.log('📈 Timing Analysis:');
        this.messageReceiveTimes.forEach((item, index) => {
            const pongRelation = item.timeSinceLastPong < 1000 ? 
                `🏓 ${item.timeSinceLastPong}ms after pong` : 
                `⏰ ${item.timeSinceLastPong}ms after pong`;
            
            this.log(`   Message ${index + 1}: ${pongRelation}`);
        });
        
        // Check if messages consistently arrive after pongs
        const messagesAfterRecentPong = this.messageReceiveTimes.filter(item => item.timeSinceLastPong < 1000);
        if (messagesAfterRecentPong.length > 0) {
            this.log('⚠️  PATTERN DETECTED: Some messages received shortly after pongs!');
        }
    }

    async waitForCondition(condition, timeout = 5000, errorMessage = 'Condition timeout') {
        const startTime = Date.now();
        
        while (!condition() && (Date.now() - startTime) < timeout) {
            await this.sleep(100);
        }
        
        if (!condition()) {
            throw new Error(errorMessage);
        }
    }

    sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    cleanup() {
        this.log('🧹 Cleaning up...');
        
        if (this.client1 && this.client1.readyState === WebSocket.OPEN) {
            this.client1.close();
        }
        
        if (this.client2 && this.client2.readyState === WebSocket.OPEN) {
            this.client2.close();
        }
    }
}

// Run the test
const test = new BroadcastTimingTest();
test.runTest().then(() => {
    console.log('✅ Test completed successfully');
    process.exit(0);
}).catch(error => {
    console.error('❌ Test failed:', error);
    process.exit(1);
});

// Handle Ctrl+C
process.on('SIGINT', () => {
    console.log('\n🔴 Test interrupted by user');
    process.exit(0);
});
