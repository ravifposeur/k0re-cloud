
# Pengembangan K0re

Dokumen ini untuk kontributor yang ingin memahami arsitektur internal, kontrak API, dan cara melakukan build serta testing.

## Arsitektur Teknis

K0re terdiri dari dua binary independen:

- **`k0re` (CLI)**: Parser YAML, HTTP client.
- **`k0red` (Daemon)**: REST server, orchestrator, eksekutor OaC.

Komunikasi: HTTP REST over loopback (default `127.0.0.1:3000`).

## Tech Stack

| Komponen | Teknologi |
|----------|-----------|
| CLI & Daemon | Go 1.21+ (Fiber v2 untuk REST) |
| Operations as Code | Bash 5+ (scripts/modules/) |
| Isolasi & container | Docker Engine API, cgroups |
| Tuning kernel | sysctl, taskset (Linux) |

## Struktur Repositori

```
k0re-paas/
├── cmd/
│   ├── cli/main.go
│   └── daemon/main.go
├── internal/orchestrator/factory.go
├── scripts/│   ├── entrypoint.sh
│   └── modules/
│       ├── 01_env_prep.sh
│       ├── 02_kernel_tuning.sh
│       └── 03_docker_runner.sh
├── examples/
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## Kontrak API (Daemon)

### Endpoint: `POST /v1/provision`

**Request Body (JSON):**

```json
{
  "name": "arena-ugm-01",
  "game": "minecraft",
  "flavor": "pvp-competitive",
  "ram": "2G"
}
```

**Response Sukses (200):**

```json
{
  "status": "success",
  "message": "Server provisioned"
}
```

**Response Gagal (400/500):**

```json
{
  "status": "error",
  "message": "Deskripsi error"
}
```

### Endpoint: `GET /v1/status/{name}`

**Response (200):**

```json
{
  "name": "arena-ugm-01",
  "status": "running",
  "cpu_usage": "12%",
  "mem_usage": "512MB/2GB"
}
```

## Kontrak Eksekusi Bash (OaC)

`factory.go` akan memanggil `entrypoint.sh` dengan urutan argumen berikut:

```bash
./scripts/entrypoint.sh <NAME> <GAME> <FLAVOR> <RAM>
```

Contoh:

```bash
./scripts/entrypoint.sh arena-ugm-01 minecraft pvp-competitive 2G
```

### Wajib dipatuhi oleh semua modul Bash:

- Setiap modul harus mengembalikan exit code `0` jika sukses, bukan `0` jika gagal.
- Semua output log diawali dengan `[Module: nama_module]`.
- Direktori persistensi: `/var/k0re/<NAME>`.

## Build & Testing

### Build binary

```bash
make build
# atau manual:
go build -o bin/k0re cmd/cli/main.go
go build -o bin/k0red cmd/daemon/main.go
```

### Testing integrasi

```bash
make test
# atau bash test/integration.sh
```

### Simulasi lokal

1. Jalankan daemon: `./bin/k0red`
2. Di terminal lain: `./bin/k0re apply -f examples/arena-pvp.yaml`

## Panduan Kontribusi

1. Ikuti struktur direktori yang sudah ditetapkan.
2. Gunakan `gofmt` untuk kode Go.
3. Setiap modul Bash harus idempoten (dapat dijalankan berulang kali tanpa efek samping).
4. Perbarui dokumentasi jika menambah fitur.