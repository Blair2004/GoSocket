package websocket

import (
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"socket-server/internal/auth"
	"socket-server/internal/models"
	"socket-server/internal/services"
	"socket-server/pkg/logger"
)

// Server manages WebSocket connections and channels
type Server struct {
	clients     map[string]*models.Client
	channels    map[string]*models.Channel
	upgrader    websocket.Upgrader
	authService *auth.Service
	laravelSvc  *services.LaravelService
	logger      *logger.Logger
	mutex       sync.RWMutex
}

// New creates a new WebSocket server
func New(authService *auth.Service, laravelSvc *services.LaravelService, logger *logger.Logger) *Server {
	return &Server{
		clients:     make(map[string]*models.Client),
		channels:    make(map[string]*models.Channel),
		authService: authService,
		laravelSvc:  laravelSvc,
		logger:      logger,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
			ReadBufferSize:    4096, // Increased from 1024
			WriteBufferSize:   4096, // Increased from 1024
			EnableCompression: true, // Enable compression for better performance
		},
	}
}

// HandleConnection handles a new WebSocket connection
func (s *Server) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("WebSocket upgrade error: %v", err)
		return
	}

	// Set connection timeouts and limits
	conn.SetReadLimit(512 * 1024) // 512KB max message size
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	client := models.NewClient(uuid.New().String(), conn)
	client.RemoteAddr = r.RemoteAddr
	client.UserAgent = r.UserAgent()

	s.mutex.Lock()
	s.clients[client.ID] = client
	s.mutex.Unlock()

	s.logger.ClientConnected(client.ID, client.RemoteAddr, client.UserAgent)

	// Send welcome message
	welcome := models.Message{
		ID:        uuid.New().String(),
		Event:     "connected",
		Data:      map[string]string{"client_id": client.ID},
		Timestamp: time.Now(),
	}
	client.SendMessage(welcome)

	// Start ping ticker for connection health
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Handle client messages and ping in separate goroutines
	done := make(chan bool, 2)
	go s.handleClientMessages(client, done)
	go s.handleClientPing(client, pingTicker, done)

	// Wait for either handler to finish
	<-done

	// Handle client disconnection - this happens after goroutines finish
	s.disconnectClient(client)
}

// GetClients returns all connected clients
func (s *Server) GetClients() map[string]*models.Client {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	clients := make(map[string]*models.Client)
	for k, v := range s.clients {
		clients[k] = v
	}
	return clients
}

// GetChannels returns all channels
func (s *Server) GetChannels() map[string]*models.Channel {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	channels := make(map[string]*models.Channel)
	for k, v := range s.channels {
		channels[k] = v
	}
	return channels
}

// GetClient returns a specific client
func (s *Server) GetClient(clientID string) (*models.Client, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	client, exists := s.clients[clientID]
	return client, exists
}

// GetChannel returns a specific channel
func (s *Server) GetChannel(channelName string) (*models.Channel, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	channel, exists := s.channels[channelName]
	return channel, exists
}

// KickClient forcefully disconnects a client
func (s *Server) KickClient(clientID string) error {
	client, exists := s.GetClient(clientID)
	if !exists {
		return models.ErrClientNotFound
	}

	// Send kick message
	kickMessage := models.Message{
		ID:        uuid.New().String(),
		Event:     "kicked",
		Data:      map[string]string{"reason": "Kicked by admin"},
		Timestamp: time.Now(),
	}
	client.SendMessage(kickMessage)

	// Close connection
	client.Close()

	return nil
}

