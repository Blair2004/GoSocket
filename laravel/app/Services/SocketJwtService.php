<?php

namespace App\Services;

use Firebase\JWT\JWT;
use Firebase\JWT\Key;
use Illuminate\Support\Facades\Auth;

class SocketJwtService
{
    protected $secret;
    protected $algorithm = 'HS256';
    
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
            $user = Auth::user();
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
        
        return JWT::encode($payload, $this->secret, $this->algorithm);
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
        
        return JWT::encode($payload, $this->secret, $this->algorithm);
    }
    
    /**
     * Verify JWT token
     */
    public function verifyToken($token)
    {
        try {
            $decoded = JWT::decode($token, new Key($this->secret, $this->algorithm));
            return (array) $decoded;
        } catch (\Exception $e) {
            throw new \Exception('Invalid token: ' . $e->getMessage());
        }
    }
    
    /**
     * Get JWT secret for socket server
     */
    public function getSecret()
    {
        return $this->secret;
    }
}
