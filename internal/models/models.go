package models

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// NewClient creates a new client
func NewClient(id string, conn *websocket.Conn) *Client {
	return &Client{
		ID:         id,
		Conn:       conn,
		Channels:   make(map[string]bool),
		LastSeen:   time.Now(),
		RemoteAddr: "",
		UserAgent:  "",
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

// Client represents a connected WebSocket client
type Client struct {
	ID         string          `json:"id"`
	Conn       *websocket.Conn `json:"-"`
	UserID     string          `json:"user_id,omitempty"`
	Username   string          `json:"username,omitempty"`
	Email      string          `json:"email,omitempty"`
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

// SendMessage sends a message to the client
func (c *Client) SendMessage(message Message) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Conn == nil {
		return ErrNilConnection
	}

	// Set write deadline to prevent hanging
	c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	return c.Conn.WriteJSON(message)
}

// AddToChannel adds the client to a channel
func (c *Client) AddToChannel(channelName string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.Channels[channelName] = true
}

// RemoveFromChannel removes the client from a channel
func (c *Client) RemoveFromChannel(channelName string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.Channels, channelName)
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

// SendPing sends a ping message to the client
func (c *Client) SendPing() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Conn == nil {
		return ErrNilConnection
	}

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
