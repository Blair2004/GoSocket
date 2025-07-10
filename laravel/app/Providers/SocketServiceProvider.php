<?php

namespace App\Providers;

use App\Listeners\SocketEventListener;
use App\Traits\InteractsWithSockets;

class SocketServiceProvider
{
    /**
     * Register the service provider.
     *
     * @return void
     */
    public function register()
    {
        // Register socket event listener for all events using InteractsWithSockets trait
        if (function_exists('app') && method_exists(app(), 'make')) {
            $events = app('events');
            
            // Listen to all events and filter in the listener
            $events->listen('*', SocketEventListener::class);
        }
    }

    /**
     * Bootstrap the service provider.
     *
     * @return void
     */
    public function boot()
    {
        // Publish configuration
        if (method_exists($this, 'publishes')) {
            $this->publishes([
                __DIR__.'/../../config/socket.php' => config_path('socket.php'),
            ], 'socket-config');
        }
    }
}
