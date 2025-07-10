<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;
use App\Events\ClientMessageReceived;

class ProcessClientMessage extends Command
{
    /**
     * The name and signature of the console command.
     *
     * @var string
     */
    protected $signature = 'socket:process-client-message {file}';

    /**
     * The console command description.
     *
     * @var string
     */
    protected $description = 'Process a client message received from the socket server';

    /**
     * Execute the console command.
     *
     * @return int
     */
    public function handle()
    {
        $filePath = $this->argument('file');
        
        if (!file_exists($filePath)) {
            $this->error("File not found: {$filePath}");
            return 1;
        }
        
        try {
            $eventData = json_decode(file_get_contents($filePath), true);
            
            if (!$eventData) {
                $this->error("Invalid JSON in file: {$filePath}");
                return 1;
            }
            
            // Dispatch Laravel event
            event(new ClientMessageReceived($eventData));
            
            $this->info("Client message processed successfully");
            return 0;
            
        } catch (\Exception $e) {
            $this->error("Error processing client message: " . $e->getMessage());
            return 1;
        }
    }
}
