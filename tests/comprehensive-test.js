const WebSocket = require('ws');

class ComprehensiveTest {
    constructor() {
        this.clients = [];
        this.totalPings = 0;
        this.totalPongs = 0;
        this.connections = 0;
        this.disconnections = 0;
        this.startTime = Date.now();
    }

    log(message, ...args) {
        const elapsed = ((Date.now() - this.startTime) / 1000).toFixed(1);
        const timestamp = new Date().toISOString().substr(11, 8);
        console.log(`[${timestamp}] [${elapsed}s] ${message}`, ...args);
    }

    createClient(clientId) {
        const client = {
            id: clientId,
            ws: null,
            pings: 0,
            pongs: 0,
            connected: false,
            pingInterval: null
        };

        client.ws = new WebSocket('ws://localhost:8083/ws');

        client.ws.on('open', () => {
            this.log(`Client ${clientId}: ✅ Connected`);
            client.connected = true;
            this.connections++;
            this.startPingLoop(client);
        });

        client.ws.on('message', (data) => {
            try {
                const msg = JSON.parse(data);
                if (msg.event === 'pong') {
                    client.pongs++;
                    this.totalPongs++;
                    this.log(`Client ${clientId}: 🏓 Pong received (${client.pongs}/${client.pings})`);
                } else if (msg.event === 'connected') {
                    this.log(`Client ${clientId}: 🎉 Welcome message`);
                }
            } catch (error) {
                this.log(`Client ${clientId}: ❌ Parse error:`, error.message);
            }
        });

        client.ws.on('close', (code, reason) => {
            this.log(`Client ${clientId}: ❌ Disconnected (${code}): ${reason}`);
            client.connected = false;
            this.disconnections++;
            this.stopPingLoop(client);
        });

        client.ws.on('error', (error) => {
            this.log(`Client ${clientId}: ❌ Error:`, error.message);
        });

        return client;
    }

    startPingLoop(client) {
        client.pingInterval = setInterval(() => {
            if (client.connected && client.ws.readyState === WebSocket.OPEN) {
                client.pings++;
                this.totalPings++;
                this.log(`Client ${client.id}: 📤 Sending ping #${client.pings}`);
                client.ws.send(JSON.stringify({ action: 'ping' }));
            }
        }, 5000); // Ping every 5 seconds
    }

    stopPingLoop(client) {
        if (client.pingInterval) {
            clearInterval(client.pingInterval);
            client.pingInterval = null;
        }
    }

    async runTest() {
        this.log('🚀 Starting comprehensive connection test...');
        
        // Create 3 clients
        for (let i = 1; i <= 3; i++) {
            const client = this.createClient(i);
            this.clients.push(client);
            await new Promise(resolve => setTimeout(resolve, 1000)); // Stagger connections
        }

        // Let them run for 60 seconds
        await new Promise(resolve => setTimeout(resolve, 60000));

        // Disconnect all clients
        this.log('🔴 Disconnecting all clients...');
        for (const client of this.clients) {
            if (client.connected) {
                client.ws.close(1000, 'Test complete');
            }
        }

        // Wait for disconnections
        await new Promise(resolve => setTimeout(resolve, 2000));

        this.showResults();
    }

    showResults() {
        const duration = (Date.now() - this.startTime) / 1000;
        const successRate = this.totalPings > 0 ? (this.totalPongs / this.totalPings * 100).toFixed(1) : 0;
        
        this.log('📊 Test Results:');
        this.log(`   Duration: ${duration.toFixed(1)}s`);
        this.log(`   Clients created: ${this.clients.length}`);
        this.log(`   Connections: ${this.connections}`);
        this.log(`   Disconnections: ${this.disconnections}`);
        this.log(`   Total pings sent: ${this.totalPings}`);
        this.log(`   Total pongs received: ${this.totalPongs}`);
        this.log(`   Success rate: ${successRate}%`);
        
        // Per-client stats
        for (const client of this.clients) {
            const clientRate = client.pings > 0 ? (client.pongs / client.pings * 100).toFixed(1) : 0;
            this.log(`   Client ${client.id}: ${client.pongs}/${client.pings} (${clientRate}%)`);
        }
    }
}

// Run the test
const test = new ComprehensiveTest();
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
