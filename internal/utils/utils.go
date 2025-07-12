package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"socket-server/internal/models"
)

// FileUtils provides file-related utilities
type FileUtils struct{}

// ReadJSONFile reads and parses a JSON file
func (f *FileUtils) ReadJSONFile(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("error parsing JSON from %s: %w", filePath, err)
	}

	return result, nil
}

// WriteJSONFile writes data to a JSON file
func (f *FileUtils) WriteJSONFile(filePath string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling data: %w", err)
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory %s: %w", dir, err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing file %s: %w", filePath, err)
	}

	return nil
}

// HTTPClient provides HTTP-related utilities
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Get performs a GET request
func (h *HTTPClient) Get(endpoint string) ([]byte, error) {
	url := h.baseURL + endpoint
	resp, err := h.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making GET request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

// Post performs a POST request
func (h *HTTPClient) Post(endpoint string, data interface{}) ([]byte, error) {
	url := h.baseURL + endpoint

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request data: %w", err)
	}

	resp, err := h.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error making POST request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error %d from %s: %s", resp.StatusCode, url, string(body))
	}

	return body, nil
}

// MessageBuilder helps build socket messages
type MessageBuilder struct{}

// BuildMessage creates a socket message
func (m *MessageBuilder) BuildMessage(channel, event string, data interface{}) models.Message {
	return models.Message{
		Channel:   channel,
		Event:     event,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// BuildAuthMessage creates an authentication message
func (m *MessageBuilder) BuildAuthMessage(token string) map[string]interface{} {
	return map[string]interface{}{
		"action": "authenticate",
		"token":  token,
	}
}

// BuildJoinChannelMessage creates a join channel message
func (m *MessageBuilder) BuildJoinChannelMessage(channel string) map[string]interface{} {
	return map[string]interface{}{
		"action":  "join_channel",
		"channel": channel,
	}
}

// BuildLeaveChannelMessage creates a leave channel message
func (m *MessageBuilder) BuildLeaveChannelMessage(channel string) map[string]interface{} {
	return map[string]interface{}{
		"action":  "leave_channel",
		"channel": channel,
	}
}

// BuildSendMessage creates a send message
func (m *MessageBuilder) BuildSendMessage(channel, event string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"action":  "send_message",
		"channel": channel,
		"event":   event,
		"data":    data,
	}
}

// BuildPingMessage creates a ping message
func (m *MessageBuilder) BuildPingMessage() map[string]interface{} {
	return map[string]interface{}{
		"action": "ping",
	}
}
