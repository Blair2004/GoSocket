package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

// Client represents a connected WebSocket client
type Client struct {
	ID         string          `json:"id"`
	Conn       *websocket.Conn `json:"-"`
	UserID     string          `json:"user_id,omitempty"`
	Username   string          `json:"username,omitempty"`
	Email      string          `json:"email,omitempty"` // Use email as user ID if available
	Channels   map[string]bool `json:"channels"`
	LastSeen   time.Time       `json:"last_seen"`
	RemoteAddr string          `json:"remote_addr"`
	UserAgent  string          `json:"user_agent"`
	mutex      sync.RWMutex    `json:"-"`
}

// Channel represents a communication channel
type Channel struct {
	Name        string             `json:"name"`
	Clients     map[string]*Client `json:"-"`
	IsPrivate   bool               `json:"is_private"`
	RequireAuth bool               `json:"require_auth"`
	CreatedAt   time.Time          `json:"created_at"`
	mutex       sync.RWMutex       `json:"-"`
}

// Message represents a message to be sent
type Message struct {
	ID        string      `json:"id"`
	Channel   string      `json:"channel"`
	Event     string      `json:"event"`
	Data      interface{} `json:"data"`
	UserID    string      `json:"user_id,omitempty"`
	Username  string      `json:"username,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Server represents the socket server
type Server struct {
	clients    map[string]*Client
	channels   map[string]*Channel
	upgrader   websocket.Upgrader
	jwtSecret  []byte
	port       string
	workingDir string
	phpBinary  string
	laravelCmd string
	tempDir    string
	mutex      sync.RWMutex
}

// NewServer creates a new socket server instance
func NewServer(port string, jwtSecret string, workingDir string, phpBinary string, laravelCmd string, tempDir string) *Server {
	return &Server{
		clients:    make(map[string]*Client),
		channels:   make(map[string]*Channel),
		jwtSecret:  []byte(jwtSecret),
		port:       port,
		workingDir: workingDir,
		phpBinary:  phpBinary,
		laravelCmd: laravelCmd,
		tempDir:    tempDir,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// Start starts the socket server
func (s *Server) Start() {
	// Initialize temp directory and start cleanup routine
	s.initializeTempDirectory()
	s.startCleanupRoutine()

	r := mux.NewRouter()

	// WebSocket endpoint
	r.HandleFunc("/ws", s.handleWebSocket)

	// REST API endpoints
	r.HandleFunc("/api/clients", s.handleGetClients).Methods("GET")
	r.HandleFunc("/api/channels", s.handleGetChannels).Methods("GET")
	r.HandleFunc("/api/channels/{channel}/clients", s.handleGetChannelClients).Methods("GET")
	r.HandleFunc("/api/clients/{client}/kick", s.handleKickClient).Methods("POST")
	r.HandleFunc("/api/broadcast", s.handleBroadcast).Methods("POST")
	r.HandleFunc("/api/health", s.handleHealth).Methods("GET")

	// Static file serving for admin interface
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

	log.Printf("Socket server starting on port %s", s.port)
	log.Fatal(http.ListenAndServe(":"+s.port, r))
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Set connection timeouts and limits
	conn.SetReadLimit(512 * 1024) // 512KB max message size
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	client := &Client{
		ID:         uuid.New().String(),
		Conn:       conn,
		Channels:   make(map[string]bool),
		LastSeen:   time.Now(),
		RemoteAddr: r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	}

	s.mutex.Lock()
	s.clients[client.ID] = client
	s.mutex.Unlock()

	log.Printf("Client connected: %s from %s (User-Agent: %s)", client.ID, client.RemoteAddr, client.UserAgent)

	// Send welcome message
	welcome := Message{
		ID:        uuid.New().String(),
		Event:     "connected",
		Data:      map[string]string{"client_id": client.ID},
		Timestamp: time.Now(),
	}
	client.SendMessage(welcome)

	// Start ping ticker for connection health
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Handle client messages and wait for completion
	done := make(chan bool)
	go s.handleClientMessages(client, done)
	go s.handleClientPing(client, pingTicker, done)

	// Handle client disconnection
	defer func() {
		s.disconnectClient(client)
		conn.Close()
	}()

	// Wait for either message handler or ping handler to finish
	<-done
}

// handleClientMessages processes messages from a client
func (s *Server) handleClientMessages(client *Client, done chan bool) {
	defer func() {
		log.Printf("Client %s message handler exiting", client.ID)
		done <- true
	}()

	for {
		var msg map[string]interface{}
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			// Handle different types of disconnection errors
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				log.Printf("⚠️ Client %s unexpected disconnection: %v", client.ID, err)
			} else if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("🔌 Client %s disconnected abnormally (code 1006 - network issue, browser closed, etc.): %v", client.ID, err)
			} else if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("✅ Client %s disconnected normally", client.ID)
			} else {
				log.Printf("❌ Client %s disconnected with error: %v", client.ID, err)
			}
			break
		}

		// Reset read deadline on successful message
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		client.LastSeen = time.Now()

		// Log incoming message with more details
		actionStr := "unknown"
		if action, ok := msg["action"].(string); ok {
			actionStr = action
		}

		log.Printf("📥 INCOMING MESSAGE from client %s (user: %s): action=%s, data=%v",
			client.ID, client.Username, actionStr, msg)

		switch msg["action"] {
		case "authenticate":
			s.handleAuthentication(client, msg)
		case "join_channel":
			s.handleJoinChannel(client, msg)
		case "leave_channel":
			s.handleLeaveChannel(client, msg)
		case "send_message":
			s.handleSendMessage(client, msg)
		case "ping":
			log.Printf("📥 Client %s sent ping", client.ID)
			client.SendMessage(Message{
				ID:        uuid.New().String(),
				Event:     "pong",
				Timestamp: time.Now(),
			})
		default:
			log.Printf("⚠️ Unknown action '%s' from client %s", actionStr, client.ID)
		}
	}
}

// handleAuthentication processes client authentication
func (s *Server) handleAuthentication(client *Client, msg map[string]interface{}) {
	tokenStr, ok := msg["token"].(string)
	if !ok {
		log.Printf("Client %s sent invalid token format", client.ID)
		client.SendError("Invalid token format")
		return
	}

	log.Printf("🔐 Client %s attempting JWT authentication", client.ID)

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		log.Printf("❌ Client %s JWT authentication failed: %v", client.ID, err)
		client.SendError("Invalid token")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Printf("❌ Client %s JWT authentication failed: invalid token claims", client.ID)
		client.SendError("Invalid token claims")

		// Dispatch authentication failure to Laravel
		s.dispatchAuthenticationToLaravel(client, "failed", tokenStr)

		return
	}

	client.mutex.Lock()
	if userID, exists := claims["user_id"]; exists {
		client.UserID = fmt.Sprintf("%v", userID)
	}
	if username, exists := claims["username"]; exists {
		client.Username = fmt.Sprintf("%v", username)
	}
	if email, exists := claims["email"]; exists {
		client.Email = fmt.Sprintf("%v", email) // Use email as user ID if available
	}
	client.mutex.Unlock()

	log.Printf("✅ Client %s authenticated successfully as user %s (%s)",
		client.ID, client.Username, client.UserID)

	// Dispatch successful authentication to Laravel
	s.dispatchAuthenticationToLaravel(client, "success", tokenStr)
}

// handleJoinChannel adds client to a channel
func (s *Server) handleJoinChannel(client *Client, msg map[string]interface{}) {
	channelName, ok := msg["channel"].(string)
	if !ok {
		log.Printf("Client %s sent invalid channel name for join", client.ID)
		client.SendError("Invalid channel name")
		return
	}

	log.Printf("Client %s (%s) attempting to join channel '%s'",
		client.ID, client.Username, channelName)

	s.mutex.Lock()
	channel, exists := s.channels[channelName]
	if !exists {
		log.Printf("Creating new channel '%s'", channelName)
		channel = &Channel{
			Name:        channelName,
			Clients:     make(map[string]*Client),
			IsPrivate:   false,
			RequireAuth: false,
			CreatedAt:   time.Now(),
		}
		s.channels[channelName] = channel
	}
	s.mutex.Unlock()

	// Check if channel requires authentication
	if channel.RequireAuth && client.UserID == "" {
		log.Printf("Client %s denied access to channel '%s': authentication required",
			client.ID, channelName)
		client.SendError("Channel requires authentication")
		return
	}

	channel.mutex.Lock()
	channel.Clients[client.ID] = client
	channel.mutex.Unlock()

	client.mutex.Lock()
	client.Channels[channelName] = true
	client.mutex.Unlock()

	log.Printf("Client %s (%s) successfully joined channel '%s'",
		client.ID, client.Username, channelName)

	client.SendMessage(Message{
		ID:        uuid.New().String(),
		Event:     "joined_channel",
		Data:      map[string]string{"channel": channelName},
		Timestamp: time.Now(),
	})

	log.Printf("Client %s joined channel %s", client.ID, channelName)
}

// handleLeaveChannel removes client from a channel
func (s *Server) handleLeaveChannel(client *Client, msg map[string]interface{}) {
	channelName, ok := msg["channel"].(string)
	if !ok {
		log.Printf("Client %s sent invalid channel name for leave", client.ID)
		client.SendError("Invalid channel name")
		return
	}

	log.Printf("Client %s (%s) attempting to leave channel '%s'",
		client.ID, client.Username, channelName)

	s.mutex.RLock()
	channel, exists := s.channels[channelName]
	s.mutex.RUnlock()

	if !exists {
		log.Printf("Client %s tried to leave non-existent channel '%s'",
			client.ID, channelName)
		client.SendError("Channel not found")
		return
	}

	channel.mutex.Lock()
	delete(channel.Clients, client.ID)
	channel.mutex.Unlock()

	client.mutex.Lock()
	delete(client.Channels, channelName)
	client.mutex.Unlock()

	log.Printf("Client %s (%s) successfully left channel '%s'",
		client.ID, client.Username, channelName)

	client.SendMessage(Message{
		ID:        uuid.New().String(),
		Event:     "left_channel",
		Data:      map[string]string{"channel": channelName},
		Timestamp: time.Now(),
	})

	log.Printf("Client %s left channel %s", client.ID, channelName)
}

// handleSendMessage processes messages sent by clients
func (s *Server) handleSendMessage(client *Client, msg map[string]interface{}) {
	channelName, ok := msg["channel"].(string)
	if !ok {
		log.Printf("Client %s sent message with invalid channel name", client.ID)
		client.SendError("Invalid channel name")
		return
	}

	event, ok := msg["event"].(string)
	if !ok {
		event = "message"
	}

	data := msg["data"]

	// Log the message details with more visibility
	log.Printf("📤 MESSAGE SENT by client %s (%s) to channel '%s': event=%s, data=%v",
		client.ID, client.Username, channelName, event, data)

	message := Message{
		ID:        uuid.New().String(),
		Channel:   channelName,
		Event:     event,
		Data:      data,
		UserID:    client.UserID,
		Username:  client.Username,
		Timestamp: time.Now(),
	}

	// Dispatch to Laravel if configured
	s.dispatchToLaravel(message, client)

	// Broadcast to all clients in channel
	s.BroadcastToChannel(channelName, message)

	log.Printf("Message %s broadcasted to channel '%s'", message.ID, channelName)
}

// BroadcastToChannel sends a message to all clients in a channel
func (s *Server) BroadcastToChannel(channelName string, message Message) {
	s.mutex.RLock()
	channel, exists := s.channels[channelName]
	s.mutex.RUnlock()

	if !exists {
		log.Printf("Channel %s not found for broadcast", channelName)
		return
	}

	channel.mutex.RLock()
	clients := make([]*Client, 0, len(channel.Clients))
	for _, client := range channel.Clients {
		clients = append(clients, client)
	}
	channel.mutex.RUnlock()

	for _, client := range clients {
		client.SendMessage(message)
	}

	log.Printf("Broadcasted message to %d clients in channel %s", len(clients), channelName)
}

// disconnectClient removes a client from the server
func (s *Server) disconnectClient(client *Client) {
	log.Printf("Client %s (%s) disconnecting from %s",
		client.ID, client.Username, client.RemoteAddr)

	s.mutex.Lock()
	delete(s.clients, client.ID)
	s.mutex.Unlock()

	// Remove client from all channels
	client.mutex.RLock()
	channels := make([]string, 0, len(client.Channels))
	for channelName := range client.Channels {
		channels = append(channels, channelName)
	}
	client.mutex.RUnlock()

	if len(channels) > 0 {
		log.Printf("Removing client %s from channels: %v", client.ID, channels)
	}

	for _, channelName := range channels {
		s.mutex.RLock()
		channel, exists := s.channels[channelName]
		s.mutex.RUnlock()

		if exists {
			channel.mutex.Lock()
			delete(channel.Clients, client.ID)
			channel.mutex.Unlock()
		}
	}

	log.Printf("Client %s disconnected", client.ID)
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(message Message) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Conn == nil {
		log.Printf("⚠️ Attempted to send message to client %s with nil connection", c.ID)
		return
	}

	// Set write deadline to prevent hanging
	c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	err := c.Conn.WriteJSON(message)
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Printf("🔌 Client %s connection closed while sending message: %v", c.ID, err)
		} else {
			log.Printf("❌ Error sending message to client %s: %v", c.ID, err)
		}
	}
}

// SendError sends an error message to the client
func (c *Client) SendError(errorMsg string) {
	message := Message{
		ID:        uuid.New().String(),
		Event:     "error",
		Data:      map[string]string{"error": errorMsg},
		Timestamp: time.Now(),
	}
	c.SendMessage(message)
}

// dispatchToLaravel sends client message to Laravel for processing
func (s *Server) dispatchToLaravel(message Message, client *Client) {
	payloadFile, err := s.createTempPayloadFile(message, client, "ClientMessageReceived")
	if err != nil {
		log.Printf("Error creating temp payload file: %v", err)
		return
	}

	// Let's log the full command that will be executed
	cmdString := fmt.Sprintf("%s artisan %s --payload %s", s.phpBinary, s.laravelCmd, payloadFile)
	log.Printf("🚀 Executing Laravel command: %s", cmdString)

	// Execute Laravel artisan command with payload file
	cmd := exec.Command(s.phpBinary, "artisan", s.laravelCmd, "--payload", payloadFile)
	cmd.Dir = s.workingDir // Set Laravel project path

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing Laravel command '%s': %v, Output: %s", s.laravelCmd, err, string(output))
	} else {
		log.Printf("Laravel command '%s' executed successfully: %s", s.laravelCmd, string(output))
	}
}

// dispatchAuthenticationToLaravel sends authentication events to Laravel
func (s *Server) dispatchAuthenticationToLaravel(client *Client, status string, token string) {
	// Create standardized authentication payload following the desired format
	standardizedPayload := map[string]interface{}{
		"message_id": uuid.New().String(),
		"timestamp":  time.Now().Format(time.RFC3339),
		"action":     "client_authentication",
		"client": map[string]interface{}{
			"id":       client.ID,
			"type":     "websocket",
			"version":  "1.0.0",
			"user_id":  client.UserID,
			"username": client.Username,
			"email":    client.Email,
		},
		"data": map[string]interface{}{
			"authentication_status": status,
			"token_provided":        token != "",
		},
	}

	payloadFile, err := s.createTempPayloadFileFromData(standardizedPayload)
	if err != nil {
		log.Printf("Error creating temp authentication payload file: %v", err)
		return
	}

	// Log the full command that will be executed
	cmdString := fmt.Sprintf("%s artisan %s --payload %s", s.phpBinary, s.laravelCmd, payloadFile)
	log.Printf("🚀 Executing Laravel authentication command: %s", cmdString)

	// Execute Laravel artisan command with payload file
	cmd := exec.Command(s.phpBinary, "artisan", s.laravelCmd, "--payload", payloadFile)
	cmd.Dir = s.workingDir // Set Laravel project path

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing Laravel authentication command '%s': %v, Output: %s", s.laravelCmd, err, string(output))
	} else {
		log.Printf("Laravel authentication event '%s' executed successfully: %s", s.laravelCmd, string(output))
	}
}

// initializeTempDirectory ensures the temp directory exists with proper permissions
func (s *Server) initializeTempDirectory() {
	if s.tempDir == "" {
		s.tempDir = filepath.Join(os.TempDir(), "socket-server-payloads")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(s.tempDir, 0755); err != nil {
		log.Printf("Error creating temp directory %s: %v", s.tempDir, err)
		return
	}

	log.Printf("Temp directory initialized: %s", s.tempDir)
}

// createTempPayloadFile creates a temporary file with message data
func (s *Server) createTempPayloadFile(message Message, client *Client, eventType string) (string, error) {
	// Create standardized message payload following the desired format
	standardizedPayload := map[string]interface{}{
		"message_id": uuid.New().String(),
		"timestamp":  time.Now().Format(time.RFC3339),
		"action":     eventType,
		"client": map[string]interface{}{
			"id":          client.ID,
			"user_id":     client.UserID,
			"username":    client.Username,
			"remote_addr": client.RemoteAddr,
		},
		"auth": map[string]interface{}{
			"user_id":    client.UserID,
			"user_email": "", // Can be populated from JWT claims if available
			"logged_at":  time.Now().Format(time.RFC3339),
		},
		"data": map[string]interface{}{
			"id":        message.ID,
			"channel":   message.Channel,
			"event":     message.Event,
			"data":      message.Data,
			"timestamp": message.Timestamp,
		},
	}

	return s.createTempPayloadFileFromData(standardizedPayload)
}

// createTempPayloadFileFromData creates a temporary file with the given data
func (s *Server) createTempPayloadFileFromData(data interface{}) (string, error) {
	// Convert to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error marshaling payload data: %v", err)
	}

	// Create filename with timestamp for expiration tracking
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("payload_%d_%s.json", timestamp, uuid.New().String()[:8])
	filepath := filepath.Join(s.tempDir, filename)

	// Write file with permissions readable by Laravel (0644)
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("error writing payload file: %v", err)
	}

	log.Printf("Created temp payload file: %s", filepath)
	return filepath, nil
}

// startCleanupRoutine starts a background routine to clean up expired temp files
func (s *Server) startCleanupRoutine() {
	go func() {
		// Run cleanup every hour
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		// Run initial cleanup
		s.cleanupExpiredFiles()

		for {
			select {
			case <-ticker.C:
				s.cleanupExpiredFiles()
			}
		}
	}()

	log.Printf("Started temp file cleanup routine (runs every hour)")
}

// cleanupExpiredFiles removes temp files older than 24 hours
func (s *Server) cleanupExpiredFiles() {
	expireTime := time.Now().Add(-24 * time.Hour) // 24 hours ago

	files, err := os.ReadDir(s.tempDir)
	if err != nil {
		log.Printf("Error reading temp directory: %v", err)
		return
	}

	cleaned := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Check if file follows our naming pattern
		if !strings.HasPrefix(file.Name(), "payload_") || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(s.tempDir, file.Name())
		info, err := file.Info()
		if err != nil {
			log.Printf("Error getting file info for %s: %v", filePath, err)
			continue
		}

		// Remove if older than 24 hours
		if info.ModTime().Before(expireTime) {
			if err := os.Remove(filePath); err != nil {
				log.Printf("Error removing expired file %s: %v", filePath, err)
			} else {
				cleaned++
				log.Printf("Removed expired temp file: %s", filePath)
			}
		}
	}

	if cleaned > 0 {
		log.Printf("Cleaned up %d expired temp files", cleaned)
	}
}

// REST API Handlers

func (s *Server) handleGetClients(w http.ResponseWriter, r *http.Request) {
	s.mutex.RLock()
	clients := make([]*Client, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	s.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clients": clients,
		"total":   len(clients),
	})
}

func (s *Server) handleGetChannels(w http.ResponseWriter, r *http.Request) {
	s.mutex.RLock()
	channels := make(map[string]interface{})
	for name, channel := range s.channels {
		channel.mutex.RLock()
		channels[name] = map[string]interface{}{
			"name":         channel.Name,
			"is_private":   channel.IsPrivate,
			"require_auth": channel.RequireAuth,
			"client_count": len(channel.Clients),
			"created_at":   channel.CreatedAt,
		}
		channel.mutex.RUnlock()
	}
	s.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channels)
}

func (s *Server) handleGetChannelClients(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelName := vars["channel"]

	s.mutex.RLock()
	channel, exists := s.channels[channelName]
	s.mutex.RUnlock()

	if !exists {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	channel.mutex.RLock()
	clients := make([]*Client, 0, len(channel.Clients))
	for _, client := range channel.Clients {
		clients = append(clients, client)
	}
	channel.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"channel": channelName,
		"clients": clients,
		"total":   len(clients),
	})
}

func (s *Server) handleKickClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["client"]

	s.mutex.RLock()
	client, exists := s.clients[clientID]
	s.mutex.RUnlock()

	if !exists {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	client.SendMessage(Message{
		ID:        uuid.New().String(),
		Event:     "kicked",
		Data:      map[string]string{"reason": "Kicked by admin"},
		Timestamp: time.Now(),
	})

	client.Conn.Close()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Client %s kicked", clientID),
	})
}

func (s *Server) handleBroadcast(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Channel string      `json:"channel"`
		Event   string      `json:"event"`
		Data    interface{} `json:"data"`
	}

	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if payload.Channel == "" {
		http.Error(w, "Channel is required", http.StatusBadRequest)
		return
	}

	if payload.Event == "" {
		payload.Event = "broadcast"
	}

	message := Message{
		ID:        uuid.New().String(),
		Channel:   payload.Channel,
		Event:     payload.Event,
		Data:      payload.Data,
		Timestamp: time.Now(),
	}

	s.BroadcastToChannel(payload.Channel, message)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Message broadcasted",
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.mutex.RLock()
	clientCount := len(s.clients)
	channelCount := len(s.channels)
	s.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "healthy",
		"clients":  clientCount,
		"channels": channelCount,
		"uptime":   time.Since(time.Now()).String(),
		"version":  "1.0.0",
	})
}

var (
	port       string
	jwtSecret  string
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
	Run: func(cmd *cobra.Command, args []string) {
		// Use command-line flags first, then fall back to environment variables
		if port == "" {
			port = os.Getenv("SOCKET_PORT")
			if port == "" {
				port = "8080"
			}
		}

		if jwtSecret == "" {
			jwtSecret = os.Getenv("JWT_SECRET")
			if jwtSecret == "" {
				jwtSecret = "default-secret-key-change-in-production"
			}
		}

		if workingDir == "" {
			workingDir = os.Getenv("LARAVEL_PATH")
			if workingDir == "" {
				workingDir = "." // Current directory as fallback
			}
		}

		if phpBinary == "" {
			phpBinary = os.Getenv("PHP_BINARY")
			if phpBinary == "" {
				phpBinary = "php" // Default PHP binary
			}
		}

		if laravelCmd == "" {
			laravelCmd = os.Getenv("LARAVEL_COMMAND")
			if laravelCmd == "" {
				laravelCmd = "ns:socket-handler" // Default Laravel command
			}
		}

		if tempDir == "" {
			tempDir = os.Getenv("SOCKET_TEMP_DIR")
			if tempDir == "" {
				tempDir = filepath.Join(os.TempDir(), "socket-server-payloads") // Default temp directory
			}
		}

		fmt.Printf("Starting Socket Server on port %s\n", port)

		// Safely display JWT secret (first few characters)
		secretDisplay := jwtSecret
		if len(secretDisplay) > 10 {
			secretDisplay = secretDisplay[:10] + "..."
		} else if len(secretDisplay) > 3 {
			secretDisplay = secretDisplay[:3] + "..."
		}
		fmt.Printf("JWT Secret: %s\n", secretDisplay)
		fmt.Printf("Working Directory: %s\n", workingDir)
		fmt.Printf("PHP Binary: %s\n", phpBinary)
		fmt.Printf("Laravel Command: %s\n", laravelCmd)
		fmt.Printf("Temp Directory: %s\n", tempDir)

		server := NewServer(port, jwtSecret, workingDir, phpBinary, laravelCmd, tempDir)
		server.Start()
	},
}

func init() {
	rootCmd.Flags().StringVarP(&port, "port", "p", "", "Port to run the server on (default: 8080 or SOCKET_PORT env var)")
	rootCmd.Flags().StringVarP(&jwtSecret, "token", "t", "", "JWT secret for authentication (default: JWT_SECRET env var)")
	rootCmd.Flags().StringVarP(&workingDir, "dir", "d", "", "Working directory for Laravel commands (default: LARAVEL_PATH env var)")
	rootCmd.Flags().StringVar(&phpBinary, "php", "", "PHP binary path (default: 'php' or PHP_BINARY env var)")
	rootCmd.Flags().StringVar(&laravelCmd, "command", "", "Laravel artisan command to execute (default: 'ns:socket-handler' or LARAVEL_COMMAND env var)")
	rootCmd.Flags().StringVar(&tempDir, "temp", "", "Temporary directory for payload files (default: system temp/socket-server-payloads or SOCKET_TEMP_DIR env var)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// handleClientPing manages ping/pong for connection health
func (s *Server) handleClientPing(client *Client, pingTicker *time.Ticker, done chan bool) {
	defer func() {
		log.Printf("Client %s ping handler exiting", client.ID)
		done <- true
	}()

	for range pingTicker.C {
		// Send ping to client
		client.mutex.Lock()
		if client.Conn != nil {
			err := client.Conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Printf("Failed to send ping to client %s: %v", client.ID, err)
				client.mutex.Unlock()
				return
			}
			log.Printf("📍 Sent ping to client %s", client.ID)
		}
		client.mutex.Unlock()
	}
}
