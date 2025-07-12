package websocket

import (
	"time"

	"github.com/google/uuid"

	"socket-server/internal/models"
)

// handleClientMessages processes messages from a client
func (s *Server) handleClientMessages(client *models.Client, done chan bool) {
	defer func() {
		s.logger.Debug("Client %s message handler exiting", client.ID)
		done <- true
	}()

	for {
		var msg map[string]interface{}
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			s.logger.WebSocketError(client.ID, err)
			break
		}

		// Reset read deadline on successful message
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		client.LastSeen = time.Now()

		// Log incoming message
		actionStr := "unknown"
		if action, ok := msg["action"].(string); ok {
			actionStr = action
		}

		s.logger.MessageReceived(client.ID, client.Username, actionStr, msg)

		// Handle different message types
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
			s.handlePing(client)
		default:
			s.handleMessage(client, msg)
		}
	}
}

func (s *Server) handleMessage(client *models.Client, msg map[string]interface{}) {
	// Forward unsupported messages to Laravel
	s.logger.Debug("Forwarding unsupported message to Laravel from client %s", client.ID)

	// Convert raw message to models.Message
	message := models.Message{
		ID:        uuid.New().String(),
		Event:     getStringFromMap(msg, "action", "unknown"),
		Channel:   getStringFromMap(msg, "channel", ""),
		Data:      msg,
		UserID:    client.UserID,
		Username:  client.Username,
		Timestamp: time.Now(),
	}

	s.laravelSvc.DispatchMessage(message, client)
}

// getStringFromMap safely extracts a string value from a map
func getStringFromMap(m map[string]interface{}, key string, defaultValue string) string {
	if value, ok := m[key].(string); ok {
		return value
	}
	return defaultValue
}

// handleClientPing manages ping/pong for connection health
func (s *Server) handleClientPing(client *models.Client, pingTicker *time.Ticker, done chan bool) {
	defer func() {
		s.logger.Debug("Client %s ping handler exiting", client.ID)
		done <- true
	}()

	for range pingTicker.C {
		// Send ping to client
		err := client.SendPing()
		if err != nil {
			s.logger.Error("Failed to send ping to client %s: %v", client.ID, err)
			return
		}
		s.logger.PingSent(client.ID)
	}
}

// handleAuthentication processes client authentication
func (s *Server) handleAuthentication(client *models.Client, msg map[string]interface{}) {
	tokenStr, ok := msg["token"].(string)
	if !ok {
		s.logger.Error("Client %s sent invalid token format", client.ID)
		s.sendError(client, "Invalid token format")
		return
	}

	s.logger.Debug("Client %s attempting JWT authentication", client.ID)

	claims, err := s.authService.ValidateToken(tokenStr)
	if err != nil {
		s.logger.ClientAuthenticationFailed(client.ID, err)
		s.sendError(client, "Invalid token")
		s.laravelSvc.DispatchAuthentication(client, "failed", tokenStr)
		return
	}

	// Extract user info from claims
	userID, username, email := s.authService.ExtractUserInfo(claims)
	client.SetUserInfo(userID, username, email)

	s.logger.ClientAuthenticated(client.ID, client.Username, client.UserID)
	s.laravelSvc.DispatchAuthentication(client, "success", tokenStr)
}

// handleJoinChannel adds client to a channel
func (s *Server) handleJoinChannel(client *models.Client, msg map[string]interface{}) {
	channelName, ok := msg["channel"].(string)
	if !ok {
		s.logger.Error("Client %s sent invalid channel name for join", client.ID)
		s.sendError(client, "Invalid channel name")
		return
	}

	s.logger.Debug("Client %s (%s) attempting to join channel '%s'", client.ID, client.Username, channelName)

	// Get or create channel
	channel := s.getOrCreateChannel(channelName)

	// Check if channel requires authentication
	if channel.RequireAuth && client.UserID == "" {
		s.logger.Warn("Client %s denied access to channel '%s': authentication required", client.ID, channelName)
		s.sendError(client, "Channel requires authentication")
		return
	}

	// Add client to channel
	channel.AddClient(client)
	client.AddToChannel(channelName)

	s.logger.ChannelJoined(client.ID, client.Username, channelName)

	// Send confirmation
	confirmation := models.Message{
		ID:        uuid.New().String(),
		Event:     "joined_channel",
		Data:      map[string]string{"channel": channelName},
		Timestamp: time.Now(),
	}
	client.SendMessage(confirmation)
}

