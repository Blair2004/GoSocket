# HTTP API Security

The socket server now includes HTTP API authentication to secure all REST endpoints. This ensures only authorized clients can access the server's administrative and broadcast functionality.

## Server Configuration

### Starting the Server with Authentication

When starting the socket server, you must provide an HTTP API token:

```bash
# Using command line flag
./bin/socket-server --server-token "your-secure-token-here"

# Using environment variable
export HTTP_TOKEN="your-secure-token-here"
./bin/socket-server

# Using both JWT token (for WebSocket) and HTTP token
./bin/socket-server --jwt-secret "jwt-secret" --server-token "http-api-token"
```

### Configuration Options

- `--server-token`: HTTP API authentication token (required for API access)
- Environment variable: `HTTP_TOKEN`

The server will refuse to start if no HTTP token is provided.

## Protected Endpoints

All HTTP API endpoints now require authentication:

- `GET /api/health` - Server health status
- `GET /api/clients` - List connected clients
- `GET /api/channels` - List active channels
- `GET /api/channels/{channel}/clients` - List clients in a specific channel
- `POST /api/clients/{client}/kick` - Kick a specific client
- `POST /api/broadcast` - Broadcast message to a channel

### Unprotected Endpoints

- `GET /ws` - WebSocket connection (uses JWT authentication internally)
- `GET /` - Static file serving for admin interface

## Client Authentication

### CLI Client Usage

The CLI client now requires the HTTP token for all API operations:

```bash
# Using command line flag
./bin/socket --server-token "your-secure-token-here" health

# Using environment variable
export HTTP_TOKEN="your-secure-token-here"
./bin/socket health

# Send a message with authentication
./bin/socket --server-token "your-token" send --channel "orders" --data '{"message": "Hello"}'

# List clients with authentication
./bin/socket --server-token "your-token" list clients

# List channels with authentication
./bin/socket --server-token "your-token" list channels
```

### HTTP Request Format

When making direct HTTP requests, include the token in the Authorization header:

```bash
curl -H "Authorization: Bearer your-secure-token-here" \
     http://localhost:8080/api/health
```

### Broadcasting Messages

```bash
curl -X POST \
     -H "Authorization: Bearer your-secure-token-here" \
     -H "Content-Type: application/json" \
     -d '{"channel": "orders", "event": "update", "data": {"order_id": 123}}' \
     http://localhost:8080/api/broadcast
```

## Error Responses

### Missing Authorization Header

```json
{
  "error": "Unauthorized: Missing Authorization header"
}
```
HTTP Status: 401

### Invalid Token Format

```json
{
  "error": "Unauthorized: Invalid Authorization header format. Use 'Bearer <token>'"
}
```
HTTP Status: 401

### Invalid Token

```json
{
  "error": "Unauthorized: Invalid token"
}
```
HTTP Status: 401

## Security Best Practices

1. **Use Strong Tokens**: Generate cryptographically secure random tokens
   ```bash
   # Generate a secure token
   openssl rand -hex 32
   ```

2. **Environment Variables**: Store tokens in environment variables, not in command line history
   ```bash
   export HTTP_TOKEN="$(openssl rand -hex 32)"
   ```

3. **Rotate Tokens**: Regularly change your HTTP API tokens

4. **Secure Transport**: Always use HTTPS in production

5. **Access Control**: Limit which systems have access to the HTTP API token

## Migration from Previous Versions

If you're upgrading from a version without HTTP authentication:

1. Generate a secure token:
   ```bash
   export HTTP_TOKEN="$(openssl rand -hex 32)"
   echo "Your HTTP token: $HTTP_TOKEN"
   ```

2. Update your server startup scripts to include the token:
   ```bash
   ./bin/socket-server --server-token "$HTTP_TOKEN"
   ```

3. Update any existing API clients to include the Authorization header:
   ```bash
   # Old way (will now fail)
   curl http://localhost:8080/api/health
   
   # New way (required)
   curl -H "Authorization: Bearer $HTTP_TOKEN" http://localhost:8080/api/health
   ```

4. Update CLI usage to include the token:
   ```bash
   # Old way (will now fail)
   ./bin/socket health
   
   # New way (required)
   ./bin/socket --server-token "$HTTP_TOKEN" health
   ```

## Logging

The server logs authentication attempts:

- Successful authentications are logged at DEBUG level
- Failed authentication attempts are logged at WARN level with the client IP address

Enable debug logging to see successful authentications:
```bash
export SOCKET_DEBUG=true
./bin/socket-server --server-token "your-token"
```
