# Socket Server

A high-performance standalone socket server with Laravel integration, built as an alternative to Laravel Reverb.

## Features

- **Standalone Binary**: Compiled Go binary that can run independently
- **WebSocket Support**: Real-time bidirectional communication
- **Channel Management**: Secure isolated channels with authentication
- **Client Management**: List, monitor, and kick connected clients
- **Laravel Integration**: Seamless integration with Laravel events
- **CLI Interface**: Command-line tool for sending messages
- **Web Dashboard**: Real-time monitoring interface
- **JWT Authentication**: Secure user sessions
- **Rate Limiting**: Protection against abuse
- **RESTful API**: HTTP endpoints for management
- **HTTP API Security**: Token-based authentication for REST endpoints

## Quick Start

### 1. Build the Server

```bash
chmod +x build.sh
./build.sh
```

Go must be available globally. You might need to export it if it's not yet the case:

```bash
export PATH=/usr/local/go/bin:$PATH
```

### 2. Start the Server

```bash
# Using command-line options (HTTP token is required)
./bin/socket-server --port 8080 --token "your-jwt-secret" --http-token "your-api-token" --dir /path/to/laravel --php /usr/bin/php8.2 --command "socket:handle"

# Or using environment variables (optional)
export SOCKET_PORT=8080
export JWT_SECRET="your-jwt-secret"
export HTTP_TOKEN="your-api-token"
export LARAVEL_PATH="/path/to/laravel"
export PHP_BINARY="/usr/bin/php8.2"
export LARAVEL_COMMAND="socket:handle"
./bin/socket-server

# Or mix both (command-line flags take precedence)
export JWT_SECRET="fallback-secret"
export HTTP_TOKEN="fallback-api-token"
./bin/socket-server --port 9000 --dir /var/www/laravel --command "custom:socket-handler"
```

**Security Note**: The `--http-token` parameter is required for HTTP API security. Generate a secure token:
```bash
export HTTP_TOKEN="$(openssl rand -hex 32)"
```

### 3. Test with CLI

```bash
# Send a message to a channel
./bin/socket send --channel "test" --event "message" --data '{"text":"Hello World"}'

# List connected clients
./bin/socket list clients

# List active channels
./bin/socket list channels

# Check server health
./bin/socket health
```

### 4. Laravel Integration

Copy the Laravel integration files to your Laravel project:

```bash
# Copy trait
cp laravel/app/Traits/InteractsWithSockets.php app/Traits/

# Copy listener
cp laravel/app/Listeners/SocketEventListener.php app/Listeners/

# Copy service provider
cp laravel/app/Providers/SocketServiceProvider.php app/Providers/

# Copy config
cp laravel/config/socket.php config/

# Copy example events
cp laravel/app/Events/* app/Events/
```

Add to your `config/app.php`:

```php
'providers' => [
    // ... other providers
    App\Providers\SocketServiceProvider::class,
],
```

### 5. Create Laravel Events

```php
<?php

namespace App\Events;

use App\Traits\InteractsWithSockets;

class OrderCreated
{
    use InteractsWithSockets;

    public $order;

    public function __construct($order)
    {
        $this->order = $order;
    }

    public function broadcastOn()
    {
        return ['orders', 'user.' . $this->order->user_id];
    }
}
```

Then dispatch events as usual:

```php
event(new OrderCreated($order));
```

## Configuration

### Command-Line Options

```bash
# View all available options
./bin/socket-server --help

# Start with custom configuration
./bin/socket-server --port 9000 --token "my-secret-key" --dir /var/www/laravel --php /usr/bin/php8.2 --command "socket:handle"

# Short form flags (where available)
./bin/socket-server -p 9000 -t "my-secret-key" -d /var/www/laravel
```

Available flags:
- `--port, -p`: Server port (default: 8080 or SOCKET_PORT env var)
- `--token, -t`: JWT secret for authentication (default: JWT_SECRET env var)
- `--dir, -d`: Working directory for Laravel commands (default: LARAVEL_PATH env var or current directory)
- `--php`: PHP binary path (default: 'php' or PHP_BINARY env var)
- `--command`: Laravel artisan command to execute (default: 'socket:handle' or LARAVEL_COMMAND env var)

