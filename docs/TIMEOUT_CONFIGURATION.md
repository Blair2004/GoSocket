# WebSocket Server Timeout Configuration

Your WebSocket server now supports configurable timeouts to control when clients are considered disconnected.

## Available Timeout Settings

### 1. **Read Timeout** (`SOCKET_READ_TIMEOUT`)
- **Default**: `60s`
- **Purpose**: Maximum time to wait for client messages before considering the client disconnected
- **Examples**: `30s`, `1m`, `2m30s`

### 2. **Ping Interval** (`SOCKET_PING_INTERVAL`) 
- **Default**: `30s`
- **Purpose**: How often the server sends ping messages to check client connectivity
- **Examples**: `15s`, `45s`, `1m`

### 3. **Broadcast Timeout** (`SOCKET_BROADCAST_TIMEOUT`)
- **Default**: `200ms`
- **Purpose**: Maximum time to wait for broadcast operations to complete
- **Examples**: `100ms`, `500ms`, `1s`

### 4. **Write Timeout** (`SOCKET_WRITE_TIMEOUT`)
- **Default**: `10s`
- **Purpose**: Maximum time to wait for write operations to clients
- **Examples**: `5s`, `15s`, `30s`

## Usage Examples

### Via Environment Variables:
```bash
# Set a shorter read timeout for faster disconnection detection
SOCKET_READ_TIMEOUT=30s ./bin/socket-server

# Set longer ping interval for less network traffic
SOCKET_PING_INTERVAL=60s ./bin/socket-server

# Set faster broadcast timeout for better performance
SOCKET_BROADCAST_TIMEOUT=100ms ./bin/socket-server

# Combined configuration
SOCKET_READ_TIMEOUT=45s \
SOCKET_PING_INTERVAL=20s \
SOCKET_BROADCAST_TIMEOUT=300ms \
SOCKET_WRITE_TIMEOUT=15s \
./bin/socket-server
```

### Docker Environment:
```yaml
environment:
  - SOCKET_READ_TIMEOUT=30s
  - SOCKET_PING_INTERVAL=15s
  - SOCKET_BROADCAST_TIMEOUT=500ms
  - SOCKET_WRITE_TIMEOUT=10s
```

## How Disconnection Detection Works

1. **Read Timeout**: If no message is received from a client within this time, the connection is considered dead
2. **Ping/Pong**: Server sends ping every `SOCKET_PING_INTERVAL`, expecting a pong response
3. **Combined Effect**: A client is disconnected if:
   - No message received within `SOCKET_READ_TIMEOUT`, OR
   - Ping fails to send/receive pong response

## Recommended Settings

### Development:
```bash
SOCKET_READ_TIMEOUT=30s
SOCKET_PING_INTERVAL=15s
SOCKET_BROADCAST_TIMEOUT=500ms
```

### Production (High Traffic):
```bash
SOCKET_READ_TIMEOUT=60s
SOCKET_PING_INTERVAL=30s
SOCKET_BROADCAST_TIMEOUT=200ms
```

### Production (Low Latency):
```bash
SOCKET_READ_TIMEOUT=20s
SOCKET_PING_INTERVAL=10s
SOCKET_BROADCAST_TIMEOUT=100ms
```

## Monitoring

The server logs will now show the configured timeout values on startup:
```
Read Timeout: 30s
Ping Interval: 15s
Broadcast Timeout: 200ms
Write Timeout: 10s
```

## Performance Impact

- **Shorter timeouts**: Faster disconnection detection, higher CPU usage
- **Longer timeouts**: Slower disconnection detection, lower CPU usage
- **Broadcast timeout**: Affects how quickly broadcasts give up on slow clients

Choose values based on your application's requirements for responsiveness vs. resource usage.
