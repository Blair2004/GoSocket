<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\SocketTokenController;

/*
|--------------------------------------------------------------------------
| Socket Token Routes
|--------------------------------------------------------------------------
|
| These routes handle JWT token generation for the socket server.
| Add these routes to your web.php or api.php file.
|
*/

// For authenticated users
Route::middleware(['auth'])->group(function () {
    Route::get('/api/socket/token', [SocketTokenController::class, 'getUserToken']);
    Route::post('/api/socket/token', [SocketTokenController::class, 'getUserToken']);
});

// For guest users (no authentication required)
Route::post('/api/socket/guest-token', [SocketTokenController::class, 'getGuestToken']);

// For token verification
Route::post('/api/socket/verify-token', [SocketTokenController::class, 'verifyToken']);
