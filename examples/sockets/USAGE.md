# Socket Server Message Examples

This directory contains comprehensive examples of messages that can be sent to and received from the socket server.

## Client-to-Server Messages

### Authentication
- **authenticate.json** - Authenticate using JWT token

### Channel Management
- **join_channel.json** - Join a general channel
- **join_channel_orders.json** - Join an orders channel
- **join_channel_private.json** - Join a private user channel
- **leave_channel.json** - Leave a channel

### Message Sending
- **send_message.json** - Send a basic text message
- **send_message_order_update.json** - Send an order status update
- **send_message_notification.json** - Send a system notification
- **send_message_payment.json** - Send a payment received notification

### Health Check
- **ping.json** - Ping the server for connection health

## Server-to-Client Responses

The `server-responses/` directory contains examples of messages the server sends to clients:

- **connected.json** - Welcome message when client connects
- **joined_channel.json** - Confirmation of joining a channel
- **left_channel.json** - Confirmation of leaving a channel
- **error.json** - Error message for invalid requests
- **pong.json** - Response to ping messages
- **message_broadcast.json** - Broadcast message from other clients

## Usage Flow Example

1. **Connect to WebSocket**: `ws://localhost:8080/ws`
2. **Receive welcome message**: Server sends `connected` event
3. **Authenticate** (optional): Send `authenticate.json` message
4. **Join channel**: Send `join_channel.json` message
5. **Receive confirmation**: Server sends `joined_channel` event
6. **Send messages**: Send `send_message.json` messages
7. **Receive broadcasts**: Server forwards messages to all channel members
8. **Health check**: Send `ping.json`, receive `pong` response
9. **Leave channel**: Send `leave_channel.json` message

## JWT Token Example

The authentication example uses a JWT token with the following structure:

```json
{
  "user_id": "123",
  "username": "john_doe",
  "email": "john.doe@example.com",
  "exp": 1736635200
}
```

**Note**: The token in the example is for demonstration purposes only. In production, generate proper JWT tokens with your secret key.

## Channel Types

- **Public channels**: `general`, `notifications`, `announcements`
- **Business channels**: `orders`, `payments`, `inventory`
- **Private channels**: `user-{user_id}`, `admin-{admin_id}`
- **Terminal channels**: `pos-terminal-{terminal_id}`

## Common Error Messages

- `"Invalid token format"` - Authentication token is malformed
- `"Invalid token"` - JWT token is invalid or expired
- `"Invalid channel name"` - Channel name is missing or invalid
- `"Channel requires authentication"` - Trying to join auth-required channel without authentication
- `"Channel not found"` - Trying to leave a non-existent channel