### Environment Variables

- `SOCKET_PORT`: Server port (default: 8080)
- `JWT_SECRET`: JWT signing secret
- `LARAVEL_PATH`: Working directory for Laravel commands
- `PHP_BINARY`: PHP binary path (default: 'php')
- `LARAVEL_COMMAND`: Laravel artisan command to execute (default: 'socket:handle')
- `SOCKET_TEMP_DIR`: Temporary directory for payload files (default: system temp/socket-server-payloads)
- `SOCKET_BINARY_PATH`: Path to socket CLI binary
- `SOCKET_SERVER_URL`: Socket server URL for CLI

### Laravel Configuration

Update your `.env`:

```env
SOCKET_BINARY_PATH=/path/to/socket
SOCKET_SERVER_URL=http://localhost:8080
SOCKET_JWT_SECRET=your-jwt-secret
SOCKET_DEBUG=false
LARAVEL_PATH=/var/www/laravel
PHP_BINARY=/usr/bin/php8.2
LARAVEL_COMMAND=socket:handle
```

## API Endpoints

### WebSocket
- `GET /ws` - WebSocket connection endpoint

### REST API
- `GET /api/health` - Server health check
- `GET /api/clients` - List connected clients
- `GET /api/channels` - List active channels
- `GET /api/channels/{channel}/clients` - List clients in channel
- `POST /api/clients/{client}/kick` - Kick a client
- `POST /api/broadcast` - Broadcast message to channel

### Dashboard
- `GET /` - Web dashboard for monitoring

## WebSocket Protocol

### Client Messages

#### Authentication
```json
{
    "action": "authenticate",
    "token": "jwt-token"
}
```

#### Join Channel
```json
{
    "action": "join_channel",
    "channel": "channel-name"
}
```

#### Leave Channel
```json
{
    "action": "leave_channel",
    "channel": "channel-name"
}
```

#### Send Message
```json
{
    "action": "send_message",
    "channel": "channel-name",
    "event": "event-name",
    "data": {"key": "value"}
}
```

#### Ping
```json
{
    "action": "ping"
}
```

### Server Messages

#### Connected
```json
{
    "id": "message-id",
    "event": "connected",
    "data": {"client_id": "client-id"},
    "timestamp": "2025-01-01T00:00:00Z"
}
```

#### Authenticated
```json
{
    "id": "message-id",
    "event": "authenticated",
    "data": {"user_id": "123", "username": "john"},
    "timestamp": "2025-01-01T00:00:00Z"
}
```

#### Channel Events
```json
{
    "id": "message-id",
    "channel": "channel-name",
    "event": "event-name",
    "data": {"key": "value"},
    "user_id": "123",
    "username": "john",
    "timestamp": "2025-01-01T00:00:00Z"
}
```

#### Errors
```json
{
    "id": "message-id",
    "event": "error",
    "data": {"error": "Error message"},
    "timestamp": "2025-01-01T00:00:00Z"
}
```

## CLI Usage

**Note**: All CLI commands now require an HTTP API token for authentication.

### Send Messages

```bash
# Send from JSON file (with authentication)
./bin/socket --token "your-api-token" send --file /path/to/message.json

# Send with flags (with authentication)
./bin/socket --token "your-api-token" send --channel "notifications" --event "alert" --data '{"message":"Server maintenance"}'

# Using environment variable for token
export HTTP_TOKEN="your-api-token"
./bin/socket send --channel "notifications" --event "alert" --data '{"message":"Server maintenance"}'
```

### Management

```bash
# List all clients
./bin/socket --token "your-api-token" list clients

# List all channels
./bin/socket --token "your-api-token" list channels

# Kick a client
./bin/socket --token "your-api-token" kick client-id

# Check server health
./bin/socket --token "your-api-token" health

# Using environment variable for all commands
export HTTP_TOKEN="your-api-token"
./bin/socket list clients
./bin/socket health
```

### Custom Server URL

```bash
./bin/socket --server http://localhost:9000 send --channel test --data '{"test":true}'
```

### 🔐 TLS Certificate Verification Bypass

