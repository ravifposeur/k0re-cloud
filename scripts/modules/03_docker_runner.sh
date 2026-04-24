#!/bin/bash
# scripts/modules/03_docker_runner.sh

NAME=$1
GAME=$2
RAM=$3
FLAVOR=$4

# Mengambil variabel dari Modul 1 dan 2
BASE_DIR=$(cat "/tmp/k0re_dir_$NAME")
JVM_OPTS=$(cat "/tmp/k0re_opts_$NAME" 2>/dev/null || echo "")

echo "  -> [Module: run] Merakit container engine..."

# Memetakan Game Engine ke Docker Image
DOCKER_IMAGE="alpine:latest" # Default fallback
if [ "$GAME" == "minecraft" ]; then
    DOCKER_IMAGE="itzg/minecraft-server:java21"
fi

# Mengecek apakah container dengan nama ini sudah ada
if docker ps -a --format '{{.Names}}' | grep -Eq "^${NAME}\$"; then
    echo "  -> [Module: run] ⚠️  Peringatan: Container $NAME sudah ada! Melakukan override..."
    docker rm -f "$NAME" >/dev/null
fi

# Mengeksekusi Docker dengan Port Statis dan Env Vars yang aman
docker run -d \
    --name "$NAME" \
    -e EULA=TRUE \
    -e TYPE=PAPER \
    -e MEMORY="$RAM" \
    -e JVM_OPTS="$JVM_OPTS" \
    -e ONLINE_MODE=FALSE \
    -v "$BASE_DIR:/data" \
    -p 25565:25565 \
    "$DOCKER_IMAGE" >/dev/null

echo "  -> [Module: run] ✅ Container berhasil diisolasi di Kernel!"

# Activity Logging (Audit Trail)
LOG_FILE="./k0re-data/audit.log"
echo "[$(date +"%Y-%m-%d %H:%M:%S")] ACTION=PROVISION SERVER=$NAME GAME=$GAME STATUS=SUCCESS" >> "$LOG_FILE"

echo "  -> [Module: run] ✅ Metrik dan Audit Log telah diperbarui."