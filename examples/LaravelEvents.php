<?php

namespace App\Events;

use App\Traits\InteractsWithSockets;

/**
 * Example event for a real-time chat application
 */
class ChatMessage
{
    use InteractsWithSockets;

    public $roomId;
    public $userId;
    public $username;
    public $message;
    public $messageType;

    public function __construct($roomId, $userId, $username, $message, $messageType = 'text')
    {
        $this->roomId = $roomId;
        $this->userId = $userId;
        $this->username = $username;
        $this->message = $message;
        $this->messageType = $messageType;
    }

    public function broadcastOn()
    {
        return [
            'chat.room.' . $this->roomId,
            'user.' . $this->userId . '.chat'
        ];
    }

    public function broadcastAs()
    {
        return 'chat.message';
    }

    public function broadcastWith()
    {
        return [
            'room_id' => $this->roomId,
            'user_id' => $this->userId,
            'username' => $this->username,
            'message' => $this->message,
            'type' => $this->messageType,
            'timestamp' => now()->toISOString(),
        ];
    }
}

/**
 * Example event for user notifications
 */
class UserNotification
{
    use InteractsWithSockets;

    public $userId;
    public $title;
    public $message;
    public $type;
    public $actionUrl;

    public function __construct($userId, $title, $message, $type = 'info', $actionUrl = null)
    {
        $this->userId = $userId;
        $this->title = $title;
        $this->message = $message;
        $this->type = $type;
        $this->actionUrl = $actionUrl;
    }

    public function broadcastOn()
    {
        return [
            'user.' . $this->userId . '.notifications',
            'notifications.all'
        ];
    }

    public function broadcastAs()
    {
        return 'notification.new';
    }

    public function broadcastWith()
    {
        return [
            'user_id' => $this->userId,
            'title' => $this->title,
            'message' => $this->message,
            'type' => $this->type,
            'action_url' => $this->actionUrl,
            'id' => uniqid(),
            'read' => false,
        ];
    }
}

/**
 * Example event for live updates
 */
class LiveDataUpdate
{
    use InteractsWithSockets;

    public $dataType;
    public $data;
    public $filters;

    public function __construct($dataType, $data, $filters = [])
    {
        $this->dataType = $dataType;
        $this->data = $data;
        $this->filters = $filters;
    }

    public function broadcastOn()
    {
        $channels = ['live.updates.' . $this->dataType];
        
        // Add filtered channels if filters are provided
        foreach ($this->filters as $key => $value) {
            $channels[] = 'live.updates.' . $this->dataType . '.' . $key . '.' . $value;
        }
        
        return $channels;
    }

    public function broadcastAs()
    {
        return 'data.updated';
    }

    public function broadcastWith()
    {
        return [
            'type' => $this->dataType,
            'data' => $this->data,
            'filters' => $this->filters,
            'updated_at' => now()->toISOString(),
        ];
    }
}

/**
 * Example event for system alerts
 */
class SystemAlert
{
    use InteractsWithSockets;

    public $level;
    public $message;
    public $component;
    public $details;

    public function __construct($level, $message, $component = 'system', $details = [])
    {
        $this->level = $level;
        $this->message = $message;
        $this->component = $component;
        $this->details = $details;
    }

    public function broadcastOn()
    {
        return [
            'system.alerts',
            'system.alerts.' . $this->level,
            'system.alerts.component.' . $this->component
        ];
    }

    public function broadcastAs()
    {
        return 'system.alert';
    }

    public function broadcastWith()
    {
        return [
            'level' => $this->level,
            'message' => $this->message,
            'component' => $this->component,
            'details' => $this->details,
            'alert_id' => uniqid(),
            'created_at' => now()->toISOString(),
        ];
    }

    /**
     * This event requires authentication
     */
    protected function requiresAuthentication()
    {
        return true;
    }
}
