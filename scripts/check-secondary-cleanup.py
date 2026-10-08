#!/usr/bin/env python3
"""Native dual-output Niri check: python3 scripts/check-secondary-cleanup.py BINARY.

Requires the desktop's WAYLAND_DISPLAY, XDG_RUNTIME_DIR and NIRI_SOCKET.
Uses only --test; closes only the Kitty and helpers created by this check.
"""
import json
import os
from pathlib import Path
import signal
import socket
import subprocess
import sys
import tempfile
import time


def wait_for(check, seconds=8):
    deadline = time.monotonic() + seconds
    while time.monotonic() < deadline:
        result = check()
        if result:
            return result
        time.sleep(0.05)
    raise AssertionError("condition timed out")


def processes():
    for entry in Path("/proc").iterdir():
        if entry.name.isdigit():
            try:
                args = (entry / "cmdline").read_bytes().split(b"\0")
                yield int(entry.name), [os.fsdecode(arg) for arg in args if arg]
            except (FileNotFoundError, PermissionError, ProcessLookupError):
                pass


def running(pid):
    try:
        return Path(f"/proc/{pid}/stat").read_text().rsplit(")", 1)[1].split()[0] != "Z"
    except FileNotFoundError:
        return False


def main():
    binary = str(Path(sys.argv[1]).resolve(strict=True))
    env = os.environ.copy()
    env.pop("GREETD_SOCK", None)
    outputs = json.loads(subprocess.check_output(["niri", "msg", "--json", "outputs"], env=env))
    assert sum(output.get("logical") is not None for output in outputs.values()) >= 2, "needs two active outputs"
    assert not any(args and args[0] == binary for _, args in processes()), "candidate already running"
    for mode in ("terminal-close", "SIGHUP", "SIGTERM", "SIGINT", "ctrl-c"):
        with tempfile.TemporaryDirectory(prefix="sysc-cleanup-check-") as directory:
            address = "unix:" + directory + "/kitty.sock"
            kitty = subprocess.Popen([
                "kitty", "-c", "NONE", "--start-as=fullscreen", "--class=sysc-greet",
                "--listen-on=" + address, "--override", "allow_remote_control=yes",
                binary, "--test", "--secondary-backgrounds=true",
            ], env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            owner = helper = None
            helper_socket = None

            def remote(*args):
                return subprocess.check_output(["kitty", "@", "--to", address, *args], env=env)

            try:
                wait_for(lambda: Path(directory + "/kitty.sock").exists())
                owner = wait_for(lambda: next((pid for pid, args in processes() if args and args[0] == binary), None))
                helper, helper_socket = wait_for(lambda: next((
                    (pid, args[args.index("--ipc-socket") + 1]) for pid, args in processes()
                    if "--ipc-socket" in args and f"PPid:\t{owner}\n" in Path(f"/proc/{pid}/status").read_text()
                ), None))
                wait_for(lambda: Path(helper_socket).exists())
                if mode == "terminal-close":
                    remote("close-window")
                elif mode == "ctrl-c":
                    remote("send-key", "ctrl+c")
                else:
                    os.kill(owner, getattr(signal, mode))
                wait_for(lambda: not running(owner))
                assert not running(helper), f"{mode}: helper {helper} outlived greeter {owner}"
                assert not Path(helper_socket).exists(), f"{mode}: helper socket leaked"
                assert not Path(helper_socket).parent.exists(), f"{mode}: runtime directory leaked"
                print(f"PASS {mode}: greeter and helper exited; runtime directory removed", flush=True)
            finally:
                if owner and running(owner):
                    os.kill(owner, signal.SIGTERM)
                    wait_for(lambda: not running(owner))
                if helper and running(helper):
                    with socket.socket(socket.AF_UNIX) as connection:
                        connection.settimeout(2)
                        connection.connect(helper_socket)
                        connection.sendall(b"quit\n")
                    wait_for(lambda: not running(helper))
                if helper_socket:
                    try:
                        Path(helper_socket).parent.rmdir()
                    except FileNotFoundError:
                        pass
                if kitty.poll() is None:
                    kitty.terminate()
                kitty.wait(timeout=8)


if __name__ == "__main__":
    main()
