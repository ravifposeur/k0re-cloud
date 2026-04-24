package orchestrator

import (
	"fmt"
	"os"
	"os/exec"
)

// ServerPayload adalah kontrak struct yang wajib sama dengan Anggota 1 & 2
type ServerPayload struct {
	Name   string `json:"name"`
	Game   string `json:"game"`
	Flavor string `json:"flavor"`
	Ram    string `json:"ram"`
}

// Provision adalah fungsi utama yang akan dipanggil oleh Daemon REST API
func Provision(payload ServerPayload) error {
	fmt.Printf("🛠️  [Orchestrator] Merakit perintah OaC untuk server: %s\n", payload.Name)

	// Memanggil skrip Bash entrypoint.sh dan mengirimkan 4 argumen berurutan
	cmd := exec.Command(
		"bash",
		"scripts/entrypoint.sh",
		payload.Name,
		payload.Game,
		payload.Flavor,
		payload.Ram,
	)

	// Mengarahkan output Bash (stdout/stderr) agar tampil di terminal Go
	// Ini sangat penting agar proses OaC terlihat log-nya
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Jalankan perintah dan tunggu sampai selesai
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("eksekusi OaC gagal: %v", err)
	}

	fmt.Printf("✅ [Orchestrator] Provisioning %s selesai tanpa error.\n", payload.Name)
	return nil
}
