#!/usr/bin/env bash
# verify-cage-greeter.sh — smoke checks for Cage lite greeter integration
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LAUNCHER="${ROOT}/scripts/cage-greeter-session.sh"
FAIL=0

pass() { echo "✓ $*"; }
fail() { echo "✗ $*"; FAIL=1; }

echo "=== Cage greeter verification ==="

if [[ -x "${LAUNCHER}" ]] || [[ -f "${LAUNCHER}" ]]; then
  pass "launcher script exists: ${LAUNCHER}"
else
  fail "launcher script missing: ${LAUNCHER}"
fi

if bash -n "${LAUNCHER}" 2>/dev/null; then
  pass "launcher script syntax (bash -n)"
else
  fail "launcher script syntax check failed"
fi

if command -v cage >/dev/null 2>&1; then
  pass "cage binary found: $(command -v cage)"
  cage -v 2>/dev/null | head -1 || true
else
  echo "⚠ cage not installed — skip runtime test (install: pacman -S cage / nix profile install nixpkgs#cage)"
fi

if [[ -f "${ROOT}/docs-src/compositors/cage.md" ]]; then
  pass "cage compositor docs present"
else
  fail "docs-src/compositors/cage.md missing"
fi

if [[ -f "${ROOT}/docs/superpowers/specs/2026-06-18-cage-compositor-design.md" ]]; then
  pass "design spec present"
else
  fail "design spec missing"
fi

# Grep installer for cage integration (draft PR scaffold)
if rg -q '"cage"' "${ROOT}/cmd/installer/main.go" 2>/dev/null; then
  pass "installer references cage compositor"
else
  fail "installer does not yet reference cage (expected after integration)"
fi

echo
if [[ "${FAIL}" -eq 0 ]]; then
  echo "All static checks passed."
  echo "Manual: SYSC_COMPOSITOR=cage ./install.sh → systemctl restart greetd → test login"
  exit 0
else
  echo "Some checks failed."
  exit 1
fi
