#!/usr/bin/env bash
# verify-void-installer.sh — static checks for Void Linux installer support
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
FAIL=0

pass() { echo "✓ $*"; }
fail() { echo "✗ $*"; FAIL=1; }

echo "=== Void installer verification ==="

for f in cmd/installer/platform.go cmd/installer/main.go; do
  if [[ -f "${ROOT}/${f}" ]]; then
    pass "found ${f}"
  else
    fail "missing ${f}"
  fi
done

if rg -q 'xbps' "${ROOT}/cmd/installer/main.go" && rg -q 'initRunit' "${ROOT}/cmd/installer/platform.go"; then
  pass "xbps + runit references present"
else
  fail "xbps or runit support missing in installer"
fi

if rg -q '_greeter' "${ROOT}/cmd/installer/platform.go"; then
  pass "Void _greeter account handling"
else
  fail "_greeter handling missing"
fi

if rg -q 'selectedCompositor == "cage"' "${ROOT}/cmd/installer/main.go"; then
  pass "cage gslapper skip logic"
else
  fail "cage gslapper skip missing"
fi

if go build -o /dev/null "${ROOT}/cmd/installer/" 2>/dev/null; then
  pass "installer compiles"
else
  fail "installer compile failed"
  go build -o /dev/null "${ROOT}/cmd/installer/" || true
fi

if [[ -f "${ROOT}/docs-src/getting-started/void-linux.md" ]]; then
  pass "Void docs present"
else
  fail "Void docs missing"
fi

echo
if [[ "${FAIL}" -eq 0 ]]; then
  echo "All static checks passed."
  echo "Runtime: run on Void with SYSC_COMPOSITOR=cage sudo ./install.sh"
  exit 0
fi
echo "Some checks failed."
exit 1
