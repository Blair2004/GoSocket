package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"socket-server/internal/models"
	"socket-server/internal/websocket"
	"socket-server/pkg/logger"
)

// HTTPHandlers contains all HTTP handlers
type HTTPHandlers struct {
	wsServer *websocket.Server
	logger   *logger.Logger
}

// New creates new HTTP handlers
func New(wsServer *websocket.Server, logger *logger.Logger) *HTTPHandlers {
	return &HTTPHandlers{
		wsServer: wsServer,
		logger:   logger,
	}
}

// GetClients returns all connected clients
func (h *HTTPHandlers) GetClients(w http.ResponseWriter, r *http.Request) {
	clients := h.wsServer.GetClients()

	// Convert to slice for JSON response
	clientSlice := make([]*models.Client, 0, len(clients))
	for _, client := range clients {
		clientSlice = append(clientSlice, client)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"clients": clientSlice,
		"total":   len(clientSlice),
	})
}

// GetChannels returns all channels
func (h *HTTPHandlers) GetChannels(w http.ResponseWriter, r *http.Request) {
	channels := h.wsServer.GetChannels()

	// Convert to response format
	channelResponse := make(map[string]interface{})
	for name, channel := range channels {
		channelResponse[name] = map[string]interface{}{
			"name":         channel.Name,
			"is_private":   channel.IsPrivate,
			"require_auth": channel.RequireAuth,
			"client_count": channel.GetClientCount(),
			"created_at":   channel.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(channelResponse)
}

// GetChannelClients returns clients in a specific channel
func (h *HTTPHandlers) GetChannelClients(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelName := vars["channel"]

	channel, exists := h.wsServer.GetChannel(channelName)
	if !exists {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	clients := channel.GetClients()

	// Convert to slice for JSON response
	clientSlice := make([]*models.Client, 0, len(clients))
	for _, client := range clients {
		clientSlice = append(clientSlice, client)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"channel": channelName,
		"clients": clientSlice,
		"total":   len(clientSlice),
	})
}

// KickClient kicks a specific client
func (h *HTTPHandlers) KickClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clientID := vars["client"]

	err := h.wsServer.KickClient(clientID)
	if err != nil {
		if err == models.ErrClientNotFound {
			http.Error(w, "Client not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Client " + clientID + " kicked",
	})
}

// Broadcast sends a message to a channel
func (h *HTTPHandlers) Broadcast(w http.ResponseWriter, r *http.Request) {
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

	message := models.Message{
		ID:        uuid.New().String(),
		Channel:   payload.Channel,
		Event:     payload.Event,
		Data:      payload.Data,
		Timestamp: time.Now(),
	}

	h.wsServer.BroadcastToChannel(payload.Channel, message)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Message broadcasted",
	})
}

// Health returns server health status
func (h *HTTPHandlers) Health(w http.ResponseWriter, r *http.Request) {
	clients := h.wsServer.GetClients()
	channels := h.wsServer.GetChannels()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "healthy",
		"clients":  len(clients),
		"channels": len(channels),
		"version":  "1.0.0",
	})
}
