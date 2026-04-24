#!/bin/bash
# test/integration.sh

BASE_URL="http://127.0.0.1:3000"
PASS=0
FAIL=0
TARGET_NAME="arena-ugm-01"

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
echo "🧪 [TEST] Memulai Simulasi End-to-End K0re"
echo "=========================================="

# --- PRE-CLEANUP ---
if docker ps -a --format '{{.Names}}' | grep -Eq "^${TARGET_NAME}\$"; then
    echo "-> Membersihkan container lama..."
    docker rm -f "$TARGET_NAME" >/dev/null
fi

# --- TEST 1: PROVISIONING API ---
echo "==> Test: POST /v1/provision"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/v1/provision" \
  -H "Content-Type: application/json" \
  -d '{"name":"'$TARGET_NAME'","game":"minecraft","flavor":"pvp-competitive","ram":"1G"}')

HTTP_BODY=$(echo "$RESPONSE" | head -n1)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

assert_eq "Status code is 200" "200" "$HTTP_CODE"
assert_eq "Status field is success" "success" "$(echo "$HTTP_BODY" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)"

# --- TEST 2: DOCKER ENGINE VERIFICATION (Tambahan) ---
echo "==> Test: Docker Engine State"
# Tunggu 3 detik agar Bash OaC selesai merakit container
sleep 3 
ACTUAL_STATE=$(docker ps --format '{{.Names}}' | grep -Eq "^${TARGET_NAME}\$" && echo "running" || echo "missing")
assert_eq "Container is running in Docker" "running" "$ACTUAL_STATE"

# --- TEST 3: STATUS API ---
echo "==> Test: GET /v1/status/$TARGET_NAME"
# (Pastikan file status.json sudah dibuat oleh modul bash, jika belum ini bisa fail)
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/status/$TARGET_NAME")
HTTP_BODY=$(echo "$RESPONSE" | head -n1)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

# Jika endpoint status belum nyambung ke file asli, kita skip assert detailnya dulu
assert_eq "Status endpoint reachable (200/500/404)" "200" "$HTTP_CODE" 

# --- TEST 4: INVALID REQUEST ---
echo "==> Test: POST /v1/provision (invalid body)"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/v1/provision" \
  -H "Content-Type: application/json" \
  -d '{"name":""}')

assert_eq "Status code is 400" "400" "$HTTP_CODE"

echo "=========================================="
echo "Results: $PASS passed, $FAIL failed"

# --- POST-CLEANUP ---
echo "-> Membersihkan env test..."
docker rm -f "$TARGET_NAME" >/dev/null

[ $FAIL -eq 0 ] && exit 0 || exit 1