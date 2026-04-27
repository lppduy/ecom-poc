#!/usr/bin/env bash
set -euo pipefail

CATALOG_URL="${CATALOG_URL:-http://localhost:8081}"
CART_URL="${CART_URL:-http://localhost:8082}"
ORDER_URL="${ORDER_URL:-http://localhost:8083}"
USER_ID="smoke-user-$(date +%s)"
IDEM_KEY="idem-smoke-$(date +%s)"

PASS=0
FAIL=0

# ─── Helpers ─────────────────────────────────────────────────────────────────

green() { printf "\033[32m✔ %s\033[0m\n" "$*"; }
red()   { printf "\033[31m✘ %s\033[0m\n" "$*"; }

assert_field() {
  local label="$1" json="$2" field="$3" expected="$4"
  local actual
  actual=$(echo "$json" | python3 -c "import sys,json; print(json.load(sys.stdin).get('$field',''))" 2>/dev/null || echo "")
  if [ "$actual" = "$expected" ]; then
    green "$label: $field=$actual"
    PASS=$((PASS+1))
  else
    red "$label: expected $field='$expected', got '$actual'"
    red "  response: $json"
    FAIL=$((FAIL+1))
  fi
}

assert_empty_items() {
  local label="$1" json="$2"
  local count
  count=$(echo "$json" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('items',[])))" 2>/dev/null || echo "-1")
  if [ "$count" = "0" ]; then
    green "$label: items=[]"
    PASS=$((PASS+1))
  else
    red "$label: expected empty items, got $count item(s)"
    red "  response: $json"
    FAIL=$((FAIL+1))
  fi
}

assert_error() {
  local label="$1" json="$2" expected="$3"
  local actual
  actual=$(echo "$json" | python3 -c "import sys,json; print(json.load(sys.stdin).get('error',''))" 2>/dev/null || echo "")
  if [ "$actual" = "$expected" ]; then
    green "$label: error='$actual'"
    PASS=$((PASS+1))
  else
    red "$label: expected error='$expected', got '$actual'"
    red "  response: $json"
    FAIL=$((FAIL+1))
  fi
}

# ─── Steps ───────────────────────────────────────────────────────────────────

echo ""
echo "╔══════════════════════════════════════════╗"
echo "║        ecom-poc  smoke-flow test         ║"
echo "╚══════════════════════════════════════════╝"
echo "  user=$USER_ID"
echo "  idem=$IDEM_KEY"
echo ""

# 1. Health checks
echo "── Health checks ──────────────────────────"
for svc in "$CATALOG_URL" "$CART_URL" "$ORDER_URL"; do
  r=$(curl -sf "$svc/health" 2>/dev/null || echo '{"status":"down"}')
  assert_field "health $svc" "$r" "status" "ok"
done
echo ""

# 2. Add cart item
echo "── Cart ────────────────────────────────────"
r=$(curl -sf -X POST "$CART_URL/cart/items" \
  -H "Content-Type: application/json" \
  -d "{\"userId\":\"$USER_ID\",\"productId\":\"prod-smoke\",\"quantity\":3}" 2>/dev/null \
  || echo '{}')
assert_field "add cart item" "$r" "message" "item added to cart"

# 3. Get cart — should have 1 item
r=$(curl -sf "$CART_URL/cart?userId=$USER_ID" 2>/dev/null || echo '{"items":[]}')
count=$(echo "$r" | python3 -c "import sys,json; print(len(json.load(sys.stdin).get('items',[])))" 2>/dev/null || echo "0")
if [ "$count" -ge 1 ]; then
  green "get cart: $count item(s)"
  PASS=$((PASS+1))
else
  red "get cart: expected >=1 item, got $count"
  FAIL=$((FAIL+1))
fi
echo ""

# 4. Create order
echo "── Create order ────────────────────────────"
r=$(curl -sf -X POST "$ORDER_URL/orders" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $IDEM_KEY" \
  -d "{\"userId\":\"$USER_ID\"}" 2>/dev/null \
  || echo '{}')
assert_field "create order" "$r" "status" "PENDING"
ORDER_ID=$(echo "$r" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))" 2>/dev/null || echo "")

# 5. Idempotency — same key returns same order
r2=$(curl -sf -X POST "$ORDER_URL/orders" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $IDEM_KEY" \
  -d "{\"userId\":\"$USER_ID\"}" 2>/dev/null \
  || echo '{}')
id2=$(echo "$r2" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))" 2>/dev/null || echo "")
if [ "$ORDER_ID" = "$id2" ] && [ -n "$ORDER_ID" ]; then
  green "idempotency: same order id=$ORDER_ID"
  PASS=$((PASS+1))
else
  red "idempotency: expected id=$ORDER_ID, got id=$id2"
  FAIL=$((FAIL+1))
fi

# 6. Cart cleared after order
r=$(curl -sf "$CART_URL/cart?userId=$USER_ID" 2>/dev/null || echo '{"items":[]}')
assert_empty_items "cart cleared after order" "$r"
echo ""

# 7. Get order by id
echo "── Order state machine ─────────────────────"
r=$(curl -sf "$ORDER_URL/orders/$ORDER_ID" 2>/dev/null || echo '{}')
assert_field "get order" "$r" "status" "PENDING"

# 8. Confirm order → CONFIRMED
r=$(curl -sf -X PATCH "$ORDER_URL/orders/$ORDER_ID/confirm" 2>/dev/null || echo '{}')
assert_field "confirm order" "$r" "status" "CONFIRMED"

# 9. Confirm again → invalid transition
r=$(curl -s -X PATCH "$ORDER_URL/orders/$ORDER_ID/confirm" 2>/dev/null || echo '{}')
assert_error "confirm again (invalid)" "$r" "invalid status transition"

# 10. Create second order → FAIL it
echo ""
echo "── Fail flow ───────────────────────────────"
IDEM_FAIL="idem-fail-$(date +%s)"
curl -sf -X POST "$CART_URL/cart/items" \
  -H "Content-Type: application/json" \
  -d "{\"userId\":\"${USER_ID}-b\",\"productId\":\"prod-smoke\",\"quantity\":1}" > /dev/null 2>&1 || true
r=$(curl -sf -X POST "$ORDER_URL/orders" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $IDEM_FAIL" \
  -d "{\"userId\":\"${USER_ID}-b\"}" 2>/dev/null \
  || echo '{}')
ORDER_ID_B=$(echo "$r" | python3 -c "import sys,json; print(json.load(sys.stdin).get('id',''))" 2>/dev/null || echo "")
r=$(curl -sf -X PATCH "$ORDER_URL/orders/$ORDER_ID_B/fail" 2>/dev/null || echo '{}')
assert_field "fail order" "$r" "status" "FAILED"

# 11. Fail again → invalid transition
r=$(curl -s -X PATCH "$ORDER_URL/orders/$ORDER_ID_B/fail" 2>/dev/null || echo '{}')
assert_error "fail again (invalid)" "$r" "invalid status transition"

# ─── Summary ─────────────────────────────────────────────────────────────────
echo ""
echo "══════════════════════════════════════════════"
TOTAL=$((PASS+FAIL))
if [ "$FAIL" -eq 0 ]; then
  printf "\033[32m  ALL PASSED  (%d/%d)\033[0m\n" "$PASS" "$TOTAL"
else
  printf "\033[31m  FAILED: %d/%d passed\033[0m\n" "$PASS" "$TOTAL"
fi
echo "══════════════════════════════════════════════"
echo ""

[ "$FAIL" -eq 0 ]
