const WebSocket = require('ws');

class LongRunningTest {
    constructor() {
        this.ws = null;
        this.startTime = Date.now();
        this.pingCount = 0;
        this.pongCount = 0;
        this.isRunning = false;
        this.pingInterval = null;
        this.lastPongTime = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 3;
    }

    log(message, ...args) {
        const elapsed = ((Date.now() - this.startTime) / 1000).toFixed(1);
        const timestamp = new Date().toISOString().substr(11, 8);
        console.log(`[${timestamp}] [${elapsed}s] ${message}`, ...args);
    }

    connect() {
        this.log('🔄 Connecting to ws://localhost:8083/ws');
        
        this.ws = new WebSocket('ws://localhost:8083/ws');
        
        this.ws.on('open', () => {
            this.log('✅ Connected to server');
            this.isRunning = true;
            this.reconnectAttempts = 0;
            this.startPingLoop();
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
            this.isRunning = false;
            this.stopPingLoop();
            
            // Try to reconnect if not a normal close
            if (code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
                this.attemptReconnect();
            } else {
                this.log('🔴 Connection test ended');
                this.showFinalStats();
            }
        });

        this.ws.on('error', (error) => {
            this.log('❌ WebSocket error:', error.message);
        });
    }

    handleMessage(msg) {
        switch (msg.event) {
            case 'connected':
                this.log('🎉 Welcome message received');
                break;
            case 'pong':
                this.pongCount++;
                this.lastPongTime = Date.now();
                this.log(`🏓 Pong #${this.pongCount} received (${this.pongCount}/${this.pingCount} success rate)`);
                break;
            case 'error':
                this.log('❌ Server error:', msg.data);
                break;
            default:
                this.log('📨 Other message:', msg.event);
        }
    }

    startPingLoop() {
        this.stopPingLoop();
        this.log('🏓 Starting ping loop (every 10 seconds)');
        
        this.pingInterval = setInterval(() => {
            if (this.isRunning && this.ws && this.ws.readyState === WebSocket.OPEN) {
                this.sendPing();
            }
        }, 10000); // Ping every 10 seconds
    }

    stopPingLoop() {
        if (this.pingInterval) {
            clearInterval(this.pingInterval);
            this.pingInterval = null;
        }
    }

    sendPing() {
        this.pingCount++;
        this.log(`📤 Sending ping #${this.pingCount}...`);
        
        try {
            this.ws.send(JSON.stringify({ action: 'ping' }));
        } catch (error) {
            this.log('❌ Error sending ping:', error.message);
        }
    }

    attemptReconnect() {
        this.reconnectAttempts++;
        const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts - 1), 10000);
        this.log(`🔄 Attempting reconnect ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${delay}ms...`);
        
        setTimeout(() => {
            this.connect();
        }, delay);
    }

    showFinalStats() {
        const totalTime = (Date.now() - this.startTime) / 1000;
        const successRate = this.pingCount > 0 ? (this.pongCount / this.pingCount * 100).toFixed(1) : 0;
        
        this.log('📊 Final Statistics:');
        this.log(`   Total time: ${totalTime.toFixed(1)}s`);
        this.log(`   Pings sent: ${this.pingCount}`);
        this.log(`   Pongs received: ${this.pongCount}`);
        this.log(`   Success rate: ${successRate}%`);
        this.log(`   Reconnect attempts: ${this.reconnectAttempts}`);
        
        if (this.lastPongTime) {
            const timeSinceLastPong = Date.now() - this.lastPongTime;
            this.log(`   Time since last pong: ${timeSinceLastPong}ms`);
        }
    }

    disconnect() {
        this.log('🔴 Manually disconnecting...');
        this.isRunning = false;
        this.stopPingLoop();
        if (this.ws) {
            this.ws.close(1000, 'Manual disconnect');
        }
    }
}

// Start the test
const test = new LongRunningTest();
test.connect();

// Graceful shutdown
process.on('SIGINT', () => {
    console.log('\n🔴 Received SIGINT, shutting down...');
    test.disconnect();
    process.exit(0);
});

// Auto-shutdown after 3 minutes
setTimeout(() => {
    test.log('⏰ 3 minute test complete');
    test.disconnect();
    process.exit(0);
}, 180000);
