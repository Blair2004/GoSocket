#!/bin/bash

# Socket Server HTTP API Security Test Script
# This script demonstrates the new HTTP API authentication features

echo "=== Socket Server HTTP API Security Test ==="
echo

# Generate a secure token
HTTP_TOKEN=$(openssl rand -hex 16)
JWT_SECRET=$(openssl rand -hex 16)
echo "Generated HTTP Token: $HTTP_TOKEN"
echo "Generated JWT Secret: $JWT_SECRET"
echo

# Test 1: Try to start server without HTTP token (should fail)
echo "Test 1: Starting server without HTTP token (should fail)..."
if ./bin/socket-server --token "$JWT_SECRET" 2>/dev/null; then
    echo "❌ FAIL: Server should not start without HTTP token"
else
    echo "✅ PASS: Server correctly refuses to start without HTTP token"
fi
echo

# Test 2: Try CLI commands without token (should fail)
echo "Test 2: CLI commands without token (should fail)..."
if ./bin/socket health 2>/dev/null; then
    echo "❌ FAIL: CLI should require token"
else
    echo "✅ PASS: CLI correctly requires token"
fi
echo

# Test 3: Start server with proper tokens (in background)
echo "Test 3: Starting server with proper authentication..."
./bin/socket-server --token "$JWT_SECRET" --http-token "$HTTP_TOKEN" --port 18080 > server.log 2>&1 &
SERVER_PID=$!
sleep 2

# Check if server started successfully
if ps -p $SERVER_PID > /dev/null; then
    echo "✅ PASS: Server started successfully with authentication"
    
    # Test 4: Try API call without token (should fail)
    echo
    echo "Test 4: HTTP API call without token (should fail)..."
    if curl -s http://localhost:18080/api/health > /dev/null 2>&1; then
        echo "❌ FAIL: API should require authentication"
    else
        echo "✅ PASS: API correctly requires authentication"
    fi
    
    # Test 5: Try API call with wrong token (should fail)
    echo
    echo "Test 5: HTTP API call with wrong token (should fail)..."
    if curl -s -H "Authorization: Bearer wrong-token" http://localhost:18080/api/health > /dev/null 2>&1; then
        echo "❌ FAIL: API should reject wrong token"
    else
        echo "✅ PASS: API correctly rejects wrong token"
    fi
    
    # Test 6: Try API call with correct token (should succeed)
    echo
    echo "Test 6: HTTP API call with correct token (should succeed)..."
    if curl -s -H "Authorization: Bearer $HTTP_TOKEN" http://localhost:18080/api/health | grep -q "healthy"; then
        echo "✅ PASS: API correctly accepts valid token"
    else
        echo "❌ FAIL: API should accept valid token"
    fi
    
    # Test 7: Try CLI with correct token (should succeed)
    echo
    echo "Test 7: CLI with correct token (should succeed)..."
    if ./bin/socket --server http://localhost:18080 --token "$HTTP_TOKEN" health | grep -q "Server Status"; then
        echo "✅ PASS: CLI works with valid token"
    else
        echo "❌ FAIL: CLI should work with valid token"
    fi
    
    # Clean up
    echo
    echo "Cleaning up..."
    kill $SERVER_PID 2>/dev/null
    wait $SERVER_PID 2>/dev/null
    rm -f server.log
    echo "✅ Test completed"
else
    echo "❌ FAIL: Server failed to start"
fi

echo
echo "=== Test Summary ==="
echo "The socket server now includes HTTP API authentication that:"
echo "1. Requires an HTTP token to start the server"
echo "2. Protects all /api/* endpoints with Bearer token authentication"
echo "3. Provides secure CLI access with token validation"
echo "4. Logs authentication attempts for security monitoring"
echo
echo "Use: ./bin/socket-server --http-token 'your-token' to start the server"
echo "Use: ./bin/socket --token 'your-token' to use the CLI"
echo "Or set HTTP_TOKEN environment variable for both"
