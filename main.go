package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/exec"
    "sync"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "github.com/gorilla/mux"
    "github.com/gorilla/websocket"
)

// Client represents a connected WebSocket client
type Client struct {
    ID         string          `json:"id"`
    Conn       *websocket.Conn `json:"-"`
    UserID     string          `json:"user_id,omitempty"`
    Username   string          `json:"username,omitempty"`
    Channels   map[string]bool `json:"channels"`
    LastSeen   time.Time       `json:"last_seen"`
    RemoteAddr string          `json:"remote_addr"`
    UserAgent  string          `json:"user_agent"`
    mutex      sync.RWMutex    `json:"-"`
}

// Channel represents a communication channel
type Channel struct {
    Name        string            `json:"name"`
    Clients     map[string]*Client `json:"-"`
    IsPrivate   bool              `json:"is_private"`
    RequireAuth bool              `json:"require_auth"`
    CreatedAt   time.Time         `json:"created_at"`
    mutex       sync.RWMutex      `json:"-"`
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
    mutex      sync.RWMutex
}

// NewServer creates a new socket server instance
func NewServer(port string, jwtSecret string) *Server {
    return &Server{
        clients:   make(map[string]*Client),
        channels:  make(map[string]*Channel),
        jwtSecret: []byte(jwtSecret),
        port:      port,
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool {
                return true // Allow all origins for now
            },
        },
    }
}

// Start starts the socket server
func (s *Server) Start() {
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
    
    log.Printf("Client connected: %s from %s", client.ID, client.RemoteAddr)
    
    // Send welcome message
    welcome := Message{
        ID:        uuid.New().String(),
        Event:     "connected",
        Data:      map[string]string{"client_id": client.ID},
        Timestamp: time.Now(),
    }
    client.SendMessage(welcome)
    
    // Handle client messages
    go s.handleClientMessages(client)
    
    // Handle client disconnection
    defer func() {
        s.disconnectClient(client)
        conn.Close()
    }()
    
    // Keep connection alive
    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            log.Printf("Client %s disconnected: %v", client.ID, err)
            break
        }
        client.LastSeen = time.Now()
    }
}

// handleClientMessages processes messages from a client
func (s *Server) handleClientMessages(client *Client) {
    for {
        var msg map[string]interface{}
        err := client.Conn.ReadJSON(&msg)
        if err != nil {
            log.Printf("Error reading message from client %s: %v", client.ID, err)
            break
        }
        
        client.LastSeen = time.Now()
        
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
            client.SendMessage(Message{
                ID:        uuid.New().String(),
                Event:     "pong",
                Timestamp: time.Now(),
            })
        }
    }
}

// handleAuthentication processes client authentication
func (s *Server) handleAuthentication(client *Client, msg map[string]interface{}) {
    tokenStr, ok := msg["token"].(string)
    if !ok {
        client.SendError("Invalid token format")
        return
    }
    
    token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
        return s.jwtSecret, nil
    })
    
    if err != nil || !token.Valid {
        client.SendError("Invalid token")
        return
    }
    
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        client.SendError("Invalid token claims")
        return
    }
    
    client.mutex.Lock()
    client.UserID = claims["user_id"].(string)
    if username, exists := claims["username"]; exists {
        client.Username = username.(string)
    }
    client.mutex.Unlock()
    
    client.SendMessage(Message{
        ID:        uuid.New().String(),
        Event:     "authenticated",
        Data:      map[string]string{"user_id": client.UserID, "username": client.Username},
        Timestamp: time.Now(),
    })
    
    log.Printf("Client %s authenticated as user %s", client.ID, client.UserID)
}

// handleJoinChannel adds client to a channel
func (s *Server) handleJoinChannel(client *Client, msg map[string]interface{}) {
    channelName, ok := msg["channel"].(string)
    if !ok {
        client.SendError("Invalid channel name")
        return
    }
    
    s.mutex.Lock()
    channel, exists := s.channels[channelName]
    if !exists {
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
        client.SendError("Channel requires authentication")
        return
    }
    
    channel.mutex.Lock()
    channel.Clients[client.ID] = client
    channel.mutex.Unlock()
    
    client.mutex.Lock()
    client.Channels[channelName] = true
    client.mutex.Unlock()
    
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
        client.SendError("Invalid channel name")
        return
    }
    
    s.mutex.RLock()
    channel, exists := s.channels[channelName]
    s.mutex.RUnlock()
    
    if !exists {
        client.SendError("Channel not found")
        return
    }
    
    channel.mutex.Lock()
    delete(channel.Clients, client.ID)
    channel.mutex.Unlock()
    
    client.mutex.Lock()
    delete(client.Channels, channelName)
    client.mutex.Unlock()
    
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
        client.SendError("Invalid channel name")
        return
    }
    
    event, ok := msg["event"].(string)
    if !ok {
        event = "message"
    }
    
    data := msg["data"]
    
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
    
    err := c.Conn.WriteJSON(message)
    if err != nil {
        log.Printf("Error sending message to client %s: %v", c.ID, err)
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
    // Create Laravel event payload
    laravelEvent := map[string]interface{}{
        "event_type":    "ClientMessageReceived",
        "socket_client": map[string]interface{}{
            "id":         client.ID,
            "user_id":    client.UserID,
            "username":   client.Username,
            "remote_addr": client.RemoteAddr,
        },
        "message": map[string]interface{}{
            "id":        message.ID,
            "channel":   message.Channel,
            "event":     message.Event,
            "data":      message.Data,
            "timestamp": message.Timestamp,
        },
    }
    
    // Save to temporary file
    tempFile := fmt.Sprintf("/tmp/socket_event_%s.json", message.ID)
    jsonData, err := json.MarshalIndent(laravelEvent, "", "  ")
    if err != nil {
        log.Printf("Error marshaling Laravel event: %v", err)
        return
    }
    
    err = os.WriteFile(tempFile, jsonData, 0644)
    if err != nil {
        log.Printf("Error writing Laravel event file: %v", err)
        return
    }
    
    // Execute Laravel artisan command to process the event
    cmd := exec.Command("php", "artisan", "socket:process-client-message", tempFile)
    cmd.Dir = os.Getenv("LARAVEL_PATH") // Set Laravel project path
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        log.Printf("Error executing Laravel command: %v, Output: %s", err, string(output))
    } else {
        log.Printf("Laravel event dispatched successfully: %s", string(output))
    }
    
    // Clean up temp file
    os.Remove(tempFile)
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
        "status":        "healthy",
        "clients":       clientCount,
        "channels":      channelCount,
        "uptime":        time.Since(time.Now()).String(),
        "version":       "1.0.0",
    })
}

func main() {
    port := os.Getenv("SOCKET_PORT")
    if port == "" {
        port = "8080"
    }
    
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        jwtSecret = "default-secret-key-change-in-production"
    }
    
    server := NewServer(port, jwtSecret)
    server.Start()
}
