<?php
// Test JWT generation - save this as test-jwt-generation.php in your Laravel project root

require_once 'vendor/autoload.php';

use Lcobucci\JWT\Encoding\ChainedFormatter;
use Lcobucci\JWT\Encoding\JoseEncoder;
use Lcobucci\JWT\Signer\Key\InMemory;
use Lcobucci\JWT\Signer\Hmac\Sha256;
use Lcobucci\JWT\Token\Builder;

// Use the same secret as your Go server
$signingKey = 'your-jwt-secret-here'; // Replace with your actual secret

$builder = Builder::new(new JoseEncoder(), ChainedFormatter::default());
$algorithm = new Sha256();
$key = InMemory::plainText($signingKey);

$token = $builder
    ->issuedBy('localhost')
    ->permittedFor('localhost')
    ->issuedAt(new DateTimeImmutable())
    ->expiresAt((new DateTimeImmutable())->add(new DateInterval('P1W')))
    ->withClaim('user_id', 123)
    ->withClaim('username', 'testuser')
    ->withClaim('email', 'test@example.com')
    ->getToken($algorithm, $key);

echo "Generated JWT Token:\n";
echo $token->toString() . "\n\n";

// Decode and display the token parts
$parts = explode('.', $token->toString());
echo "Header: " . json_encode(json_decode(base64_decode($parts[0])), JSON_PRETTY_PRINT) . "\n";
echo "Payload: " . json_encode(json_decode(base64_decode($parts[1])), JSON_PRETTY_PRINT) . "\n";
echo "Signature length: " . strlen($parts[2]) . "\n";