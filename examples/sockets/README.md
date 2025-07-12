# Socket Server Message Examples

This directory contains example messages that the socket server understands. Each JSON file represents a different type of message that can be sent via WebSocket connection.

## Message Types

1. **authenticate.json** - JWT authentication message
2. **join_channel.json** - Join a channel message
3. **leave_channel.json** - Leave a channel message
4. **send_message.json** - Send a message to a channel
5. **ping.json** - Ping message for connection health check

## Usage

These examples can be used with WebSocket clients to test the socket server functionality. The server expects messages in JSON format with specific fields for each action.

## Connection Flow

1. **Connect** to WebSocket endpoint: `ws://localhost:8080/ws`
2. **Authenticate** (optional but recommended): Send authenticate message
3. **Join Channel**: Send join_channel message
4. **Send Messages**: Send send_message messages
5. **Leave Channel**: Send leave_channel message (optional)

## Server Responses

The server will respond with appropriate messages for each action:
- **connected**: Welcome message upon connection
- **authenticated**: Confirmation of successful authentication
- **joined_channel**: Confirmation of joining a channel
- **left_channel**: Confirmation of leaving a channel
- **error**: Error messages for invalid requests
- **pong**: Response to ping messages
- **message**: Broadcast messages from other clients
