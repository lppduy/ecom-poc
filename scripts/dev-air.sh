#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
AIR_BIN="${AIR_BIN:-$(go env GOPATH)/bin/air}"

if [[ ! -x "$AIR_BIN" ]]; then
  echo "air binary not found at: $AIR_BIN"
  echo "Install with: go install github.com/air-verse/air@latest"
  exit 1
fi

cleanup() {
  trap - INT TERM EXIT
  if [[ -n "${AUTH_PID:-}" ]]; then kill "$AUTH_PID" 2>/dev/null || true; fi
  if [[ -n "${CATALOG_PID:-}" ]]; then kill "$CATALOG_PID" 2>/dev/null || true; fi
  if [[ -n "${CART_PID:-}" ]]; then kill "$CART_PID" 2>/dev/null || true; fi
  if [[ -n "${ORDER_PID:-}" ]]; then kill "$ORDER_PID" 2>/dev/null || true; fi
  if [[ -n "${INVENTORY_PID:-}" ]]; then kill "$INVENTORY_PID" 2>/dev/null || true; fi
  if [[ -n "${SEARCH_PID:-}" ]]; then kill "$SEARCH_PID" 2>/dev/null || true; fi
  if [[ -n "${PAYMENT_PID:-}" ]]; then kill "$PAYMENT_PID" 2>/dev/null || true; fi
}
trap cleanup INT TERM EXIT

(
  cd "$ROOT_DIR/services/auth"
  "$AIR_BIN"
) &
AUTH_PID=$!

(
  cd "$ROOT_DIR/services/catalog"
  "$AIR_BIN"
) &
CATALOG_PID=$!

(
  cd "$ROOT_DIR/services/cart"
  "$AIR_BIN"
) &
CART_PID=$!

(
  cd "$ROOT_DIR/services/order"
  "$AIR_BIN"
) &
ORDER_PID=$!

(
  cd "$ROOT_DIR/services/inventory"
  "$AIR_BIN"
) &
INVENTORY_PID=$!

(
  cd "$ROOT_DIR/services/search"
  "$AIR_BIN"
) &
SEARCH_PID=$!

(
  cd "$ROOT_DIR/services/payment"
  "$AIR_BIN"
) &
PAYMENT_PID=$!

echo "Air dev servers running:"
echo "  auth:      http://localhost:8087"
echo "  catalog:   http://localhost:8081"
echo "  cart:      http://localhost:8082"
echo "  order:     http://localhost:8083"
echo "  inventory: http://localhost:8084"
echo "  search:    http://localhost:8085"
echo "  payment:   http://localhost:8086"
echo ""
echo "Press Ctrl+C to stop all."

wait
