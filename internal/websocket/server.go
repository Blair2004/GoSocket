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
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
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

	client := &models.Client{
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

	// Handle client disconnection
	defer func() {
		s.disconnectClient(client)
		conn.Close()
	}()

	// Wait for either handler to finish
	<-done
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
	client.Conn.Close()

	return nil
}

// BroadcastToChannel sends a message to all clients in a channel
func (s *Server) BroadcastToChannel(channelName string, message models.Message) {
	channel, exists := s.GetChannel(channelName)
	if !exists {
		s.logger.Warn("Channel %s not found for broadcast", channelName)
		return
	}

	clients := channel.GetClients()
	for _, client := range clients {
		if err := client.SendMessage(message); err != nil {
			s.logger.Error("Failed to send message to client %s: %v", client.ID, err)
		}
	}

	s.logger.Info("Broadcasted message to %d clients in channel %s", len(clients), channelName)
}
