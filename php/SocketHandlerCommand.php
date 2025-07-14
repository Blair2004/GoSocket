<?php
namespace Ns\Console\Commands;

use Illuminate\Console\Command;
use Ns\Services\ModulesService;
use Ns\Socket\Authenticator;
use Illuminate\Pipeline\Pipeline;

class SocketHandlerCommand extends Command
{
    public $signature = 'ns:socket-handler {--payload=}';

    public $description = 'Handle socket connections for NexoPOS Core.';

    public function handle()
    {
        $payloadPath = $this->option( 'payload' );

        /**
         * Payload Structure:
         * {
         *      "type": "connection" | "message" | "disconnect",
         *      "data": {
         *          "token" : int,
         *          "user_id": 123,
         *          "handler": "someHandler",
         *          "message": "Hello, World!"
         *      }
         * }
         */

        if ( file_exists( $payloadPath ) ) {
            $payload = json_decode( file_get_contents( $payloadPath ), true );

            if ( ! is_array( $payload ) || ! isset( $payload['action'], $payload['data'], $payload[ 'auth' ] ) ) {
                return $this->error( 'Invalid JSON payload.' );
            }

            /**
             * @var ModulesService
             */
            $moduleService   =   app()->make( ModulesService::class );
            $modules    =   $moduleService->getEnabledAndAutoloadedModules();

            $handlers    =   collect( $modules )->map( fn( $module ) => $module[ 'socket-handlers' ] )->flatten();

            $action = $handlers
                ->map( function( $filePath ) {
                    $className = str_replace( '/', '\\', $filePath );
                    $className = 'Modules\\' . $className;
                    $className = substr( $className, 0, -4 ); // Remove .php extension

                    return $className;
                })->filter( function( $className ) use ( $payload ) {

                    if ( $payload[ 'action' ] === $className ) {
                        return true;
                    }

                    $object = new $className;

                    /*
                     * We believe every socket action should use the trait `Ns\Traits\SocketAction`
                     */
                    if ( method_exists( $object, 'getName' ) && is_callable( [ $object, 'getName' ] ) ) {
                        return $payload[ 'action' ] === $object->getName();
                    } else {
                        return $payload[ 'action' ] === $object::class;
                    }
                }
            );

            if ( $action->isEmpty() ) {
                return $this->error( sprintf( 'No action found for the provided action: %s', $payload[ 'action' ] ) );
            }

            /**
             * Let's load the middleware that are defined on the socket configuration
             * file of the package.
             */
            $middlewares = config( 'socket.middlewares', [] );

            try {
                // Process payload through middleware pipeline
                $processedPayload = app(Pipeline::class)
                    ->send($payload)
                    ->through($this->resolveMiddlewares($middlewares))
                    ->then(function ($payload) {
                        return $payload;
                    });

                /**
                 * let's initialize the Socket Action handler.
                 */
                $object = new ($action->first());
                
                // Execute the action with the processed payload
                if (method_exists($object, 'handle')) {
                    $object->handle($processedPayload);
                } else {
                    return $this->error('Action handler does not have a handle method.');
                }

                return $this->info( 'Payload processed successfully.' );
            } catch (\Exception $e) {
                return $this->error( 'Middleware processing failed: ' . $e->getMessage() );
            }
        } else {
            return $this->error( 'No JSON payload provided.' );
        }
    }

    /**
     * Resolve middleware instances from configuration
     */
    protected function resolveMiddlewares(array $middlewares): array
    {
        return collect($middlewares)->map(function ($middleware) {
            // If it's already an instance or closure, return as is
            if (is_object($middleware) || is_callable($middleware)) {
                return $middleware;
            }
            
            // If it's an array with class name and parameters
            if (is_array($middleware) && isset($middleware['class'])) {
                $class = $middleware['class'];
                $parameters = $middleware['parameters'] ?? [];
                
                // Create instance with parameters
                return new $class(...$parameters);
            }
            
            // If it's an array with 'middleware' and 'parameters' keys (Laravel style)
            if (is_array($middleware) && isset($middleware['middleware'])) {
                $class = $middleware['middleware'];
                $parameters = $middleware['parameters'] ?? [];
                
                return new $class(...$parameters);
            }

            // If it's a string with parameters (like 'MiddlewareClass:param1=foo,param2=bar')
            if (is_string($middleware) && strpos($middleware, ':') !== false) {
                [$class, $paramString] = explode(':', $middleware, 2);
                $parameters = [];
                foreach (explode(',', $paramString) as $param) {
                    [$key, $value] = explode('=', $param);
                    $parameters[$key] = $value;
                }

                return new $class(...$parameters);
            }
            
            return $middleware;
        })->toArray();
    }
}