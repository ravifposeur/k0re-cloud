#!/bin/bash

TARGET_NAME=$1
FLAVOR=$3

echo "  -> [Module: tune] Menganalisis parameter untuk profil: $FLAVOR"

touch "/tmp/k0re_opts_${TARGET_NAME}"

# 1. Pengecekan Akses Root (Penting!)
if [ "$EUID" -ne 0 ]; then
    echo "  -> [Module: tune] [WARNING] Berjalan tanpa 'sudo'. Menggunakan mode simulasi."
    echo "  -> [Module: tune] Menerapkan Low-Latency Network Profile (BBR & TCP Fast Open)... [SIMULATED]"
    echo "  -> [Module: tune] Menyiapkan environment flag: ZGC (Zero Garbage Collector)... [SIMULATED]"
    
    echo "-e JAVA_OPTS='-XX:+UseZGC -Xmx1G'" > "/tmp/k0re_opts_${TARGET_NAME}"
    
    exit 0
fi

echo "  -> [Module: tune] ⚠️ Akses Root terdeteksi! Memulai Injeksi Kernel (Gacor Mode)..."

# Mengganti algoritma antrean jaringan standar menjadi BBR
echo "  -> [Module: tune] Menginjeksi TCP BBR Congestion Control..."
sysctl -w net.core.default_qdisc=fq > /dev/null
sysctl -w net.ipv4.tcp_congestion_control=bbr > /dev/null

# Memperbesar buffer jaringan agar tidak ada paket hilang saat pemain ramai (TPS drop)
sysctl -w net.core.rmem_max=16777216 > /dev/null
sysctl -w net.core.wmem_max=16777216 > /dev/null
sysctl -w net.ipv4.tcp_rmem="4096 87380 16777216" > /dev/null
sysctl -w net.ipv4.tcp_wmem="4096 65536 16777216" > /dev/null

# Memaksa Linux menggunakan RAM murni untuk Minecraft dan menghindari Hard-Disk (Swap)
echo "  -> [Module: tune] Menekan Swappiness ke titik terendah (OOM Killer optimation)..."
sysctl -w vm.swappiness=1 > /dev/null
sysctl -w vm.dirty_ratio=10 > /dev/null
sysctl -w vm.dirty_background_ratio=5 > /dev/null
sysctl -w vm.max_map_count=262144 > /dev/null

# Mengaktifkan Transparent Huge Pages (THP) jika didukung kernel
if [ -f /sys/kernel/mm/transparent_hugepage/enabled ]; then
    echo "  -> [Module: tune] Mengaktifkan Transparent Huge Pages (THP)..."
    echo "madvise" > /sys/kernel/mm/transparent_hugepage/enabled
fi

# 4. CPU Scheduler (Khusus Profil Kompetitif)
if [ "$FLAVOR" = "pvp-competitive" ]; then
    echo "  -> [Module: tune] ⚡ Profil Kompetitif Terdeteksi! Memaksa CPU ke mode Performa..."
    # Memaksa CPU tidak bolak-balik istirahat (C-States) jika governor tersedia
    for cpu in /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor; do
        [ -f "$cpu" ] && echo "performance" > "$cpu" 2>/dev/null
    done
fi

# Menyimpan konfigurasi agar permanen
sysctl -p > /dev/null 2>&1

echo "  -> [Module: tune] ✅ OS Kernel Tuning selesai dan terkunci!"