#!/bin/bash
# test/integration.sh

BASE_URL="http://127.0.0.1:3000"
PASS=0
FAIL=0

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

echo "==> Test: POST /v1/provision"
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/v1/provision" \
  -H "Content-Type: application/json" \
  -d '{"name":"arena-ugm-01","game":"minecraft","flavor":"pvp-competitive","ram":"2G"}')

HTTP_BODY=$(echo "$RESPONSE" | head -n1)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

assert_eq "Status code is 200" "200" "$HTTP_CODE"
assert_eq "Status field is success" "success" "$(echo "$HTTP_BODY" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)"

echo "==> Test: GET /v1/status/arena-ugm-01"
RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/v1/status/arena-ugm-01")
HTTP_BODY=$(echo "$RESPONSE" | head -n1)
HTTP_CODE=$(echo "$RESPONSE" | tail -n1)

assert_eq "Status code is 200" "200" "$HTTP_CODE"
assert_eq "Name matches" "arena-ugm-01" "$(echo "$HTTP_BODY" | grep -o '"name":"[^"]*"' | cut -d'"' -f4)"

echo "==> Test: POST /v1/provision (invalid body)"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE_URL/v1/provision" \
  -H "Content-Type: application/json" \
  -d '{"name":""}')

assert_eq "Status code is 400" "400" "$HTTP_CODE"

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1