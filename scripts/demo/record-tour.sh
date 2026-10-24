#!/usr/bin/env bash
# Record the README demo tour on the laptop testbench and encode assets/showcase.webp.
#
# Usage: scripts/demo/record-tour.sh [--check] [--timeline FILE] [--out FILE]
#   --check  play the timeline in the real Kitty window without recording
# Runs inside the user's niri session (locally or over SSH). See
# docs/plans/2026-10-23-demo-tour-design.md.
set -euo pipefail

here=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
repo=$(cd "$here/../.." && pwd)
CHECK=0
TIMELINE=$here/tour.timeline
OUT=$repo/assets/showcase.webp
while [ $# -gt 0 ]; do
    case "$1" in
        --check) CHECK=1 ;;
        --timeline) TIMELINE=$(realpath "$2"); shift ;;
        --out) OUT=$(realpath -m "$2"); shift ;;
        -h | --help) sed -n 2,7p "$0"; exit 0 ;;
        *) echo "unknown argument: $1" >&2; exit 2 ;;
    esac
    shift
done

SYSC_GREET_BIN=${SYSC_GREET_BIN:-/usr/local/bin/sysc-greet}
DEMO_FONT_SIZE=${DEMO_FONT_SIZE:-13}
WANT_VERSION=v1.3.0
OPENING_THEME=tokyo-night
OPENING_SESSION=mango
FORM_MODE=frosted
OUTPUT_NAME=eDP-1
GRID_COLUMNS=(140 160)
GRID_LINES=(34 42)
MAX_BYTES=20971520
MIN_SECONDS=${DEMO_MIN_SECONDS:-40}
MAX_SECONDS=90
ENCODE_FPS=20
ENCODE_QUALITIES=(70 65 60 55 50)

fail() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }
say() { printf '==> %s\n' "$*"; }

# Re-exec once under an idle inhibitor so dimming or locking cannot start mid-run.
if [ -z "${DEMO_INHIBITED:-}" ]; then
    command -v systemd-inhibit >/dev/null || fail "systemd-inhibit not found"
    export DEMO_INHIBITED=1
    exec systemd-inhibit --what=idle --who=sysc-greet-demo --why="recording demo" "$0" \
        $([ "$CHECK" -eq 1 ] && echo --check) --timeline "$TIMELINE" --out "$OUT"
fi

say "Preflight"
# Read all of --version: piping into head can SIGPIPE it under pipefail.
version=$("$SYSC_GREET_BIN" --version 2>/dev/null) || fail "$SYSC_GREET_BIN does not run"
version=${version%%$'\n'*}
[[ $version == *"$WANT_VERSION"* ]] || fail "$SYSC_GREET_BIN reports '${version:-nothing}', need $WANT_VERSION"
tools=(kitty kitten niri ffmpeg python3)
[ "$CHECK" -eq 1 ] || tools+=(gpu-screen-recorder webpmux)
for tool in "${tools[@]}"; do
    if ! command -v "$tool" >/dev/null; then
        [ "$tool" = webpmux ] && fail "webpmux not found: install libwebp-utils"
        fail "$tool not found"
    fi
done
[ -f "$TIMELINE" ] || fail "timeline $TIMELINE not found"
runtime=/run/user/$(id -u)
if [ -z "${WAYLAND_DISPLAY:-}" ]; then
    WAYLAND_DISPLAY=$(basename "$(ls -1 "$runtime"/wayland-[0-9]* 2>/dev/null | grep -v '\.lock$' | head -1)" 2>/dev/null) || true
fi
[ -n "${NIRI_SOCKET:-}" ] || NIRI_SOCKET=$(ls -1 "$runtime"/niri.*.sock 2>/dev/null | head -1) || true
[ -n "${WAYLAND_DISPLAY:-}" ] || fail "no Wayland display found in $runtime"
[ -n "${NIRI_SOCKET:-}" ] || fail "no niri socket found in $runtime"
export WAYLAND_DISPLAY NIRI_SOCKET XDG_RUNTIME_DIR=${XDG_RUNTIME_DIR:-$runtime}
focused=$(niri msg --json focused-output | python3 -c 'import json,sys; print((json.load(sys.stdin) or {}).get("name",""))') ||
    fail "niri msg failed"
[ "$focused" = "$OUTPUT_NAME" ] || fail "focused output is '$focused', need $OUTPUT_NAME"

work=$(mktemp -d)
socket=$work/kitty.sock
recorder_pid=
dnd_restore=

