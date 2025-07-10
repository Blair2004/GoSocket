# Socket Server Makefile

.PHONY: build clean run test install deps help server cli dashboard

# Default target
all: build

# Build all binaries
build: deps
	@echo "Building socket server binaries..."
	@mkdir -p bin
	@go build -o bin/socket-server main.go
	@go build -o bin/socket cmd/cli/main.go
	@chmod +x bin/socket-server bin/socket
	@echo "Build completed!"

# Install dependencies
deps:
	@echo "Installing Go dependencies..."
	@go mod tidy
	@go mod download

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@go clean

# Run the server
run: build
	@echo "Starting socket server..."
	@./bin/socket-server

# Run the server in development mode
dev: build
	@echo "Starting socket server in development mode..."
	@SOCKET_DEBUG=true ./bin/socket-server

# Test the build
test: build
	@echo "Testing server health..."
	@./bin/socket health || echo "Server not running"

# Install binaries to system PATH
install: build
	@echo "Installing binaries to /usr/local/bin..."
	@sudo cp bin/socket-server /usr/local/bin/
	@sudo cp bin/socket /usr/local/bin/
	@echo "Installation completed!"

# Start server as daemon
daemon: build
	@echo "Starting socket server as daemon..."
	@nohup ./bin/socket-server > server.log 2>&1 &
	@echo "Server started in background. PID: $$!"

# Stop daemon
stop:
	@echo "Stopping socket server daemon..."
	@pkill -f socket-server || echo "No server process found"

# View server logs
logs:
	@tail -f server.log

# Open dashboard in browser
dashboard:
	@echo "Opening dashboard..."
	@python3 -c "import webbrowser; webbrowser.open('http://localhost:8080')" 2>/dev/null || \
	 open http://localhost:8080 2>/dev/null || \
	 xdg-open http://localhost:8080 2>/dev/null || \
	 echo "Please open http://localhost:8080 in your browser"

# CLI shortcuts
cli-help: build
	@./bin/socket --help

cli-health: build
	@./bin/socket health

cli-clients: build
	@./bin/socket list clients

cli-channels: build
	@./bin/socket list channels

# Development helpers
fmt:
	@echo "Formatting Go code..."
	@go fmt ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

lint: fmt vet
	@echo "Code formatting and vetting completed"

# Build for different platforms
build-linux: deps
	@echo "Building for Linux..."
	@mkdir -p bin
	@GOOS=linux GOARCH=amd64 go build -o bin/socket-server-linux main.go
	@GOOS=linux GOARCH=amd64 go build -o bin/socket-linux cmd/cli/main.go

build-macos: deps
	@echo "Building for macOS..."
	@mkdir -p bin
	@GOOS=darwin GOARCH=amd64 go build -o bin/socket-server-macos main.go
	@GOOS=darwin GOARCH=amd64 go build -o bin/socket-macos cmd/cli/main.go

build-windows: deps
	@echo "Building for Windows..."
	@mkdir -p bin
	@GOOS=windows GOARCH=amd64 go build -o bin/socket-server.exe main.go
	@GOOS=windows GOARCH=amd64 go build -o bin/socket.exe cmd/cli/main.go

build-all: build-linux build-macos build-windows
	@echo "Built for all platforms"

# Docker targets
docker-build:
	@echo "Building Docker image..."
	@docker build -t socket-server .

docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p 8080:8080 socket-server

# Laravel integration
laravel-install:
	@echo "Installing Laravel integration files..."
	@if [ ! -d "../../app" ]; then echo "Error: Not in Laravel project root"; exit 1; fi
	@cp -r laravel/app/* ../../app/
	@cp -r laravel/config/* ../../config/
	@echo "Laravel integration files copied. Please add SocketServiceProvider to config/app.php"

# Help
help:
	@echo "Socket Server Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  build         - Build all binaries"
	@echo "  clean         - Clean build artifacts"
	@echo "  run           - Build and run server"
	@echo "  dev           - Run server in development mode"
	@echo "  test          - Test the build"
	@echo "  install       - Install binaries to system PATH"
	@echo "  daemon        - Start server as background daemon"
	@echo "  stop          - Stop daemon server"
	@echo "  logs          - View server logs"
	@echo "  dashboard     - Open web dashboard"
	@echo ""
	@echo "CLI shortcuts:"
	@echo "  cli-help      - Show CLI help"
	@echo "  cli-health    - Check server health"
	@echo "  cli-clients   - List connected clients"
	@echo "  cli-channels  - List active channels"
	@echo ""
	@echo "Development:"
	@echo "  fmt           - Format Go code"
	@echo "  vet           - Run go vet"
	@echo "  lint          - Format and vet code"
	@echo ""
	@echo "Cross-platform builds:"
	@echo "  build-linux   - Build for Linux"
	@echo "  build-macos   - Build for macOS"
	@echo "  build-windows - Build for Windows"
	@echo "  build-all     - Build for all platforms"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run in Docker container"
	@echo ""
	@echo "Laravel:"
	@echo "  laravel-install - Install Laravel integration files"
