<?php

return [
    /*
    |--------------------------------------------------------------------------
    | Socket Server Configuration
    |--------------------------------------------------------------------------
    |
    | This file contains configuration options for the custom socket server
    | integration with Laravel. These settings control how Laravel communicates
    | with the standalone socket server.
    |
    */

    /*
    |--------------------------------------------------------------------------
    | Socket Binary Path
    |--------------------------------------------------------------------------
    |
    | The path to the socket binary executable. This can be an absolute path
    | or a command available in the system PATH.
    |
    */
    'binary_path' => env('SOCKET_BINARY_PATH', 'socket'),

    /*
    |--------------------------------------------------------------------------
    | Socket Server URL
    |--------------------------------------------------------------------------
    |
    | The URL where the socket server is running. This is used by the CLI
    | binary to communicate with the socket server's HTTP API.
    |
    */
    'server_url' => env('SOCKET_SERVER_URL', 'http://localhost:8080'),

    /*
    |--------------------------------------------------------------------------
    | JWT Secret
    |--------------------------------------------------------------------------
    |
    | The secret key used for JWT token generation and validation.
    | This should match the JWT_SECRET used by the socket server.
    |
    */
    'jwt_secret' => env('SOCKET_JWT_SECRET', env('APP_KEY')),

    /*
    |--------------------------------------------------------------------------
    | Queue Configuration
    |--------------------------------------------------------------------------
    |
    | The queue to use for processing socket events. Set to null to process
    | events synchronously.
    |
    */
    'queue' => env('SOCKET_QUEUE', 'default'),

    /*
    |--------------------------------------------------------------------------
    | Debug Mode
    |--------------------------------------------------------------------------
    |
    | When enabled, additional logging will be performed for socket operations.
    |
    */
    'debug' => env('SOCKET_DEBUG', false),

    /*
    |--------------------------------------------------------------------------
    | Default Channels
    |--------------------------------------------------------------------------
    |
    | Default channels that should be created when the server starts.
    |
    */
    'default_channels' => [
        'public' => [
            'require_auth' => false,
            'is_private' => false,
        ],
        'notifications' => [
            'require_auth' => true,
            'is_private' => false,
        ],
    ],

    /*
    |--------------------------------------------------------------------------
    | Authentication
    |--------------------------------------------------------------------------
    |
    | Settings for user authentication with the socket server.
    |
    */
    'auth' => [
        /*
        |--------------------------------------------------------------------------
        | User Model
        |--------------------------------------------------------------------------
        |
        | The model to use for user authentication.
        |
        */
        'user_model' => env('SOCKET_USER_MODEL', 'App\\Models\\User'),

        /*
        |--------------------------------------------------------------------------
        | JWT Token TTL
        |--------------------------------------------------------------------------
        |
        | The time-to-live for JWT tokens in minutes.
        |
        */
        'token_ttl' => env('SOCKET_TOKEN_TTL', 60),

        /*
        |--------------------------------------------------------------------------
        | Token Claims
        |--------------------------------------------------------------------------
        |
        | Additional claims to include in JWT tokens.
        |
        */
        'token_claims' => [
            'user_id' => 'id',
            'username' => 'name',
            'email' => 'email',
        ],
    ],

    /*
    |--------------------------------------------------------------------------
    | Rate Limiting
    |--------------------------------------------------------------------------
    |
    | Rate limiting configuration for socket events.
    |
    */
    'rate_limiting' => [
        'enabled' => env('SOCKET_RATE_LIMITING', true),
        'max_events_per_minute' => env('SOCKET_MAX_EVENTS_PER_MINUTE', 60),
        'max_connections_per_ip' => env('SOCKET_MAX_CONNECTIONS_PER_IP', 10),
    ],

    /*
    |--------------------------------------------------------------------------
    | CORS Configuration
    |--------------------------------------------------------------------------
    |
    | Cross-Origin Resource Sharing configuration for WebSocket connections.
    |
    */
    'cors' => [
        'allowed_origins' => explode(',', env('SOCKET_ALLOWED_ORIGINS', '*')),
        'allowed_headers' => ['Content-Type', 'Authorization'],
        'allowed_methods' => ['GET', 'POST', 'OPTIONS'],
    ],
];
