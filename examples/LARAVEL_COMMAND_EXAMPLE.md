# Laravel Socket Handler Command Example

This example shows how to create a custom Laravel command that handles socket events with the new `--json` payload format.

## 1. Create the Laravel Command

```bash
php artisan make:command SocketHandler --command="ns:socket-handler"
```

## 2. Example Command Implementation

```php
<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;
use App\Events\ClientMessageReceived;

class SocketHandler extends Command
{
    protected $signature = 'ns:socket-handler {--json=}';
    protected $description = 'Handle socket events from the Go server';

    public function handle()
    {
        $json = $this->option('json');
        
        if (!$json) {
            $this->error('No JSON payload provided. Use --json option.');
            return 1;
        }

        try {
            $eventData = json_decode($json, true);
            
            if (json_last_error() !== JSON_ERROR_NONE) {
                $this->error('Invalid JSON payload: ' . json_last_error_msg());
                return 1;
            }

            // Process the event
            switch ($eventData['event_type']) {
                case 'ClientMessageReceived':
                    event(new ClientMessageReceived($eventData));
                    break;
                    
                default:
                    $this->warn('Unknown event type: ' . $eventData['event_type']);
                    break;
            }

            $this->info('Socket event processed successfully');
            return 0;
            
        } catch (\Exception $e) {
            $this->error('Error processing socket event: ' . $e->getMessage());
            return 1;
        }
    }
}
```

## 3. JSON Payload Structure

The Go server will send a JSON payload with this structure:

```json
{
  "event_type": "ClientMessageReceived",
  "socket_client": {
    "id": "client-123",
    "user_id": "456",
    "username": "john",
    "remote_addr": "127.0.0.1"
  },
  "message": {
    "id": "msg-789",
    "channel": "chat.room.1",
    "event": "chat_message",
    "data": {
      "message": "Hello everyone!",
      "room_id": 1,
      "message_type": "text"
    },
    "timestamp": "2025-01-01T00:00:00Z"
  }
}
```

## 4. Start the Socket Server

```bash
# With custom command
./bin/socket-server --port 8080 --dir /var/www/laravel --command "ns:socket-handler"

# Or with environment variables
export LARAVEL_COMMAND=ns:socket-handler
export LARAVEL_PATH=/var/www/laravel
./bin/socket-server
```

## 5. Benefits of This Approach

### ✅ **Advantages:**
- **No Temporary Files**: Direct JSON payload, no file system operations
- **Faster Processing**: No file I/O overhead
- **Flexible Commands**: Use any Laravel command name
- **Better Security**: No temporary files to secure
- **Easier Debugging**: JSON payload is directly visible in logs
- **Container Friendly**: No shared file system requirements

### ✅ **Customization Options:**
- **Different Commands**: Use different commands for different message types
- **Namespace Support**: Use commands like `chat:handle`, `notifications:process`, etc.
- **Environment Specific**: Different commands for dev/staging/production
- **Team Workflows**: Each team can have their own handler commands

## 6. Advanced Usage Examples

### Multiple Event Types Handler
```php
public function handle()
{
    $eventData = json_decode($this->option('json'), true);
    
    switch ($eventData['message']['event']) {
        case 'chat_message':
            $this->handleChatMessage($eventData);
            break;
            
        case 'user_action':
            $this->handleUserAction($eventData);
            break;
            
        case 'system_request':
            $this->handleSystemRequest($eventData);
            break;
            
        default:
            $this->warn('Unknown message event: ' . $eventData['message']['event']);
    }
}
```

### Async Processing with Queues
```php
public function handle()
{
    $eventData = json_decode($this->option('json'), true);
    
    // Queue heavy processing
    ProcessSocketMessage::dispatch($eventData);
    
    $this->info('Socket event queued for processing');
}
```

### Logging and Monitoring
```php
public function handle()
{
    $eventData = json_decode($this->option('json'), true);
    
    // Log the event
    Log::info('Socket event received', [
        'client_id' => $eventData['socket_client']['id'],
        'event' => $eventData['message']['event'],
        'channel' => $eventData['message']['channel']
    ]);
    
    // Process the event
    event(new ClientMessageReceived($eventData));
    
    $this->info('Socket event processed successfully');
}
```

This new approach provides much more flexibility and better performance! 🚀
