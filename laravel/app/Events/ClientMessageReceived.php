<?php

namespace App\Events;

use App\Traits\InteractsWithSockets;

class ClientMessageReceived
{
    use InteractsWithSockets;

    public $socketClient;
    public $message;
    public $originalData;

    /**
     * Create a new event instance.
     *
     * @param  array  $eventData
     * @return void
     */
    public function __construct($eventData)
    {
        $this->originalData = $eventData;
        $this->socketClient = $eventData['socket_client'] ?? [];
        $this->message = $eventData['message'] ?? [];
    }

    /**
     * Get the channels the event should broadcast on.
     *
     * @return array
     */
    public function broadcastOn()
    {
        // Broadcast back to admins/monitoring channels
        return [
            'admin.socket.events',
            'monitoring.client.messages'
        ];
    }

    /**
     * Get the event name for broadcasting.
     *
     * @return string
     */
    public function broadcastAs()
    {
        return 'client.message.received';
    }

    /**
     * Get the data to broadcast.
     *
     * @return array
     */
    public function broadcastWith()
    {
        return [
            'client' => $this->socketClient,
            'message' => $this->message,
            'processed_at' => date('c'),
        ];
    }

    /**
     * Determine if this event should broadcast.
     *
     * @return bool
     */
    public function shouldBroadcast()
    {
        // Only broadcast admin events, not regular client messages
        return false; // Set to true if you want to monitor client messages
    }
}
