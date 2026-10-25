#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
printf '%s\n' \
  'Installer unit tests only. Void support also requires a clean VM run:' \
  'boot on tty1, desktop login, then F4 reboot and shutdown.' \
  'See docs-site/content/docs/getting-started/void-linux.md.'
exec go test -count=1 ./cmd/installer/
