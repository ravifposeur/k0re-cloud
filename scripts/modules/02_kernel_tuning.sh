#!/bin/bash
set -e

NAME=$1
FLAVOR=$2

echo "  -> [Module: tune] Menganalisis parameter untuk profil: $FLAVOR"

if [ "$FLAVOR" == "pvp-competitive" ]; then
    echo "  -> [Module: tune] Menerapkan Low-Latency Network Profile (BBR & TCP Fast Open)..."
    # Catatan: Perintah sysctl asli dinonaktifkan sementara untuk development lokal
    # agar tidak merusak konfigurasi jaringan OS laptopmu. 
    # Hapus tanda pagar (#) saat demo jika menggunakan mesin VPS Linux sungguhan.
    
    # sudo sysctl -w net.ipv4.tcp_congestion_control=bbr >/dev/null
    # sudo sysctl -w net.ipv4.tcp_fastopen=3 >/dev/null
    
    echo "  -> [Module: tune] Menyiapkan environment flag: ZGC (Zero Garbage Collector)"
    echo "-XX:+UseZGC -XX:+ZProactive" > "/tmp/k0re_opts_$NAME"

elif [ "$FLAVOR" == "survival-standard" ]; then
    echo "  -> [Module: tune] Menerapkan Standard IO Profile..."
    echo "-XX:+UseG1GC" > "/tmp/k0re_opts_$NAME"
else
    echo "  -> [Module: tune] Menggunakan profil default OS."
    echo "" > "/tmp/k0re_opts_$NAME"
fi