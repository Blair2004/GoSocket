package config

import (
	"os"
	"path/filepath"
)

// Config holds all configuration for the socket server
type Config struct {
	Port       string
	JWTSecret  string
	HTTPToken  string
	WorkingDir string
	PHPBinary  string
	LaravelCmd string
	TempDir    string
	Debug      bool
}

// New creates a new configuration with default values
func New() *Config {
	return &Config{
		Port:       getEnv("SOCKET_PORT", "8080"),
		JWTSecret:  getEnv("JWT_SECRET", "default-secret-key-change-in-production"),
		HTTPToken:  getEnv("HTTP_TOKEN", ""),
		WorkingDir: getEnv("LARAVEL_PATH", "."),
		PHPBinary:  getEnv("PHP_BINARY", "php"),
		LaravelCmd: getEnv("LARAVEL_COMMAND", "ns:socket-handler"),
		TempDir:    getEnv("SOCKET_TEMP_DIR", filepath.Join(os.TempDir(), "socket-server-payloads")),
		Debug:      getEnv("SOCKET_DEBUG", "false") == "true",
	}
}

// LoadFromFlags updates configuration from command line flags
func (c *Config) LoadFromFlags(port, jwtSecret, httpToken, workingDir, phpBinary, laravelCmd, tempDir string) {
	if port != "" {
		c.Port = port
	}
	if jwtSecret != "" {
		c.JWTSecret = jwtSecret
	}
	if httpToken != "" {
		c.HTTPToken = httpToken
	}
	if workingDir != "" {
		c.WorkingDir = workingDir
	}
	if phpBinary != "" {
		c.PHPBinary = phpBinary
	}
	if laravelCmd != "" {
		c.LaravelCmd = laravelCmd
	}
	if tempDir != "" {
		c.TempDir = tempDir
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Port == "" {
		return ErrEmptyPort
	}
	if c.JWTSecret == "" {
		return ErrEmptyJWTSecret
	}
	if c.HTTPToken == "" {
		return ErrEmptyHTTPToken
	}
	return nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
