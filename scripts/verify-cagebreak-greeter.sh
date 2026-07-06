#!/usr/bin/env bash
# verify-cagebreak-greeter.sh — smoke checks for the cagebreak greeter integration
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG="${ROOT}/config/cagebreak-greeter-config"
FAIL=0

pass() { echo "✓ $*"; }
fail() { echo "✗ $*"; FAIL=1; }

echo "=== Cagebreak greeter verification ==="

if [[ -f "${CONFIG}" ]]; then
  pass "greeter config exists: ${CONFIG}"
else
  fail "greeter config missing: ${CONFIG}"
fi

if command -v cagebreak >/dev/null 2>&1; then
  pass "cagebreak binary found: $(command -v cagebreak) ($(cagebreak -v 2>&1))"
  rc=0
  WLR_BACKENDS=headless WLR_LIBINPUT_NO_DEVICES=1 timeout 5 cagebreak -e -c "${CONFIG}" >/dev/null 2>&1 || rc=$?
  if [[ "${rc}" -eq 0 || "${rc}" -eq 124 ]]; then
    pass "config parses under headless cagebreak"
  else
    fail "cagebreak rejected the greeter config (exit ${rc})"
  fi
else
  echo "⚠ cagebreak not installed — skip runtime test (Arch: paru -S cagebreak)"
fi

if command -v socat >/dev/null 2>&1; then
  pass "socat available (needed for quit-via-socket)"
else
  fail "socat missing — greeter cannot quit cagebreak after login"
fi

if grep -q '"cagebreak"' "${ROOT}/cmd/installer/main.go"; then
  pass "installer references cagebreak"
else
  fail "installer does not reference cagebreak"
fi

if [[ -f "${ROOT}/docs-src/compositors/cagebreak.md" ]]; then
  pass "cagebreak docs present"
else
  fail "docs-src/compositors/cagebreak.md missing"
fi

echo
if [[ "${FAIL}" -eq 0 ]]; then
  echo "All static checks passed."
  echo "Manual: SYSC_COMPOSITOR=cagebreak sudo ./install.sh → systemctl restart greetd → test login"
else
  echo "Some checks failed."
  exit 1
fi
