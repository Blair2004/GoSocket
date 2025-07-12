package models

import "errors"

var (
	// ErrNilConnection indicates a nil WebSocket connection
	ErrNilConnection = errors.New("websocket connection is nil")

	// ErrChannelNotFound indicates a channel was not found
	ErrChannelNotFound = errors.New("channel not found")

	// ErrClientNotFound indicates a client was not found
	ErrClientNotFound = errors.New("client not found")

	// ErrInvalidToken indicates an invalid JWT token
	ErrInvalidToken = errors.New("invalid token")

	// ErrUnauthorized indicates unauthorized access
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInvalidMessage indicates an invalid message format
	ErrInvalidMessage = errors.New("invalid message format")
)
