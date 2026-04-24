#!/bin/bash


# Fail-fast
set -e

# Argumen dari Go
NAME=$1
GAME=$2
FLAVOR=$3
RAM=$4

echo "=========================================="
echo "🚀 [OaC Master] Memulai K0re-Cloud"
echo "📦 Target : $NAME"
echo "🎮 Game   : $GAME"
echo "⚡ Flavor : $FLAVOR"
echo "💾 RAM    : $RAM"
echo "=========================================="

# Validasi argumen kosong
if [ -z "$NAME" ] || [ -z "$GAME" ] || [ -z "$FLAVOR" ] || [ -z "$RAM" ]; then
    echo "[Error] Argumen tidak lengkap! Format: entrypoint.sh <NAME> <GAME> <FLAVOR> <RAM>"
    exit 1
fi

# Persiapan Environment
echo "[Module: Prep] Mengeksekusi 01_env_prep.sh..."
bash scripts/modules/01_env_prep.sh "$NAME"

# Tuning Kernel (CPU/Network)
echo "[Module: Tune] Mengeksekusi 02_kernel_tuning.sh..."
bash scripts/modules/02_kernel_tuning.sh "$NAME" "$FLAVOR"

# Docker Runner
echo "[Module: Run] Mengeksekusi 03_docker_runner.sh..."
bash scripts/modules/03_docker_runner.sh "$NAME" "$GAME" "$RAM" "$FLAVOR"

echo "=========================================="
echo "🎯 [OaC Master] Rantai Eksekusi Selesai!"
echo "=========================================="
exit 0