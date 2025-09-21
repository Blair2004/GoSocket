/**
 * Socket Client for Laravel Socket Server
 * 
 * A comprehensive WebSocket client for interacting with the custom socket server.
 * Works in both browser and Node.js environments.
 * 
 * @version 1.0.0
 * @author Your Name
 * @license MIT
 */

(function(global) {
    'use strict';

    // Environment detection
    const isBrowser = typeof window !== 'undefined';
    const isNode = typeof process !== 'undefined' && process.versions && process.versions.node;
    
    // WebSocket implementation
    let WebSocketImpl;
    if (isBrowser) {
        WebSocketImpl = window.WebSocket || window.MozWebSocket;
    } else if (isNode) {
        try {
            WebSocketImpl = require('ws');
        } catch (e) {
            console.error('WebSocket module not found. Install with: npm install ws');
        }
    }

    /**
     * Laravel Socket Client Class
     */
    class LaravelSocketClient {
        constructor(options = {}) {
            // Configuration
            this.config = {
                url: options.url || 'ws://localhost:8080/ws',
                reconnect: options.reconnect !== false,
                reconnectInterval: options.reconnectInterval || 1000,
                reconnectMaxAttempts: options.reconnectMaxAttempts || 5,
                heartbeat: options.heartbeat !== false,
                heartbeatInterval: options.heartbeatInterval || 30000,
                debug: options.debug || false,
                token: options.token || null,
                autoConnect: options.autoConnect !== false,
                ...options
            };

            // State
            this.ws = null;
            this.connected = false;
            this.authenticated = false;
            this.connecting = false;
            this.reconnectAttempts = 0;
            this.lastPing = null;
            this.lastPong = null;
            
            // Event management
            this.listeners = new Map();
            this.channels = new Set();
            this.messageQueue = [];
            
            // Timers
            this.reconnectTimer = null;
            this.heartbeatTimer = null;
            this.pingTimer = null;

            // Auto-connect if enabled
            if (this.config.autoConnect) {
                this.connect();
            }

            this.log('LaravelSocketClient initialized', this.config);
        }

        /**
         * Connect to the socket server
         */
        connect() {
            if (this.connecting || this.connected) {
                this.log('Already connecting or connected');
                return Promise.resolve();
            }

            return new Promise((resolve, reject) => {
                try {
                    this.connecting = true;
                    this.log('Connecting to', this.config.url);

                    this.ws = new WebSocketImpl(this.config.url);
                    
                    // Connection opened
                    this.ws.onopen = (event) => {
                        this.connecting = false;
                        this.connected = true;
                        this.reconnectAttempts = 0;
                        
                        this.log('Connected to socket server');
                        this.emit('connected', event);
                        
                        // Start heartbeat
                        if (this.config.heartbeat) {
                            this.startHeartbeat();
                        }
                        
                        // Authenticate if token provided
                        if (this.config.token) {
                            this.authenticate(this.config.token);
                        }
                        
                        // Process queued messages
                        this.processMessageQueue();
                        
                        resolve();
                    };

                    // Message received
                    this.ws.onmessage = (event) => {
                        try {
                            const message = JSON.parse(event.data);
                            this.handleMessage(message);
                        } catch (error) {
                            this.log('Error parsing message:', error, event.data);
                            this.emit('error', { type: 'parse_error', error, data: event.data });
                        }
                    };

                    // Connection closed
                    this.ws.onclose = (event) => {
                        this.connecting = false;
                        this.connected = false;
                        this.authenticated = false;
                        
                        this.log('Disconnected from socket server', event.code, event.reason);
                        this.emit('disconnected', event);
                        
                        this.stopHeartbeat();
                        
                        // Auto-reconnect if enabled
                        if (this.config.reconnect && !event.wasClean) {
                            this.scheduleReconnect();
                        }
                    };

                    // Connection error
                    this.ws.onerror = (error) => {
                        this.log('WebSocket error:', error);
                        this.emit('error', { type: 'connection_error', error });
                        
                        if (this.connecting) {
                            reject(error);
                        }
                    };

                } catch (error) {
                    this.connecting = false;
                    this.log('Connection error:', error);
                    reject(error);
                }
            });
        }

        /**
         * Disconnect from the socket server
         */
        disconnect() {
            this.config.reconnect = false; // Disable auto-reconnect
            this.clearTimers();
            
            if (this.ws) {
                this.ws.close(1000, 'Client disconnect');
                this.ws = null;
            }
            
            this.connected = false;
            this.authenticated = false;
            this.channels.clear();
            
            this.log('Disconnected');
            this.emit('disconnected', { wasClean: true });
        }

        /**
         * Send a message to the server
         */
        send(message) {
            if (!this.connected) {
                this.log('Not connected, queueing message:', message);
                this.messageQueue.push(message);
                return false;
            }

            try {
                const jsonMessage = JSON.stringify(message);
                this.ws.send(jsonMessage);
                this.log('Sent message:', message);
                this.emit('message_sent', message);
                return true;
            } catch (error) {
                this.log('Error sending message:', error);
                this.emit('error', { type: 'send_error', error, message });
                return false;
            }
        }

        /**
         * Authenticate with JWT token
         */
        authenticate(token) {
            this.config.token = token;
            return this.send({
                action: 'authenticate',
                token: token
            });
        }

        /**
         * Join a channel
         */
        join(channel) {
            this.channels.add(channel);
            return this.send({
                action: 'join_channel',
                channel: channel
            });
        }

        /**
         * Leave a channel
         */
        leave(channel) {
            this.channels.delete(channel);
            return this.send({
                action: 'leave_channel',
                channel: channel
            });
        }

        /**
         * Send a message to a channel
         */
        sendToChannel(channel, event, data = {}) {
            return this.send({
                action: 'send_message',
                channel: channel,
                event: event,
                data: data
            });
        }

        /**
         * Send a ping to the server
         */
        ping() {
            this.lastPing = Date.now();
            return this.send({ action: 'ping' });
        }

        /**
         * Handle incoming messages
         */
        handleMessage(message) {
            this.log('Received message:', message);
            
            // Handle system events
            switch (message.event) {
                case 'connected':
                    this.emit('server_connected', message);
                    break;
                    
                case 'authenticated':
                    this.authenticated = true;
                    this.emit('authenticated', message.data);
                    break;
                    
                case 'joined_channel':
                    this.emit('channel_joined', message.data);
                    break;
                    
                case 'left_channel':
                    this.emit('channel_left', message.data);
                    break;
                    
                case 'pong':
                    this.lastPong = Date.now();
                    this.emit('pong', message);
                    break;
                    
                case 'error':
                    this.emit('server_error', message.data);
                    break;
                    
                case 'kicked':
                    this.emit('kicked', message.data);
                    this.disconnect();
                    break;
                    
                default:
                    // Channel-specific events
                    if (message.channel) {
                        this.emit(`channel:${message.channel}`, message);
                        this.emit(`channel:${message.channel}:${message.event}`, message);
                    }
                    
                    // Event-specific listeners
                    this.emit(message.event, message);
                    break;
            }
            
            // General message event
            this.emit('message', message);
        }

        /**
         * Process queued messages
         */
        processMessageQueue() {
            while (this.messageQueue.length > 0 && this.connected) {
                const message = this.messageQueue.shift();
                this.send(message);
            }
        }

        /**
         * Schedule reconnection attempt
         */
        scheduleReconnect() {
            if (this.reconnectAttempts >= this.config.reconnectMaxAttempts) {
                this.log('Max reconnection attempts reached');
                this.emit('reconnect_failed');
                return;
            }

            const delay = this.config.reconnectInterval * Math.pow(2, this.reconnectAttempts);
            this.reconnectAttempts++;

            this.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
            
            this.reconnectTimer = setTimeout(() => {
                this.emit('reconnecting', this.reconnectAttempts);
                this.connect().catch(error => {
                    this.log('Reconnection failed:', error);
                });
            }, delay);
        }

        /**
         * Start heartbeat mechanism
         */
        startHeartbeat() {
            this.heartbeatTimer = setInterval(() => {
                if (this.connected) {
                    this.ping();
                }
            }, this.config.heartbeatInterval);
        }

        /**
         * Stop heartbeat mechanism
         */
        stopHeartbeat() {
            if (this.heartbeatTimer) {
                clearInterval(this.heartbeatTimer);
                this.heartbeatTimer = null;
            }
        }

        /**
         * Clear all timers
         */
        clearTimers() {
            if (this.reconnectTimer) {
                clearTimeout(this.reconnectTimer);
                this.reconnectTimer = null;
            }
            this.stopHeartbeat();
        }

        /**
         * Add event listener
         */
        on(event, callback) {
            if (!this.listeners.has(event)) {
                this.listeners.set(event, []);
            }
            this.listeners.get(event).push(callback);
            return this;
        }

        /**
         * Remove event listener
         */
        off(event, callback) {
            if (this.listeners.has(event)) {
                const callbacks = this.listeners.get(event);
                const index = callbacks.indexOf(callback);
                if (index > -1) {
                    callbacks.splice(index, 1);
                }
            }
            return this;
        }

        /**
         * Add one-time event listener
         */
        once(event, callback) {
            const wrapper = (...args) => {
                callback(...args);
                this.off(event, wrapper);
            };
            return this.on(event, wrapper);
        }

        /**
         * Emit event to listeners
         */
        emit(event, data = null) {
            if (this.listeners.has(event)) {
                this.listeners.get(event).forEach(callback => {
                    try {
                        callback(data);
                    } catch (error) {
                        this.log('Error in event listener:', error);
                    }
                });
            }
        }

        /**
         * Get connection status
         */
        getStatus() {
            return {
                connected: this.connected,
                authenticated: this.authenticated,
                connecting: this.connecting,
                channels: Array.from(this.channels),
                reconnectAttempts: this.reconnectAttempts,
                queuedMessages: this.messageQueue.length,
                lastPing: this.lastPing,
                lastPong: this.lastPong,
                latency: this.lastPing && this.lastPong ? this.lastPong - this.lastPing : null
            };
        }

        /**
         * Get list of joined channels
         */
        getChannels() {
            return Array.from(this.channels);
        }

        /**
         * Check if connected to a specific channel
         */
        inChannel(channel) {
            return this.channels.has(channel);
        }

        /**
         * Set authentication token
         */
        setToken(token) {
            this.config.token = token;
            if (this.connected) {
                this.authenticate(token);
            }
        }

        /**
         * Log messages (if debug enabled)
         */
        log(...args) {
            if (this.config.debug) {
                const timestamp = new Date().toISOString();
                console.log(`[SocketClient ${timestamp}]`, ...args);
            }
        }

        /**
         * Destroy the client instance
         */
        destroy() {
            this.disconnect();
            this.clearTimers();
            this.listeners.clear();
            this.messageQueue = [];
            this.channels.clear();
        }
    }

    /**
     * Helper methods for common use cases
     */
    LaravelSocketClient.prototype.chat = {
        // Send a chat message
        send: function(roomId, message, type = 'text') {
            return this.sendToChannel(`chat.room.${roomId}`, 'chat_message', {
                message: message,
                room_id: roomId,
                type: type
            });
        }.bind(this),

        // Join a chat room
        join: function(roomId) {
            return this.join(`chat.room.${roomId}`);
        }.bind(this),

        // Leave a chat room
        leave: function(roomId) {
            return this.leave(`chat.room.${roomId}`);
        }.bind(this),

        // Send typing indicator
        typing: function(roomId, isTyping = true) {
            return this.sendToChannel(`chat.room.${roomId}`, 'user_action', {
                action: 'typing',
                is_typing: isTyping,
                room_id: roomId
            });
        }.bind(this)
    };

    /**
     * Utility methods
     */
    LaravelSocketClient.createAuthToken = function(userData, secret) {
        // WARNING: This is a simplified JWT implementation for demo purposes only!
        // In production, use a proper JWT library like jsonwebtoken or jose
        
        console.warn('LaravelSocketClient.createAuthToken: This is a demo implementation. Use a proper JWT library in production!');
        
        // JWT requires proper HMAC-SHA256 signature, which is not available in browser without crypto library
        // This demo creates a token that will NOT pass server validation
        // You should generate JWT tokens server-side in your Laravel application
        
        const header = btoa(JSON.stringify({ typ: 'JWT', alg: 'HS256' }))
            .replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
        
        const payload = btoa(JSON.stringify({
            ...userData,
            iat: Math.floor(Date.now() / 1000),
            exp: Math.floor(Date.now() / 1000) + (60 * 60) // 1 hour
        })).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
        
        // This creates an INVALID signature - it's just for demo purposes
        // Real JWT requires HMAC-SHA256(header.payload, secret)
        const fakeSignature = btoa('DEMO_SIGNATURE_NOT_VALID')
            .replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
            
        return `${header}.${payload}.${fakeSignature}`;
    };

    /**
     * Generate a proper JWT token server-side (Laravel example)
     * 
     * In your Laravel application, use this code to generate valid JWT tokens:
     * 
     * use Firebase\JWT\JWT;
     * use Firebase\JWT\Key;
     * 
     * $payload = [
     *     'user_id' => auth()->id(),
     *     'username' => auth()->user()->name,
     *     'email' => auth()->user()->email,
     *     'iat' => time(),
     *     'exp' => time() + (60 * 60) // 1 hour
     * ];
     * 
     * $jwtSecret = config('app.jwt_secret'); // Same as your Go server
     * $token = JWT::encode($payload, $jwtSecret, 'HS256');
     * 
     * Then pass this token to your JavaScript client:
     * const client = new LaravelSocketClient({
     *     url: 'ws://localhost:8080',
     *     token: token // Use the server-generated token
     * });
     */
    LaravelSocketClient.generateProperTokenExample = function() {
        return {
            laravel_example: `
// In your Laravel controller:
use Firebase\\JWT\\JWT;

public function generateSocketToken() {
    $payload = [
        'user_id' => auth()->id(),
        'username' => auth()->user()->name,
        'email' => auth()->user()->email,
        'iat' => time(),
        'exp' => time() + (60 * 60)
    ];
    
    $jwtSecret = env('JWT_SECRET'); // Same as Go server
    return JWT::encode($payload, $jwtSecret, 'HS256');
}`,
            javascript_usage: `
// In your JavaScript:
fetch('/api/socket-token')
    .then(response => response.json())
    .then(data => {
        const client = new LaravelSocketClient({
            url: 'ws://localhost:8080',
            token: data.token
        });
    });`
        };
    };

    LaravelSocketClient.parseMessage = function(data) {
        try {
            return typeof data === 'string' ? JSON.parse(data) : data;
        } catch (error) {
            console.error('Error parsing message:', error);
            return null;
        }
    };

    // Export for different environments
    if (isBrowser) {
        window.LaravelSocketClient = LaravelSocketClient;
    } else if (isNode) {
        module.exports = LaravelSocketClient;
    }

    // Also support UMD pattern
    if (typeof define === 'function' && define.amd) {
        define([], function() { return LaravelSocketClient; });
    }

    return LaravelSocketClient;

})(typeof self !== 'undefined' ? self : this);

/**
 * Usage Examples:
 * 
 * // Basic connection
 * const client = new LaravelSocketClient({
 *     url: 'ws://localhost:8080/ws',
 *     debug: true
 * });
 * 
 * // With authentication
 * const client = new LaravelSocketClient({
 *     url: 'wss://your-domain.com/ws',
 *     token: 'your-jwt-token',
 *     debug: false
 * });
 * 
 * // Event listeners
 * client.on('connected', () => console.log('Connected!'));
 * client.on('authenticated', (data) => console.log('Authenticated:', data));
 * client.on('message', (message) => console.log('Message:', message));
 * 
 * // Channel operations
 * client.join('notifications');
 * client.sendToChannel('chat.room.1', 'message', {text: 'Hello!'});
 * 
 * // Chat helpers
 * client.chat.join(1);
 * client.chat.send(1, 'Hello everyone!');
 * client.chat.typing(1, true);
 * 
 * // Status monitoring
 * console.log(client.getStatus());
 */
