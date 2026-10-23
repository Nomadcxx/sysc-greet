#!/usr/bin/env python3
"""Play a demo timeline against a sysc-greet Kitty window.

Timeline lines: "<seconds> keys <key>...", "<seconds> expect <label>",
"<seconds> mark <name>". "#" starts a comment.
"""
import argparse
import json
import re
import subprocess
import sys
import time
from typing import Callable, NamedTuple

KEY_GAP = 0.2
ACTIONS = ("keys", "expect", "mark")
SGR = re.compile(r"\x1b\[([0-9;:]*)m")
OTHER_ESCAPE = re.compile(r"\x1b(\[[0-9;:?]*[A-Za-z]|\][^\x07\x1b]*(\x07|\x1b\\))")


class TimelineError(Exception):
    pass


class ExpectError(Exception):
    pass


class Step(NamedTuple):
    at: float
    action: str
    args: tuple
    line: int


def parse_timeline(text: str) -> list:
    steps = []
    marks = set()
    busy_until = 0.0
    for number, raw in enumerate(text.splitlines(), start=1):
        line = raw.split("#", 1)[0].strip()
        if not line:
            continue
        fields = line.split(None, 2)
        try:
            at = float(fields[0])
        except ValueError:
            raise TimelineError(f"line {number}: time {fields[0]!r} is not a number") from None
        action = fields[1] if len(fields) > 1 else ""
        if action not in ACTIONS:
            raise TimelineError(f"line {number}: unknown action {action!r}")
        rest = fields[2] if len(fields) > 2 else ""
        args = (rest,) if action == "expect" else tuple(rest.split())
        if not args or not args[0]:
            raise TimelineError(f"line {number}: {action} needs an argument")
        if action == "mark" and len(args) != 1:
            raise TimelineError(f"line {number}: mark takes one name")
        if steps and at < steps[-1].at:
            raise TimelineError(f"line {number}: time {at} goes backwards")
        if at < busy_until:
            raise TimelineError(f"line {number}: starts at {at} before the previous keys finish at {busy_until:.2f}")
        if action == "mark":
            if args[0] in marks:
                raise TimelineError(f"line {number}: mark {args[0]!r} is already defined")
            marks.add(args[0])
        if action == "keys":
            busy_until = at + KEY_GAP * (len(args) - 1)
        steps.append(Step(at, action, args, number))
    return steps


def plain_text(screen_ansi: str) -> str:
    return OTHER_ESCAPE.sub("", SGR.sub("", screen_ansi))


def highlighted(screen_ansi: str, label: str) -> bool:
    for line in screen_ansi.split("\n"):
        bold = False
        visible = []
        position = 0
        for match in SGR.finditer(line):
            visible.append((plain_text(line[position:match.start()]), bold))
            for code in (match.group(1) or "0").replace(":", ";").split(";"):
                if code == "1":
                    bold = True
                elif code in ("0", "", "22"):
                    bold = False
            position = match.end()
        visible.append((plain_text(line[position:]), bold))
        text = "".join(part for part, _ in visible)
        start = text.find(label)
        if start < 0:
            continue
        offset = 0
        for part, part_bold in visible:
            if offset <= start < offset + len(part):
                return part_bold
            offset += len(part)
    return False


def run(
    steps: list,
    send_key: Callable[[str], None],
    screen_ansi: Callable[[], str],
    clock: Callable[[], float] = time.monotonic,
    sleep: Callable[[float], None] = time.sleep,
) -> dict:
    start = clock()
    marks = {}

    def wait_until(offset):
        remaining = start + offset - clock()
        if remaining > 0:
            sleep(remaining)

    for step in steps:
        wait_until(step.at)
        if step.action == "mark":
            marks[step.args[0]] = round(clock() - start, 3)
        elif step.action == "keys":
            for index, key in enumerate(step.args):
                wait_until(step.at + index * KEY_GAP)
                send_key(key)
        else:
            screen = screen_ansi()
            if not highlighted(screen, step.args[0]):
                raise ExpectError(
                    f"line {step.line}: expected highlighted row {step.args[0]!r}; screen shows:\n{plain_text(screen)}"
                )
    return marks


class Kitty:
    def __init__(self, socket: str):
        self.socket = socket

    def _kitten(self, *args):
        return subprocess.run(
            ["kitten", "@", "--to", self.socket, *args], check=True, capture_output=True, text=True
        ).stdout

    def send_key(self, key: str) -> None:
        self._kitten("send-key", key)

    def screen_ansi(self) -> str:
        return self._kitten("get-text", "--ansi", "--extent", "screen")


def main(argv=None) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--socket", required=True)
    parser.add_argument("--timeline", required=True)
    parser.add_argument("--marks", required=True)
    options = parser.parse_args(argv)
    try:
        with open(options.timeline, encoding="utf-8") as handle:
            steps = parse_timeline(handle.read())
        kitty = Kitty(options.socket)
        marks = run(steps, kitty.send_key, kitty.screen_ansi)
    except (TimelineError, ExpectError) as error:
        print(f"{options.timeline}: {error}", file=sys.stderr)
        return 1
    with open(options.marks, "w", encoding="utf-8") as handle:
        json.dump(marks, handle)
    return 0


if __name__ == "__main__":
    sys.exit(main())
