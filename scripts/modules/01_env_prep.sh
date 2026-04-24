#!/bin/bash
set -e

NAME=$1
# Sesuai Kontrak Spesifikasi di TEAM.md
BASE_DIR="/var/k0re/$NAME"

echo "  -> [Module: prep] Menyiapkan isolasi direktori untuk $NAME..."

# Membuat direktori (butuh akses sudo/root jika menggunakan /var)
# Untuk development lokal yang aman, kita buat fallback ke /tmp jika akses /var ditolak
if mkdir -p "$BASE_DIR" 2>/dev/null; then
    chmod 777 "$BASE_DIR"
    echo "  -> [Module: prep] Direktori persisten siap di $BASE_DIR"
else
    # Fallback untuk development lokal (tanpa sudo)
    BASE_DIR="./k0re-data/$NAME"
    mkdir -p "$BASE_DIR"
    chmod 777 "$BASE_DIR"
    echo "  -> [Module: prep] Direktori persisten (Dev Mode) siap di $BASE_DIR"
fi

# Menyimpan lokasi direktori ini agar bisa dibaca oleh modul selanjutnya
echo "$BASE_DIR" > "/tmp/k0re_dir_$NAME"