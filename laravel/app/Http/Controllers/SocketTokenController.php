<?php

namespace App\Http\Controllers;

use App\Services\SocketJwtService;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;

class SocketTokenController extends Controller
{
    protected $jwtService;
    
    public function __construct(SocketJwtService $jwtService)
    {
        $this->jwtService = $jwtService;
    }
    
    /**
     * Get JWT token for authenticated user
     */
    public function getUserToken(Request $request)
    {
        try {
            $user = Auth::user();
            if (!$user) {
                return response()->json(['error' => 'Not authenticated'], 401);
            }
            
            $token = $this->jwtService->generateUserToken($user);
            
            return response()->json([
                'token' => $token,
                'user' => [
                    'id' => $user->id,
                    'name' => $user->name ?? $user->username ?? 'User',
                    'email' => $user->email ?? '',
                ],
                'expires_in' => 60 * 60 * 24, // 24 hours
            ]);
        } catch (\Exception $e) {
            return response()->json(['error' => $e->getMessage()], 500);
        }
    }
    
    /**
     * Get JWT token for guest user
     */
    public function getGuestToken(Request $request)
    {
        try {
            $guestId = $request->input('guest_id', 'guest_' . uniqid());
            $guestName = $request->input('guest_name', 'Guest');
            
            $token = $this->jwtService->generateGuestToken($guestId, $guestName);
            
            return response()->json([
                'token' => $token,
                'user' => [
                    'id' => $guestId,
                    'name' => $guestName,
                    'email' => '',
                    'is_guest' => true,
                ],
                'expires_in' => 60 * 60 * 2, // 2 hours
            ]);
        } catch (\Exception $e) {
            return response()->json(['error' => $e->getMessage()], 500);
        }
    }
    
    /**
     * Verify JWT token
     */
    public function verifyToken(Request $request)
    {
        try {
            $token = $request->input('token');
            if (!$token) {
                return response()->json(['error' => 'Token is required'], 400);
            }
            
            $payload = $this->jwtService->verifyToken($token);
            
            return response()->json([
                'valid' => true,
                'payload' => $payload,
            ]);
        } catch (\Exception $e) {
            return response()->json([
                'valid' => false,
                'error' => $e->getMessage(),
            ], 401);
        }
    }
}
