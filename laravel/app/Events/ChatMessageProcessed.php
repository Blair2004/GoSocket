<?php

namespace App\Events;

use App\Traits\InteractsWithSockets;

class ChatMessageProcessed
{
    use InteractsWithSockets;

    public $roomId;
    public $userId;
    public $username;
    public $message;
    public $processedAt;

    /**
     * Create a new event instance.
     *
     * @param  string  $roomId
     * @param  string  $userId
     * @param  string  $username
     * @param  string  $message
     * @return void
     */
    public function __construct($roomId, $userId, $username, $message)
    {
        $this->roomId = $roomId;
        $this->userId = $userId;
        $this->username = $username;
        $this->message = $message;
        $this->processedAt = date('c');
    }

    /**
     * Get the channels the event should broadcast on.
     *
     * @return array
     */
    public function broadcastOn()
    {
        return [
            'chat.room.' . $this->roomId,
            'user.' . $this->userId . '.chat'
        ];
    }

    /**
     * Get the event name for broadcasting.
     *
     * @return string
     */
    public function broadcastAs()
    {
        return 'chat.message.processed';
    }

    /**
     * Get the data to broadcast.
     *
     * @return array
     */
    public function broadcastWith()
    {
        return [
            'room_id' => $this->roomId,
            'user_id' => $this->userId,
            'username' => $this->username,
            'message' => $this->message,
            'processed_at' => $this->processedAt,
            'message_id' => uniqid('msg_'),
        ];
    }
}