For development environments or when using self-signed certificates, you can bypass TLS certificate verification:

```bash
# Skip certificate verification for HTTPS endpoints
./bin/socket --server https://localhost:8443 --insecure health

# This is equivalent to curl -k
./bin/socket --server https://self-signed.example.com --insecure send --channel test --data '{"test":true}'

# Useful for:
# - Development with self-signed certificates
# - Testing environments
# - Internal networks with custom CAs
```

**⚠️ Security Warning**: Only use `--insecure` in development environments. Never use this in production as it disables certificate validation and makes connections vulnerable to man-in-the-middle attacks.

## Client Examples

### JavaScript (Browser)

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = function() {
    // Join a channel
    ws.send(JSON.stringify({
        action: 'join_channel',
        channel: 'notifications'
    }));
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    console.log('Received:', message);
};

// Send a message
ws.send(JSON.stringify({
    action: 'send_message',
    channel: 'chat',
    event: 'message',
    data: {text: 'Hello everyone!'}
}));
```

### JavaScript (Node.js)

```javascript
const WebSocket = require('ws');

const ws = new WebSocket('ws://localhost:8080/ws');

ws.on('open', function() {
    ws.send(JSON.stringify({
        action: 'join_channel',
        channel: 'updates'
    }));
});

ws.on('message', function(data) {
    const message = JSON.parse(data);
    console.log('Received:', message);
});
```

### PHP

```php
// Using ReactPHP/Socket
$connector = new React\Socket\Connector();
$connector->connect('tcp://localhost:8080')
    ->then(function (React\Socket\ConnectionInterface $connection) {
        $connection->write(json_encode([
            'action' => 'join_channel',
            'channel' => 'events'
        ]));
        
        $connection->on('data', function ($data) {
            echo "Received: " . $data;
        });
    });
```

## Security Features

### JWT Authentication
- Secure user sessions with JWT tokens
- Configurable token expiration
- User information embedded in tokens

### Channel Authorization
- Private channels requiring authentication
- Custom authorization logic support
- User-specific channels

### Rate Limiting
- Connection limits per IP
- Message rate limiting
- Configurable thresholds

### CORS Support
- Configurable allowed origins
- Header and method restrictions
- WebSocket upgrade protection

## Performance

### Benchmarks
- Handles 10,000+ concurrent connections
- Sub-millisecond message latency
- Minimal memory footprint
- Efficient channel management

### Scalability
- Horizontal scaling with load balancer
- Redis support for multi-instance setups
- Persistent connection management
- Efficient message broadcasting

## Troubleshooting

### Connection Issues

If you encounter WebSocket disconnection errors (like error code 1006), see the comprehensive troubleshooting guide:

```bash
# View troubleshooting guide
cat WEBSOCKET_TROUBLESHOOTING.md

# Test with resilient client
open resilient-websocket-client.html
```

### Common Issues

#### Server Won't Start
```bash
# Check if port is available
lsof -i :8080

# Check logs for errors
./bin/socket-server 2>&1 | tee server.log
```

#### Laravel Events Not Broadcasting
```bash
# Test CLI connectivity
./bin/socket health

# Check Laravel logs
tail -f storage/logs/laravel.log

# Verify binary path
which socket
```

#### WebSocket Connection Fails
```bash
# Test WebSocket endpoint
curl -I http://localhost:8080/ws

# Check CORS settings
curl -H "Origin: http://localhost:3000" http://localhost:8080/api/health
```

### Connection Improvements

The server now includes several improvements for handling unstable connections:

- **Automatic ping/pong**: Keeps connections alive and detects dead connections
- **Connection timeouts**: Prevents hanging connections
- **Better error classification**: Distinguishes between normal, abnormal, and unexpected disconnections
- **Graceful error handling**: Proper cleanup and resource management
- **Detailed logging**: Better visibility into connection issues

### Debug Mode

Enable debug logging:

```env
SOCKET_DEBUG=true
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Support

- GitHub Issues: Report bugs and feature requests
- Documentation: See the `/docs` directory
  - [HTTP API Security](docs/HTTP_API_SECURITY.md) - Authentication setup and usage
- Examples: Check the `/examples` directory
