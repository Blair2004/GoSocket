package models

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// NewClient creates a new client
func NewClient(id string, conn *websocket.Conn) *Client {
	return &Client{
		ID:              id,
		Conn:            conn,
		Channels:        make(map[string]bool),
		ChannelMetadata: make(map[string]*ChannelMetadata),
		LastSeen:        time.Now(),
		RemoteAddr:      "",
		UserAgent:       "",
	}
}

// NewChannel creates a new channel
func NewChannel(name string) *Channel {
	return &Channel{
		Name:        name,
		Clients:     make(map[string]*Client),
		IsPrivate:   false,
		RequireAuth: false,
		CreatedAt:   time.Now(),
	}
}

// ChannelMetadata represents metadata stored for each channel a client joins
type ChannelMetadata struct {
	Data     interface{} `json:"data"`
	JoinedAt time.Time   `json:"joined_at"`
}

// Client represents a connected WebSocket client
type Client struct {
	ID              string                      `json:"id"`
	Conn            *websocket.Conn             `json:"-"`
	UserID          string                      `json:"user_id,omitempty"`
	Username        string                      `json:"username,omitempty"`
	Email           string                      `json:"email,omitempty"`
	Channels        map[string]bool             `json:"channels"`
	ChannelMetadata map[string]*ChannelMetadata `json:"channel_metadata"`
	LastSeen        time.Time                   `json:"last_seen"`
	RemoteAddr      string                      `json:"remote_addr"`
	UserAgent       string                      `json:"user_agent"`
	mutex           sync.RWMutex                `json:"-"`
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
	Private   *bool       `json:"private,omitempty"`
	Event     string      `json:"event"`
	Data      interface{} `json:"data"`
	UserID    string      `json:"user_id,omitempty"`
	Username  string      `json:"username,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// SendMessage sends a message to the client
func (c *Client) SendMessage(message Message) error {
	start := time.Now()

	// Track mutex wait time
	mutexStart := time.Now()
	c.mutex.Lock()
	mutexTime := time.Since(mutexStart)
	defer c.mutex.Unlock()

	if c.Conn == nil {
		return ErrNilConnection
	}

	// Set a very short write deadline for local environment (500ms)
	deadlineStart := time.Now()
	c.Conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))
	deadlineTime := time.Since(deadlineStart)

	// Track actual write time
	writeStart := time.Now()
	err := c.Conn.WriteJSON(message)
	writeTime := time.Since(writeStart)

	totalTime := time.Since(start)

	// Log all operations that take more than 50ms in local environment
	if totalTime > 50*time.Millisecond {
		// Log detailed timing breakdown for slow operations
		// This will help identify if it's mutex contention, network, or JSON serialization
		_ = mutexTime    // Mutex wait time
		_ = deadlineTime // Deadline setting time
		_ = writeTime    // Actual write time
		_ = totalTime    // Total time
	}

	return err
}

// SafeReadJSON safely reads a JSON message from the client connection
func (c *Client) SafeReadJSON(v interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Conn == nil {
		return ErrNilConnection
	}

	return c.Conn.ReadJSON(v)
}

// SafeSetReadDeadline safely sets the read deadline on the client connection
func (c *Client) SafeSetReadDeadline(t time.Time) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Conn == nil {
		return ErrNilConnection
	}

	return c.Conn.SetReadDeadline(t)
}

// AddToChannel adds the client to a channel
func (c *Client) AddToChannel(channelName string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Channels[channelName] = true
}

// AddToChannelWithMetadata adds the client to a channel with metadata
func (c *Client) AddToChannelWithMetadata(channelName string, data interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Channels[channelName] = true
	c.ChannelMetadata[channelName] = &ChannelMetadata{
		Data:     data,
		JoinedAt: time.Now(),
	}
}

// RemoveFromChannel removes the client from a channel
func (c *Client) RemoveFromChannel(channelName string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.Channels, channelName)
	delete(c.ChannelMetadata, channelName)
}

// GetChannelMetadata returns the metadata for a specific channel
func (c *Client) GetChannelMetadata(channelName string) *ChannelMetadata {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.ChannelMetadata[channelName]
}

// GetAllChannelMetadata returns a copy of all channel metadata
func (c *Client) GetAllChannelMetadata() map[string]*ChannelMetadata {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	metadata := make(map[string]*ChannelMetadata)
	for k, v := range c.ChannelMetadata {
		metadata[k] = &ChannelMetadata{
			Data:     v.Data,
			JoinedAt: v.JoinedAt,
		}
	}
	return metadata
}

// GetChannels returns a copy of the client's channels
func (c *Client) GetChannels() map[string]bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	channels := make(map[string]bool)
	for k, v := range c.Channels {
		channels[k] = v
	}
	return channels
}

// SetUserInfo sets the user information in a thread-safe manner
func (c *Client) SetUserInfo(userID, username, email string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.UserID = userID
	c.Username = username
	c.Email = email
}

// Close safely closes the client connection
func (c *Client) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.Conn != nil {
		c.Conn.Close()
		c.Conn = nil
	}
}

// IsConnected safely checks if the client connection is still valid
func (c *Client) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.Conn != nil
}

// SendPing sends a ping message to the client
func (c *Client) SendPing() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Conn == nil {
		return ErrNilConnection
	}

	// Set write deadline for ping (same as SendMessage)
	c.Conn.SetWriteDeadline(time.Now().Add(500 * time.Millisecond))

	return c.Conn.WriteMessage(websocket.PingMessage, nil)
}

// AddClient adds a client to the channel
func (ch *Channel) AddClient(client *Client) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	ch.Clients[client.ID] = client
}

// RemoveClient removes a client from the channel
func (ch *Channel) RemoveClient(clientID string) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	delete(ch.Clients, clientID)
}

// GetClients returns a copy of the channel's clients
func (ch *Channel) GetClients() map[string]*Client {
	ch.mutex.RLock()
	defer ch.mutex.RUnlock()

	clients := make(map[string]*Client)
	for k, v := range ch.Clients {
		clients[k] = v
	}
	return clients
}

// GetClientCount returns the number of clients in the channel
func (ch *Channel) GetClientCount() int {
	ch.mutex.RLock()
	defer ch.mutex.RUnlock()
	return len(ch.Clients)
}
