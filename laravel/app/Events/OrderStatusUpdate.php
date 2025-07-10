<?php

namespace App\Events;

use App\Traits\InteractsWithSockets;

class OrderStatusUpdate
{
    use InteractsWithSockets;

    public $orderId;
    public $status;
    public $customerId;
    public $details;

    /**
     * Create a new event instance.
     *
     * @param  int  $orderId
     * @param  string  $status
     * @param  int  $customerId
     * @param  array  $details
     * @return void
     */
    public function __construct($orderId, $status, $customerId, $details = [])
    {
        $this->orderId = $orderId;
        $this->status = $status;
        $this->customerId = $customerId;
        $this->details = $details;
    }

    /**
     * Get the channels the event should broadcast on.
     *
     * @return array
     */
    public function broadcastOn()
    {
        return [
            'orders',
            'customer.' . $this->customerId,
        ];
    }

    /**
     * Get the event name for broadcasting.
     *
     * @return string
     */
    public function broadcastAs()
    {
        return 'order.status.updated';
    }

    /**
     * Get the data to broadcast.
     *
     * @return array
     */
    public function broadcastWith()
    {
        return [
            'order_id' => $this->orderId,
            'status' => $this->status,
            'customer_id' => $this->customerId,
            'details' => $this->details,
            'updated_at' => date('c'),
        ];
    }
}
