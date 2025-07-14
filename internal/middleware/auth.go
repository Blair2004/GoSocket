package middleware

import (
	"net/http"
	"strings"

	"socket-server/pkg/logger"
)

// HTTPAuth provides HTTP API authentication middleware
type HTTPAuth struct {
	token  string
	logger *logger.Logger
}

// NewHTTPAuth creates a new HTTP authentication middleware
func NewHTTPAuth(token string, logger *logger.Logger) *HTTPAuth {
	return &HTTPAuth{
		token:  token,
		logger: logger,
	}
}

// Authenticate is a middleware that validates HTTP API token
func (a *HTTPAuth) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			a.logger.Warn("HTTP API request without Authorization header from %s", r.RemoteAddr)
			http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Check for Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			a.logger.Warn("HTTP API request with invalid Authorization header format from %s", r.RemoteAddr)
			http.Error(w, "Unauthorized: Invalid Authorization header format. Use 'Bearer <token>'", http.StatusUnauthorized)
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != a.token {
			a.logger.Warn("HTTP API request with invalid token from %s", r.RemoteAddr)
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Token is valid, proceed to the next handler
		a.logger.Debug("HTTP API request authenticated successfully from %s", r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// AuthenticateFunc is a middleware function that validates HTTP API token
func (a *HTTPAuth) AuthenticateFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			a.logger.Warn("HTTP API request without Authorization header from %s", r.RemoteAddr)
			http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Check for Bearer token format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			a.logger.Warn("HTTP API request with invalid Authorization header format from %s", r.RemoteAddr)
			http.Error(w, "Unauthorized: Invalid Authorization header format. Use 'Bearer <token>'", http.StatusUnauthorized)
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token != a.token {
			a.logger.Warn("HTTP API request with invalid token from %s", r.RemoteAddr)
			http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
			return
		}

		// Token is valid, proceed to the next handler
		a.logger.Debug("HTTP API request authenticated successfully from %s", r.RemoteAddr)
		next(w, r)
	}
}
