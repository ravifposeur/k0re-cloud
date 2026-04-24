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

// StatusResponse represents the daemon's status response
type StatusResponse struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	CPUUsage    string `json:"cpu_usage"`
	MemUsage    string `json:"mem_usage"`
	LastUpdated string `json:"last_updated"`
}

var daemonBaseURL = "http://127.0.0.1:3000"

func main() {
	// Check if no arguments provided
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Route to appropriate subcommand
	switch os.Args[1] {
	case "apply":
		applyCmd := flag.NewFlagSet("apply", flag.ExitOnError)
		filePath := applyCmd.String("f", "", "Path to YAML configuration file (required)")
		applyCmd.Parse(os.Args[2:])

		// Validate that file flag is provided
		if *filePath == "" {
			log.Fatal("Error: -f flag is required. Usage: k0re apply -f <config.yaml>")
		}

		// Execute the apply command
		if err := applyConfig(*filePath); err != nil {
			log.Fatalf("Error: %v", err)
		}

	case "status":
		if len(os.Args) < 3 {
			log.Fatal("Error: server name is required. Usage: k0re status <server-name>")
		}
		serverName := os.Args[2]
		if err := getServerStatus(serverName); err != nil {
			log.Fatalf("Error: %v", err)
		}

	case "audit":
		if err := getAuditLog(); err != nil {
			log.Fatalf("Error: %v", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: k0re <command> [options]")
	fmt.Println("Commands:")
	fmt.Println("  apply    Apply configuration from YAML file (-f <config.yaml>)")
	fmt.Println("  status   View live metrics of a specific server (status <server-name>)")
	fmt.Println("  audit    View activity logs and audit trail")
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
	url := fmt.Sprintf("%s/v1/provision", daemonBaseURL)
	return sendProvisionRequest(url, provisionReq)
}

// sendProvisionRequest sends the provision request to the daemon and handles the response
func sendProvisionRequest(url string, req ProvisionRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("Sending request to %s: %s", url, string(jsonData))

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request to daemon: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var provisionResp ProvisionResponse
	if err := json.Unmarshal(respBody, &provisionResp); err != nil {
		return fmt.Errorf("failed to parse daemon response: %w", err)
	}

	log.Printf("Daemon response (HTTP %d): status=%s, message=%s",
		resp.StatusCode, provisionResp.Status, provisionResp.Message)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("daemon returned error (HTTP %d): %s", resp.StatusCode, provisionResp.Message)
	}

	fmt.Printf("✓ Server '%s' provisioned successfully\n", req.Name)
	return nil
}

// getServerStatus fetches and prints the live metrics of a server
func getServerStatus(serverName string) error {
	url := fmt.Sprintf("%s/v1/status/%s", daemonBaseURL, serverName)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to contact daemon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("❌ Server '%s' tidak ditemukan.\n", serverName)
		return nil
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("daemon returned error (HTTP %d)", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var status StatusResponse
	if err := json.Unmarshal(body, &status); err != nil {
		return fmt.Errorf("failed to parse daemon response: %w", err)
	}

	fmt.Println("==========================================")
	fmt.Printf("📊 STATUS SERVER : %s\n", status.Name)
	fmt.Println("==========================================")
	fmt.Printf("State       : %s\n", status.Status)
	fmt.Printf("CPU Usage   : %s\n", status.CPUUsage)
	fmt.Printf("RAM Usage   : %s\n", status.MemUsage)
	fmt.Printf("Last Update : %s\n", status.LastUpdated)
	fmt.Println("==========================================")

	return nil
}

// getAuditLog fetches and prints the audit trail
func getAuditLog() error {
	url := fmt.Sprintf("%s/v1/audit", daemonBaseURL)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to contact daemon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("daemon returned error (HTTP %d)", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Println("==========================================")
	fmt.Println("📜 K0RE AUDIT TRAIL")
	fmt.Println("==========================================")
	fmt.Print(string(body))

	return nil
}
