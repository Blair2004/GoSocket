const WebSocket = require('ws');

class TestClient {
    constructor(url) {
        this.url = url;
        this.ws = null;
        this.isConnected = false;
        this.pingInterval = null;
        this.startTime = Date.now();
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.pingIntervalMs = 30000; // 30 seconds like the actual client
        this.lastPongTime = null;
        this.missedPongs = 0;
        this.maxMissedPongs = 3;
    }

    log(message, ...args) {
        const elapsed = ((Date.now() - this.startTime) / 1000).toFixed(1);
        console.log(`[${elapsed}s] ${message}`, ...args);
    }

    connect() {
        this.log('🔄 Connecting to:', this.url);
        
        try {
            this.ws = new WebSocket(this.url);
            this.setupEventHandlers();
        } catch (error) {
            this.log('❌ Connection error:', error.message);
            this.handleReconnect();
        }
    }

    setupEventHandlers() {
        this.ws.on('open', () => {
            this.log('✅ Connected to server');
            this.isConnected = true;
            this.reconnectAttempts = 0;
            this.missedPongs = 0;
            this.startPing();
        });

        this.ws.on('message', (data) => {
            try {
                const msg = JSON.parse(data);
                this.handleMessage(msg);
            } catch (error) {
                this.log('❌ Error parsing message:', error.message);
            }
        });

        this.ws.on('close', (code, reason) => {
            this.log('❌ Connection closed:', { code, reason: reason.toString() });
            this.isConnected = false;
            this.stopPing();
            
            if (code !== 1000) { // Not a normal closure
                this.handleReconnect();
            }
        });

        this.ws.on('error', (error) => {
            this.log('❌ WebSocket error:', error.message);
        });

        this.ws.on('ping', () => {
            this.log('🏓 Received server ping (WebSocket frame)');
            // WebSocket automatically handles pong response
        });

        this.ws.on('pong', () => {
            this.log('🏓 Received server pong (WebSocket frame)');
        });
    }

    handleMessage(msg) {
        this.log('📥 Received message:', JSON.stringify(msg, null, 2));
        
        switch (msg.event) {
            case 'connected':
                this.log('🎉 Welcome message received:', msg.data);
                break;
            case 'pong':
                this.log('🏓 Received pong response to our ping');
                this.lastPongTime = Date.now();
                this.missedPongs = 0;
                break;
            case 'error':
                this.log('❌ Server error:', msg.data);
                break;
            default:
                this.log('📨 Other message:', msg.event, msg.data);
        }
    }

    startPing() {
        this.stopPing();
        this.log('🏓 Starting ping mechanism (every 30s)');
        
        this.pingInterval = setInterval(() => {
            if (this.isConnected) {
                this.sendPing();
                this.checkPongTimeout();
            }
        }, this.pingIntervalMs);
    }

    stopPing() {
        if (this.pingInterval) {
            clearInterval(this.pingInterval);
            this.pingInterval = null;
            this.log('⏹️ Stopped ping mechanism');
        }
    }

    sendPing() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.log('📤 Sending ping...');
            this.ws.send(JSON.stringify({ action: 'ping' }));
        }
    }

    checkPongTimeout() {
        if (this.lastPongTime) {
            const timeSinceLastPong = Date.now() - this.lastPongTime;
            if (timeSinceLastPong > this.pingIntervalMs * 2) { // 60 seconds
                this.missedPongs++;
                this.log(`⚠️ Missed pong #${this.missedPongs} (${timeSinceLastPong}ms since last pong)`);
                
                if (this.missedPongs >= this.maxMissedPongs) {
                    this.log('💀 Too many missed pongs, connection likely dead');
                    this.ws.close(1000, 'Ping timeout');
                }
            }
        }
    }

    handleReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts - 1), 30000);
            this.log(`🔄 Attempting reconnect ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${delay}ms...`);
            
            setTimeout(() => {
                this.connect();
            }, delay);
        } else {
            this.log('💀 Max reconnect attempts reached, giving up');
            process.exit(1);
        }
    }

    disconnect() {
        this.log('🔴 Disconnecting...');
        this.stopPing();
        if (this.ws) {
            this.ws.close(1000, 'Manual disconnect');
        }
    }

    // Test methods
    testConnectionHealth() {
        this.log('🧪 Testing connection health...');
        
        // Send a few pings in quick succession
        for (let i = 0; i < 3; i++) {
            setTimeout(() => {
                this.sendPing();
            }, i * 1000);
        }
    }

    sendCustomMessage(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.log('📤 Sending custom message:', message);
            this.ws.send(JSON.stringify(message));
        }
    }
}

// Create and start the test client
const client = new TestClient('ws://localhost:8083/ws');
client.connect();

// Test scenarios
setTimeout(() => {
    client.log('🧪 Running connection health test...');
    client.testConnectionHealth();
}, 5000);

// Test custom message
setTimeout(() => {
    client.sendCustomMessage({
        action: 'custom_test',
        data: { test: 'message' }
    });
}, 10000);

// Monitor for a longer period
setTimeout(() => {
    client.log('🕐 Test running for 2 minutes to monitor connection stability...');
}, 15000);

// Graceful shutdown
process.on('SIGINT', () => {
    console.log('\n🔴 Shutting down test client...');
    client.disconnect();
    process.exit(0);
});

// Auto-shutdown after 5 minutes
setTimeout(() => {
    client.log('⏰ 5 minute test complete, shutting down...');
    client.disconnect();
    process.exit(0);
}, 300000);
