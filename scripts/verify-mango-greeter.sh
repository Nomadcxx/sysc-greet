#!/usr/bin/env bash
# verify-mango-greeter.sh — smoke checks for the mango greeter integration
#
# Static checks only. Do NOT start mango from a logged-in graphical session:
# on quit mango runs `systemctl --user stop graphical-session.target`, which
# ends your real session. The runtime test is a greetd login on a spare machine.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CONFIG="${ROOT}/config/mango-greeter-config.conf"
SESSION="${ROOT}/config/mango-greeter-session.sh"
FAIL=0

pass() { echo "✓ $*"; }
fail() { echo "✗ $*"; FAIL=1; }

echo "=== Mango greeter verification ==="

if [[ -f "${CONFIG}" ]]; then
  pass "greeter config exists: ${CONFIG}"
else
  fail "greeter config missing: ${CONFIG}"
fi

if grep -nE '^\s*(bind|mousebind|axisbind|source)\s*=' "${CONFIG}"; then
  fail "greeter config has bind/source lines — compositor keybindings would be reachable"
else
  pass "no bind/mousebind/axisbind/source lines"
fi

if [[ -x "${SESSION}" ]]; then
  pass "session script is executable: ${SESSION}"
else
  fail "session script missing or not executable: ${SESSION}"
fi

if grep -q '^mmsg dispatch quit' "${SESSION}"; then
  pass "session script quits mango with mmsg"
else
  fail "session script does not run 'mmsg dispatch quit' — login would hang"
fi

if command -v mango >/dev/null 2>&1; then
  pass "mango binary found: $(command -v mango) ($(mango -v 2>&1))"
  if mango -c "${CONFIG}" -p >/dev/null 2>&1; then
    pass "config parses (mango -p)"
  else
    fail "mango rejected the greeter config: run mango -c ${CONFIG} -p"
  fi
  if command -v mmsg >/dev/null 2>&1; then
    pass "mmsg available (needed to quit mango after login)"
  else
    fail "mmsg missing — greeter cannot quit mango after login"
  fi
else
  echo "⚠ mango not installed — skip parse check (Arch: pacman -S mangowm)"
fi

if grep -q '"mango"' "${ROOT}/cmd/installer/main.go"; then
  pass "installer references mango"
else
  fail "installer does not reference mango"
fi

script_version=$(sed -n 's/^MANGO_VERSION=//p' "${ROOT}/scripts/build-mango.sh")
installer_version=$(sed -n 's/^const mangoPackageVersion = "\(.*\)"/\1/p' "${ROOT}/cmd/installer/main.go")
if [[ -n "${script_version}" && "${script_version}" == "${installer_version}" ]]; then
  pass "installer expects the mango packages build-mango.sh builds (${script_version})"
else
  fail "mango version mismatch: build-mango.sh=${script_version} installer=${installer_version}"
fi

if [[ -f "${ROOT}/docs-site/content/docs/compositors/mango.md" ]]; then
  pass "mango docs present"
else
  fail "docs-site/content/docs/compositors/mango.md missing"
fi

echo
if [[ "${FAIL}" -eq 0 ]]; then
  echo "All static checks passed."
  echo "Manual: sudo SYSC_COMPOSITOR=mango ./install.sh → systemctl restart greetd → test login"
else
  echo "Some checks failed."
  exit 1
fi
