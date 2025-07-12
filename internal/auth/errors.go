package auth

import "errors"

var (
	// ErrInvalidToken indicates an invalid JWT token
	ErrInvalidToken = errors.New("invalid token")

	// ErrInvalidClaims indicates invalid token claims
	ErrInvalidClaims = errors.New("invalid token claims")
)
