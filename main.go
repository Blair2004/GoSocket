package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	"socket-server/internal/auth"
	"socket-server/internal/config"
	"socket-server/internal/handlers"
	"socket-server/internal/middleware"
	"socket-server/internal/services"
	"socket-server/internal/websocket"
	"socket-server/pkg/logger"
)

var (
	port       string
	jwtSecret  string
	httpToken  string
	workingDir string
	phpBinary  string
	laravelCmd string
	tempDir    string
)

var rootCmd = &cobra.Command{
	Use:   "socket-server",
	Short: "High-performance WebSocket server for Laravel integration",
	Long: `A standalone WebSocket server that provides real-time bidirectional communication 
for Laravel applications. Features include channel management, JWT authentication, 
client management, and Laravel event integration.`,
	Run: runServer,
}

func init() {
	rootCmd.Flags().StringVarP(&port, "port", "p", "", "Port to run the server on (default: 8080 or SOCKET_PORT env var)")
	rootCmd.Flags().StringVarP(&jwtSecret, "token", "t", "", "JWT secret for authentication (default: JWT_SECRET env var)")
	rootCmd.Flags().StringVar(&httpToken, "http-token", "", "HTTP API authentication token (required for API access)")
	rootCmd.Flags().StringVarP(&workingDir, "dir", "d", "", "Working directory for Laravel commands (default: LARAVEL_PATH env var)")
	rootCmd.Flags().StringVar(&phpBinary, "php", "", "PHP binary path (default: 'php' or PHP_BINARY env var)")
	rootCmd.Flags().StringVar(&laravelCmd, "command", "", "Laravel artisan command to execute (default: 'socket:handle' or LARAVEL_COMMAND env var)")
	rootCmd.Flags().StringVar(&tempDir, "temp", "", "Temporary directory for payload files (default: system temp/socket-server-payloads or SOCKET_TEMP_DIR env var)")
}

func runServer(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg := config.New()
	cfg.LoadFromFlags(port, jwtSecret, httpToken, workingDir, phpBinary, laravelCmd, tempDir)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Initialize logger
	logger := logger.New(cfg.Debug)

	// Display configuration
	logger.Info("Starting Socket Server on port %s", cfg.Port)

	// Safely display JWT secret (first few characters)
	secretDisplay := cfg.JWTSecret
	if len(secretDisplay) > 10 {
		secretDisplay = secretDisplay[:10] + "..."
	} else if len(secretDisplay) > 3 {
		secretDisplay = secretDisplay[:3] + "..."
	}

	// Safely display HTTP token (first few characters)
	tokenDisplay := cfg.HTTPToken
	if len(tokenDisplay) > 10 {
		tokenDisplay = tokenDisplay[:10] + "..."
	} else if len(tokenDisplay) > 3 {
		tokenDisplay = tokenDisplay[:3] + "..."
	}

	logger.Info("JWT Secret: %s", secretDisplay)
	logger.Info("HTTP API Token: %s", tokenDisplay)
	logger.Info("Working Directory: %s", cfg.WorkingDir)
	logger.Info("PHP Binary: %s", cfg.PHPBinary)
	logger.Info("Laravel Command: %s", cfg.LaravelCmd)
	logger.Info("Temp Directory: %s", cfg.TempDir)

	// Initialize services
	authService := auth.New(cfg.JWTSecret)
	laravelSvc := services.NewLaravelService(cfg.WorkingDir, cfg.PHPBinary, cfg.LaravelCmd, cfg.TempDir, logger)

	// Initialize temp directory and start cleanup routine
	if err := laravelSvc.InitializeTempDirectory(); err != nil {
		logger.Fatal("Failed to initialize temp directory: %v", err)
	}
	laravelSvc.StartCleanupRoutine()

	// Initialize WebSocket server
	wsServer := websocket.New(authService, laravelSvc, logger)

	// Initialize HTTP handlers
	httpHandlers := handlers.New(wsServer, logger)

	// Initialize HTTP authentication middleware
	httpAuth := middleware.NewHTTPAuth(cfg.HTTPToken, logger)

	// Setup routes
	r := mux.NewRouter()

	// WebSocket endpoint (no authentication required for WebSocket - handled internally)
	r.HandleFunc("/ws", wsServer.HandleConnection)

	// REST API endpoints (all require authentication)
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", httpAuth.AuthenticateFunc(httpHandlers.Health)).Methods("GET")
	api.HandleFunc("/clients", httpAuth.AuthenticateFunc(httpHandlers.GetClients)).Methods("GET")
	api.HandleFunc("/channels", httpAuth.AuthenticateFunc(httpHandlers.GetChannels)).Methods("GET")
	api.HandleFunc("/channels/{channel}/clients", httpAuth.AuthenticateFunc(httpHandlers.GetChannelClients)).Methods("GET")
	api.HandleFunc("/clients/{client}/kick", httpAuth.AuthenticateFunc(httpHandlers.KickClient)).Methods("POST")
	api.HandleFunc("/broadcast", httpAuth.AuthenticateFunc(httpHandlers.Broadcast)).Methods("POST")

	// Static file serving for admin interface (no authentication required)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	// Start server
	logger.Info("Socket server starting on port %s", cfg.Port)
	logger.Fatal("Server error: %v", http.ListenAndServe(":"+cfg.Port, r))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
