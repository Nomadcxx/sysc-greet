#!/usr/bin/env python3
"""Check effect containment in native fullscreen Kitty; run with a built binary.

Requires Kitty remote control and a graphical session. Always uses --test.
Captures terminal text at 5, 10, and 15 seconds after each effect change.
"""
from pathlib import Path
import argparse
import subprocess
import tempfile
import time

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("binary", type=Path)
parser.add_argument("--theme", choices=("dracula", "blue"), default="dracula")
args = parser.parse_args()
binary = str(args.binary.resolve(strict=True))
with tempfile.TemporaryDirectory(prefix="sysc-effects-") as directory:
    socket = "unix:" + directory + "/kitty.sock"

    def remote(*args):
        return subprocess.check_output(
            ["kitty", "@", "--to", socket, *args], text=True
        )

    def key(value):
        remote("send-key", "--match", "id:1", value)
        time.sleep(0.2)

    def choose(menu, item):
        key("f1")
        for _ in range(menu):
            key("down")
        key("enter")
        for _ in range(item):
            key("down")
        key("enter")

    subprocess.run(
        ["kitty", "-c", "NONE", "--detach", "--start-as=fullscreen",
         "--class=sysc-effects-check", "--listen-on=" + socket,
         "--override", "allow_remote_control=yes", binary, "--test",
         "--theme", args.theme], check=True
    )
    try:
        # Wait for the socket and the fullscreen resize before driving menus.
        deadline = time.monotonic() + 10
        while not Path(directory + "/kitty.sock").exists():
            if time.monotonic() > deadline:
                raise RuntimeError("Kitty remote control did not start")
            time.sleep(0.1)
        time.sleep(3)
        choose(1, {"dracula": 1, "blue": 15}[args.theme])
        choose(2, 1)  # Borders: Classic, with a static enclosing inner border.
        for name, item in [("typewriter", 1), ("print", 2), ("beams", 3), ("pour", 4)]:
            choose(4, item)
            key("escape")
            key("enter")
            for elapsed in (5, 10, 15):
                time.sleep(5)
                lines = remote("get-text", "--match", "id:1").splitlines()
                # The widest rounded border encloses the art and login form.
                spans = [(y, line.index("╭"), line.rindex("╮"))
                         for y, line in enumerate(lines) if "╭" in line and "╮" in line]
                assert spans, "enclosing login border not found"
                top, left, right = max(spans, key=lambda span: span[2] - span[1])
                bottom = next(y for y in range(top + 1, len(lines))
                              if len(lines[y]) > right and lines[y][left] == "╰"
                              and lines[y][right] == "╯")
                for y in range(top + 1, bottom):
                    outside = lines[y][:left] + lines[y][right + 1:]
                    assert all(char in " ║" for char in outside), (
                        f"{name} at {elapsed}s: displaced text on terminal row {y + 1}: {outside!r}"
                    )
            print(name + ": contained at 5, 10 and 15 seconds", flush=True)
            choose(4, item)  # Toggle the effect off before the next case.
            key("escape")
            key("enter")
    finally:
        remote("close-window", "--match", "id:1")
