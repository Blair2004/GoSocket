<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;
use App\Events\ClientMessageReceived;

class SocketHandler extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'ns:socket-handler {--payload= : Path to JSON payload file}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'Handle socket events from the Go server via payload file';

    /**
     * Execute the console command.
     */
    public function handle()
    {
        $payloadFile = $this->option('payload');
        
        if (!$payloadFile) {
            $this->error('No payload file provided. Use --payload option.');
            return 1;
        }
        
        if (!file_exists($payloadFile)) {
            $this->error("Payload file not found: {$payloadFile}");
            return 1;
        }

        try {
            $json = file_get_contents($payloadFile);
            $message = json_decode($json, true);
            
            if (json_last_error() !== JSON_ERROR_NONE) {
                $this->error('Invalid JSON in payload file: ' . json_last_error_msg());
                return 1;
            }

            // Handle different actions using the standardized format
            $action = $message['action'] ?? 'unknown';
            
            switch ($action) {
                case 'ClientMessageReceived':
                    $result = $this->handleClientMessage($message);
                    break;
                    
                case 'client_authentication':
                    $result = $this->handleClientAuthentication($message);
                    break;
                    
                default:
                    $this->line('Unknown action: ' . $action);
                    $result = 0;
            }
            
            // Optional: Clean up the payload file immediately after processing
            if (config('socket.cleanup_payloads_immediately', true)) {
                @unlink($payloadFile);
            }
            
            $this->info('Socket message processed successfully');
            return $result;

        } catch (\Exception $e) {
            $this->error('Error processing socket message: ' . $e->getMessage());
            return 1;
        }
    }

    /**
     * Handle client message event
     */
    protected function handleClientMessage($message)
    {
        // Validate required fields in the standardized format
        if (!isset($message['data']) || !isset($message['client'])) {
            $this->error('Missing required fields in message payload');
            return 1;
        }

        // Extract message information from standardized format
        $messageId = $message['message_id'] ?? 'unknown';
        $action = $message['action'] ?? 'unknown';
        $clientInfo = $message['client'] ?? [];
        $authInfo = $message['auth'] ?? [];
        $data = $message['data'] ?? [];

        $this->info("Processing message: {$messageId} - Action: {$action}");

        // Dispatch Laravel event with standardized format
        event(new ClientMessageReceived($message));

        $this->info('Socket message processed successfully');
        return 0;
    }

    /**
     * Handle client authentication event
     */
    protected function handleClientAuthentication($message)
    {
        // Validate required fields in the standardized format
        if (!isset($message['client']) || !isset($message['data'])) {
            $this->error('Missing required fields in authentication payload');
            return 1;
        }

        // Extract authentication information from standardized format
        $messageId = $message['message_id'] ?? 'unknown';
        $clientInfo = $message['client'] ?? [];
        $authInfo = $message['auth'] ?? [];
        $data = $message['data'] ?? [];
        
        $authStatus = $data['authentication_status'] ?? 'unknown';
        $clientId = $clientInfo['id'] ?? 'unknown';

        $this->info("Authentication event: {$messageId} - Client: {$clientId} - Status: {$authStatus}");
        
        // You can add your authentication logic here:
        // - Log successful/failed logins
        // - Update user's last_seen timestamp
        // - Send notifications
        // - Rate limiting for failed attempts
        
        return 0;
    }
}
