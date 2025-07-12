# WebSocket Connection Issues and Solutions

## Understanding WebSocket Error Code 1006

Error code 1006 indicates an "abnormal closure" of the WebSocket connection. This is a common issue that can occur due to various reasons:

### Common Causes of Error 1006

1. **Network Issues**
   - Intermittent network connectivity
   - Mobile networks switching between towers
   - WiFi dropping and reconnecting
   - Network proxy or firewall interference

2. **Client-Side Issues**
   - Browser tab closing without proper WebSocket close
   - Mobile app being backgrounded or killed
   - JavaScript errors causing connection to drop
   - Client device going to sleep mode

3. **Server-Side Issues**
   - Server restart or crash
   - Load balancer timeout
   - Connection timeout due to inactivity
   - Resource exhaustion (memory, file descriptors)

4. **Infrastructure Issues**
   - Reverse proxy timeout (nginx, Apache)
   - Load balancer configuration
   - CDN or edge server issues
   - Container orchestration platform restarts

## Improved Connection Handling

Our socket server now includes several improvements to handle these issues:

### 1. Connection Health Monitoring

```go
// Automatic ping/pong every 30 seconds
pingTicker := time.NewTicker(30 * time.Second)

// Pong handler resets read deadline
conn.SetPongHandler(func(string) error {
    conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    return nil
})
```

### 2. Connection Timeouts

```go
// Set various timeouts to prevent hanging connections
conn.SetReadLimit(512 * 1024)  // 512KB max message size
conn.SetReadDeadline(time.Now().Add(60 * time.Second))  // Read timeout
conn.SetWriteDeadline(time.Now().Add(10 * time.Second)) // Write timeout
```

### 3. Better Error Classification

The server now provides more detailed logging for different disconnect scenarios:

- **Normal Disconnection**: Client properly closes connection
- **Abnormal Disconnection**: Network issues, browser closed, etc.
- **Unexpected Disconnection**: Unexpected protocol errors

### 4. Graceful Error Handling

- Null connection checks before sending messages
- Proper mutex locking for concurrent access
- Graceful cleanup of client resources

## Client-Side Best Practices

### 1. Automatic Reconnection

```javascript
class ReconnectingWebSocket {
    constructor(url, options = {}) {
        this.url = url;
        this.options = {
            maxReconnectAttempts: 5,
            reconnectInterval: 1000,
            maxReconnectInterval: 30000,
            reconnectDecay: 1.5,
            ...options
        };
        this.reconnectAttempts = 0;
        this.connect();
    }

    connect() {
        this.ws = new WebSocket(this.url);
        
        this.ws.onopen = () => {
            console.log('Connected to server');
            this.reconnectAttempts = 0;
            this.onopen?.();
        };

        this.ws.onmessage = (event) => {
            this.onmessage?.(event);
        };

        this.ws.onclose = (event) => {
            console.log('Disconnected from server:', event.code, event.reason);
            this.handleReconnect();
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    handleReconnect() {
        if (this.reconnectAttempts < this.options.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const timeout = Math.min(
                this.options.reconnectInterval * Math.pow(this.options.reconnectDecay, this.reconnectAttempts),
                this.options.maxReconnectInterval
            );
            
            console.log(`Reconnecting in ${timeout}ms... (attempt ${this.reconnectAttempts})`);
            setTimeout(() => this.connect(), timeout);
        } else {
            console.error('Max reconnection attempts reached');
        }
    }

    send(data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(data);
        } else {
            console.warn('WebSocket not open, message queued');
            // You could implement message queuing here
        }
    }
}
```

### 2. Connection State Management

