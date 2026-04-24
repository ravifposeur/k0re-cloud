package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the YAML configuration structure
type Config struct {
	Server struct {
		Name      string `yaml:"name"`
		Game      string `yaml:"game"`
		Flavor    string `yaml:"flavor"`
		Resources struct {
			RAM string `yaml:"ram"`
		} `yaml:"resources"`
	} `yaml:"server"`
}

// ProvisionRequest represents the JSON payload sent to the daemon
type ProvisionRequest struct {
	Name   string `json:"name"`
	Game   string `json:"game"`
	Flavor string `json:"flavor"`
	RAM    string `json:"ram"`
}

// ProvisionResponse represents the daemon's response
type ProvisionResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	// Define the 'apply' subcommand
	applyCmd := flag.NewFlagSet("apply", flag.ExitOnError)
	filePath := applyCmd.String("f", "", "Path to YAML configuration file (required)")

	// Check if no arguments provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: k0re apply -f <config.yaml>")
		fmt.Println("Commands:")
		fmt.Println("  apply    Apply configuration from YAML file")
		os.Exit(1)
	}

	// Route to appropriate subcommand
	switch os.Args[1] {
	case "apply":
		applyCmd.Parse(os.Args[2:])

		// Validate that file flag is provided
		if *filePath == "" {
			log.Fatal("Error: -f flag is required. Usage: k0re apply -f <config.yaml>")
		}

		// Execute the apply command
		if err := applyConfig(*filePath); err != nil {
			log.Fatalf("Error: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		fmt.Println("Usage: k0re apply -f <config.yaml>")
		os.Exit(1)
	}
}

// applyConfig reads, parses, validates, and sends the configuration to the daemon
func applyConfig(filePath string) error {
	// Step 1: Read the YAML file
	yamlData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Step 2: Parse YAML into Config struct
	var config Config
	if err := yaml.Unmarshal(yamlData, &config); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Step 3: Validate required fields
	if config.Server.Name == "" {
		return fmt.Errorf("validation error: server.name is required")
	}
	if config.Server.Resources.RAM == "" {
		return fmt.Errorf("validation error: server.resources.ram is required")
	}

	log.Printf("Parsed config: name=%s, game=%s, flavor=%s, ram=%s",
		config.Server.Name, config.Server.Game, config.Server.Flavor, config.Server.Resources.RAM)

	// Step 4: Create the provision request payload
	provisionReq := ProvisionRequest{
		Name:   config.Server.Name,
		Game:   config.Server.Game,
		Flavor: config.Server.Flavor,
		RAM:    config.Server.Resources.RAM,
	}

	// Step 5: Send HTTP POST request to daemon
	daemonURL := "http://127.0.0.1:3000/v1/provision"
	return sendProvisionRequest(daemonURL, provisionReq)
}

// sendProvisionRequest sends the provision request to the daemon and handles the response
func sendProvisionRequest(url string, req ProvisionRequest) error {
	// Marshal the request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("Sending request to %s: %s", url, string(jsonData))

	// Create HTTP POST request
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request to daemon: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var provisionResp ProvisionResponse
	if err := json.Unmarshal(respBody, &provisionResp); err != nil {
		return fmt.Errorf("failed to parse daemon response: %w", err)
	}

	// Log the response
	log.Printf("Daemon response (HTTP %d): status=%s, message=%s",
		resp.StatusCode, provisionResp.Status, provisionResp.Message)

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("daemon returned error (HTTP %d): %s", resp.StatusCode, provisionResp.Message)
	}

	fmt.Printf("✓ Server '%s' provisioned successfully\n", req.Name)
	return nil
}
