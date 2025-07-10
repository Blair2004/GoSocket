<?php

namespace App\Events;

use App\Traits\InteractsWithSockets;
use Illuminate\Broadcasting\Channel;
use Illuminate\Broadcasting\PrivateChannel;

class UserNotification
{
    use InteractsWithSockets;

    public $user;
    public $message;
    public $type;
    public $data;

    /**
     * Create a new event instance.
     *
     * @param  mixed  $user
     * @param  string  $message
     * @param  string  $type
     * @param  array  $data
     * @return void
     */
    public function __construct($user, $message, $type = 'info', $data = [])
    {
        $this->user = $user;
        $this->message = $message;
        $this->type = $type;
        $this->data = $data;
    }

    /**
     * Get the channels the event should broadcast on.
     *
     * @return \Illuminate\Broadcasting\Channel|array
     */
    public function broadcastOn()
    {
        return new PrivateChannel('user.' . $this->user->id);
    }

    /**
     * Get the data to broadcast.
     *
     * @return array
     */
    public function broadcastWith()
    {
        return [
            'user_id' => $this->user->id,
            'message' => $this->message,
            'type' => $this->type,
            'data' => $this->data,
        ];
    }
}
