<?php

namespace App\Traits;

trait InteractsWithSockets
{
    /**
     * Get the channels the event should broadcast on.
     *
     * @return \Illuminate\Broadcasting\Channel|array
     */
    abstract public function broadcastOn();

    /**
     * Get the event name for broadcasting.
     *
     * @return string
     */
    public function broadcastAs()
    {
        $className = get_class($this);
        return substr($className, strrpos($className, '\\') + 1);
    }

    /**
     * Get the data to broadcast.
     *
     * @return array
     */
    public function broadcastWith()
    {
        $data = [];
        
        // Get all public properties
        $reflection = new \ReflectionClass($this);
        $properties = $reflection->getProperties(\ReflectionProperty::IS_PUBLIC);
        
        foreach ($properties as $property) {
            $data[$property->getName()] = $property->getValue($this);
        }
        
        return $data;
    }

    /**
     * Determine if this event should broadcast.
     *
     * @return bool
     */
    public function shouldBroadcast()
    {
        return true;
    }

    /**
     * Get the queue on which to broadcast the event.
     *
     * @return string|null
     */
    public function broadcastQueue()
    {
        if (function_exists('config')) {
            return config('socket.queue', 'default');
        }
        return 'default';
    }

    /**
     * Get additional socket options.
     *
     * @return array
     */
    public function getSocketOptions()
    {
        return [
            'require_auth' => $this->requiresAuthentication(),
            'is_private' => $this->isPrivate(),
        ];
    }

    /**
     * Check if the event requires authentication.
     *
     * @return bool
     */
    protected function requiresAuthentication()
    {
        $channels = $this->broadcastOn();
        if (!is_array($channels)) {
            $channels = [$channels];
        }

        foreach ($channels as $channel) {
            $channelType = get_class($channel);
            if (strpos($channelType, 'PrivateChannel') !== false || 
                strpos($channelType, 'PresenceChannel') !== false) {
                return true;
            }
        }

        return false;
    }

    /**
     * Check if the event uses private channels.
     *
     * @return bool
     */
    protected function isPrivate()
    {
        $channels = $this->broadcastOn();
        if (!is_array($channels)) {
            $channels = [$channels];
        }

        foreach ($channels as $channel) {
            $channelType = get_class($channel);
            if (strpos($channelType, 'PrivateChannel') !== false) {
                return true;
            }
        }

        return false;
    }

    /**
     * Get the channel names as strings.
     *
     * @return array
     */
    public function getChannelNames()
    {
        $channels = $this->broadcastOn();
        if (!is_array($channels)) {
            $channels = [$channels];
        }

        $names = [];
        foreach ($channels as $channel) {
            if (is_object($channel) && property_exists($channel, 'name')) {
                $names[] = $channel->name;
            } else {
                $names[] = (string) $channel;
            }
        }

        return $names;
    }
}
