package models

import (
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		conn      *websocket.Conn
		expectNil bool
	}{
		{
			name:      "Valid client creation",
			id:        "client-123",
			conn:      nil, // Using nil for test simplicity
			expectNil: false,
		},
		{
			name:      "Empty client ID",
			id:        "",
			conn:      nil,
			expectNil: false, // Constructor doesn't validate, so it shouldn't be nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.id, tt.conn)

			if tt.expectNil && client != nil {
				t.Errorf("Expected nil client but got %v", client)
				return
			}

			if !tt.expectNil && client == nil {
				t.Error("Expected non-nil client but got nil")
				return
			}

			if client != nil {
				if client.ID != tt.id {
					t.Errorf("Expected client ID %s, got %s", tt.id, client.ID)
				}

				if client.Conn != tt.conn {
					t.Errorf("Expected connection %v, got %v", tt.conn, client.Conn)
				}

				if client.Channels == nil {
					t.Error("Expected Channels map to be initialized")
				}

				if len(client.Channels) != 0 {
					t.Errorf("Expected empty channels map, got %d channels", len(client.Channels))
				}

				if client.LastSeen.IsZero() {
					t.Error("Expected LastSeen to be set")
				}
			}
		})
	}
}

func TestNewChannel(t *testing.T) {
	tests := []struct {
		name        string
		channelName string
		expectNil   bool
	}{
		{
			name:        "Valid channel creation",
			channelName: "test-channel",
			expectNil:   false,
		},
		{
			name:        "Empty channel name",
			channelName: "",
			expectNil:   false, // Constructor doesn't validate, so it shouldn't be nil
		},
		{
			name:        "Channel with special characters",
			channelName: "test-channel-123_with-special.chars",
			expectNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel := NewChannel(tt.channelName)

			if tt.expectNil && channel != nil {
				t.Errorf("Expected nil channel but got %v", channel)
				return
			}

			if !tt.expectNil && channel == nil {
				t.Error("Expected non-nil channel but got nil")
				return
			}

			if channel != nil {
				if channel.Name != tt.channelName {
					t.Errorf("Expected channel name %s, got %s", tt.channelName, channel.Name)
				}

				if channel.Clients == nil {
					t.Error("Expected Clients map to be initialized")
				}

				if len(channel.Clients) != 0 {
					t.Errorf("Expected empty clients map, got %d clients", len(channel.Clients))
				}

				if channel.CreatedAt.IsZero() {
					t.Error("Expected CreatedAt to be set")
				}
			}
		})
	}
}

func TestChannelAddClient(t *testing.T) {
	channel := NewChannel("test-channel")
	client := NewClient("client-123", nil)

	// Test adding client
	channel.AddClient(client)

	if len(channel.Clients) != 1 {
		t.Errorf("Expected 1 client, got %d", len(channel.Clients))
	}

	if _, exists := channel.Clients[client.ID]; !exists {
		t.Error("Client was not added to channel")
	}

	// Test adding same client again (should not duplicate)
	channel.AddClient(client)
	if len(channel.Clients) != 1 {
		t.Errorf("Expected 1 client after duplicate add, got %d", len(channel.Clients))
	}
}

func TestChannelRemoveClient(t *testing.T) {
	channel := NewChannel("test-channel")
	client := NewClient("client-123", nil)

	// Add client first
	channel.AddClient(client)

	// Test removing client
	channel.RemoveClient(client.ID)

	if len(channel.Clients) != 0 {
		t.Errorf("Expected 0 clients after removal, got %d", len(channel.Clients))
	}

	if _, exists := channel.Clients[client.ID]; exists {
		t.Error("Client was not removed from channel")
	}

	// Test removing non-existent client (should not panic)
	channel.RemoveClient("non-existent-client")
}

func TestChannelGetClients(t *testing.T) {
	channel := NewChannel("test-channel")

	// Test empty channel
	clients := channel.GetClients()
	if len(clients) != 0 {
		t.Errorf("Expected 0 clients for empty channel, got %d", len(clients))
	}

	// Add multiple clients
	for i := 0; i < 3; i++ {
		client := NewClient(fmt.Sprintf("client-%d", i), nil)
		channel.AddClient(client)
	}

	clients = channel.GetClients()
	if len(clients) != 3 {
		t.Errorf("Expected 3 clients, got %d", len(clients))
	}
}

