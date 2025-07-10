<?php

/**
 * Laravel Socket Server Integration Examples
 * 
 * This file shows how to use the socket server in your Laravel application
 */

// 1. Basic usage in a controller
class ChatController extends Controller
{
    public function sendMessage(Request $request)
    {
        $request->validate([
            'room_id' => 'required|integer',
            'message' => 'required|string|max:1000',
        ]);

        $user = auth()->user();
        
        // Save message to database
        $chatMessage = ChatMessage::create([
            'room_id' => $request->room_id,
            'user_id' => $user->id,
            'message' => $request->message,
        ]);

        // Broadcast to socket server
        event(new \App\Events\ChatMessage(
            $request->room_id,
            $user->id,
            $user->name,
            $request->message
        ));

        return response()->json(['status' => 'sent']);
    }
}

// 2. Usage in a job
class ProcessOrderJob implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    protected $order;

    public function __construct(Order $order)
    {
        $this->order = $order;
    }

    public function handle()
    {
        // Process the order
        $this->order->update(['status' => 'processing']);
        
        // Notify user via socket
        event(new \App\Events\OrderStatusUpdate(
            $this->order->id,
            'processing',
            $this->order->user_id,
            ['estimated_completion' => now()->addMinutes(30)]
        ));

        // ... processing logic ...

        $this->order->update(['status' => 'completed']);
        
        // Final notification
        event(new \App\Events\OrderStatusUpdate(
            $this->order->id,
            'completed',
            $this->order->user_id,
            ['completed_at' => now()]
        ));
    }
}

// 3. Usage in an observer
class UserObserver
{
    public function created(User $user)
    {
        // Send welcome notification
        event(new \App\Events\UserNotification(
            $user->id,
            'Welcome!',
            'Welcome to our platform, ' . $user->name,
            'success',
            '/dashboard'
        ));
    }

    public function updated(User $user)
    {
        if ($user->wasChanged('email')) {
            // Notify about email change
            event(new \App\Events\UserNotification(
                $user->id,
                'Email Updated',
                'Your email address has been successfully updated.',
                'info'
            ));
        }
    }
}

// 4. Usage in a command
class SendSystemAlertCommand extends Command
{
    protected $signature = 'socket:alert {level} {message} {--component=system}';
    protected $description = 'Send a system alert via socket server';

    public function handle()
    {
        $level = $this->argument('level');
        $message = $this->argument('message');
        $component = $this->option('component');

        event(new \App\Events\SystemAlert($level, $message, $component));

        $this->info("Alert sent: [$level] $message");
    }
}

// 5. Usage in middleware for real-time activity tracking
class ActivityTrackingMiddleware
{
    public function handle($request, Closure $next)
    {
        $response = $next($request);

        if (auth()->check() && $request->isMethod('POST')) {
            // Track user activity
            event(new \App\Events\LiveDataUpdate(
                'user_activity',
                [
                    'user_id' => auth()->id(),
                    'action' => $request->route()->getName(),
                    'ip' => $request->ip(),
                ],
                ['user_id' => auth()->id()]
            ));
        }

        return $response;
    }
}

// 6. Service class for centralized socket management
class SocketService
{
    public function notifyUser($userId, $title, $message, $type = 'info', $actionUrl = null)
    {
        event(new \App\Events\UserNotification($userId, $title, $message, $type, $actionUrl));
    }

    public function broadcastToRoom($roomId, $userId, $username, $message, $type = 'text')
    {
        event(new \App\Events\ChatMessage($roomId, $userId, $username, $message, $type));
    }

    public function systemAlert($level, $message, $component = 'system', $details = [])
    {
        event(new \App\Events\SystemAlert($level, $message, $component, $details));
    }

    public function updateLiveData($type, $data, $filters = [])
    {
        event(new \App\Events\LiveDataUpdate($type, $data, $filters));
    }

    public function getUserChannels($userId)
    {
        return [
            'user.' . $userId . '.notifications',
            'user.' . $userId . '.chat',
            'user.' . $userId . '.updates',
        ];
    }

    public function generateUserToken($user)
    {
        $payload = [
            'user_id' => $user->id,
            'username' => $user->name,
            'email' => $user->email,
            'iat' => time(),
            'exp' => time() + (60 * 60), // 1 hour
        ];

        $secret = config('socket.jwt_secret', config('app.key'));
        
        // Simple JWT implementation (use a proper library in production)
        $header = json_encode(['typ' => 'JWT', 'alg' => 'HS256']);
        $payload = json_encode($payload);
        
        $headerEncoded = str_replace(['+', '/', '='], ['-', '_', ''], base64_encode($header));
        $payloadEncoded = str_replace(['+', '/', '='], ['-', '_', ''], base64_encode($payload));
        
        $signature = hash_hmac('sha256', $headerEncoded . "." . $payloadEncoded, $secret, true);
        $signatureEncoded = str_replace(['+', '/', '='], ['-', '_', ''], base64_encode($signature));
        
        return $headerEncoded . "." . $payloadEncoded . "." . $signatureEncoded;
    }
}

// 7. Controller for socket management
class SocketManagementController extends Controller
{
    protected $socketService;

    public function __construct(SocketService $socketService)
    {
        $this->socketService = $socketService;
    }

    public function getUserToken()
    {
        $user = auth()->user();
        $token = $this->socketService->generateUserToken($user);
        
        return response()->json([
            'token' => $token,
            'channels' => $this->socketService->getUserChannels($user->id),
            'expires_in' => 3600,
        ]);
    }

