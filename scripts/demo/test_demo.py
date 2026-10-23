import os
import shutil
import subprocess
import sys
import tempfile
import unittest

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import tour  # noqa: E402
import webp_report  # noqa: E402


class FakeTime:
    def __init__(self):
        self.now = 0.0

    def clock(self):
        return self.now

    def sleep(self, seconds):
        self.now += max(0.0, seconds)


class TimelineParseTest(unittest.TestCase):
    def test_parse_valid(self):
        steps = tour.parse_timeline("0 mark start\n1.5 keys f1 down  # open\n2.4 expect Themes\n")
        self.assertEqual(
            steps,
            [
                tour.Step(0.0, "mark", ("start",), 1),
                tour.Step(1.5, "keys", ("f1", "down"), 2),
                tour.Step(2.4, "expect", ("Themes",), 3),
            ],
        )
        self.assertEqual(tour.parse_timeline("3 expect Theme: Tokyo Night")[0].args, ("Theme: Tokyo Night",))

    def test_parse_rejects(self):
        cases = {
            "1 jump": "line 1:",
            "1 keys": "line 1:",
            "x keys f1": "line 1:",
            "2 mark a\n1 mark b": "line 2:",
            "1 keys f1 down down\n1.3 keys enter": "line 2:",
            "1 mark a\n2 mark a": "line 2:",
        }
        for text, prefix in cases.items():
            with self.subTest(text=text):
                with self.assertRaises(tour.TimelineError) as caught:
                    tour.parse_timeline(text)
                self.assertTrue(str(caught.exception).startswith(prefix), str(caught.exception))


class HighlightTest(unittest.TestCase):
    screen = "  ← Back\n\x1b[1;38;2;189;147;249m  Theme: RAMA  \x1b[0m\n  Theme: Dark"

    def test_highlighted(self):
        self.assertTrue(tour.highlighted(self.screen, "Theme: RAMA"))
        self.assertFalse(tour.highlighted(self.screen, "Theme: Dark"))
        self.assertFalse(tour.highlighted(self.screen, "Theme: Nord"))
        self.assertFalse(tour.highlighted("\x1b[1m\x1b[22m Theme: Dark", "Theme: Dark"))


class RunTest(unittest.TestCase):
    def test_run_schedule(self):
        fake = FakeTime()
        sent = []
        steps = tour.parse_timeline("0 mark start\n1 keys f1 down\n2 mark end")
        marks = tour.run(
            steps,
            send_key=lambda key: sent.append((key, round(fake.now, 3))),
            screen_ansi=lambda: "",
            clock=fake.clock,
            sleep=fake.sleep,
        )
        self.assertEqual(sent, [("f1", 1.0), ("down", 1.2)])
        self.assertEqual(marks, {"start": 0.0, "end": 2.0})

    def test_run_expect_fails(self):
        fake = FakeTime()
        steps = tour.parse_timeline("0 expect Theme: RAMA")
        with self.assertRaises(tour.ExpectError) as caught:
            tour.run(
                steps,
                send_key=lambda key: None,
                screen_ansi=lambda: "\x1b[1m  Theme: Dark\x1b[0m\n  Theme: RAMA",
                clock=fake.clock,
                sleep=fake.sleep,
            )
        message = str(caught.exception)
        self.assertIn("Theme: RAMA", message)
        self.assertIn("  Theme: Dark\n  Theme: RAMA", message)
        self.assertNotIn("\x1b", message)

    def test_main_interrupt_exits_130(self):
        def interrupted(*args, **kwargs):
            raise KeyboardInterrupt

        with tempfile.TemporaryDirectory() as directory:
            timeline = os.path.join(directory, "t.timeline")
            with open(timeline, "w", encoding="utf-8") as handle:
                handle.write("0 mark start\n")
            original = tour.run
            tour.run = interrupted
            try:
                code = tour.main(["--socket", "unix:/none", "--timeline", timeline, "--marks", os.path.join(directory, "m")])
            finally:
                tour.run = original
        self.assertEqual(code, 130)


# Real `webpmux -info` output (libwebp-utils 1.6.0) for ffmpeg testsrc 64x48, 3 frames at 10 fps;
# libwebp_anim merged the first two identical frames into one 200 ms frame.
WEBPMUX_ANIM = """Canvas size: 64 x 48
Features present: animation
Background color : 0xFFFFFFFF  Loop Count : 0
Number of frames: 2
No.: width height alpha x_offset y_offset duration   dispose blend image_size  compression
  1:    64    48    no        0        0      200       none    no        250       lossy
  2:    63     7    no        0       36      100       none   yes         88       lossy
"""
WEBPMUX_STILL = """Canvas size: 64 x 48
No features present.
"""


class WebPReportTest(unittest.TestCase):
    def test_parse_webpmux_info(self):
        info = webp_report.parse_webpmux_info(WEBPMUX_ANIM, 500)
        self.assertEqual(info, webp_report.WebPInfo(64, 48, 2, 300, 500, 0))
        self.assertAlmostEqual(info.fps, 2 / 0.3)
        with self.assertRaises(ValueError):
            webp_report.parse_webpmux_info(WEBPMUX_STILL, 100)
        with self.assertRaises(ValueError):
            webp_report.parse_webpmux_info(WEBPMUX_ANIM.replace("Number of frames: 2\n", ""), 500)

    @unittest.skipUnless(shutil.which("ffmpeg") and shutil.which("webpmux"), "needs ffmpeg and webpmux")
    def test_read_webp(self):
        with tempfile.TemporaryDirectory() as directory:
            path = os.path.join(directory, "out.webp")
            subprocess.run(
                ["ffmpeg", "-v", "error", "-f", "lavfi", "-i", "testsrc=size=64x48:rate=10", "-t", "1",
                 "-c:v", "libwebp_anim", "-loop", "0", path],
                check=True,
            )
            info = webp_report.read_webp(path)
            self.assertEqual((info.width, info.height, info.loop_count), (64, 48, 0))
            self.assertTrue(1 <= info.frames <= 10, info.frames)
            self.assertLessEqual(abs(info.duration_ms - 1000), 100)
            self.assertEqual(info.size_bytes, os.path.getsize(path))

            result = subprocess.run(
                [sys.executable, webp_report.__file__, path, "--max-bytes", "10"], capture_output=True, text=True
            )
            self.assertEqual(result.returncode, 1)
            self.assertIn("size", result.stdout + result.stderr)


if __name__ == "__main__":
    unittest.main()
