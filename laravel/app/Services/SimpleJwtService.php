<?php

namespace App\Services;

class SimpleJwtService
{
    protected $secret;
    
    public function __construct()
    {
        $this->secret = config('socket.jwt_secret', env('SOCKET_JWT_SECRET', 'default-secret-key-change-in-production'));
    }
    
    /**
     * Generate JWT token for authenticated user
     */
    public function generateUserToken($user = null)
    {
        if (!$user) {
            $user = auth()->user();
        }
        
        if (!$user) {
            throw new \Exception('No authenticated user found');
        }
        
        $payload = [
            'user_id' => (string) $user->id,
            'username' => $user->name ?? $user->username ?? 'User',
            'email' => $user->email ?? '',
            'iat' => time(),
            'exp' => time() + (60 * 60 * 24), // 24 hours
        ];
        
        return $this->createJWT($payload);
    }
    
    /**
     * Generate JWT token for guest user
     */
    public function generateGuestToken($guestId = null, $guestName = null)
    {
        $payload = [
            'user_id' => $guestId ?? 'guest_' . uniqid(),
            'username' => $guestName ?? 'Guest',
            'email' => '',
            'is_guest' => true,
            'iat' => time(),
            'exp' => time() + (60 * 60 * 2), // 2 hours for guests
        ];
        
        return $this->createJWT($payload);
    }
    
    /**
     * Create JWT token using simple implementation
     */
    private function createJWT($payload)
    {
        $header = json_encode(['typ' => 'JWT', 'alg' => 'HS256']);
        $payload = json_encode($payload);
        
        $headerEncoded = $this->base64UrlEncode($header);
        $payloadEncoded = $this->base64UrlEncode($payload);
        
        $signature = hash_hmac('sha256', $headerEncoded . "." . $payloadEncoded, $this->secret, true);
        $signatureEncoded = $this->base64UrlEncode($signature);
        
        return $headerEncoded . "." . $payloadEncoded . "." . $signatureEncoded;
    }
    
    /**
     * Base64 URL encode
     */
    private function base64UrlEncode($data)
    {
        return rtrim(strtr(base64_encode($data), '+/', '-_'), '=');
    }
    
    /**
     * Base64 URL decode
     */
    private function base64UrlDecode($data)
    {
        return base64_decode(str_pad(strtr($data, '-_', '+/'), strlen($data) % 4, '=', STR_PAD_RIGHT));
    }
    
    /**
     * Verify JWT token
     */
    public function verifyToken($token)
    {
        $parts = explode('.', $token);
        if (count($parts) !== 3) {
            throw new \Exception('Invalid token format');
        }
        
        [$headerEncoded, $payloadEncoded, $signatureEncoded] = $parts;
        
        // Verify signature
        $signature = hash_hmac('sha256', $headerEncoded . "." . $payloadEncoded, $this->secret, true);
        $expectedSignature = $this->base64UrlEncode($signature);
        
        if (!hash_equals($expectedSignature, $signatureEncoded)) {
            throw new \Exception('Invalid token signature');
        }
        
        // Decode payload
        $payload = json_decode($this->base64UrlDecode($payloadEncoded), true);
        
        // Check expiration
        if (isset($payload['exp']) && $payload['exp'] < time()) {
            throw new \Exception('Token has expired');
        }
        
        return $payload;
    }
    
    /**
     * Get JWT secret for socket server
     */
    public function getSecret()
    {
        return $this->secret;
    }
}