// BroadcastToChannel sends a message to all clients in a channel
func (s *Server) BroadcastToChannel(channelName string, message models.Message) {
	start := time.Now()
	s.logger.Info("📺 BroadcastToChannel started for channel: %s", channelName)

	lookupStart := time.Now()
	channel, exists := s.GetChannel(channelName)
	if !exists {
		s.logger.Warn("Channel %s not found for broadcast", channelName)
		return
	}
	lookupTime := time.Since(lookupStart)
	s.logger.Info("⏱️ Channel lookup took: %v", lookupTime)

	clientsStart := time.Now()
	clients := channel.GetClients()
	clientsTime := time.Since(clientsStart)
	s.logger.Info("⏱️ Getting clients took: %v", clientsTime)

	sendStart := time.Now()

	// Use goroutines for non-blocking sends with timeout
	type clientResult struct {
		clientID string
		err      error
		duration time.Duration
	}

	results := make(chan clientResult, len(clients))

	// Send to all clients concurrently
	for _, client := range clients {
		go func(c *models.Client) {
			clientStart := time.Now()
			err := c.SendMessage(message)
			results <- clientResult{
				clientID: c.ID,
				err:      err,
				duration: time.Since(clientStart),
			}
		}(client)
	}

	// Collect results with timeout
	successCount := 0
	timeout := time.After(1 * time.Second) // Max 1 second for all sends in local env

collectLoop:
	for i := 0; i < len(clients); i++ {
		select {
		case result := <-results:
			if result.err != nil {
				s.logger.Error("Failed to send message to client %s: %v", result.clientID, result.err)
			} else {
				successCount++
			}
			if result.duration > 10*time.Millisecond {
				s.logger.Warn("⚠️ Slow client send to %s took: %v", result.clientID, result.duration)
			}
		case <-timeout:
			s.logger.Warn("⏰ Broadcast timeout - %d/%d clients completed", i, len(clients))
			break collectLoop
		}
	}

	sendTime := time.Since(sendStart)
	s.logger.Info("⏱️ Concurrent sending to %d clients took: %v (success: %d)", len(clients), sendTime, successCount)

	// After collecting results, remove clients that consistently failed
	go func() {
		for i := 0; i < len(clients); i++ {
			select {
			case result := <-results:
				// If a client took too long, it's likely dead - remove it
				if result.duration > 500*time.Millisecond && result.err != nil {
					s.logger.Info("🗑️ Removing slow/dead client: %s (took %v)", result.clientID, result.duration)
					s.mutex.Lock()
					delete(s.clients, result.clientID)
					s.mutex.Unlock()
				}
			case <-time.After(100 * time.Millisecond):
				// Any remaining clients are likely dead
				return
			}
		}
	}()

	totalTime := time.Since(start)
	s.logger.Info("🏁 BroadcastToChannel total time: %v", totalTime)
	s.logger.Info("Broadcasted message to %d clients in channel %s", len(clients), channelName)
}

// BroadcastToAll sends a message to all connected clients
func (s *Server) BroadcastToAll(message models.Message) {
	start := time.Now()
	s.logger.Info("🌍 BroadcastToAll started")

	lockStart := time.Now()
	s.mutex.RLock()
	clients := make([]*models.Client, 0, len(s.clients))
	for _, client := range s.clients {
		clients = append(clients, client)
	}
	s.mutex.RUnlock()
	lockTime := time.Since(lockStart)
	s.logger.Info("⏱️ Client collection took: %v", lockTime)

	sendStart := time.Now()

	// Use goroutines for non-blocking sends with timeout
	type clientResult struct {
		clientID string
		err      error
		duration time.Duration
	}

	results := make(chan clientResult, len(clients))

	// Send to all clients concurrently
	for _, client := range clients {
		s.logger.Info("🎇 Starting goroutine for client %s", client.ID)
		go func(c *models.Client, s *Server) {
			s.logger.Info("🏤 Sending message to client %s", c.ID)
			clientStart := time.Now()
			err := c.SendMessage(message)
			results <- clientResult{
				clientID: c.ID,
				err:      err,
				duration: time.Since(clientStart),
			}
			s.logger.Info("💌 Sent message to client %s", c.ID)
		}(client, s)
		s.logger.Info("🎆 Ended goroutine for client %s", client.ID)
	}

	// Collect results with timeout
	successCount := 0
	timeout := time.After(1 * time.Second) // Max 1 second for all sends in local env

collectLoop:
	for i := 0; i < len(clients); i++ {
		select {
		case result := <-results:
			if result.err != nil {
				s.logger.Error("Failed to send message to client %s: %v", result.clientID, result.err)
			} else {
				successCount++
			}
			if result.duration > 10*time.Millisecond {
				s.logger.Warn("⚠️ Slow global client send to %s took: %v", result.clientID, result.duration)
			}
		case <-timeout:
			s.logger.Warn("⏰ Global broadcast timeout - %d/%d clients completed", i, len(clients))
			break collectLoop
		}
	}

	sendTime := time.Since(sendStart)
	s.logger.Info("⏱️ Concurrent global sending to %d clients took: %v (success: %d)", len(clients), sendTime, successCount)

	// After collecting results, remove clients that consistently failed
	go func() {
		for i := 0; i < len(clients); i++ {
			select {
			case result := <-results:
				// If a client took too long, it's likely dead - remove it
				if result.duration > 500*time.Millisecond && result.err != nil {
					s.logger.Info("🗑️ Removing slow/dead client: %s (took %v)", result.clientID, result.duration)
					s.mutex.Lock()
					delete(s.clients, result.clientID)
					s.mutex.Unlock()
				}
			case <-time.After(100 * time.Millisecond):
				// Any remaining clients are likely dead
				return
			}
		}
	}()

	totalTime := time.Since(start)
	s.logger.Info("🏁 BroadcastToAll total time: %v", totalTime)
	s.logger.Info("Broadcasted message to %d/%d clients globally", successCount, len(clients))
}

