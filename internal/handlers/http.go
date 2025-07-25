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
	startTime := time.Now()
	h.logger.Info("🚀 Broadcast request started")

	var payload struct {
		Channel             string      `json:"channel"`
		Event               string      `json:"event"`
		Data                interface{} `json:"data"`
		BroadcastToEveryone bool        `json:"broadcast_to_everyone"`
		ExcludeCurrentUser  bool        `json:"exclude_current_user"`
		UserID              *string     `json:"user_id"`
		ClientID            *string     `json:"client_id"`
		BroadcastType       string      `json:"broadcast_type"` // "channel", "global", "authenticated", "user", "user_except", "client"
	}

	decodeStart := time.Now()
	err := json.NewDecoder(r.Body).Decode(&payload)
	decodeTime := time.Since(decodeStart)
	h.logger.Info("⏱️ JSON decode took: %v", decodeTime)
	if err != nil {
		h.logger.Error("Failed to decode JSON payload", "error", err.Error())

		// Provide more specific error messages for common issues
		errorMsg := "Invalid JSON payload: " + err.Error()
		if jsonErr, ok := err.(*json.UnmarshalTypeError); ok {
			switch jsonErr.Field {
			case "channel":
				errorMsg = "Invalid 'channel' field: expected string, got " + jsonErr.Value + ". Example: \"channel\": \"my-channel\""
			case "event":
				errorMsg = "Invalid 'event' field: expected string, got " + jsonErr.Value + ". Example: \"event\": \"my-event\""
			case "user_id":
				errorMsg = "Invalid 'user_id' field: expected string, got " + jsonErr.Value + ". Example: \"user_id\": \"123\""
			case "client_id":
				errorMsg = "Invalid 'client_id' field: expected string, got " + jsonErr.Value + ". Example: \"client_id\": \"abc123\""
			case "broadcast_type":
				errorMsg = "Invalid 'broadcast_type' field: expected string, got " + jsonErr.Value + ". Valid values: \"global\", \"authenticated\", \"user\", \"user_except\", \"client\", \"channel\""
			case "broadcast_to_everyone":
				errorMsg = "Invalid 'broadcast_to_everyone' field: expected boolean, got " + jsonErr.Value + ". Example: \"broadcast_to_everyone\": true"
			case "exclude_current_user":
				errorMsg = "Invalid 'exclude_current_user' field: expected boolean, got " + jsonErr.Value + ". Example: \"exclude_current_user\": false"
			default:
				errorMsg = "Invalid field '" + jsonErr.Field + "': expected " + jsonErr.Type.String() + ", got " + jsonErr.Value
			}
		}

		http.Error(w, errorMsg, http.StatusBadRequest)
		return
	}

	if payload.Event == "" {
		payload.Event = "broadcast"
	}

	msgCreateStart := time.Now()
	message := models.Message{
		ID:        uuid.New().String(),
		Channel:   payload.Channel,
		Event:     payload.Event,
		Data:      payload.Data,
		Timestamp: time.Now(),
	}
	msgCreateTime := time.Since(msgCreateStart)
	h.logger.Info("⏱️ Message creation took: %v", msgCreateTime)

	// Determine broadcast type based on payload
	typeDetectStart := time.Now()
	broadcastType := payload.BroadcastType
	if broadcastType == "" {
		// Legacy behavior: determine from other fields
		if payload.BroadcastToEveryone {
			broadcastType = "global"
		} else if payload.ExcludeCurrentUser && payload.UserID != nil && *payload.UserID != "" {
			broadcastType = "user_except"
		} else if payload.UserID != nil && *payload.UserID != "" {
			broadcastType = "user"
		} else if payload.Channel != "" {
			broadcastType = "channel"
		} else {
			broadcastType = "global"
		}
	}
	typeDetectTime := time.Since(typeDetectStart)
	h.logger.Info("⏱️ Broadcast type detection took: %v", typeDetectTime)

	broadcastStart := time.Now()
	var responseMessage string
	switch broadcastType {
	case "global":
		h.logger.Info("🌍 Starting global broadcast")
		h.wsServer.BroadcastToAll(message)
		responseMessage = "Message broadcasted to all clients"

	case "authenticated":
		h.logger.Info("🔐 Starting authenticated broadcast")
		h.wsServer.BroadcastToAuthenticated(message)
		responseMessage = "Message broadcasted to all authenticated clients"

	case "user":
		if payload.UserID == nil || *payload.UserID == "" {
			http.Error(w, "user_id is required for user broadcast", http.StatusBadRequest)
			return
		}
		h.logger.Info("👤 Starting user broadcast to user: %s", *payload.UserID)
		h.wsServer.BroadcastToUser(*payload.UserID, message)
		responseMessage = "Message broadcasted to user " + *payload.UserID

	case "user_except":
		if payload.UserID == nil || *payload.UserID == "" {
			http.Error(w, "user_id is required for user_except broadcast", http.StatusBadRequest)
			return
		}
		h.logger.Info("👥 Starting user_except broadcast excluding user: %s", *payload.UserID)
		h.wsServer.BroadcastToUsersExcept(*payload.UserID, message)
		responseMessage = "Message broadcasted to all authenticated clients except user " + *payload.UserID

	case "client":
		if payload.ClientID == nil || *payload.ClientID == "" {
			http.Error(w, "client_id is required for client broadcast", http.StatusBadRequest)
			return
		}
		h.logger.Info("🖥️ Starting client broadcast to client: %s", *payload.ClientID)
		err := h.wsServer.BroadcastToClient(*payload.ClientID, message)
		if err != nil {
			if err == models.ErrClientNotFound {
				http.Error(w, "Client not found", http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		responseMessage = "Message sent to client " + *payload.ClientID

	case "channel":
		if payload.Channel == "" {
			http.Error(w, "channel is required for channel broadcast", http.StatusBadRequest)
			return
		}
		h.logger.Info("📺 Starting channel broadcast to channel: %s", payload.Channel)
		h.wsServer.BroadcastToChannel(payload.Channel, message)
		responseMessage = "Message broadcasted to channel " + payload.Channel

	default:
		http.Error(w, "Invalid broadcast_type. Must be: global, authenticated, user, user_except, client, or channel", http.StatusBadRequest)
		return
	}
	broadcastTime := time.Since(broadcastStart)
	h.logger.Info("⏱️ Broadcast operation took: %v", broadcastTime)

	responseStart := time.Now()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": responseMessage,
		"type":    broadcastType,
	})
	responseTime := time.Since(responseStart)
	h.logger.Info("⏱️ Response generation took: %v", responseTime)

	totalTime := time.Since(startTime)
	h.logger.Info("🏁 Total broadcast request took: %v", totalTime)
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
