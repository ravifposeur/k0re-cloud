package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"k0re/internal/orchestrator" // <-- Import Jembatan Go buatanmu

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
	Name        string `json:"name"`
	Status      string `json:"status"`
	CPUUsage    string `json:"cpu_usage"`
	MemUsage    string `json:"mem_usage"`
	LastUpdated string `json:"last_updated"`
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
	v1.Get("/audit", handleAudit)

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

	// --- INTEGRASI KE ORCHESTRATOR ---
	payload := orchestrator.ServerPayload{
		Name:   req.Name,
		Game:   req.Game,
		Flavor: req.Flavor,
		Ram:    req.RAM,
	}

	if err := orchestrator.Provision(payload); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Status:  "error",
			Message: fmt.Sprintf("Eksekusi OaC Gagal: %v", err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(APIResponse{
		Status:  "success",
		Message: "Server provisioned",
	})
}

// handleStatus processes GET /v1/status/:name requests (LIVE DOCKER METRICS)
func handleStatus(c *fiber.Ctx) error {
	serverName := c.Params("name")

	if !isValidName(serverName) {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Status:  "error",
			Message: "Parameter 'name' tidak valid",
		})
	}

	// 1. Cek status container langsung ke Docker Engine
	cmdCheck := exec.Command("docker", "inspect", "-f", "{{.State.Status}}", serverName)
	stateOut, err := cmdCheck.Output()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Status:  "error",
			Message: fmt.Sprintf("Server '%s' tidak ditemukan atau mati", serverName),
		})
	}
	status := strings.TrimSpace(string(stateOut))

	// 2. Ambil metrik live dari Docker Stats
	cmdStats := exec.Command("docker", "stats", serverName, "--no-stream", "--format", "{{.CPUPerc}}|{{.MemUsage}}")
	statsOut, _ := cmdStats.Output()

	// Parsing output Docker ("2.5%|50MiB / 2GiB")
	statsRaw := strings.TrimSpace(string(statsOut))
	statsParts := strings.Split(statsRaw, "|")

	cpuUsage := "0.00%"
	memUsage := "0B"

	if len(statsParts) == 2 {
		// 1. Normalisasi CPU (Gaya Task Manager)
		cpuRaw := strings.TrimSuffix(statsParts[0], "%")
		cpuFloat, err := strconv.ParseFloat(cpuRaw, 64)
		if err == nil {
			numCores := float64(runtime.NumCPU()) // Mengambil jumlah core laptop/server
			normalizedCPU := cpuFloat / numCores
			cpuUsage = fmt.Sprintf("%.2f%%", normalizedCPU)
		} else {
			cpuUsage = statsParts[0]
		}

		// 2. Bersihkan Format RAM
		rawMem := statsParts[1]
		memSplit := strings.Split(rawMem, " / ")
		if len(memSplit) > 0 {
			memUsage = memSplit[0] // Hanya mengambil memori yang terpakai
		} else {
			memUsage = rawMem
		}
	}

	// 3. Kembalikan respons JSON secara real-time
	response := StatusResponse{
		Name:        serverName,
		Status:      status,
		CPUUsage:    cpuUsage,
		MemUsage:    memUsage,
		LastUpdated: time.Now().Format(time.RFC3339),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func handleAudit(c *fiber.Ctx) error {
	data, err := os.ReadFile("./k0re-data/audit.log")
	if err != nil {
		return c.Status(500).SendString("Gagal membaca log audit")
	}
	return c.Status(200).SendString(string(data))
}
