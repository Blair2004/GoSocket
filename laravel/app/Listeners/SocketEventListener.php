<?php

namespace App\Listeners;

use App\Traits\InteractsWithSockets;

class SocketEventListener
{
    /**
     * Handle the event.
     *
     * @param  object  $event
     * @return void
     */
    public function handle($event)
    {
        // Check if the event uses the InteractsWithSockets trait
        if (!$this->usesSocketTrait($event)) {
            return;
        }

        try {
            $this->broadcastToSocket($event);
        } catch (\Exception $e) {
            error_log('Failed to broadcast event to socket server: ' . $e->getMessage());
        }
    }

    /**
     * Check if the event uses the InteractsWithSockets trait.
     *
     * @param  object  $event
     * @return bool
     */
    protected function usesSocketTrait($event)
    {
        $reflection = new \ReflectionClass($event);
        $traits = [];
        
        // Get all traits from the class hierarchy
        do {
            $traits = array_merge($traits, $reflection->getTraitNames());
        } while ($reflection = $reflection->getParentClass());
        
        return in_array(InteractsWithSockets::class, $traits);
    }

    /**
     * Broadcast the event to the socket server.
     *
     * @param  object  $event
     * @return void
     */
    protected function broadcastToSocket($event)
    {
        $channels = $event->getChannelNames();
        $eventName = $event->broadcastAs();
        $data = $event->broadcastWith();
        $options = $event->getSocketOptions();

        foreach ($channels as $channel) {
            $payload = [
                'channel' => $channel,
                'event' => $eventName,
                'data' => array_merge($data, [
                    'socket_options' => $options,
                    'timestamp' => date('c'),
                    'laravel_event' => get_class($event)
                ])
            ];

            $this->sendToSocketServer($payload);
        }
    }

    /**
     * Send payload to the socket server using the CLI binary.
     *
     * @param  array  $payload
     * @return void
     */
    protected function sendToSocketServer(array $payload)
    {
        // Create temporary file with the payload
        $tempFile = tempnam(sys_get_temp_dir(), 'socket_payload_');
        file_put_contents($tempFile, json_encode($payload, JSON_PRETTY_PRINT));

        try {
            // Get configuration values
            $socketBinaryPath = $this->getConfig('socket.binary_path', 'socket');
            $serverUrl = $this->getConfig('socket.server_url', 'http://localhost:8080');

            // Build command
            $command = sprintf(
                '%s --server %s send --file %s',
                escapeshellarg($socketBinaryPath),
                escapeshellarg($serverUrl),
                escapeshellarg($tempFile)
            );

            // Execute the command
            $output = [];
            $returnCode = 0;
            exec($command . ' 2>&1', $output, $returnCode);

            if ($returnCode !== 0) {
                error_log('Socket binary execution failed: ' . implode("\n", $output));
            } else {
                // Optional: log success in debug mode
                if ($this->getConfig('socket.debug', false)) {
                    error_log('Successfully sent event to socket server: ' . $payload['channel'] . '/' . $payload['event']);
                }
            }
        } finally {
            // Clean up temporary file
            if (file_exists($tempFile)) {
                unlink($tempFile);
            }
        }
    }

    /**
     * Get configuration value with fallback.
     *
     * @param  string  $key
     * @param  mixed  $default
     * @return mixed
     */
    protected function getConfig($key, $default = null)
    {
        if (function_exists('config')) {
            return config($key, $default);
        }
        
        // Fallback for when config() function is not available
        $configs = [
            'socket.binary_path' => env('SOCKET_BINARY_PATH', 'socket'),
            'socket.server_url' => env('SOCKET_SERVER_URL', 'http://localhost:8080'),
            'socket.debug' => env('SOCKET_DEBUG', false),
        ];
        
        return $configs[$key] ?? $default;
    }
}
