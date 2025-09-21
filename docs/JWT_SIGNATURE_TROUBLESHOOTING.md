# JWT Signature Validation Issues - Troubleshooting Guide

## Problem Description

If you're getting "Invalid token" or signature validation errors, this document explains the common causes and solutions.

## Root Cause Analysis

### 1. Invalid JWT Generation in JavaScript Client

The JavaScript client's `createAuthToken()` function was creating **malformed JWT tokens** with incorrect signatures:

```javascript
// WRONG - This creates invalid signatures
const signature = btoa(`${header}.${payload}.${secret}`);
```

**Why this fails:**
- JWT requires HMAC-SHA256 signature: `HMACSHA256(base64UrlEncode(header) + "." + base64UrlEncode(payload), secret)`
- Simple base64 encoding is NOT the same as HMAC-SHA256
- Browser environments don't have built-in HMAC-SHA256 without crypto libraries

### 2. Common JWT Validation Issues

#### Issue: "signature is invalid" 
**Causes:**
1. **Different secrets**: Client and server using different JWT secrets
2. **Malformed tokens**: Incorrect JWT structure or signature algorithm
3. **Base64 encoding issues**: Standard base64 vs base64url encoding differences
4. **Timing issues**: Token generated with wrong timestamp

#### Issue: "unexpected signing method"
**Cause:** Token generated with different algorithm than HS256

#### Issue: "token is expired"
**Cause:** Token `exp` claim is in the past

## Solutions

### 1. Generate JWT Tokens Server-Side (Recommended)

**Never generate JWT tokens in the browser for production use.** Always generate them server-side:

#### Laravel Example (PHP):
```php
use Firebase\JWT\JWT;

public function generateSocketToken() {
    $payload = [
        'user_id' => auth()->id(),
        'username' => auth()->user()->name, 
        'email' => auth()->user()->email,
        'iat' => time(),
        'exp' => time() + (60 * 60) // 1 hour
    ];
    
    // Use the SAME secret as your Go server
    $jwtSecret = env('JWT_SECRET'); 
    return JWT::encode($payload, $jwtSecret, 'HS256');
}
```

#### Go Server Configuration:
```bash
# Make sure your Go server uses the same secret
./bin/socket-server --jwt-secret "your-secret-key" --server-token "api-token"
```

#### JavaScript Usage:
```javascript
// Get token from your Laravel API
fetch('/api/socket-token')
    .then(response => response.json())
    .then(data => {
        const client = new LaravelSocketClient({
            url: 'ws://localhost:8080',
            token: data.token // Use server-generated token
        });
    });
```

### 2. Debug JWT Issues

#### Check Token Structure:
```javascript
// Decode JWT parts (for debugging only)
function debugJWT(token) {
    const [header, payload, signature] = token.split('.');
    
    console.log('Header:', JSON.parse(atob(header)));
    console.log('Payload:', JSON.parse(atob(payload)));
    console.log('Signature length:', signature.length);
}
```

#### Verify Server Configuration:
```bash
# Test with a known good token
./bin/socket --server-token "your-api-token" send --channel "test" --data '{"test": true}'
```

#### Check Server Logs:
The Go server logs authentication failures:
```
Client abc123 sent invalid token format
Client def456 authentication failed: signature is invalid
```

### 3. Correct JWT Generation (If You Must Do Client-Side)

If you absolutely need client-side JWT generation, use a proper library:

#### Using `jsonwebtoken` library (Node.js):
```javascript
const jwt = require('jsonwebtoken');

const token = jwt.sign({
    user_id: 'user123',
    username: 'john',
    email: 'john@example.com',
    iat: Math.floor(Date.now() / 1000),
    exp: Math.floor(Date.now() / 1000) + (60 * 60)
}, 'your-secret-key', { algorithm: 'HS256' });
```

#### Using `jose` library (Browser):
```javascript
import { SignJWT } from 'jose';

const secret = new TextEncoder().encode('your-secret-key');
const token = await new SignJWT({
    user_id: 'user123',
    username: 'john',
    email: 'john@example.com'
})
.setProtectedHeader({ alg: 'HS256' })
.setIssuedAt()
.setExpirationTime('1h')
.sign(secret);
```

## Testing Your Fix

### 1. Test Token Generation:
```bash
# Start server with known secret
./bin/socket-server --jwt-secret "test-secret" --server-token "test-token"
```

### 2. Generate Test Token (PHP):
```php
$token = JWT::encode([
    'user_id' => '123',
    'username' => 'testuser',
    'iat' => time(),
    'exp' => time() + 3600
], 'test-secret', 'HS256');

echo $token;
```

### 3. Test Authentication:
```javascript
const client = new LaravelSocketClient({
    url: 'ws://localhost:8080',
    token: 'your-generated-token'
});

client.on('authenticated', () => {
    console.log('Authentication successful!');
});

client.on('error', (error) => {
    console.log('Authentication failed:', error);
});
```

## Key Points

1. **Always use server-side JWT generation for production**
2. **Ensure client and server use the same JWT secret**
3. **Use proper JWT libraries, not manual base64 encoding**
4. **Check token expiration times**
5. **Use HMAC-SHA256 algorithm (HS256)**
6. **Use base64url encoding, not standard base64**

## Environment Variables

Make sure your secrets match:

```bash
# Laravel (.env)
JWT_SECRET=your-secret-key

# Go Server
export JWT_SECRET=your-secret-key
./bin/socket-server --jwt-secret "$JWT_SECRET"
```