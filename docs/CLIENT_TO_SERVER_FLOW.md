# Client-to-Server Message Flow

## Complete Flow: Client → Socket Server → Laravel → Back to Clients

### 1. **Client Sends Message via WebSocket**

```javascript
// Client sends a chat message
const ws = new WebSocket('ws://localhost:8080/ws');

ws.send(JSON.stringify({
    action: 'send_message',
    channel: 'chat.room.1',
    event: 'chat_message',
    data: {
        message: 'Hello everyone!',
        room_id: 1,
        message_type: 'text'
    }
}));
```

### 2. **Go Server Receives & Processes**

```go
// In main.go - handleSendMessage function
func (s *Server) handleSendMessage(client *Client, msg map[string]interface{}) {
    // Create message object
    message := Message{
        ID:        uuid.New().String(),
        Channel:   channelName,
        Event:     event,
        Data:      data,
        UserID:    client.UserID,
        Username:  client.Username,
        Timestamp: time.Now(),
    }
    
    // 1. Dispatch to Laravel for processing
    s.dispatchToLaravel(message, client)
    
    // 2. Broadcast to all clients in channel
    s.BroadcastToChannel(channelName, message)
}
```

### 3. **Laravel Processes via Artisan Command**

```bash
# Go server executes (customizable command):
php artisan ns:socket-handler --json '{"event_type":"ClientMessageReceived","socket_client":{"id":"client-123","user_id":"456","username":"john","remote_addr":"127.0.0.1"},"message":{"id":"msg-789","channel":"chat.room.1","event":"chat_message","data":{"message":"Hello everyone!","room_id":1,"message_type":"text"},"timestamp":"2025-01-01T00:00:00Z"}}'
```

```php
// App\Console\Commands\SocketHandler
class SocketHandler extends Command
{
    protected $signature = 'ns:socket-handler {--json=}';
    protected $description = 'Handle socket events from the Go server';

    public function handle()
    {
        $json = $this->option('json');
        $eventData = json_decode($json, true);
        
        // Dispatch Laravel event
        event(new ClientMessageReceived($eventData));
        
        $this->info('Socket event processed successfully');
    }
}
```

### 4. **Laravel Event Listener Processes Message**

```php
// App\Listeners\ProcessClientMessageListener
public function handle(ClientMessageReceived $event)
{
    switch ($event->message['event']) {
        case 'chat_message':
            $this->handleChatMessage($event->message, $event->socketClient);
            break;
    }
}

protected function handleChatMessage($message, $client)
{
    // Save to database
    // ChatMessage::create([...]);
    
    // Dispatch processed event (goes back to socket server)
    event(new ChatMessageProcessed(
        $data['room_id'],
        $client['user_id'],
        $client['username'],
        $data['message']
    ));
}
```

### 5. **Socket Event Listener Sends Back to Clients**

```php
// App\Listeners\SocketEventListener automatically catches ChatMessageProcessed
// and sends it back through the socket server
```

### 6. **Clients Receive Processed Message**

```javascript
ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    
    if (message.event === 'chat.message.processed') {
        // Display the processed chat message
        addMessageToChat(message.data);
    }
};
```

## Complete Flow Diagram

```
CLIENT A                    GO SERVER                   LARAVEL                    CLIENT B
   |                           |                          |                          |
   | 1. WebSocket message      |                          |                          |
   |-------------------------->|                          |                          |
   |   {action: 'send_message'}| 2. Process & validate    |                          |
   |                           |                          |                          |
   |                           | 3. dispatchToLaravel()   |                          |
   |                           |------------------------->| 4. Artisan command       |
   |                           |                          |    ProcessClientMessage  |
   |                           |                          |                          |
   |                           |                          | 5. Event: ClientMessageReceived
   |                           |                          | 6. Listener processes    |
   |                           |                          | 7. Save to DB           |
   |                           |                          | 8. Event: ChatMessageProcessed
   |                           |                          |                          |
   |                           | 9. SocketEventListener   |                          |
   |                           |<-------------------------|    (catches event)      |
   |                           |                          |                          |
   |                           | 10. Broadcast to clients |                          |
   | 11. Receive processed msg |                          |                          | 12. Receive processed msg
   |<--------------------------|                          |                          |<--------
   |                           |                          |                          |
```

## Example Client Message Types

### 1. **Chat Message**
```javascript
ws.send(JSON.stringify({
    action: 'send_message',
    channel: 'chat.room.1',
    event: 'chat_message',
    data: {
        message: 'Hello!',
        room_id: 1,
        reply_to: null
    }
}));
```

### 2. **User Action (Typing)**
```javascript
ws.send(JSON.stringify({
    action: 'send_message',
    channel: 'chat.room.1',
    event: 'user_action',
    data: {
        action: 'typing',
        is_typing: true,
        room_id: 1
    }
}));
```

### 3. **System Request**
```javascript
ws.send(JSON.stringify({
    action: 'send_message',
    channel: 'system.requests',
    event: 'system_request',
    data: {
        request: 'get_online_users',
        room_id: 1,
        request_id: 'req_123'
    }
}));
```

### 4. **Private Room Join**
```javascript
ws.send(JSON.stringify({
    action: 'send_message',
    channel: 'system.requests',
    event: 'system_request',
    data: {
        request: 'join_private_room',
        room_id: 'private_123',
        access_code: 'secret123'
    }
}));
```

## Laravel Event Dispatching Benefits

### ✅ **Advantages of This Approach:**

1. **Real-time Processing**: Client messages are processed immediately
2. **Database Integration**: Messages can be saved, validated, transformed
3. **Business Logic**: Complex processing rules in Laravel
4. **Event Broadcasting**: Processed results sent back to all clients
5. **Audit Trail**: All client interactions logged and trackable
6. **Security**: Server-side validation and authorization
7. **Scalability**: Laravel events can be queued for heavy processing

### ✅ **Use Cases:**

- **Chat Applications**: Message validation, spam filtering, user mentions
- **Live Gaming**: Move validation, game state updates
- **Collaboration Tools**: Document editing, conflict resolution
- **Real-time Trading**: Order validation, market updates
- **Customer Support**: Ticket updates, agent assignment

## Configuration

### Environment Variables
```bash
# In your .env
LARAVEL_PATH=/path/to/your/laravel/project
PHP_BINARY=/usr/bin/php8.2
LARAVEL_COMMAND=ns:socket-handler
SOCKET_PORT=8080
```

### Command-Line Options
```bash
# Start server with custom configuration
./bin/socket-server --port 8080 --dir /var/www/laravel --php /usr/bin/php8.2 --command "ns:socket-handler"

# Or use environment variables
export LARAVEL_PATH=/var/www/laravel
export PHP_BINARY=/usr/bin/php8.2
export LARAVEL_COMMAND=ns:socket-handler
./bin/socket-server
```

### Laravel Event Registration
```php
// In EventServiceProvider.php
protected $listen = [
    ClientMessageReceived::class => [
        ProcessClientMessageListener::class,
    ],
];
```

This creates a **true bidirectional communication** system where clients can send messages that trigger server-side processing and get results back in real-time! 🚀
