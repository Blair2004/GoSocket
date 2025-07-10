<?php

namespace App\Listeners;

use App\Events\ClientMessageReceived;
use App\Events\ChatMessageProcessed;
use App\Events\UserNotification;

class ProcessClientMessageListener
{
    /**
     * Handle the event.
     *
     * @param  \App\Events\ClientMessageReceived  $event
     * @return void
     */
    public function handle(ClientMessageReceived $event)
    {
        $message = $event->message;
        $client = $event->socketClient;
        
        // Process different types of client messages
        switch ($message['event']) {
            case 'chat_message':
                $this->handleChatMessage($message, $client);
                break;
                
            case 'user_action':
                $this->handleUserAction($message, $client);
                break;
                
            case 'system_request':
                $this->handleSystemRequest($message, $client);
                break;
                
            default:
                error_log("Unknown client message event: " . $message['event']);
        }
    }
    
    /**
     * Handle chat messages from clients
     */
    protected function handleChatMessage($message, $client)
    {
        $data = $message['data'];
        
        // Validate and save chat message
        if (isset($data['message']) && isset($data['room_id'])) {
            // Here you would typically save to database
            // ChatMessage::create([...]);
            
            // Dispatch processed chat event (will be sent back via socket)
            $processedEvent = new ChatMessageProcessed(
                $data['room_id'],
                $client['user_id'] ?? 'anonymous',
                $client['username'] ?? 'Anonymous',
                $data['message']
            );
            
            // This will be caught by SocketEventListener and sent back to clients
            event($processedEvent);
        }
    }
    
    /**
     * Handle user actions from clients
     */
    protected function handleUserAction($message, $client)
    {
        $data = $message['data'];
        
        switch ($data['action'] ?? '') {
            case 'typing':
                // Broadcast typing indicator
                $this->broadcastTypingStatus($data, $client);
                break;
                
            case 'seen':
                // Mark messages as seen
                $this->markMessagesAsSeen($data, $client);
                break;
                
            case 'presence':
                // Update user presence
                $this->updateUserPresence($data, $client);
                break;
        }
    }
    
    /**
     * Handle system requests from clients
     */
    protected function handleSystemRequest($message, $client)
    {
        $data = $message['data'];
        
        switch ($data['request'] ?? '') {
            case 'get_online_users':
                // Send online users list back to client
                $this->sendOnlineUsers($data, $client);
                break;
                
            case 'join_private_room':
                // Handle private room join request
                $this->handlePrivateRoomJoin($data, $client);
                break;
        }
    }
    
    /**
     * Broadcast typing status
     */
    protected function broadcastTypingStatus($data, $client)
    {
        // Create typing event that will be sent back via socket
        $typingEvent = new \App\Events\LiveDataUpdate(
            'typing_status',
            [
                'user_id' => $client['user_id'],
                'username' => $client['username'],
                'is_typing' => $data['is_typing'] ?? true,
                'room_id' => $data['room_id'] ?? null,
            ],
            ['room_id' => $data['room_id'] ?? null]
        );
        
        event($typingEvent);
    }
    
    /**
     * Mark messages as seen
     */
    protected function markMessagesAsSeen($data, $client)
    {
        // Update database with seen status
        // Then notify other users
        
        $seenEvent = new \App\Events\LiveDataUpdate(
            'message_seen',
            [
                'user_id' => $client['user_id'],
                'message_ids' => $data['message_ids'] ?? [],
                'seen_at' => date('c'),
            ]
        );
        
        event($seenEvent);
    }
    
    /**
     * Update user presence
     */
    protected function updateUserPresence($data, $client)
    {
        // Update user presence in database/cache
        
        $presenceEvent = new \App\Events\LiveDataUpdate(
            'user_presence',
            [
                'user_id' => $client['user_id'],
                'status' => $data['status'] ?? 'online',
                'last_seen' => date('c'),
            ]
        );
        
        event($presenceEvent);
    }
    
    /**
     * Send online users list back to requesting client
     */
    protected function sendOnlineUsers($data, $client)
    {
        // Get online users (from cache/database)
        $onlineUsers = []; // Implement your logic here
        
        // Send notification back to requesting client only
        $responseEvent = new UserNotification(
            $client['user_id'],
            'Online Users',
            'Online users list',
            'system_response',
            null
        );
        
        // Add online users data
        $responseEvent->data = [
            'response_type' => 'online_users',
            'users' => $onlineUsers,
            'request_id' => $data['request_id'] ?? null,
        ];
        
        event($responseEvent);
    }
    
    /**
     * Handle private room join request
     */
    protected function handlePrivateRoomJoin($data, $client)
    {
        $roomId = $data['room_id'] ?? null;
        $userId = $client['user_id'] ?? null;
        
        if ($roomId && $userId) {
            // Check if user has permission to join private room
            $canJoin = true; // Implement your authorization logic
            
            if ($canJoin) {
                // Grant access (this could update database)
                
                // Notify client of successful join
                $joinEvent = new UserNotification(
                    $userId,
                    'Room Access Granted',
                    "You've been granted access to room {$roomId}",
                    'success'
                );
                
                $joinEvent->data = [
                    'response_type' => 'room_access',
                    'room_id' => $roomId,
                    'access_granted' => true,
                ];
                
                event($joinEvent);
            } else {
                // Deny access
                $denyEvent = new UserNotification(
                    $userId,
                    'Room Access Denied',
                    "Access to room {$roomId} was denied",
                    'error'
                );
                
                event($denyEvent);
            }
        }
    }
}
