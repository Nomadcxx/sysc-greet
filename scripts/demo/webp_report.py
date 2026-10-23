#!/usr/bin/env python3
"""Report an animated WebP's size, frames, and duration via `webpmux -info`.

libwebp_anim merges identical consecutive frames, so the average fps printed
here can be lower than the encode fps on static stretches.
"""
import argparse
import os
import re
import subprocess
import sys
from typing import NamedTuple


class WebPInfo(NamedTuple):
    width: int
    height: int
    frames: int
    duration_ms: int
    size_bytes: int
    loop_count: int

    @property
    def fps(self) -> float:
        return self.frames / (self.duration_ms / 1000) if self.duration_ms else 0.0


FRAME_ROW = re.compile(r"^\s*\d+:\s+(.*)$")


def _field(text: str, pattern: str) -> re.Match:
    match = re.search(pattern, text, re.MULTILINE)
    if not match:
        raise ValueError(f"webpmux output lacks {pattern!r}")
    return match


def parse_webpmux_info(text: str, size_bytes: int) -> WebPInfo:
    features = _field(text, r"^(?:Features present: (.*)|No features present\.)$")
    if "animation" not in (features.group(1) or ""):
        raise ValueError("not an animated WebP")
    canvas = _field(text, r"^Canvas size: (\d+) x (\d+)$")
    loop = _field(text, r"Loop Count : (\d+)")
    frames = int(_field(text, r"^Number of frames: (\d+)$").group(1))
    header = _field(text, r"^No\.:(.*)$").group(1).split()
    column = header.index("duration")
    durations = [int(m.group(1).split()[column]) for m in map(FRAME_ROW.match, text.splitlines()) if m]
    if len(durations) != frames:
        raise ValueError(f"expected {frames} frame rows, found {len(durations)}")
    return WebPInfo(int(canvas.group(1)), int(canvas.group(2)), frames, sum(durations), size_bytes, int(loop.group(1)))


def read_webp(path: str) -> WebPInfo:
    try:
        result = subprocess.run(["webpmux", "-info", path], capture_output=True, text=True)
    except FileNotFoundError:
        raise ValueError("webpmux not found: install libwebp-utils") from None
    if result.returncode != 0:
        raise ValueError(f"webpmux failed on {path}: {result.stderr.strip()}")
    return parse_webpmux_info(result.stdout, os.path.getsize(path))


def main(argv=None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("path")
    parser.add_argument("--max-bytes", type=int)
    parser.add_argument("--min-seconds", type=float)
    parser.add_argument("--max-seconds", type=float)
    parser.add_argument("--size", help="required WxH")
    options = parser.parse_args(argv)
    try:
        info = read_webp(options.path)
    except ValueError as error:
        print(f"{options.path}: {error}", file=sys.stderr)
        return 1
    seconds = info.duration_ms / 1000
    print(
        f"{os.path.basename(options.path)}: {info.width}x{info.height}, {info.frames} frames, "
        f"{seconds:.1f} s, {info.fps:.1f} avg fps, {info.size_bytes / 1048576:.1f} MB, loop={info.loop_count}"
    )
    problems = []
    if options.max_bytes is not None and info.size_bytes > options.max_bytes:
        problems.append(f"size {info.size_bytes} bytes exceeds {options.max_bytes}")
    if options.min_seconds is not None and seconds < options.min_seconds:
        problems.append(f"duration {seconds:.1f} s is under {options.min_seconds:g} s")
    if options.max_seconds is not None and seconds > options.max_seconds:
        problems.append(f"duration {seconds:.1f} s is over {options.max_seconds:g} s")
    if options.size and f"{info.width}x{info.height}" != options.size:
        problems.append(f"dimensions {info.width}x{info.height} are not {options.size}")
    for problem in problems:
        print(f"FAIL {problem}", file=sys.stderr)
    return 1 if problems else 0


if __name__ == "__main__":
    sys.exit(main())
