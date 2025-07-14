<?php
namespace Ns\Listeners;

use Illuminate\Auth\Events\Validated;
use Illuminate\Support\Facades\Log;
use Lcobucci\JWT\Encoding\ChainedFormatter;
use Lcobucci\JWT\Encoding\JoseEncoder;
use Lcobucci\JWT\Signer\Key\InMemory;
use Lcobucci\JWT\Signer\Hmac\Sha256;
use Lcobucci\JWT\Token\Builder;

class ValidatedListener
{
    /**
     * Handle the event.
     *
     * @param  mixed  $event
     * @return void
     */
    public function handle( Validated $event)
    {
        $builder = Builder::new( new JoseEncoder, ChainedFormatter::default() );
        $algorithm = new Sha256();
        $key = InMemory::plainText( env( 'SOCKET_SIGNINKEY' ) );
        $token = $builder->issuedBy( env( 'APP_URL' ) )
            ->permittedFor( env( 'APP_URL' ) )
            ->expiresAt( now()->addWeek()->toDateTimeImmutable() )
            ->withClaim( 'user_id', $event->user->id )
            ->withClaim( 'username', $event->user->username )
            ->withClaim( 'email', $event->user->email )
            ->getToken( $algorithm, $key );

        $event->user->socket_jwt = $token->toString();
        $event->user->save();
    }
}