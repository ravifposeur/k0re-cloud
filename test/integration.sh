#!/bin/bash

BASE_URL="http://127.0.0.1:3000"
PASS=0
FAIL=0
TARGET_NAME="arena-cli-test"

assert_eq() {
  local desc="$1" expected="$2" actual="$3"
  if [ "$actual" = "$expected" ]; then
    echo "  [PASS] $desc"
    ((PASS++))
  else
    echo "  [FAIL] $desc - expected '$expected', got '$actual'"
    ((FAIL++))
  fi
}

echo "=========================================="
echo "🧪 [FINAL TEST]"
echo "=========================================="

if docker ps -a --format '{{.Names}}' | grep -Eq "^${TARGET_NAME}\$"; then
    echo "-> Membersihkan container lama..."
    docker rm -f "$TARGET_NAME" >/dev/null
fi

echo "==> Test 1: Menulis file server.yaml (Simulasi User)"
cat <<EOF > /tmp/test-server.yaml
server:
  name: "$TARGET_NAME"
  game: "minecraft"
  flavor: "pvp-competitive"
  resources:
    ram: "1G"
EOF
assert_eq "File YAML terbuat" "true" "$([ -f /tmp/test-server.yaml ] && echo 'true' || echo 'false')"

echo "==> Test 2: Menjalankan perintah 'k0re apply -f'"
CLI_OUTPUT=$(go run cmd/cli/main.go apply -f /tmp/test-server.yaml 2>&1)
CLI_EXIT_CODE=$?

assert_eq "CLI berjalan tanpa error (Exit 0)" "0" "$CLI_EXIT_CODE"

SUCCESS_MSG=$(echo "$CLI_OUTPUT" | grep -o "provisioned successfully" || echo "not found")
assert_eq "CLI menerima balasan sukses dari Daemon" "provisioned successfully" "$SUCCESS_MSG"

echo "==> Test 3: Cek status mesin Docker di Kernel"
# Tunggu 3 detik agar Bash OaC selesai merakit container
sleep 3 
ACTUAL_STATE=$(docker ps --format '{{.Names}}' | grep -Eq "^${TARGET_NAME}\$" && echo "running" || echo "missing")
assert_eq "Container benar-benar menyala" "running" "$ACTUAL_STATE"

echo "==> Test 4: Membaca status.json via Daemon"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/status/$TARGET_NAME")
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

assert_eq "Endpoint status merespons (200 OK)" "200" "$HTTP_CODE" 

echo "=========================================="
echo "Results: $PASS passed, $FAIL failed"

echo "-> Membersihkan env test..."
docker rm -f "$TARGET_NAME" >/dev/null
rm -f /tmp/test-server.yaml

[ $FAIL -eq 0 ] && exit 0 || exit 1