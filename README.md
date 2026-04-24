
# K0re Cloud: Orchestrator PaaS untuk Komunitas Akar Rumput

**K0re Cloud** adalah sebuah *Platform-as-a-Service* (PaaS) minimalis yang dirancang untuk mengorkestrasi *game server* (fokus saat ini: Minecraft) di atas infrastruktur Linux menggunakan Docker. Proyek ini dibangun dengan filosofi efisiensi sumber daya, kemandirian teknologi, dan penolakan terhadap ketergantungan pada penyedia layanan awan komersial yang membebani komunitas dengan biaya *overhead* tinggi.

---

## 🛠️ Arsitektur Sistem

K0re Cloud menggunakan pendekatan **Operations-as-Code (OaC)** dengan memisahkan logika orkestrasi menjadi tiga lapisan utama:

1.  **Core Daemon (Go/Fiber):** Pusat kendali yang mengelola permintaan API, validasi parameter, dan koordinasi modul secara asinkron.
2.  **Logic Modules (Bash):** Skrip *low-level* yang berinteraksi langsung dengan Kernel Linux dan Docker Engine untuk optimasi sistem dan manajemen kontainer.
3.  **CLI Controller (Go):** Antarmuka baris perintah tunggal untuk manajemen infrastruktur secara efisien.

---

## 🚀 Panduan Instalasi & Prasyarat

Sebelum melakukan kompilasi, pastikan sistem Anda telah terpasang paket-paket dasar berikut (Instruksi untuk distribusi berbasis Debian/Ubuntu):

```bash
# Update repository dan install prasyarat sistem
sudo apt update
sudo apt install -y docker.io docker-compose-v2 make golang-go
```

Pastikan pengguna aktif memiliki izin akses ke Docker Engine tanpa `sudo`:
```bash
sudo usermod -aG docker $USER
# Harap log out dan log in kembali untuk menerapkan perubahan grup
```

### Langkah-langkah Build
1. Kloning repositori:
   ```bash
   git clone https://github.com/username/k0re-cloud.git
   cd k0re-cloud
   ```
2. Jalankan proses build melalui Makefile:
   ```bash
   make build
   sudo make install
   ```

---

## 🛠️ Alur Kerja Makefile

Proyek ini menggunakan `Makefile` untuk menstandardisasi proses pengembangan dan operasional:

| Perintah | Deskripsi |
| :--- | :--- |
| `make build` | Mengompilasi kode sumber Go menjadi biner `k0red` (Daemon) dan `k0re` (CLI). |
| `make install` | Memasang biner CLI ke `/usr/local/bin` agar dapat diakses secara global. |
| `make test` | Menjalankan pengujian *End-to-End* terintegrasi untuk memvalidasi seluruh siklus hidup server. |
| `make run` | Menjalankan Daemon secara langsung di terminal untuk keperluan pemantauan log/debugging. |
| `make clean` | Menghapus biner hasil kompilasi, folder data sementara, dan log audit. |

---

## 📖 Penggunaan CLI

| Perintah | Deskripsi | Contoh |
| :--- | :--- | :--- |
| **apply** | Deployment server berdasarkan file YAML | `k0re apply -f config.yaml` |
| **status** | Memantau metrik live (CPU/RAM) | `k0re status arena-ugm-01` |
| **stop** | Menghentikan operasional server | `k0re stop arena-ugm-01` |
| **start** | Menyalakan kembali server yang terhenti | `k0re start arena-ugm-01` |
| **destroy** | Menghapus server dan seluruh datanya | `k0re destroy arena-ugm-01` |
| **audit** | Melihat riwayat aktivitas aktivitas | `k0re audit` |

---

## 📄 Dokumentasi API

Sistem ini menggunakan standar **OpenAPI 3.0**. Definisi teknis mengenai *endpoint* dan skema data tersedia pada file:
👉 [**api.yaml](./api.yaml)**

---

## ⚖️ Perspektif dan Filosofi

K0re Cloud membuktikan bahwa skalabilitas tidak harus selalu berarti kompleksitas. Dengan mengoptimalkan perkakas bawaan sistem operasi (Bash/Docker) dan efisiensi bahasa Go, kita dapat membangun infrastruktur mandiri yang stabil tanpa harus tunduk pada ekosistem *proprietary* yang mahal. Teknologi ini adalah bentuk kedaulatan digital bagi komunitas akar rumput.

---

**Kontribusi & Lisensi**
Proyek ini bersifat terbuka untuk pengembangan komunitas yang berorientasi pada kemandirian digital.

---

### Perlu bantuan tambahan?
Gunakan perintah `k0re --help` untuk panduan cepat di terminal.