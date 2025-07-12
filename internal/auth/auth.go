package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Service handles JWT authentication
type Service struct {
	jwtSecret []byte
}

// New creates a new auth service
func New(jwtSecret string) *Service {
	return &Service{
		jwtSecret: []byte(jwtSecret),
	}
}

// GenerateToken generates a JWT token for a user
func (s *Service) GenerateToken(userID, channel string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"channel": channel,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 24 hours expiration
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// ExtractUserInfo extracts user information from JWT claims
func (s *Service) ExtractUserInfo(claims jwt.MapClaims) (userID, username, email string) {
	if uid, exists := claims["user_id"]; exists {
		userID = fmt.Sprintf("%v", uid)
	}
	if uname, exists := claims["username"]; exists {
		username = fmt.Sprintf("%v", uname)
	}
	if uemail, exists := claims["email"]; exists {
		email = fmt.Sprintf("%v", uemail)
	}
	return userID, username, email
}
