# Pembagian Tugas Tim K0re

Dokumen ini berisi alokasi 5 paket kerja kepada 5 anggota tim. Semua paket wajib mengikuti kontrak teknis yang tercantum di `DEVELOPMENT.md`.

## Ringkasan Tugas

| ID | Paket Kerja | Penanggung Jawab |
|----|-------------|------------------|
| 1 | CLI & Parser YAML | Anggota 1 |
| 2 | Daemon & REST API | Anggota 2 |
| 3 | Orchestrator + Operations as Code (OaC) | Ravif |
| 4 | Integrasi & Setup Proyek | Anggota 4 |
| 5 | Testing & Dokumentasi | Anggota 5 |

---

## Rincian per Paket

### Paket 1: CLI & Parser YAML (Anggota 1)

- **Output**: Binary `k0re` dengan perintah `apply -f <file>`
- **Fungsi utama**:
  - Parse YAML ke struct Go menggunakan `gopkg.in/yaml.v3`
  - Validasi minimal (wajib ada `server.name`, `server.resources.ram`)
  - Kirim POST request ke `http://daemon:3000/v1/provision` dengan body JSON sesuai kontrak
- **Catatan**: Tidak perlu menangani logika container atau tuning.

### Paket 2: Daemon & REST API (Anggota 2)

- **Output**: Binary `k0red` dengan endpoint REST
- **Fungsi utama**:
  - Endpoint `POST /v1/provision` → terima JSON, validasi, lalu panggil `orchestrator/factory.go`
  - Endpoint `GET /v1/status/{name}` → ambil status dari file `/var/k0re/<name>/status.json`
  - Logging setiap request ke stdout

### Paket 3: Orchestrator & Operations as Code (OaC) (Ravif)

- **Output**:
  - File `internal/orchestrator/factory.go`
  - Seluruh skrip Bash di `scripts/` dan `scripts/modules/`
- **Fungsi utama**:
  - `factory.go`: fungsi `BuildEnv(flavor, ram) []string` menyusun variabel lingkungan berdasarkan flavor, lalu memanggil `exec.Command("scripts/entrypoint.sh", name, game, flavor, ram)`
  - `entrypoint.sh`: master script yang memanggil modul OaC secara berurutan (01, 02, 03)
  - `01_env_prep.sh`: membuat direktori `/var/k0re/<NAME>` dan mengatur permission
  - `02_kernel_tuning.sh`: menerapkan `sysctl` untuk TCP buffer dan CPU pinning via `taskset` (jika flavor `pvp-competitive`)
  - `03_docker_runner.sh`: menjalankan `docker run -d --name <NAME> --memory <RAM> -p 25565:25565 alpine:latest` (atau image game yang sesuai)

### Paket 4: Integrasi & Setup Proyek (Anggota 4)

- **Output**:
  - `go.mod`, `go.sum`, struktur direktori proyek
  - `Makefile` dengan target `build`, `test`, `clean`
  - Skrip `build.sh` jika diperlukan
- **Fungsi utama**:
  - Menyusun `go.mod` dengan dependensi: `github.com/gofiber/fiber/v2`, `gopkg.in/yaml.v3`
  - Membuat `Makefile` untuk mengkompilasi `cmd/cli/main.go` dan `cmd/daemon/main.go` ke `bin/k0re` dan `bin/k0red`
  - Menyiapkan environment agar daemon dan CLI dapat berjalan di satu mesin (simulasi lokal)
- **Catatan**: Pastikan semua path relatif sesuai struktur.

### Paket 5: Testing & Dokumentasi (Anggota 5)

- **Output**:
  - Skrip `test/integration.sh` (menggunakan `curl` dan `assert`)
  - `README.md` (sudah disediakan, tetapi Anggota 5 memastikan kontennya sesuai dengan hasil akhir)
  - Dokumentasi API (file `docs/api.yaml` atau `docs/api.md` dalam format OpenAPI)
  - Laporan akhir proyek (PDF atau Markdown) yang berisi ringkasan, kendala, dan demo
- **Fungsi utama**:
  - Menulis integration test yang menjalankan daemon, CLI apply, lalu memverifikasi container benar-benar berjalan
  - Membuat dokumentasi API dari kontrak yang sudah ditetapkan
  - Menggabungkan laporan akhir dari semua anggota

---

## Batasan & Aturan

- Dilarang menambahkan GUI, dashboard web, atau antarmuka grafis apapun.
- Dilarang menggunakan database eksternal (MySQL, PostgreSQL, dll).
- Semua konfigurasi harus melalui YAML dan environment variable, bukan hardcode.
- Setiap anggota wajib mengikuti kontrak API dan kontrak Bash yang sudah ditetapkan di `DEVELOPMENT.md`.
