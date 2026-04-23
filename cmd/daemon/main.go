package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// ProvisionRequest matches the API contract in DEVELOPMENT.md.
type ProvisionRequest struct {
	Name   string `json:"name"`
	Game   string `json:"game"`
	Flavor string `json:"flavor"`
	RAM    string `json:"ram"`
}

// APIResponse is the standard envelope for all non-status responses.
type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// StatusResponse matches the GET /v1/status/:name contract in DEVELOPMENT.md.
type StatusResponse struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	CPUUsage string `json:"cpu_usage"`
	MemUsage string `json:"mem_usage"`
}

// safeNamePattern enforces strict naming to prevent path traversal attacks.
var safeNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]*[a-z0-9]$`)

func isValidName(name string) bool {
	return safeNamePattern.MatchString(name)
}

// SetupApp initializes Fiber, configures middleware, and registers routes.
func SetupApp() *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var fiberErr *fiber.Error

			if ok := errors.As(err, &fiberErr); ok {
				code = fiberErr.Code
			}
			return c.Status(code).JSON(APIResponse{
				Status:  "error",
				Message: err.Error(),
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
		Output: os.Stdout,
	}))

	v1 := app.Group("/v1")
	v1.Post("/provision", handleProvision)
	v1.Get("/status/:name", handleStatus)

	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Status:  "error",
			Message: fmt.Sprintf("Route %s %s not found", c.Method(), c.Path()),
		})
	})

	return app
}

func main() {
	app := SetupApp()

	host := os.Getenv("K0RED_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	port := os.Getenv("K0RED_PORT")
	if port == "" {
		port = "3000"
	}
	address := fmt.Sprintf("%s:%s", host, port)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("k0red daemon starting on http://%s\n", address)
		if err := app.Listen(address); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	<-sigCh
	fmt.Println("\nShutting down k0red daemon gracefully...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	fmt.Println("k0red stopped.")
}

// handleProvision processes POST /v1/provision requests.
func handleProvision(c *fiber.Ctx) error {
	var req ProvisionRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Status:  "error",
			Message: "Request body bukan JSON yang valid",
		})
	}

	if req.Name == "" || req.Game == "" || req.Flavor == "" || req.RAM == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Status:  "error",
			Message: "Semua field (name, game, flavor, ram) wajib diisi",
		})
	}

	if !isValidName(req.Name) {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Status:  "error",
			Message: "Field 'name' hanya boleh mengandung huruf kecil, angka, dan tanda hubung",
		})
	}

	// Idempotency and state check
	targetDir := fmt.Sprintf("/var/k0re/%s", req.Name)
	if _, err := os.Stat(targetDir); err == nil {
		return c.Status(fiber.StatusConflict).JSON(APIResponse{
			Status:  "error",
			Message: fmt.Sprintf("Server '%s' sudah ada", req.Name),
		})
	} else if !errors.Is(err, os.ErrNotExist) {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Status:  "error",
			Message: fmt.Sprintf("Gagal memeriksa direktori server: %v", err),
		})
	}

	// TODO: Replace stub with Package 3 Orchestrator call
	log.Printf("[Orchestrator Stub] Provisioning name=%s game=%s flavor=%s ram=%s\n",
		req.Name, req.Game, req.Flavor, req.RAM)

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Status:  "success",
		Message: "Server provisioned",
	})
}

// handleStatus processes GET /v1/status/:name requests.
func handleStatus(c *fiber.Ctx) error {
	serverName := c.Params("name")

	if !isValidName(serverName) {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Status:  "error",
			Message: "Parameter 'name' tidak valid",
		})
	}

	filePath := fmt.Sprintf("/var/k0re/%s/status.json", serverName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).JSON(APIResponse{
				Status:  "error",
				Message: fmt.Sprintf("Server '%s' tidak ditemukan", serverName),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Status:  "error",
			Message: "Gagal membaca status file",
		})
	}

	// Validate JSON structure before returning to client
	var status StatusResponse
	if err := json.Unmarshal(data, &status); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Status:  "error",
			Message: "Status file rusak atau format tidak valid",
		})
	}

	return c.Status(fiber.StatusOK).JSON(status)
}
