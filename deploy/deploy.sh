#!/bin/bash

# Socket Server Deployment Script

set -e

INSTALL_DIR="/opt/socket-server"
SERVICE_FILE="/etc/systemd/system/socket-server.service"
USER="socket"
GROUP="socket"

echo "Socket Server Deployment Script"
echo "================================"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root (use sudo)"
    exit 1
fi

# Create user and group
echo "Creating user and group..."
if ! id "$USER" &>/dev/null; then
    useradd --system --shell /bin/false --home "$INSTALL_DIR" --create-home "$USER"
fi

# Create installation directory
echo "Creating installation directory..."
mkdir -p "$INSTALL_DIR"/{bin,logs,web}

# Build the application
echo "Building application..."
if [ ! -f "bin/socket-server" ] || [ ! -f "bin/socket" ]; then
    echo "Building binaries..."
    make build
fi

# Copy files
echo "Installing files..."
cp bin/socket-server "$INSTALL_DIR/bin/"
cp bin/socket "$INSTALL_DIR/bin/"
cp -r web/* "$INSTALL_DIR/web/"

# Set permissions
echo "Setting permissions..."
chown -R "$USER:$GROUP" "$INSTALL_DIR"
chmod +x "$INSTALL_DIR/bin/socket-server"
chmod +x "$INSTALL_DIR/bin/socket"

# Install systemd service
echo "Installing systemd service..."
cp deploy/socket-server.service "$SERVICE_FILE"

# Prompt for configuration
echo ""
echo "Configuration:"
read -p "Enter port (default: 8080): " PORT
PORT=${PORT:-8080}

read -p "Enter JWT secret (leave empty to generate): " JWT_SECRET
if [ -z "$JWT_SECRET" ]; then
    JWT_SECRET=$(openssl rand -hex 32)
    echo "Generated JWT secret: $JWT_SECRET"
fi

read -p "Enable debug mode? (y/N): " DEBUG
if [[ $DEBUG =~ ^[Yy]$ ]]; then
    DEBUG_VALUE="true"
else
    DEBUG_VALUE="false"
fi

# Update service file with configuration
sed -i "s/SOCKET_PORT=8080/SOCKET_PORT=$PORT/" "$SERVICE_FILE"
sed -i "s/JWT_SECRET=your-production-secret-here/JWT_SECRET=$JWT_SECRET/" "$SERVICE_FILE"
sed -i "s/SOCKET_DEBUG=false/SOCKET_DEBUG=$DEBUG_VALUE/" "$SERVICE_FILE"

# Create environment file
echo "Creating environment file..."
cat > "$INSTALL_DIR/.env" << EOF
SOCKET_PORT=$PORT
JWT_SECRET=$JWT_SECRET
SOCKET_DEBUG=$DEBUG_VALUE
EOF

chown "$USER:$GROUP" "$INSTALL_DIR/.env"
chmod 600 "$INSTALL_DIR/.env"

# Reload systemd and enable service
echo "Configuring systemd..."
systemctl daemon-reload
systemctl enable socket-server

# Start the service
echo "Starting socket server..."
systemctl start socket-server

# Wait a moment and check status
sleep 2
if systemctl is-active --quiet socket-server; then
    echo ""
    echo "✅ Socket server deployed successfully!"
    echo ""
    echo "Service status:"
    systemctl status socket-server --no-pager -l
    echo ""
    echo "Server is running on port $PORT"
    echo "Dashboard: http://localhost:$PORT"
    echo ""
    echo "Useful commands:"
    echo "  systemctl status socket-server   - Check status"
    echo "  systemctl restart socket-server  - Restart service"
    echo "  systemctl stop socket-server     - Stop service"
    echo "  journalctl -u socket-server -f   - View logs"
    echo "  $INSTALL_DIR/bin/socket health   - Check health"
else
    echo ""
    echo "❌ Service failed to start!"
    echo "Check logs with: journalctl -u socket-server -l"
    exit 1
fi

# Configure firewall if available
if command -v ufw >/dev/null 2>&1; then
    read -p "Configure firewall to allow port $PORT? (y/N): " FIREWALL
    if [[ $FIREWALL =~ ^[Yy]$ ]]; then
        ufw allow "$PORT"
        echo "Firewall rule added for port $PORT"
    fi
fi

echo ""
echo "Deployment completed!"
echo "Remember to:"
echo "1. Update your Laravel .env file with SOCKET_SERVER_URL=http://localhost:$PORT"
echo "2. Install the Laravel integration files in your project"
echo "3. Add SocketServiceProvider to your config/app.php"
