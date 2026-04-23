package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestProvision_Success(t *testing.T) {
	app := SetupApp()
	payload := map[string]string{
		"name":   "arena-test-success",
		"game":   "minecraft",
		"flavor": "pvp-competitive",
		"ram":    "2G",
	}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/v1/provision", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var body APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if body.Status != "success" {
		t.Errorf("Expected status 'success', got '%s'", body.Status)
	}
}

func TestProvision_MissingFields(t *testing.T) {
	app := SetupApp()
	payload := map[string]string{
		"name": "arena-test-missing",
		"game": "minecraft",
	}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/v1/provision", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("Expected 400, got %d", resp.StatusCode)
	}
}

func TestProvision_InvalidJSON(t *testing.T) {
	app := SetupApp()
	req := httptest.NewRequest("POST", "/v1/provision", bytes.NewBufferString(`{this is not json}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("Expected 400, got %d", resp.StatusCode)
	}
}

func TestProvision_EmptyBody(t *testing.T) {
	app := SetupApp()
	req := httptest.NewRequest("POST", "/v1/provision", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("Expected 400, got %d", resp.StatusCode)
	}
}

func TestProvision_PathTraversal(t *testing.T) {
	app := SetupApp()
	maliciousNames := []string{
		"../../etc/passwd",
		"../secret",
		"/absolute/path",
		"name with spaces",
		"name;rm -rf /",
		"Name_With_Uppercase",
	}

	for _, name := range maliciousNames {
		payload := map[string]string{
			"name":   name,
			"game":   "minecraft",
			"flavor": "pvp-competitive",
			"ram":    "2G",
		}
		raw, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/v1/provision", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("app.Test failed for name=%q: %v", name, err)
		}
		if resp.StatusCode != 400 {
			t.Errorf("Path traversal: name=%q expected 400, got %d", name, resp.StatusCode)
		}
	}
}

func TestProvision_Idempotency(t *testing.T) {
	app := SetupApp()
	serverName := "arena-already-exists"

	dir := filepath.Join("/var/k0re", serverName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Skipf("Cannot create /var/k0re (requires root): %v", err)
	}
	defer os.RemoveAll(dir)

	payload := map[string]string{
		"name":   serverName,
		"game":   "minecraft",
		"flavor": "pvp-competitive",
		"ram":    "2G",
	}
	raw, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/v1/provision", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != 409 {
		t.Errorf("Expected 409 Conflict, got %d", resp.StatusCode)
	}
}

func TestStatus_NotFound(t *testing.T) {
	app := SetupApp()
	req := httptest.NewRequest("GET", "/v1/status/server-yang-tidak-ada", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("Expected 404, got %d", resp.StatusCode)
	}

	var body APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if body.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", body.Status)
	}
}

func TestStatus_PathTraversal(t *testing.T) {
	app := SetupApp()
	maliciousNames := []string{
		"..etc..passwd",
		"name;whoami",
		"Name_With_Underscore",
		"name%20with%20spaces",
	}

	for _, name := range maliciousNames {
		req := httptest.NewRequest("GET", "/v1/status/"+name, nil)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatalf("app.Test failed for name=%q: %v", name, err)
		}
		if resp.StatusCode != 400 {
			t.Errorf("Path traversal: name=%q expected 400, got %d", name, resp.StatusCode)
		}
	}
}

func TestStatus_ValidFile(t *testing.T) {
	app := SetupApp()
	serverName := "arena-status-test"

	dir := filepath.Join("/var/k0re", serverName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Skipf("Cannot create /var/k0re (requires root): %v", err)
	}
	defer os.RemoveAll(dir)

	statusContent := StatusResponse{
		Name:     serverName,
		Status:   "running",
		CPUUsage: "12%",
		MemUsage: "512MB/2GB",
	}
	raw, _ := json.Marshal(statusContent)
	if err := os.WriteFile(filepath.Join(dir, "status.json"), raw, 0644); err != nil {
		t.Fatalf("Failed to write status.json: %v", err)
	}

	req := httptest.NewRequest("GET", "/v1/status/"+serverName, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var body StatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if body.Name != serverName {
		t.Errorf("Expected name '%s', got '%s'", serverName, body.Name)
	}
	if body.Status != "running" {
		t.Errorf("Expected status 'running', got '%s'", body.Status)
	}
}

func TestStatus_CorruptFile(t *testing.T) {
	app := SetupApp()
	serverName := "arena-corrupt-test"

	dir := filepath.Join("/var/k0re", serverName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Skipf("Cannot create /var/k0re (requires root): %v", err)
	}
	defer os.RemoveAll(dir)

	if err := os.WriteFile(filepath.Join(dir, "status.json"), []byte(`{corrupt:json`), 0644); err != nil {
		t.Fatalf("Failed to write corrupt status.json: %v", err)
	}

	req := httptest.NewRequest("GET", "/v1/status/"+serverName, nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != 500 {
		t.Errorf("Expected 500 for corrupted status file, got %d", resp.StatusCode)
	}
}

func TestUnknownRoute(t *testing.T) {
	app := SetupApp()
	req := httptest.NewRequest("GET", "/v1/nonexistent", nil)

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("Expected 404 for unknown route, got %d", resp.StatusCode)
	}
}