// handleLeaveChannel removes client from a channel
func (s *Server) handleLeaveChannel(client *models.Client, msg map[string]interface{}) {
	channelName, ok := msg["channel"].(string)
	if !ok {
		s.logger.Error("Client %s sent invalid channel name for leave", client.ID)
		s.sendError(client, "Invalid channel name")
		return
	}

	s.logger.Debug("Client %s (%s) attempting to leave channel '%s'", client.ID, client.Username, channelName)

	channel, exists := s.GetChannel(channelName)
	if !exists {
		s.logger.Error("Client %s tried to leave non-existent channel '%s'", client.ID, channelName)
		s.sendError(client, "Channel not found")
		return
	}

	// Remove client from channel
	channel.RemoveClient(client.ID)
	client.RemoveFromChannel(channelName)

	s.logger.ChannelLeft(client.ID, client.Username, channelName)

	// Send confirmation
	confirmation := models.Message{
		ID:        uuid.New().String(),
		Event:     "left_channel",
		Data:      map[string]string{"channel": channelName},
		Timestamp: time.Now(),
	}
	client.SendMessage(confirmation)
}

// handleSendMessage processes messages sent by clients
func (s *Server) handleSendMessage(client *models.Client, msg map[string]interface{}) {
	channelName, ok := msg["channel"].(string)
	if !ok {
		s.logger.Error("Client %s sent message with invalid channel name", client.ID)
		s.sendError(client, "Invalid channel name")
		return
	}

	event, ok := msg["event"].(string)
	if !ok {
		event = "message"
	}

	data := msg["data"]

	s.logger.MessageSent(client.ID, client.Username, channelName, event, data)

	message := models.Message{
		ID:        uuid.New().String(),
		Channel:   channelName,
		Event:     event,
		Data:      data,
		UserID:    client.UserID,
		Username:  client.Username,
		Timestamp: time.Now(),
	}

	// Dispatch to Laravel if configured
	if err := s.laravelSvc.DispatchMessage(message, client); err != nil {
		s.logger.Error("Failed to dispatch message to Laravel: %v", err)
	}

	// Broadcast to all clients in channel
	s.BroadcastToChannel(channelName, message)
}

// handlePing processes ping messages
func (s *Server) handlePing(client *models.Client) {
	s.logger.PongReceived(client.ID)
	pong := models.Message{
		ID:        uuid.New().String(),
		Event:     "pong",
		Timestamp: time.Now(),
	}
	client.SendMessage(pong)
}

// disconnectClient removes a client from the server
func (s *Server) disconnectClient(client *models.Client) {
	s.logger.ClientDisconnected(client.ID, client.Username, client.RemoteAddr)

	s.mutex.Lock()
	delete(s.clients, client.ID)
	s.mutex.Unlock()

	// Remove client from all channels
	channels := client.GetChannels()
	for channelName := range channels {
		if channel, exists := s.GetChannel(channelName); exists {
			channel.RemoveClient(client.ID)
		}
	}
}

// getOrCreateChannel gets an existing channel or creates a new one
func (s *Server) getOrCreateChannel(channelName string) *models.Channel {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	channel, exists := s.channels[channelName]
	if !exists {
		s.logger.Debug("Creating new channel '%s'", channelName)
		channel = &models.Channel{
			Name:        channelName,
			Clients:     make(map[string]*models.Client),
			IsPrivate:   false,
			RequireAuth: false,
			CreatedAt:   time.Now(),
		}
		s.channels[channelName] = channel
	}

	return channel
}

// sendError sends an error message to a client
func (s *Server) sendError(client *models.Client, errorMsg string) {
	message := models.Message{
		ID:        uuid.New().String(),
		Event:     "error",
		Data:      map[string]string{"error": errorMsg},
		Timestamp: time.Now(),
	}
	client.SendMessage(message)
}
