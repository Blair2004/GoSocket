const WebSocket = require('ws');

// Test client for per-channel metadata storage and forwarding
class MetadataTestClient {
    constructor(id, url = 'ws://localhost:8080/ws') {
        this.id = id;
        this.url = url;
        this.ws = null;
        this.isConnected = false;
        this.messageHandlers = {};
        this.joinedChannels = new Set();
    }

    async connect() {
        return new Promise((resolve, reject) => {
            this.ws = new WebSocket(this.url);

            this.ws.on('open', () => {
                console.log(`[${this.id}] Connected to WebSocket server`);
                this.isConnected = true;
                resolve();
            });

            this.ws.on('message', (data) => {
                try {
                    const message = JSON.parse(data);
                    this.handleMessage(message);
                } catch (error) {
                    console.error(`[${this.id}] Error parsing message:`, error);
                }
            });

            this.ws.on('close', () => {
                console.log(`[${this.id}] Connection closed`);
                this.isConnected = false;
            });

            this.ws.on('error', (error) => {
                console.error(`[${this.id}] WebSocket error:`, error);
                reject(error);
            });
        });
    }

    handleMessage(message) {
        console.log(`[${this.id}] Received:`, JSON.stringify(message, null, 2));
        
        // Track joined channels
        if (message.event === 'joined_channel') {
            this.joinedChannels.add(message.data.channel);
        } else if (message.event === 'left_channel') {
            this.joinedChannels.delete(message.data.channel);
        }

        // Call specific handlers
        if (this.messageHandlers[message.event]) {
            this.messageHandlers[message.event](message);
        }
    }

    onMessage(event, handler) {
        this.messageHandlers[event] = handler;
    }

    joinChannel(channel, metadata = null) {
        if (!this.isConnected) {
            console.error(`[${this.id}] Cannot join channel: not connected`);
            return;
        }

        const message = {
            event: 'join_channel',
            channel: channel
        };

        if (metadata !== null) {
            message.data = metadata;
        }

        console.log(`[${this.id}] Joining channel '${channel}' with metadata:`, metadata);
        this.ws.send(JSON.stringify(message));
    }

    leaveChannel(channel) {
        if (!this.isConnected) {
            console.error(`[${this.id}] Cannot leave channel: not connected`);
            return;
        }

        console.log(`[${this.id}] Leaving channel '${channel}'`);
        this.ws.send(JSON.stringify({
            event: 'leave_channel',
            channel: channel
        }));
    }

    disconnect() {
        if (this.ws) {
            console.log(`[${this.id}] Disconnecting...`);
            this.ws.close();
        }
    }
}

// Test scenarios
async function runMetadataTests() {
    console.log('=== Per-Channel Metadata Test Started ===\n');

    const client1 = new MetadataTestClient('client1');
    const client2 = new MetadataTestClient('client2');

    try {
        // Connect clients
        await client1.connect();
        await client2.connect();

        // Test 1: Join channel with metadata
        console.log('\n--- Test 1: Join channel with metadata ---');
        await new Promise(resolve => setTimeout(resolve, 500));
        
        client1.joinChannel('test-channel', {
            user_type: 'premium',
            joined_from: 'web_app',
            session_id: 'abc123',
            preferences: {
                notifications: true,
                theme: 'dark'
            }
        });

        // Test 2: Join same channel with different metadata
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        client2.joinChannel('test-channel', {
            user_type: 'basic',
            joined_from: 'mobile_app',
            session_id: 'def456',
            preferences: {
                notifications: false,
                theme: 'light'
            }
        });

        // Test 3: Manual leave (should forward stored metadata)
        console.log('\n--- Test 3: Manual leave with stored metadata ---');
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        client1.leaveChannel('test-channel');

        // Test 4: Join another channel with different metadata
        console.log('\n--- Test 4: Join different channel ---');
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        client2.joinChannel('another-channel', {
            room_type: 'private',
            access_level: 'admin',
            created_by: 'client2'
        });

        // Test 5: Disconnect while in channels (should trigger leave events with metadata)
        console.log('\n--- Test 5: Disconnect with stored metadata ---');
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        console.log(`[client2] Disconnecting while in channels: ${Array.from(client2.joinedChannels).join(', ')}`);
        client2.disconnect();

        // Wait for events to propagate
        await new Promise(resolve => setTimeout(resolve, 2000));

        // Test 6: Join without metadata (should use default)
        console.log('\n--- Test 6: Join without metadata ---');
        const client3 = new MetadataTestClient('client3');
        await client3.connect();
        
        client3.joinChannel('no-metadata-channel');
        
        await new Promise(resolve => setTimeout(resolve, 1000));
        client3.leaveChannel('no-metadata-channel');

        // Cleanup
        await new Promise(resolve => setTimeout(resolve, 1000));
        client1.disconnect();
        client3.disconnect();

        console.log('\n=== Per-Channel Metadata Test Complete ===');

    } catch (error) {
        console.error('Test failed:', error);
    }
}

// Run the test
runMetadataTests().catch(console.error);
