<?php
namespace Ns\Socket;

use Closure;

interface Action
{
    /**
     * Get the name of the action.
     *
     * @return string
     */
    public function handle( array $payload ): void;

    /**
     * Define the middleware for the action
     */
    public static function middlewares(): array;

    /**
     * set if the action should be automatically loaded.
     */
    public static function autoLoad(): bool;
}