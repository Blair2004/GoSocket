#!/bin/bash

# Script to run socket-server with configuration from .env file
# Usage: ./run-socket-server.sh

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Load environment variables from .env file
if [ -f "$SCRIPT_DIR/.env" ]; then
    echo "Loading configuration from .env file..."
    export $(grep -v '^#' "$SCRIPT_DIR/.env" | xargs)
else
    echo "Warning: .env file not found at $SCRIPT_DIR/.env"
    exit 1
fi

# Build the command with arguments from environment variables
CMD="$SCRIPT_DIR/bin/socket-server"

# Add JWT secret (required for authentication)
if [ -n "$SOCKET_JWT_SECRET" ]; then
    CMD="$CMD --jwt-secret $SOCKET_JWT_SECRET"
else
    echo "Error: SOCKET_JWT_SECRET is not set in .env"
    exit 1
fi

# Add server token (required for API access)
if [ -n "$SOCKET_SERVER_TOKEN" ]; then
    CMD="$CMD --server-token $SOCKET_SERVER_TOKEN"
else
    echo "Error: SOCKET_SERVER_TOKEN is not set in .env"
    exit 1
fi

# Add working directory for Laravel commands
if [ -n "$SOCKET_DIR" ]; then
    CMD="$CMD --dir $SOCKET_DIR"
fi

# Add optional port (defaults to 8080 if not set)
if [ -n "$SOCKET_PORT" ]; then
    CMD="$CMD --port $SOCKET_PORT"
fi

# Add optional Laravel command (defaults to 'socket:handle' if not set)
if [ -n "$LARAVEL_COMMAND" ]; then
    CMD="$CMD --command $LARAVEL_COMMAND"
fi

# Add optional PHP binary path (defaults to 'php' if not set)
if [ -n "$PHP_BINARY" ]; then
    CMD="$CMD --php $PHP_BINARY"
fi

# Add optional temp directory
if [ -n "$SOCKET_TEMP_DIR" ]; then
    CMD="$CMD --temp $SOCKET_TEMP_DIR"
fi

# Add optional debug flag
if [ "$DEBUG" = "true" ]; then
    CMD="$CMD --debug"
fi

# Add optional performance logs flag
if [ "$PERFORMANCE_LOGS" = "true" ]; then
    CMD="$CMD --performance-logs"
fi

# Display the command being executed
echo "Starting socket-server..."
echo "Command: $CMD"
echo ""

# Execute the command
exec $CMD
