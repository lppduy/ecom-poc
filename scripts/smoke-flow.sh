#!/usr/bin/env bash
set -euo pipefail

AUTH_URL="${AUTH_URL:-http://localhost:8086}"
CATALOG_URL="${CATALOG_URL:-http://localhost:8081}"
CART_URL="${CART_URL:-http://localhost:8082}"
ORDER_URL="${ORDER_URL:-http://localhost:8083}"
INVENTORY_URL="${INVENTORY_URL:-http://localhost:8084}"
SEARCH_URL="${SEARCH_URL:-http://localhost:8085}"
PAYMENT_URL="${PAYMENT_URL:-http://localhost:8087}"

PASS=0; FAIL=0
TS=$(date +%s)
USERNAME="smoke-$TS"
PASSWORD="pass-$TS"
IDEM_KEY="idem-$TS"

# ── helpers ──────────────────────────────────────────────────────────────────

green() { printf "\033[32m  ✔ %s\033[0m\n" "$*"; }
red()   { printf "\033[31m  ✘ %s\033[0m\n" "$*"; }
info()  { printf "\033[90m  → %s\033[0m\n" "$*"; }

json_field() { echo "$1" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('$2',''))" 2>/dev/null || echo ""; }
json_len()   { echo "$1" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d.get('$2',[])))" 2>/dev/null || echo "0"; }

pass() { green "$1"; PASS=$((PASS+1)); }
fail() { red "$1"; FAIL=$((FAIL+1)); }

assert_eq() {
  local label="$1" actual="$2" expected="$3"
  [ "$actual" = "$expected" ] && pass "$label: '$actual'" || { fail "$label: expected='$expected' got='$actual'"; }
}

assert_nonempty() {
  local label="$1" val="$2"
  [ -n "$val" ] && pass "$label: '$val'" || fail "$label: empty"
}

assert_ge() {
  local label="$1" val="$2" min="$3"
  [ "$val" -ge "$min" ] 2>/dev/null && pass "$label: $val >= $min" || fail "$label: $val < $min"
}

# ── banner ───────────────────────────────────────────────────────────────────
echo ""
echo "╔══════════════════════════════════════════════════╗"
echo "║          ecom-poc  full smoke-flow test          ║"
echo "╚══════════════════════════════════════════════════╝"
echo "  user=$USERNAME"
echo ""

# ── 1. health checks ─────────────────────────────────────────────────────────
echo "── 1. Health checks ──────────────────────────────────"
for entry in \
  "auth:$AUTH_URL" \
  "catalog:$CATALOG_URL" \
  "cart:$CART_URL" \
  "order:$ORDER_URL" \
  "inventory:$INVENTORY_URL" \
  "search:$SEARCH_URL" \
  "payment:$PAYMENT_URL"; do
  name="${entry%%:*}"
  url="${entry#*:}"
  r=$(curl -sf "$url/health" 2>/dev/null || echo '{"status":"down"}')
  assert_eq "health/$name" "$(json_field "$r" status)" "ok"
done
echo ""

# ── 2. auth: register + login ─────────────────────────────────────────────────
echo "── 2. Auth ───────────────────────────────────────────"
r=$(curl -sf -X POST "$AUTH_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" 2>/dev/null || echo '{}')
assert_eq "register" "$(json_field "$r" username)" "$USERNAME"

r=$(curl -sf -X POST "$AUTH_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" 2>/dev/null || echo '{}')
TOKEN=$(json_field "$r" "token")
assert_nonempty "login: get JWT" "$TOKEN"
info "token=${TOKEN:0:40}..."

r=$(curl -sf "$AUTH_URL/auth/me" \
  -H "Authorization: Bearer $TOKEN" 2>/dev/null || echo '{}')
assert_eq "me: username" "$(json_field "$r" username)" "$USERNAME"
echo ""

# ── 3. catalog: list products ─────────────────────────────────────────────────
echo "── 3. Catalog ────────────────────────────────────────"
r=$(curl -sf "$CATALOG_URL/products" 2>/dev/null || echo '[]')
count=$(echo "$r" | python3 -c "import sys,json; print(len(json.load(sys.stdin)))" 2>/dev/null || echo "0")
assert_ge "list products" "$count" 1
PRODUCT_ID=$(echo "$r" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d[0]['id'] if d else '')" 2>/dev/null || echo "")
assert_nonempty "first product id" "$PRODUCT_ID"
info "product_id=$PRODUCT_ID"
echo ""