func TestChannelGetClientCount(t *testing.T) {
	channel := NewChannel("test-channel")

	// Test empty channel
	count := channel.GetClientCount()
	if count != 0 {
		t.Errorf("Expected 0 clients for empty channel, got %d", count)
	}

	// Add clients
	for i := 0; i < 5; i++ {
		client := NewClient(fmt.Sprintf("client-%d", i), nil)
		channel.AddClient(client)
	}

	count = channel.GetClientCount()
	if count != 5 {
		t.Errorf("Expected 5 clients, got %d", count)
	}
}

func TestClientAddToChannel(t *testing.T) {
	client := NewClient("client-123", nil)

	// Test adding to channel
	client.AddToChannel("test-channel")

	channels := client.GetChannels()
	if len(channels) != 1 {
		t.Errorf("Expected 1 channel, got %d", len(channels))
	}

	if !channels["test-channel"] {
		t.Error("Channel was not added to client")
	}
}

func TestClientRemoveFromChannel(t *testing.T) {
	client := NewClient("client-123", nil)

	// Add channel first
	client.AddToChannel("test-channel")

	// Test removing from channel
	client.RemoveFromChannel("test-channel")

	channels := client.GetChannels()
	if len(channels) != 0 {
		t.Errorf("Expected 0 channels after removal, got %d", len(channels))
	}

	if channels["test-channel"] {
		t.Error("Channel was not removed from client")
	}
}

func TestClientSetUserInfo(t *testing.T) {
	client := NewClient("client-123", nil)

	// Test setting user info
	client.SetUserInfo("user-456", "john_doe", "john@example.com")

	if client.UserID != "user-456" {
		t.Errorf("Expected user ID 'user-456', got '%s'", client.UserID)
	}

	if client.Username != "john_doe" {
		t.Errorf("Expected username 'john_doe', got '%s'", client.Username)
	}

	if client.Email != "john@example.com" {
		t.Errorf("Expected email 'john@example.com', got '%s'", client.Email)
	}
}

func TestClientSendMessageWithNilConnection(t *testing.T) {
	client := NewClient("client-123", nil)

	message := Message{
		ID:        "msg-123",
		Channel:   "test-channel",
		Event:     "test-event",
		Data:      "test data",
		Timestamp: time.Now(),
	}

	err := client.SendMessage(message)
	if err != ErrNilConnection {
		t.Errorf("Expected ErrNilConnection, got %v", err)
	}
}

func TestClientSendPingWithNilConnection(t *testing.T) {
	client := NewClient("client-123", nil)

	err := client.SendPing()
	if err != ErrNilConnection {
		t.Errorf("Expected ErrNilConnection, got %v", err)
	}
}

func TestClientConcurrentAccess(t *testing.T) {
	client := NewClient("client-123", nil)

	// Test concurrent access to client methods
	done := make(chan bool)

	// Simulate concurrent operations
	for i := 0; i < 10; i++ {
		go func(id int) {
			channelName := fmt.Sprintf("channel-%d", id)
			client.AddToChannel(channelName)
			_ = client.GetChannels()
			client.RemoveFromChannel(channelName)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestChannelConcurrentAccess(t *testing.T) {
	channel := NewChannel("test-channel")

	// Test concurrent client addition and removal
	done := make(chan bool)

	// Add clients concurrently
	for i := 0; i < 5; i++ {
		go func(id int) {
			client := NewClient(fmt.Sprintf("client-%d", id), nil)
			channel.AddClient(client)
			done <- true
		}(i)
	}

	// Wait for all additions to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = channel.GetClientCount()
			_ = channel.GetClients()
			done <- true
		}()
	}

	// Wait for all reads to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestChannelConcurrentAddRemove(t *testing.T) {
	channel := NewChannel("test-channel")

	// Add some initial clients
	for i := 0; i < 10; i++ {
		client := NewClient(fmt.Sprintf("client-%d", i), nil)
		channel.AddClient(client)
	}

	done := make(chan bool)

	// Concurrently add and remove clients
	for i := 0; i < 5; i++ {
		go func(id int) {
			// Add new client
			newClient := NewClient(fmt.Sprintf("new-client-%d", id), nil)
			channel.AddClient(newClient)

			// Remove an existing client
			channel.RemoveClient(fmt.Sprintf("client-%d", id))
			done <- true
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Check final state
	count := channel.GetClientCount()
	if count != 10 { // Should still have 10 clients (5 removed, 5 added)
		t.Errorf("Expected 10 clients after concurrent operations, got %d", count)
	}
}
