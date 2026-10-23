import os
import sys
import unittest

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

import tour  # noqa: E402


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


if __name__ == "__main__":
    unittest.main()