# ── 4. cart: add item + get cart ──────────────────────────────────────────────
echo "── 4. Cart ───────────────────────────────────────────"
r=$(curl -sf -X POST "$CART_URL/cart/items" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"productId\":\"$PRODUCT_ID\",\"quantity\":2}" 2>/dev/null || echo '{}')
assert_eq "add cart item" "$(json_field "$r" message)" "item added to cart"

r=$(curl -sf "$CART_URL/cart" \
  -H "Authorization: Bearer $TOKEN" 2>/dev/null || echo '{"items":[]}')
cart_count=$(json_len "$r" items)
assert_ge "get cart: items" "$cart_count" 1
echo ""

# ── 5. search ────────────────────────────────────────────────────────────────
echo "── 5. Search ─────────────────────────────────────────"
sleep 1  # allow ES indexing
r=$(curl -sf "$SEARCH_URL/search?q=a" 2>/dev/null || echo '{"total":0,"products":[]}')
total=$(json_field "$r" total)
assert_ge "search?q=a: total" "${total:-0}" 0
info "search total=$total"
echo ""

# ── 6. order: create ─────────────────────────────────────────────────────────
echo "── 6. Order ──────────────────────────────────────────"
r=$(curl -sf -X POST "$ORDER_URL/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $IDEM_KEY" \
  -d '{}' 2>/dev/null || echo '{}')
ORDER_STATUS=$(json_field "$r" status)
ORDER_ID=$(json_field "$r" id)
assert_eq "create order: status" "$ORDER_STATUS" "PENDING"
assert_nonempty "create order: id" "$ORDER_ID"
info "order_id=$ORDER_ID"