cleanup() {
    local status=$?
    trap - EXIT INT TERM
    if [ -n "$recorder_pid" ] && kill -0 "$recorder_pid" 2>/dev/null; then
        kill -INT "$recorder_pid" 2>/dev/null || true
        wait "$recorder_pid" 2>/dev/null || true
    fi
    if [ -S "$socket" ]; then
        kitten @ --to "unix:$socket" close-window >/dev/null 2>&1 || true
    fi
    # Fall back to niri if the socket never appeared or close-window failed.
    niri msg --json windows 2>/dev/null |
        python3 -c 'import json,sys; [print(w["id"]) for w in json.load(sys.stdin) if w.get("app_id")=="sysc-greet-demo"]' 2>/dev/null |
        while read -r id; do niri msg action close-window --id "$id" >/dev/null 2>&1 || true; done
    [ -z "$dnd_restore" ] || eval "$dnd_restore" >/dev/null 2>&1 || true
    rm -rf "$work"
    exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT TERM

if pgrep -x swaync >/dev/null && command -v swaync-client >/dev/null; then
    if [ "$(swaync-client --get-dnd 2>/dev/null)" != "true" ]; then
        swaync-client --dnd-on >/dev/null && dnd_restore="swaync-client --dnd-off"
    fi
elif pgrep -x mako >/dev/null && command -v makoctl >/dev/null; then
    if ! makoctl mode | grep -qx do-not-disturb; then
        makoctl mode -a do-not-disturb >/dev/null && dnd_restore="makoctl mode -r do-not-disturb"
    fi
else
    say "No swaync or mako running; silence any other notifications before recording"
fi

say "Opening Kitty"
# --test never reads or writes preferences; the throwaway HOME keeps any other
# cache files out of the user's home. The timeline sets the opening session and
# header variant in its lead-in, before "mark start", which is trimmed from the video.
mkdir -p "$work/home"
niri msg action spawn -- kitty --class sysc-greet-demo --config /etc/greetd/kitty.conf \
    -o background_opacity=1.0 -o allow_remote_control=socket-only -o mouse_hide_wait=0 \
    -o "font_size=$DEMO_FONT_SIZE" --listen-on "unix:$socket" --start-as=fullscreen \
    env -u NO_COLOR "HOME=$work/home" "$SYSC_GREET_BIN" --test --test-session "$OPENING_SESSION" --theme "$OPENING_THEME" --form-mode "$FORM_MODE" --follow-shell=false
for _ in $(seq 100); do [ -S "$socket" ] && break; sleep 0.1; done
[ -S "$socket" ] || fail "Kitty did not open its control socket within 10 s"

sleep 2 # let fullscreen settle before measuring the grid
grid=$(kitten @ --to "unix:$socket" ls | python3 -c '
import json, sys
w = json.load(sys.stdin)[0]["tabs"][0]["windows"][0]
print(w["columns"], w["lines"])')
read -r columns lines <<<"$grid"
say "Kitty grid ${columns}x${lines} at font_size $DEMO_FONT_SIZE"
if ((columns < GRID_COLUMNS[0] || columns > GRID_COLUMNS[1] || lines < GRID_LINES[0] || lines > GRID_LINES[1])); then
    fail "grid ${columns}x${lines} is outside ${GRID_COLUMNS[0]}-${GRID_COLUMNS[1]}x${GRID_LINES[0]}-${GRID_LINES[1]}; adjust DEMO_FONT_SIZE"
fi

if [ "$CHECK" -eq 0 ]; then
    say "Recording $OUTPUT_NAME"
    recorder_start=$(date +%s.%N)
    gpu-screen-recorder -w "$OUTPUT_NAME" -f 60 -cursor no -k hevc -q very_high -o "$work/capture.mkv" \
        >"$work/recorder.log" 2>&1 &
    recorder_pid=$!
    sleep 1
    kill -0 "$recorder_pid" 2>/dev/null || fail "gpu-screen-recorder exited: $(tail -3 "$work/recorder.log")"
fi

say "Playing $(basename "$TIMELINE")"
tour_start=$(date +%s.%N)
python3 "$here/tour.py" --socket "unix:$socket" --timeline "$TIMELINE" --marks "$work/marks.json" ||
    fail "timeline failed"

if [ "$CHECK" -eq 1 ]; then
    say "check passed"
    exit 0
fi

kill -0 "$recorder_pid" 2>/dev/null || fail "gpu-screen-recorder stopped during the tour: $(tail -3 "$work/recorder.log")"
kill -INT "$recorder_pid"
wait "$recorder_pid" || true
recorder_pid=
[ -s "$work/capture.mkv" ] || fail "no capture written"

read -r trim_start trim_end < <(python3 -c '
import json, sys
marks = json.load(open(sys.argv[1]))
offset = float(sys.argv[2]) - float(sys.argv[3])
print("%.3f %.3f" % (marks["start"] + offset, marks["end"] + offset))' \
    "$work/marks.json" "$tour_start" "$recorder_start")

# Over budget, step the WebP quality down before giving up; the capture is reused.
for quality in "${ENCODE_QUALITIES[@]}"; do
    say "Encoding ${trim_start}s-${trim_end}s at $ENCODE_FPS fps, quality $quality"
    ffmpeg -v error -y -ss "$trim_start" -to "$trim_end" -i "$work/capture.mkv" -vf "fps=$ENCODE_FPS" \
        -c:v libwebp_anim -lossless 0 -q:v "$quality" -compression_level 4 -loop 0 "$work/showcase.webp"
    if python3 "$here/webp_report.py" "$work/showcase.webp" --max-bytes "$MAX_BYTES" \
        --min-seconds "$MIN_SECONDS" --max-seconds "$MAX_SECONDS" --size 1920x1080; then
        ENCODE_QUALITY=$quality
        break
    fi
    ENCODE_QUALITY=
done
[ -n "$ENCODE_QUALITY" ] || fail "no quality in ${ENCODE_QUALITIES[*]} met the limits; $OUT unchanged"
mkdir -p "$(dirname "$OUT")"
mv "$work/showcase.webp" "$OUT"
say "Wrote $OUT (encode fps $ENCODE_FPS, quality $ENCODE_QUALITY)"
