package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	serverURL          string
	httpToken          string
	filePath           string
	channel            string
	event              string
	data               string
	insecureSkipVerify bool
)

var rootCmd = &cobra.Command{
	Use:   "socket",
	Short: "Socket server CLI client",
	Long:  "CLI client for communicating with the socket server",
}

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a message to the socket server",
	Long:  "Send a message to the socket server via HTTP API",
	Run:   sendMessage,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List connected clients or channels",
	Long:  "List connected clients or channels on the socket server",
}

var clientsCmd = &cobra.Command{
	Use:   "clients",
	Short: "List connected clients",
	Long:  "List all connected clients on the socket server",
	Run:   listClients,
}

var channelsCmd = &cobra.Command{
	Use:   "channels",
	Short: "List channels",
	Long:  "List all channels on the socket server",
	Run:   listChannels,
}

var kickCmd = &cobra.Command{
	Use:   "kick [client-id]",
	Short: "Kick a client",
	Long:  "Kick a specific client from the socket server",
	Args:  cobra.ExactArgs(1),
	Run:   kickClient,
}

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check server health",
	Long:  "Check the health status of the socket server",
	Run:   checkHealth,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "Socket server URL")
	rootCmd.PersistentFlags().StringVar(&httpToken, "token", os.Getenv("HTTP_TOKEN"), "HTTP API authentication token (required)")
	rootCmd.PersistentFlags().BoolVar(&insecureSkipVerify, "insecure", false, "Skip TLS certificate verification (for development/self-signed certificates)")

	// Send command flags
	sendCmd.Flags().StringVar(&filePath, "file", "", "JSON file containing message data")
	sendCmd.Flags().StringVar(&channel, "channel", "", "Channel to send message to")
	sendCmd.Flags().StringVar(&event, "event", "broadcast", "Event type")
	sendCmd.Flags().StringVar(&data, "data", "", "JSON data to send")

	// Add commands
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(kickCmd)
	rootCmd.AddCommand(healthCmd)

	listCmd.AddCommand(clientsCmd)
	listCmd.AddCommand(channelsCmd)
}

// getHTTPClient returns an HTTP client with optional TLS verification bypass
func getHTTPClient() *http.Client {
	if insecureSkipVerify {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		return &http.Client{Transport: tr}
	}
	return &http.Client{}
}

// createRequest creates an HTTP request with proper headers
func createRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// Add authentication header if token is provided
	if httpToken != "" {
		req.Header.Set("Authorization", "Bearer "+httpToken)
	}

	// Set content type for POST requests
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

// checkToken validates that the HTTP token is provided
func checkToken() {
	if httpToken == "" {
		fmt.Println("Error: HTTP API token is required. Use --token flag or set HTTP_TOKEN environment variable.")
		os.Exit(1)
	}
}

func sendMessage(cmd *cobra.Command, args []string) {
	checkToken()

	var payload map[string]interface{}

	if filePath != "" {
		// Read from file
		fileData, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}

		err = json.Unmarshal(fileData, &payload)
		if err != nil {
			fmt.Printf("Error parsing JSON file: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Build payload from flags
		if channel == "" {
			fmt.Println("Channel is required (use --channel flag)")
			os.Exit(1)
		}

		payload = map[string]interface{}{
			"channel": channel,
			"event":   event,
		}

		if data != "" {
			var jsonData interface{}
			err := json.Unmarshal([]byte(data), &jsonData)
			if err != nil {
				// If not valid JSON, treat as string
				payload["data"] = data
			} else {
				payload["data"] = jsonData
			}
		}
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling payload: %v\n", err)
		os.Exit(1)
	}

	client := getHTTPClient()
	req, err := createRequest("POST", serverURL+"/api/broadcast", bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server error (%d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Response: %s\n", string(body))
	} else {
		fmt.Printf("Status: %s\n", response["status"])
		fmt.Printf("Message: %s\n", response["message"])
	}
}

func listClients(cmd *cobra.Command, args []string) {
	checkToken()

	client := getHTTPClient()
	req, err := createRequest("GET", serverURL+"/api/clients", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server error (%d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	clients := response["clients"].([]interface{})
	total := response["total"].(float64)

	fmt.Printf("Connected Clients (%d):\n", int(total))
	fmt.Printf("%-36s %-15s %-20s %-15s %s\n", "ID", "User ID", "Username", "Channels", "Last Seen")
	fmt.Printf("%s\n", "---------------------------------------------------------------------------")

	for _, clientData := range clients {
		client := clientData.(map[string]interface{})
		id := client["id"].(string)
		userID := ""
		username := ""

		if client["user_id"] != nil {
			userID = client["user_id"].(string)
		}
		if client["username"] != nil {
			username = client["username"].(string)
		}

		channels := client["channels"].(map[string]interface{})
		channelCount := len(channels)

		lastSeen := client["last_seen"].(string)

		fmt.Printf("%-36s %-15s %-20s %-15d %s\n", id, userID, username, channelCount, lastSeen)
	}
}

func listChannels(cmd *cobra.Command, args []string) {
	checkToken()

	client := getHTTPClient()
	req, err := createRequest("GET", serverURL+"/api/channels", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server error (%d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var channels map[string]interface{}
	err = json.Unmarshal(body, &channels)
	if err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Channels (%d):\n", len(channels))
	fmt.Printf("%-30s %-10s %-12s %-10s %s\n", "Name", "Private", "Auth Required", "Clients", "Created")
	fmt.Printf("%s\n", "-------------------------------------------------------------------------------")

	for name, channelData := range channels {
		channel := channelData.(map[string]interface{})
		isPrivate := channel["is_private"].(bool)
		requireAuth := channel["require_auth"].(bool)
		clientCount := int(channel["client_count"].(float64))
		createdAt := channel["created_at"].(string)

		fmt.Printf("%-30s %-10t %-12t %-10d %s\n", name, isPrivate, requireAuth, clientCount, createdAt)
	}
}

func kickClient(cmd *cobra.Command, args []string) {
	checkToken()

	clientID := args[0]

	client := getHTTPClient()
	req, err := createRequest("POST", serverURL+"/api/clients/"+clientID+"/kick", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server error (%d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Response: %s\n", string(body))
	} else {
		fmt.Printf("Status: %s\n", response["status"])
		fmt.Printf("Message: %s\n", response["message"])
	}
}

func checkHealth(cmd *cobra.Command, args []string) {
	checkToken()

	client := getHTTPClient()
	req, err := createRequest("GET", serverURL+"/api/health", nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Server error (%d): %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Error parsing response: %v\n", err)
		fmt.Printf("Raw response: %s\n", string(body))
		os.Exit(1)
	}

	fmt.Printf("Server Status: %v\n", response["status"])

	if clients, ok := response["clients"].(float64); ok {
		fmt.Printf("Connected Clients: %.0f\n", clients)
	} else {
		fmt.Printf("Connected Clients: %v\n", response["clients"])
	}

	if channels, ok := response["channels"].(float64); ok {
		fmt.Printf("Active Channels: %.0f\n", channels)
	} else {
		fmt.Printf("Active Channels: %v\n", response["channels"])
	}

	fmt.Printf("Version: %v\n", response["version"])
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
