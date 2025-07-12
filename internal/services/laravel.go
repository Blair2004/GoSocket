package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"socket-server/internal/models"
	"socket-server/pkg/logger"
)

// LaravelService handles Laravel integration
type LaravelService struct {
	workingDir string
	phpBinary  string
	laravelCmd string
	tempDir    string
	logger     *logger.Logger
}

// NewLaravelService creates a new Laravel service
func NewLaravelService(workingDir, phpBinary, laravelCmd, tempDir string, logger *logger.Logger) *LaravelService {
	return &LaravelService{
		workingDir: workingDir,
		phpBinary:  phpBinary,
		laravelCmd: laravelCmd,
		tempDir:    tempDir,
		logger:     logger,
	}
}

// InitializeTempDirectory ensures the temp directory exists with proper permissions
func (s *LaravelService) InitializeTempDirectory() error {
	if s.tempDir == "" {
		s.tempDir = filepath.Join(os.TempDir(), "socket-server-payloads")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(s.tempDir, 0755); err != nil {
		return fmt.Errorf("error creating temp directory %s: %w", s.tempDir, err)
	}

	s.logger.Info("Temp directory initialized: %s", s.tempDir)
	return nil
}

// DispatchMessage sends a client message to Laravel for processing
func (s *LaravelService) DispatchMessage(message models.Message, client *models.Client) error {
	payloadFile, err := s.createTempPayloadFile(message, client)
	if err != nil {
		return fmt.Errorf("error creating temp payload file: %w", err)
	}

	return s.executeLaravelCommand(payloadFile)
}

// DispatchAuthentication sends authentication events to Laravel
func (s *LaravelService) DispatchAuthentication(client *models.Client, status string, token string) error {
	// Create standardized authentication payload
	standardizedPayload := map[string]interface{}{
		"message_id": uuid.New().String(),
		"timestamp":  time.Now().Format(time.RFC3339),
		"action":     "client_authentication",
		"auth": map[string]interface{}{
			"user_id":     client.UserID,
			"user_email":  client.Email,
			"logged_at":   time.Now().Format(time.RFC3339),
			"id":          client.ID,
			"username":    client.Username,
			"remote_addr": client.RemoteAddr,
		},
		"data": map[string]interface{}{
			"authentication_status": status,
			"token_provided":        token != "",
		},
	}

	payloadFile, err := s.createTempPayloadFileFromData(standardizedPayload)
	if err != nil {
		return fmt.Errorf("error creating temp authentication payload file: %w", err)
	}

	return s.executeLaravelCommand(payloadFile)
}

// createTempPayloadFile creates a temporary file with message data
func (s *LaravelService) createTempPayloadFile(message models.Message, client *models.Client) (string, error) {
	// Create standardized message payload
	standardizedPayload := map[string]interface{}{
		"message_id": uuid.New().String(),
		"timestamp":  time.Now().Format(time.RFC3339),
		"action":     message.Event,
		"auth": map[string]interface{}{
			"user_id":     client.UserID,
			"user_email":  client.Email,
			"logged_at":   time.Now().Format(time.RFC3339),
			"id":          client.ID,
			"username":    client.Username,
			"remote_addr": client.RemoteAddr,
		},
		"data": map[string]interface{}{
			"id":        message.ID,
			"channel":   message.Channel,
			"event":     message.Event,
			"data":      message.Data,
			"timestamp": message.Timestamp,
		},
	}

	return s.createTempPayloadFileFromData(standardizedPayload)
}

// createTempPayloadFileFromData creates a temporary file with the given data
func (s *LaravelService) createTempPayloadFileFromData(data interface{}) (string, error) {
	// Convert to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error marshaling payload data: %w", err)
	}

	// Create filename with timestamp for expiration tracking
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("payload_%d_%s.json", timestamp, uuid.New().String()[:8])
	filepath := filepath.Join(s.tempDir, filename)

	// Write file with permissions readable by Laravel (0644)
	if err := os.WriteFile(filepath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("error writing payload file: %w", err)
	}

	s.logger.TempFileCreated(filepath)
	return filepath, nil
}

// executeLaravelCommand executes the Laravel artisan command with payload file
func (s *LaravelService) executeLaravelCommand(payloadFile string) error {
	cmdString := fmt.Sprintf("%s artisan %s --payload %s", s.phpBinary, s.laravelCmd, payloadFile)
	s.logger.LaravelCommand(cmdString)

	cmd := exec.Command(s.phpBinary, "artisan", s.laravelCmd, "--payload", payloadFile)
	cmd.Dir = s.workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.logger.LaravelCommandError(s.laravelCmd, err, string(output))
		return fmt.Errorf("error executing Laravel command: %w", err)
	}

	s.logger.LaravelCommandSuccess(s.laravelCmd, string(output))
	return nil
}

// StartCleanupRoutine starts a background routine to clean up expired temp files
func (s *LaravelService) StartCleanupRoutine() {
	go func() {
		// Run cleanup every hour
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		// Run initial cleanup
		s.cleanupExpiredFiles()

		for range ticker.C {
			s.cleanupExpiredFiles()
		}
	}()

	s.logger.Info("Started temp file cleanup routine (runs every hour)")
}

// cleanupExpiredFiles removes temp files older than 24 hours
func (s *LaravelService) cleanupExpiredFiles() {
	expireTime := time.Now().Add(-24 * time.Hour) // 24 hours ago

	files, err := os.ReadDir(s.tempDir)
	if err != nil {
		s.logger.Error("Error reading temp directory: %v", err)
		return
	}

	cleaned := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Check if file follows our naming pattern
		if !strings.HasPrefix(file.Name(), "payload_") || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(s.tempDir, file.Name())
		info, err := file.Info()
		if err != nil {
			s.logger.Error("Error getting file info for %s: %v", filePath, err)
			continue
		}

		// Remove if older than 24 hours
		if info.ModTime().Before(expireTime) {
			if err := os.Remove(filePath); err != nil {
				s.logger.Error("Error removing expired file %s: %v", filePath, err)
			} else {
				cleaned++
				s.logger.Debug("Removed expired temp file: %s", filePath)
			}
		}
	}

	s.logger.TempFileCleanup(cleaned)
}