    public function sendNotification(Request $request)
    {
        $request->validate([
            'user_id' => 'required|exists:users,id',
            'title' => 'required|string',
            'message' => 'required|string',
            'type' => 'in:info,success,warning,error',
            'action_url' => 'nullable|url',
        ]);

        $this->socketService->notifyUser(
            $request->user_id,
            $request->title,
            $request->message,
            $request->type ?? 'info',
            $request->action_url
        );

        return response()->json(['status' => 'sent']);
    }

    public function broadcastMessage(Request $request)
    {
        $request->validate([
            'channel' => 'required|string',
            'event' => 'required|string',
            'data' => 'required|array',
        ]);

        // Use CLI to send message directly
        $tempFile = tempnam(sys_get_temp_dir(), 'socket_broadcast_');
        file_put_contents($tempFile, json_encode([
            'channel' => $request->channel,
            'event' => $request->event,
            'data' => $request->data,
        ]));

        $command = config('socket.binary_path', 'socket') . ' send --file ' . escapeshellarg($tempFile);
        exec($command, $output, $returnCode);

        unlink($tempFile);

        return response()->json([
            'status' => $returnCode === 0 ? 'sent' : 'failed',
            'output' => implode("\n", $output),
        ]);
    }
}

// 8. Frontend JavaScript integration example
?>

<script>
// Frontend Socket Client for Laravel Integration
class LaravelSocketClient {
    constructor(serverUrl, token = null) {
        this.serverUrl = serverUrl;
        this.token = token;
        this.ws = null;
        this.listeners = new Map();
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
    }

    connect() {
        this.ws = new WebSocket(this.serverUrl);
        
        this.ws.onopen = () => {
            console.log('Connected to socket server');
            this.reconnectAttempts = 0;
            
            // Authenticate if token is available
            if (this.token) {
                this.authenticate(this.token);
            }
            
            this.emit('connected');
        };
        
        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };
        
        this.ws.onclose = () => {
            console.log('Disconnected from socket server');
            this.emit('disconnected');
            this.attemptReconnect();
        };
        
        this.ws.onerror = (error) => {
            console.error('Socket error:', error);
            this.emit('error', error);
        };
    }

    authenticate(token) {
        this.send({
            action: 'authenticate',
            token: token
        });
    }

    joinChannel(channel) {
        this.send({
            action: 'join_channel',
            channel: channel
        });
    }

    leaveChannel(channel) {
        this.send({
            action: 'leave_channel',
            channel: channel
        });
    }

    sendMessage(channel, event, data) {
        this.send({
            action: 'send_message',
            channel: channel,
            event: event,
            data: data
        });
    }

    send(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        }
    }

    handleMessage(message) {
        // Emit event for specific message types
        if (message.event) {
            this.emit(message.event, message);
        }
        
        // Emit channel-specific events
        if (message.channel) {
            this.emit(`channel:${message.channel}`, message);
            this.emit(`channel:${message.channel}:${message.event}`, message);
        }
        
        // Emit general message event
        this.emit('message', message);
    }

    on(event, callback) {
        if (!this.listeners.has(event)) {
            this.listeners.set(event, []);
        }
        this.listeners.get(event).push(callback);
    }

    off(event, callback) {
        if (this.listeners.has(event)) {
            const callbacks = this.listeners.get(event);
            const index = callbacks.indexOf(callback);
            if (index > -1) {
                callbacks.splice(index, 1);
            }
        }
    }

    emit(event, data = null) {
        if (this.listeners.has(event)) {
            this.listeners.get(event).forEach(callback => {
                callback(data);
            });
        }
    }

    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            setTimeout(() => {
                console.log(`Reconnection attempt ${this.reconnectAttempts}`);
                this.connect();
            }, this.reconnectDelay * this.reconnectAttempts);
        }
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
        }
    }
}

// Usage example
document.addEventListener('DOMContentLoaded', function() {
    // Get token from Laravel
    fetch('/api/socket/token')
        .then(response => response.json())
        .then(data => {
            const socket = new LaravelSocketClient('ws://localhost:8080/ws', data.token);
            
            // Connect to server
            socket.connect();
            
            // Join user-specific channels
            socket.on('authenticated', () => {
                data.channels.forEach(channel => {
                    socket.joinChannel(channel);
                });
            });
            
            // Listen for notifications
            socket.on('notification.new', (message) => {
                showNotification(message.data.title, message.data.message, message.data.type);
            });
            
            // Listen for chat messages
            socket.on('chat.message', (message) => {
                addChatMessage(message.data);
            });
            
            // Listen for live data updates
            socket.on('data.updated', (message) => {
                updateLiveData(message.data.type, message.data.data);
            });
            
            // Store socket instance globally
            window.socketClient = socket;
        });
});

function showNotification(title, message, type) {
    // Your notification display logic here
    console.log(`[${type.toUpperCase()}] ${title}: ${message}`);
}

function addChatMessage(data) {
    // Your chat message display logic here
    console.log(`${data.username}: ${data.message}`);
}

function updateLiveData(type, data) {
    // Your live data update logic here
    console.log(`Live data update [${type}]:`, data);
}
</script>

<?php
// 9. Add to routes/web.php or routes/api.php
/*
Route::middleware('auth')->group(function() {
    Route::get('/socket/token', [SocketManagementController::class, 'getUserToken']);
    Route::post('/socket/notify', [SocketManagementController::class, 'sendNotification']);
    Route::post('/socket/broadcast', [SocketManagementController::class, 'broadcastMessage']);
});
*/

// 10. Add to config/app.php providers array
/*
'providers' => [
    // ... other providers
    App\Providers\SocketServiceProvider::class,
],
*/