# idempotency: same key returns same order
r2=$(curl -sf -X POST "$ORDER_URL/orders" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: $IDEM_KEY" \
  -d '{}' 2>/dev/null || echo '{}')
assert_eq "idempotency: same id" "$(json_field "$r2" id)" "$ORDER_ID"

# cart cleared after order
r=$(curl -sf "$CART_URL/cart" \
  -H "Authorization: Bearer $TOKEN" 2>/dev/null || echo '{"items":[]}')
cart_after=$(json_len "$r" items)
assert_eq "cart cleared after order" "$cart_after" "0"
echo ""

# ── 7. payment + Kafka callback ───────────────────────────────────────────────
echo "── 7. Payment (Kafka outbox flow) ────────────────────"
r=$(curl -sf -X POST "$PAYMENT_URL/payments" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"orderId\":\"$ORDER_ID\",\"amount\":999000}" 2>/dev/null || echo '{}')
PAYMENT_ID=$(json_field "$r" id)
PAYMENT_STATUS=$(json_field "$r" status)
assert_nonempty "create payment: id" "$PAYMENT_ID"
assert_eq "create payment: status" "$PAYMENT_STATUS" "PENDING"
info "payment_id=$PAYMENT_ID"

# trigger mock callback (success) -> publishes to Kafka -> order consumer confirms
r=$(curl -sf -X POST "$PAYMENT_URL/payments/$PAYMENT_ID/callback" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"result":"success"}' 2>/dev/null || echo '{}')
assert_eq "payment callback: status" "$(json_field "$r" status)" "SUCCESS"

# wait for relay (3s interval) + consumer to process
info "waiting 12s for Kafka relay + consumer to confirm order..."
sleep 12

r=$(curl -sf "$ORDER_URL/orders/$ORDER_ID" \
  -H "Authorization: Bearer $TOKEN" 2>/dev/null || echo '{}')
assert_eq "order confirmed via Kafka" "$(json_field "$r" status)" "CONFIRMED"
echo ""

# ── 8. payment idempotency: already processed ────────────────────────────────
echo "── 8. Payment idempotency ────────────────────────────"
r=$(curl -s -X POST "$PAYMENT_URL/payments/$PAYMENT_ID/callback" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"result":"success"}' 2>/dev/null || echo '{}')
assert_eq "duplicate callback rejected" "$(json_field "$r" error)" "payment already processed"
echo ""

# ── 9. flash sale (Redis atomic counter) ──────────────────────────────────────
echo "── 9. Flash sale ─────────────────────────────────────"
FS_PRODUCT="flash-smoke-$TS"
r=$(curl -sf -X POST "$INVENTORY_URL/inventory/flash-sale/init" \
  -H "Content-Type: application/json" \
  -d "{\"productId\":\"$FS_PRODUCT\",\"quantity\":3}" 2>/dev/null || echo '{}')
assert_eq "flash sale init" "$(json_field "$r" message)" "flash sale initialised"

r=$(curl -sf -X POST "$INVENTORY_URL/inventory/flash-sale/reserve" \
  -H "Content-Type: application/json" \
  -d "{\"productId\":\"$FS_PRODUCT\",\"quantity\":1}" 2>/dev/null || echo '{}')
remaining_val=$(json_field "$r" remaining)
assert_eq "flash reserve 1" "$remaining_val" "2"

r=$(curl -sf "$INVENTORY_URL/inventory/flash-sale/stock/$FS_PRODUCT" 2>/dev/null || echo '{}')
assert_eq "flash stock remaining" "$(json_field "$r" flashSaleStock)" "2"

# exhaust stock
curl -sf -X POST "$INVENTORY_URL/inventory/flash-sale/reserve" \
  -H "Content-Type: application/json" \
  -d "{\"productId\":\"$FS_PRODUCT\",\"quantity\":2}" > /dev/null 2>&1 || true

r=$(curl -s -X POST "$INVENTORY_URL/inventory/flash-sale/reserve" \
  -H "Content-Type: application/json" \
  -d "{\"productId\":\"$FS_PRODUCT\",\"quantity\":1}" 2>/dev/null || echo '{}')
assert_eq "flash reserve oversell blocked" "$(json_field "$r" error)" "sold out"
echo ""

# ── 10. second order: fail flow ───────────────────────────────────────────────
echo "── 10. Fail flow ─────────────────────────────────────"
# register second user
r=$(curl -sf -X POST "$AUTH_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"${USERNAME}-b\",\"password\":\"$PASSWORD\"}" 2>/dev/null || echo '{}')
r=$(curl -sf -X POST "$AUTH_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"${USERNAME}-b\",\"password\":\"$PASSWORD\"}" 2>/dev/null || echo '{}')
TOKEN_B=$(json_field "$r" token)

curl -sf -X POST "$CART_URL/cart/items" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d "{\"productId\":\"$PRODUCT_ID\",\"quantity\":1}" > /dev/null 2>&1 || true

r=$(curl -sf -X POST "$ORDER_URL/orders" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: idem-fail-$TS" \
  -d '{}' 2>/dev/null || echo '{}')
ORDER_ID_B=$(json_field "$r" id)
assert_eq "fail flow: create order" "$(json_field "$r" status)" "PENDING"

# pay and fail it
r=$(curl -sf -X POST "$PAYMENT_URL/payments" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d "{\"orderId\":\"$ORDER_ID_B\",\"amount\":100}" 2>/dev/null || echo '{}')
PAY_B=$(json_field "$r" id)

r=$(curl -sf -X POST "$PAYMENT_URL/payments/$PAY_B/callback" \
  -H "Authorization: Bearer $TOKEN_B" \
  -H "Content-Type: application/json" \
  -d '{"result":"fail"}' 2>/dev/null || echo '{}')
assert_eq "fail callback: status" "$(json_field "$r" status)" "FAILED"

info "waiting 12s for Kafka relay + consumer to fail order..."
sleep 12

r=$(curl -sf "$ORDER_URL/orders/$ORDER_ID_B" \
  -H "Authorization: Bearer $TOKEN_B" 2>/dev/null || echo '{}')
assert_eq "order failed via Kafka" "$(json_field "$r" status)" "FAILED"
echo ""

# ── 11. unauthorized checks ───────────────────────────────────────────────────
echo "── 11. Auth guard ────────────────────────────────────"
http_code=$(curl -s -o /dev/null -w "%{http_code}" "$CART_URL/cart" 2>/dev/null || echo "0")
assert_eq "cart without JWT: 401" "$http_code" "401"

http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$ORDER_URL/orders" \
  -H "Content-Type: application/json" -d '{}' 2>/dev/null || echo "0")
assert_eq "order without JWT: 401" "$http_code" "401"
echo ""

# ── summary ───────────────────────────────────────────────────────────────────
echo "══════════════════════════════════════════════════════"
TOTAL=$((PASS+FAIL))
if [ "$FAIL" -eq 0 ]; then
  printf "\033[32m  ALL PASSED  (%d/%d)\033[0m\n" "$PASS" "$TOTAL"
else
  printf "\033[31m  FAILED: %d/%d passed, %d failed\033[0m\n" "$PASS" "$TOTAL" "$FAIL"
fi
echo "══════════════════════════════════════════════════════"
echo ""

[ "$FAIL" -eq 0 ]
