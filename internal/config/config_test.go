package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	// Clear environment variables to test defaults
	originalEnv := map[string]string{
		"SOCKET_PORT":     os.Getenv("SOCKET_PORT"),
		"JWT_SECRET":      os.Getenv("JWT_SECRET"),
		"LARAVEL_PATH":    os.Getenv("LARAVEL_PATH"),
		"PHP_BINARY":      os.Getenv("PHP_BINARY"),
		"LARAVEL_COMMAND": os.Getenv("LARAVEL_COMMAND"),
		"SOCKET_TEMP_DIR": os.Getenv("SOCKET_TEMP_DIR"),
		"SOCKET_DEBUG":    os.Getenv("SOCKET_DEBUG"),
	}

	// Clear environment variables
	for key := range originalEnv {
		os.Unsetenv(key)
	}

	cfg := New()

	if cfg == nil {
		t.Error("Expected non-nil config")
	}

	// Check default values
	if cfg.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Port)
	}

	if cfg.PHPBinary != "php" {
		t.Errorf("Expected default PHP binary 'php', got %s", cfg.PHPBinary)
	}

	if cfg.LaravelCmd != "ns:socket-handler" {
		t.Errorf("Expected default Laravel command 'ns:socket-handler', got %s", cfg.LaravelCmd)
	}

	if cfg.WorkingDir != "." {
		t.Errorf("Expected default working directory '.', got %s", cfg.WorkingDir)
	}

	expectedTempDir := filepath.Join(os.TempDir(), "socket-server-payloads")
	if cfg.TempDir != expectedTempDir {
		t.Errorf("Expected default temp directory %s, got %s", expectedTempDir, cfg.TempDir)
	}

	if cfg.Debug {
		t.Error("Expected debug to be false by default")
	}

	// Restore original environment
	for key, value := range originalEnv {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

func TestNewWithEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		"SOCKET_PORT":     os.Getenv("SOCKET_PORT"),
		"JWT_SECRET":      os.Getenv("JWT_SECRET"),
		"LARAVEL_PATH":    os.Getenv("LARAVEL_PATH"),
		"PHP_BINARY":      os.Getenv("PHP_BINARY"),
		"LARAVEL_COMMAND": os.Getenv("LARAVEL_COMMAND"),
		"SOCKET_TEMP_DIR": os.Getenv("SOCKET_TEMP_DIR"),
		"SOCKET_DEBUG":    os.Getenv("SOCKET_DEBUG"),
	}

	// Set test environment variables
	os.Setenv("SOCKET_PORT", "9000")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("LARAVEL_PATH", "/test/path")
	os.Setenv("PHP_BINARY", "/usr/bin/php8.2")
	os.Setenv("LARAVEL_COMMAND", "test:command")
	os.Setenv("SOCKET_TEMP_DIR", "/tmp/test")
	os.Setenv("SOCKET_DEBUG", "true")

	cfg := New()

	// Check loaded values
	if cfg.Port != "9000" {
		t.Errorf("Expected port 9000, got %s", cfg.Port)
	}

	if cfg.JWTSecret != "test-secret" {
		t.Errorf("Expected JWT secret 'test-secret', got %s", cfg.JWTSecret)
	}

	if cfg.WorkingDir != "/test/path" {
		t.Errorf("Expected working dir '/test/path', got %s", cfg.WorkingDir)
	}

	if cfg.PHPBinary != "/usr/bin/php8.2" {
		t.Errorf("Expected PHP binary '/usr/bin/php8.2', got %s", cfg.PHPBinary)
	}

	if cfg.LaravelCmd != "test:command" {
		t.Errorf("Expected Laravel command 'test:command', got %s", cfg.LaravelCmd)
	}

	if cfg.TempDir != "/tmp/test" {
		t.Errorf("Expected temp dir '/tmp/test', got %s", cfg.TempDir)
	}

	if !cfg.Debug {
		t.Error("Expected debug to be true")
	}

	// Restore original environment
	for key, value := range originalEnv {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

func TestLoadFromFlags(t *testing.T) {
	cfg := New()

	// Test loading from flags
	cfg.LoadFromFlags(
		"7000",            // port
		"flag-secret",     // jwtSecret
		"/flag/path",      // workingDir
		"/usr/bin/php7.4", // phpBinary
		"flag:command",    // laravelCmd
		"/tmp/flag",       // tempDir
	)

	if cfg.Port != "7000" {
		t.Errorf("Expected port 7000, got %s", cfg.Port)
	}

	if cfg.JWTSecret != "flag-secret" {
		t.Errorf("Expected JWT secret 'flag-secret', got %s", cfg.JWTSecret)
	}

	if cfg.WorkingDir != "/flag/path" {
		t.Errorf("Expected working dir '/flag/path', got %s", cfg.WorkingDir)
	}

	if cfg.PHPBinary != "/usr/bin/php7.4" {
		t.Errorf("Expected PHP binary '/usr/bin/php7.4', got %s", cfg.PHPBinary)
	}

	if cfg.LaravelCmd != "flag:command" {
		t.Errorf("Expected Laravel command 'flag:command', got %s", cfg.LaravelCmd)
	}

	if cfg.TempDir != "/tmp/flag" {
		t.Errorf("Expected temp dir '/tmp/flag', got %s", cfg.TempDir)
	}
}

func TestLoadFromFlagsWithEmptyValues(t *testing.T) {
	cfg := New()
	originalPort := cfg.Port
	originalPHPBinary := cfg.PHPBinary

	// Test with empty flag values (should keep existing values)
	cfg.LoadFromFlags("", "", "", "", "", "")

	if cfg.Port != originalPort {
		t.Errorf("Expected port to remain %s, got %s", originalPort, cfg.Port)
	}

	if cfg.PHPBinary != originalPHPBinary {
		t.Errorf("Expected PHP binary to remain %s, got %s", originalPHPBinary, cfg.PHPBinary)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "Valid config",
			config: &Config{
				Port:       "8080",
				JWTSecret:  "test-secret",
				WorkingDir: "/test/path",
				PHPBinary:  "php",
				LaravelCmd: "test:command",
				TempDir:    "/tmp/test",
			},
			expectError: false,
		},
		{
			name: "Missing port",
			config: &Config{
				JWTSecret:  "test-secret",
				WorkingDir: "/test/path",
				PHPBinary:  "php",
				LaravelCmd: "test:command",
				TempDir:    "/tmp/test",
			},
			expectError: true,
		},
		{
			name: "Missing JWT secret",
			config: &Config{
				Port:       "8080",
				WorkingDir: "/test/path",
				PHPBinary:  "php",
				LaravelCmd: "test:command",
				TempDir:    "/tmp/test",
			},
			expectError: true,
		},
		{
			name: "Empty JWT secret",
			config: &Config{
				Port:       "8080",
				JWTSecret:  "",
				WorkingDir: "/test/path",
				PHPBinary:  "php",
				LaravelCmd: "test:command",
				TempDir:    "/tmp/test",
			},
			expectError: true,
		},
		{
			name: "Empty port",
			config: &Config{
				Port:       "",
				JWTSecret:  "test-secret",
				WorkingDir: "/test/path",
				PHPBinary:  "php",
				LaravelCmd: "test:command",
				TempDir:    "/tmp/test",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestEnvironmentFallback(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		"SOCKET_PORT":     os.Getenv("SOCKET_PORT"),
		"JWT_SECRET":      os.Getenv("JWT_SECRET"),
		"LARAVEL_PATH":    os.Getenv("LARAVEL_PATH"),
		"PHP_BINARY":      os.Getenv("PHP_BINARY"),
		"LARAVEL_COMMAND": os.Getenv("LARAVEL_COMMAND"),
		"SOCKET_TEMP_DIR": os.Getenv("SOCKET_TEMP_DIR"),
	}

	// Set environment variables
	os.Setenv("SOCKET_PORT", "9000")
	os.Setenv("JWT_SECRET", "env-secret")
	os.Setenv("LARAVEL_PATH", "/env/path")
	os.Setenv("PHP_BINARY", "/usr/bin/php8.1")
	os.Setenv("LARAVEL_COMMAND", "env:command")
	os.Setenv("SOCKET_TEMP_DIR", "/tmp/env")

	cfg := New()

	// Test flag override (flags should take precedence)
	cfg.LoadFromFlags(
		"7000",       // port - should override env
		"",           // jwtSecret - should keep env value
		"/flag/path", // workingDir - should override env
		"",           // phpBinary - should keep env value
		"",           // laravelCmd - should keep env value
		"",           // tempDir - should keep env value
	)

	if cfg.Port != "7000" {
		t.Errorf("Expected port 7000 (flag override), got %s", cfg.Port)
	}

	if cfg.JWTSecret != "env-secret" {
		t.Errorf("Expected JWT secret 'env-secret' (env value), got %s", cfg.JWTSecret)
	}

	if cfg.WorkingDir != "/flag/path" {
		t.Errorf("Expected working dir '/flag/path' (flag override), got %s", cfg.WorkingDir)
	}

	if cfg.PHPBinary != "/usr/bin/php8.1" {
		t.Errorf("Expected PHP binary '/usr/bin/php8.1' (env value), got %s", cfg.PHPBinary)
	}

	// Restore original environment
	for key, value := range originalEnv {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}
}

func TestDebugFlagParsing(t *testing.T) {
	tests := []struct {
		name        string
		debugValue  string
		expectDebug bool
	}{
		{"Debug true", "true", true},
		{"Debug false", "false", false},
		{"Debug empty", "", false},
		{"Debug invalid", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalDebug := os.Getenv("SOCKET_DEBUG")

			// Set test environment
			if tt.debugValue == "" {
				os.Unsetenv("SOCKET_DEBUG")
			} else {
				os.Setenv("SOCKET_DEBUG", tt.debugValue)
			}

			cfg := New()

			if cfg.Debug != tt.expectDebug {
				t.Errorf("Expected debug %v, got %v", tt.expectDebug, cfg.Debug)
			}

			// Restore original environment
			if originalDebug == "" {
				os.Unsetenv("SOCKET_DEBUG")
			} else {
				os.Setenv("SOCKET_DEBUG", originalDebug)
			}
		})
	}
}

func TestConcurrentConfigAccess(t *testing.T) {
	cfg := New()

	done := make(chan bool)

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = cfg.Port
			_ = cfg.JWTSecret
			_ = cfg.WorkingDir
			_ = cfg.PHPBinary
			_ = cfg.LaravelCmd
			_ = cfg.TempDir
			_ = cfg.Debug
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
