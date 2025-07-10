#!/bin/bash

# Build script for the socket server

set -e

echo "Building Socket Server..."

# Create bin directory if it doesn't exist
mkdir -p bin

# Build the main server binary
echo "Building server binary..."
go build -o bin/socket-server main.go

# Build the CLI binary
echo "Building CLI binary..."
go build -o bin/socket cmd/cli/main.go

# Make binaries executable
chmod +x bin/socket-server
chmod +x bin/socket

echo "Build completed successfully!"
echo ""
echo "Binaries created:"
echo "  - bin/socket-server (main server)"
echo "  - bin/socket (CLI client)"
echo ""
echo "To start the server:"
echo "  ./bin/socket-server"
echo ""
echo "To use the CLI:"
echo "  ./bin/socket --help"
