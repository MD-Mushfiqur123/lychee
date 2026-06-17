#!/bin/sh
# Lychee API Smoke Test
# Tests all major endpoints against a running Lychee server

HOST="${LYCHEE_HOST:-http://localhost:11434}"
PASS=0
FAIL=0

check() {
  local name="$1"
  local method="$2"
  local url="$3"
  local data="$4"
  local expected="${5:-200}"

  if [ -n "$data" ]; then
    status=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "$HOST$url" -H "Content-Type: application/json" -d "$data" 2>&1)
  else
    status=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "$HOST$url" 2>&1)
  fi

  if [ "$status" = "$expected" ]; then
    echo "✅ $name"
    PASS=$((PASS + 1))
  else
    echo "❌ $name (got $status, expected $expected)"
    FAIL=$((FAIL + 1))
  fi
}

echo "Lychee Smoke Test ($HOST)"
echo "=========================="

check "Health" "GET" "/"
check "Version" "GET" "/api/version"
check "List models" "GET" "/api/tags"
check "Running models" "GET" "/api/ps"

echo ""
echo "Results: $PASS passed, $FAIL failed"
[ $FAIL -eq 0 ] && exit 0 || exit 1