// BroadcastToAuthenticated sends a message to all authenticated clients
func (s *Server) BroadcastToAuthenticated(message models.Message) {
	start := time.Now()
	s.logger.Info("🔐 BroadcastToAuthenticated started")

	lockStart := time.Now()
	s.mutex.RLock()
	clients := make([]*models.Client, 0)
	for _, client := range s.clients {
		if client.UserID != "" {
			clients = append(clients, client)
		}
	}
	s.mutex.RUnlock()
	lockTime := time.Since(lockStart)
	s.logger.Info("⏱️ Authenticated client collection took: %v", lockTime)

	sendStart := time.Now()

	// Use goroutines for non-blocking sends with timeout
	type clientResult struct {
		clientID string
		err      error
		duration time.Duration
	}

	results := make(chan clientResult, len(clients))

	// Send to all clients concurrently
	for _, client := range clients {
		go func(c *models.Client) {
			clientStart := time.Now()
			err := c.SendMessage(message)
			results <- clientResult{
				clientID: c.ID,
				err:      err,
				duration: time.Since(clientStart),
			}
		}(client)
	}

	// Collect results with timeout
	successCount := 0
	timeout := time.After(1 * time.Second) // Max 1 second for all sends in local env

collectLoop:
	for i := 0; i < len(clients); i++ {
		select {
		case result := <-results:
			if result.err != nil {
				s.logger.Error("Failed to send message to authenticated client %s: %v", result.clientID, result.err)
			} else {
				successCount++
			}
			if result.duration > 10*time.Millisecond {
				s.logger.Warn("⚠️ Slow authenticated client send to %s took: %v", result.clientID, result.duration)
			}
		case <-timeout:
			s.logger.Warn("⏰ Authenticated broadcast timeout - %d/%d clients completed", i, len(clients))
			break collectLoop
		}
	}

	sendTime := time.Since(sendStart)
	s.logger.Info("⏱️ Concurrent authenticated sending to %d clients took: %v (success: %d)", len(clients), sendTime, successCount)

	// After collecting results, remove clients that consistently failed
	go func() {
		for i := 0; i < len(clients); i++ {
			select {
			case result := <-results:
				// If a client took too long, it's likely dead - remove it
				if result.duration > 500*time.Millisecond && result.err != nil {
					s.logger.Info("🗑️ Removing slow/dead client: %s (took %v)", result.clientID, result.duration)
					s.mutex.Lock()
					delete(s.clients, result.clientID)
					s.mutex.Unlock()
				}
			case <-time.After(100 * time.Millisecond):
				// Any remaining clients are likely dead
				return
			}
		}
	}()

	totalTime := time.Since(start)
	s.logger.Info("🏁 BroadcastToAuthenticated total time: %v", totalTime)
	s.logger.Info("Broadcasted message to %d authenticated clients", successCount)
}

// BroadcastToUser sends a message to all connections of a specific user
func (s *Server) BroadcastToUser(userID string, message models.Message) {
	s.mutex.RLock()
	clients := make([]*models.Client, 0)
	for _, client := range s.clients {
		if client.UserID == userID {
			clients = append(clients, client)
		}
	}
	s.mutex.RUnlock()

	successCount := 0
	for _, client := range clients {
		if err := client.SendMessage(message); err != nil {
			s.logger.Error("Failed to send message to user %s client %s: %v", userID, client.ID, err)
		} else {
			successCount++
		}
	}

	s.logger.Info("Broadcasted message to %d connections of user %s", successCount, userID)
}

// BroadcastToUsersExcept sends a message to all authenticated clients except the specified user
func (s *Server) BroadcastToUsersExcept(excludeUserID string, message models.Message) {
	s.mutex.RLock()
	clients := make([]*models.Client, 0)
	for _, client := range s.clients {
		if client.UserID != "" && client.UserID != excludeUserID {
			clients = append(clients, client)
		}
	}
	s.mutex.RUnlock()

	successCount := 0
	for _, client := range clients {
		if err := client.SendMessage(message); err != nil {
			s.logger.Error("Failed to send message to client %s: %v", client.ID, err)
		} else {
			successCount++
		}
	}

	s.logger.Info("Broadcasted message to %d authenticated clients (excluding user %s)", successCount, excludeUserID)
}

// BroadcastToClient sends a message to a specific client connection
func (s *Server) BroadcastToClient(clientID string, message models.Message) error {
	s.mutex.RLock()
	client, exists := s.clients[clientID]
	s.mutex.RUnlock()

	if !exists {
		return models.ErrClientNotFound
	}

	if err := client.SendMessage(message); err != nil {
		s.logger.Error("Failed to send message to client %s: %v", clientID, err)
		return err
	}

	s.logger.Info("Sent message to client %s", clientID)
	return nil
}

// cleanupDeadConnections removes clients that consistently fail to receive messages
func (s *Server) cleanupDeadConnections() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var deadClients []string

	for clientID, client := range s.clients {
		if !client.IsConnected() {
			deadClients = append(deadClients, clientID)
		}
	}

	for _, clientID := range deadClients {
		s.logger.Info("🧹 Cleaning up dead connection: %s", clientID)
		delete(s.clients, clientID)
	}

	if len(deadClients) > 0 {
		s.logger.Info("🧹 Cleaned up %d dead connections", len(deadClients))
	}
}