```javascript
class SocketClient {
    constructor(url) {
        this.url = url;
        this.ws = new ReconnectingWebSocket(url);
        this.authenticated = false;
        this.joinedChannels = new Set();
        this.messageQueue = [];
        
        this.ws.onopen = () => this.handleOpen();
        this.ws.onmessage = (event) => this.handleMessage(event);
    }

    handleOpen() {
        // Re-authenticate if needed
        if (this.authToken) {
            this.authenticate(this.authToken);
        }
        
        // Re-join channels
        this.joinedChannels.forEach(channel => {
            this.joinChannel(channel);
        });
        
        // Send queued messages
        this.flushMessageQueue();
    }

    authenticate(token) {
        this.authToken = token;
        this.send({
            action: 'authenticate',
            token: token
        });
    }

    joinChannel(channel) {
        this.joinedChannels.add(channel);
        this.send({
            action: 'join_channel',
            channel: channel
        });
    }

    send(message) {
        if (this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        } else {
            this.messageQueue.push(message);
        }
    }

    flushMessageQueue() {
        while (this.messageQueue.length > 0) {
            const message = this.messageQueue.shift();
            this.send(message);
        }
    }
}
```

### 3. Heartbeat Implementation

```javascript
class HeartbeatWebSocket {
    constructor(url) {
        this.url = url;
        this.connect();
    }

    connect() {
        this.ws = new WebSocket(this.url);
        
        this.ws.onopen = () => {
            this.startHeartbeat();
        };
        
        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            if (message.event === 'pong') {
                this.resetHeartbeat();
            }
        };
        
        this.ws.onclose = () => {
            this.stopHeartbeat();
        };
    }

    startHeartbeat() {
        this.heartbeatInterval = setInterval(() => {
            this.ping();
        }, 30000); // 30 seconds
    }

    stopHeartbeat() {
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
        }
    }

    resetHeartbeat() {
        this.stopHeartbeat();
        this.startHeartbeat();
    }

    ping() {
        if (this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({ action: 'ping' }));
        }
    }
}
```

## Server Configuration

### 1. Reverse Proxy Configuration (nginx)

```nginx
location /ws {
    proxy_pass http://localhost:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    
    # Important: Set longer timeouts for WebSocket connections
    proxy_read_timeout 3600s;
    proxy_send_timeout 3600s;
    proxy_connect_timeout 60s;
    
    # Disable buffering for real-time communication
    proxy_buffering off;
}
```

### 2. Load Balancer Configuration

If using a load balancer, ensure:
- Sticky sessions are enabled for WebSocket connections
- Health checks are configured properly
- Timeouts are set appropriately

### 3. Container Configuration

```yaml
# docker-compose.yml
version: '3.8'
services:
  socket-server:
    build: .
    ports:
      - "8080:8080"
    restart: unless-stopped
    environment:
      - SOCKET_PORT=8080
      - JWT_SECRET=your-secret-key
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## Monitoring and Debugging

### 1. Connection Metrics

Monitor these metrics to identify connection issues:

- Connection count over time
- Connection duration
- Disconnect rate by error code
- Message success/failure rates
- Ping/pong response times

### 2. Debugging Commands

```bash
# Check server health
curl http://localhost:8080/api/health

# Monitor connections
./bin/socket list clients

# View server logs
tail -f server.log

# Check system resources
netstat -an | grep :8080
lsof -i :8080
```

### 3. Log Analysis

Look for patterns in disconnect logs:
- Time of day when disconnects occur
- User agents of frequently disconnecting clients
- Network patterns (same IP ranges)
- Correlation with server load

## Best Practices

1. **Always implement client-side reconnection logic**
2. **Use exponential backoff for reconnection attempts**
3. **Implement proper error handling and logging**
4. **Set appropriate timeouts for your use case**
5. **Monitor connection health with ping/pong**
6. **Handle edge cases gracefully (null connections, etc.)**
7. **Test with poor network conditions**
8. **Implement proper authentication state management**
9. **Queue messages when connection is unavailable**
10. **Use proper WebSocket close codes when possible**

## Common Error Codes

- **1000**: Normal closure
- **1001**: Going away (page unload)
- **1002**: Protocol error
- **1003**: Unsupported data type
- **1006**: Abnormal closure (most common)
- **1007**: Invalid frame payload data
- **1008**: Policy violation
- **1009**: Message too big
- **1010**: Mandatory extension missing
- **1011**: Internal server error
- **1015**: TLS handshake failure
