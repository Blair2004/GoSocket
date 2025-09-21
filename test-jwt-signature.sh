#!/bin/bash

# JWT Signature Issue Demonstration
# This script shows the difference between valid and invalid JWT tokens

echo "JWT Signature Issue Demonstration"
echo "================================="
echo

# Set test secret
JWT_SECRET="test-secret-123"
echo "Using JWT Secret: $JWT_SECRET"
echo

# Start server in background
echo "Starting socket server..."
./bin/socket-server --jwt-secret "$JWT_SECRET" --server-token "test-token" --port 18081 > /tmp/server.log 2>&1 &
SERVER_PID=$!

# Wait for server to start
sleep 2

echo "Server started (PID: $SERVER_PID)"
echo

# Test 1: Try with invalid JavaScript-generated token
echo "Test 1: JavaScript-generated token (WILL FAIL)"
echo "----------------------------------------------"

# This simulates what the broken JavaScript client would generate
INVALID_TOKEN="eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoidGVzdCIsImV4cCI6MTY5OTk5OTk5OX0.REVNT19TSUdOQVRVUkVfTk9UX1ZBTElE"

echo "Invalid token: $INVALID_TOKEN"
echo

# Try to authenticate with invalid token - this should fail
echo "Attempting authentication..."
timeout 5s node -e "
const WebSocket = require('ws');
const ws = new WebSocket('ws://localhost:18081');
ws.on('open', () => {
    ws.send(JSON.stringify({
        action: 'authenticate', 
        token: '$INVALID_TOKEN'
    }));
});
ws.on('message', (data) => {
    console.log('Server response:', data.toString());
    ws.close();
});
ws.on('error', (err) => {
    console.log('Connection error:', err.message);
});
" 2>/dev/null || echo "Authentication failed as expected"

echo
echo

# Test 2: Generate valid token using Go
echo "Test 2: Server-generated token (SHOULD WORK)"
echo "--------------------------------------------"

# Create a simple Go program to generate a valid token
cat > /tmp/token_gen.go << 'EOF'
package main

import (
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

func main() {
    secret := "test-secret-123"
    claims := jwt.MapClaims{
        "user_id": "test",
        "username": "testuser",
        "email": "test@example.com",
        "exp": time.Now().Add(time.Hour).Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(secret))
    if err != nil {
        panic(err)
    }
    
    fmt.Print(tokenString)
}
EOF

echo "Generating valid token with Go..."
VALID_TOKEN=$(cd /tmp && go run token_gen.go 2>/dev/null)
echo "Valid token: $VALID_TOKEN"
echo

# Try to authenticate with valid token
echo "Attempting authentication..."
timeout 5s node -e "
const WebSocket = require('ws');
const ws = new WebSocket('ws://localhost:18081');
ws.on('open', () => {
    ws.send(JSON.stringify({
        action: 'authenticate', 
        token: '$VALID_TOKEN'
    }));
});
ws.on('message', (data) => {
    const msg = JSON.parse(data.toString());
    if (msg.type === 'authenticated') {
        console.log('✅ Authentication successful!');
    } else {
        console.log('Server response:', data.toString());
    }
    ws.close();
});
ws.on('error', (err) => {
    console.log('Connection error:', err.message);
});
" 2>/dev/null || echo "Test completed"

echo
echo

# Show server logs
echo "Server logs (last 10 lines):"
echo "-----------------------------"
tail -10 /tmp/server.log

# Cleanup
kill $SERVER_PID 2>/dev/null
rm -f /tmp/token_gen.go /tmp/server.log

echo
echo "Demonstration complete!"
echo
echo "Key findings:"
echo "1. JavaScript base64 encoding != JWT HMAC-SHA256 signature"
echo "2. Always generate JWT tokens server-side with proper libraries"
echo "3. Use the same secret key in both client-side generation and server validation"