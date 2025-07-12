package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNew(t *testing.T) {
	secret := "test-secret"
	authService := New(secret)

	if string(authService.jwtSecret) != secret {
		t.Errorf("Expected secret %s, got %s", secret, string(authService.jwtSecret))
	}
}

func TestGenerateToken(t *testing.T) {
	secret := "test-secret"
	authService := New(secret)

	userID := "user-123"
	channel := "test-channel"

	token, err := authService.GenerateToken(userID, channel)
	if err != nil {
		t.Errorf("Unexpected error generating token: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Verify the token can be parsed and contains correct claims
	claims, err := authService.ValidateToken(token)
	if err != nil {
		t.Errorf("Error validating generated token: %v", err)
	}

	if claims["user_id"] != userID {
		t.Errorf("Expected user_id %s, got %v", userID, claims["user_id"])
	}

	if claims["channel"] != channel {
		t.Errorf("Expected channel %s, got %v", channel, claims["channel"])
	}
}

func TestValidateToken(t *testing.T) {
	secret := "test-secret"
	authService := New(secret)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "Valid token",
			token:       createValidToken(secret),
			expectError: false,
		},
		{
			name:        "Invalid token",
			token:       "invalid-token",
			expectError: true,
		},
		{
			name:        "Empty token",
			token:       "",
			expectError: true,
		},
		{
			name:        "Wrong secret",
			token:       createTokenWithWrongSecret(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := authService.ValidateToken(tt.token)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if claims == nil {
				t.Error("Expected non-nil claims")
			}
		})
	}
}

func TestExtractUserInfo(t *testing.T) {
	secret := "test-secret"
	authService := New(secret)

	// Create claims with user info
	claims := jwt.MapClaims{
		"user_id":  "user-123",
		"username": "john_doe",
		"email":    "john@example.com",
		"exp":      time.Now().Add(time.Hour).Unix(),
	}

	userID, username, email := authService.ExtractUserInfo(claims)

	if userID != "user-123" {
		t.Errorf("Expected user ID 'user-123', got '%s'", userID)
	}

	if username != "john_doe" {
		t.Errorf("Expected username 'john_doe', got '%s'", username)
	}

	if email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", email)
	}
}

func TestExtractUserInfoWithMissingClaims(t *testing.T) {
	secret := "test-secret"
	authService := New(secret)

	// Create claims without user info
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	}

	userID, username, email := authService.ExtractUserInfo(claims)

	// Should return empty strings for missing claims
	if userID != "" {
		t.Errorf("Expected empty user ID, got '%s'", userID)
	}

	if username != "" {
		t.Errorf("Expected empty username, got '%s'", username)
	}

	if email != "" {
		t.Errorf("Expected empty email, got '%s'", email)
	}
}

func TestValidateTokenWithExpiredToken(t *testing.T) {
	secret := "test-secret"
	authService := New(secret)

	// Create an expired token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "user-123",
		"channel": "test-channel",
		"exp":     time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	_, err = authService.ValidateToken(tokenString)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func TestValidateTokenWithInvalidSigningMethod(t *testing.T) {
	secret := "test-secret"
	authService := New(secret)

	// Create a token with invalid signing method
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"user_id": "user-123",
		"channel": "test-channel",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	// This will fail because we're using RS256 but the service expects HS256
	tokenString := token.Raw
	if tokenString == "" {
		// If Raw is empty, we need to manually create an invalid token
		tokenString = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMiLCJjaGFubmVsIjoidGVzdC1jaGFubmVsIiwiZXhwIjoxNjAwMDAwMDAwfQ.invalid"
	}

	_, err := authService.ValidateToken(tokenString)
	if err == nil {
		t.Error("Expected error for invalid signing method")
	}
}

// Helper functions for testing

func createValidToken(secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  "user-123",
		"username": "john_doe",
		"email":    "john@example.com",
		"channel":  "test-channel",
		"exp":      time.Now().Add(time.Hour).Unix(),
	})

	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func createTokenWithWrongSecret() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "user-123",
		"channel": "test-channel",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	tokenString, _ := token.SignedString([]byte("wrong-secret"))
	return tokenString
}

func TestConcurrentTokenValidation(t *testing.T) {
	secret := "test-secret"
	authService := New(secret)

	// Create a valid token
	token := createValidToken(secret)

	// Test concurrent validation
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			_, err := authService.ValidateToken(token)
			if err != nil {
				t.Errorf("Unexpected error during concurrent validation: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestConcurrentTokenGeneration(t *testing.T) {
	secret := "test-secret"
	authService := New(secret)

	// Test concurrent token generation
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			userID := fmt.Sprintf("user-%d", id)
			channel := fmt.Sprintf("channel-%d", id)

			token, err := authService.GenerateToken(userID, channel)
			if err != nil {
				t.Errorf("Unexpected error during concurrent token generation: %v", err)
			}

			if token == "" {
				t.Error("Expected non-empty token")
			}

			// Validate the generated token
			claims, err := authService.ValidateToken(token)
			if err != nil {
				t.Errorf("Error validating generated token: %v", err)
			}

			if claims["user_id"] != userID {
				t.Errorf("Expected user_id %s, got %v", userID, claims["user_id"])
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestExtractUserInfoWithDifferentTypes(t *testing.T) {
	secret := "test-secret"
	authService := New(secret)

	// Test with different claim types
	claims := jwt.MapClaims{
		"user_id":  123,        // number
		"username": "john_doe", // string
		"email":    nil,        // nil
		"exp":      time.Now().Add(time.Hour).Unix(),
	}

	userID, username, email := authService.ExtractUserInfo(claims)

	if userID != "123" {
		t.Errorf("Expected user ID '123', got '%s'", userID)
	}

	if username != "john_doe" {
		t.Errorf("Expected username 'john_doe', got '%s'", username)
	}

	if email != "<nil>" {
		t.Errorf("Expected email '<nil>', got '%s'", email)
	}
}
