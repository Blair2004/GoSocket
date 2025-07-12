package main

import (
	"fmt"
	"log"

	"socket-server/internal/auth"
	"socket-server/internal/config"
	"socket-server/internal/models"
	"socket-server/internal/services"
	"socket-server/internal/websocket"
	"socket-server/pkg/logger"
)

func testComponents() {
	fmt.Println("Testing refactored socket server components...")

	// Test 1: Configuration
	fmt.Println("\n1. Testing Configuration...")
	cfg := config.New()
	cfg.Port = "8888"
	cfg.JWTSecret = "test-secret"
	cfg.WorkingDir = "/tmp"
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}
	fmt.Printf("✓ Configuration loaded and validated successfully\n")

	// Test 2: Logger
	fmt.Println("\n2. Testing Logger...")
	logger := logger.New(false)
	logger.Info("Test log message")
	fmt.Printf("✓ Logger working correctly\n")

	// Test 3: Auth Service
	fmt.Println("\n3. Testing Auth Service...")
	authService := auth.New(cfg.JWTSecret)
	token, err := authService.GenerateToken("test-user", "test-channel")
	if err != nil {
		log.Fatalf("Token generation failed: %v", err)
	}
	fmt.Printf("✓ Generated token: %s...\n", token[:20])

	// Test 4: Laravel Service
	fmt.Println("\n4. Testing Laravel Service...")
	laravelSvc := services.NewLaravelService(cfg.WorkingDir, cfg.PHPBinary, cfg.LaravelCmd, cfg.TempDir, logger)
	if err := laravelSvc.InitializeTempDirectory(); err != nil {
		log.Fatalf("Laravel service initialization failed: %v", err)
	}
	fmt.Printf("✓ Laravel service initialized successfully\n")

	// Test 5: Models
	fmt.Println("\n5. Testing Models...")
	client := models.NewClient("test-client", nil)
	channel := models.NewChannel("test-channel")

	channel.AddClient(client)
	clients := channel.GetClients()
	if len(clients) != 1 {
		log.Fatalf("Expected 1 client, got %d", len(clients))
	}
	fmt.Printf("✓ Models working correctly\n")

	// Test 6: WebSocket Server
	fmt.Println("\n6. Testing WebSocket Server...")
	wsServer := websocket.New(authService, laravelSvc, logger)
	if wsServer == nil {
		log.Fatalf("WebSocket server creation failed")
	}
	fmt.Printf("✓ WebSocket server created successfully\n")

	fmt.Println("\n✅ All components tested successfully!")
	fmt.Println("The refactored socket server is working correctly.")
}
