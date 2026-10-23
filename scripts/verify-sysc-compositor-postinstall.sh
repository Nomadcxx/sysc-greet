#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
test_root=$(mktemp -d)
trap 'rm -rf "$test_root"' EXIT
stub_dir="$test_root/bin"
config_file="$test_root/etc/greetd/config.toml"
mkdir -p "$stub_dir" "$(dirname "$config_file")"

for tool in id usermod mkdir chown chmod systemctl sysc-compositor; do
    printf '#!/bin/sh\nexit 0\n' > "$stub_dir/$tool"
    chmod +x "$stub_dir/$tool"
done
printf '#!/bin/sh\nexec /bin/cat "$@"\n' > "$stub_dir/cat"
chmod +x "$stub_dir/cat"

# Redirect only the generated greetd config into the temporary test directory.
sed "s|/etc/greetd/config.toml|$config_file|g" \
    "$repo_root/scripts/postinstall.sh" > "$test_root/postinstall.sh"
PATH="$stub_dir" /bin/bash "$test_root/postinstall.sh" > "$test_root/output"

expected='command = "sysc-compositor -c /etc/greetd/niri-greeter-config.kdl"'
grep -Fqx "$expected" "$config_file" 2>/dev/null || {
    echo "postinstall did not generate the sysc-compositor greetd command" >&2
    exit 1
}
echo "sysc-compositor postinstall check passed"
