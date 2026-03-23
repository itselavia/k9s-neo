import importlib.util
import json
import sys
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
REPLAY_PATH = REPO_ROOT / "hack" / "bench" / "replay.py"
SPEC = importlib.util.spec_from_file_location("bench_replay", REPLAY_PATH)
bench_replay = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
sys.modules[SPEC.name] = bench_replay
SPEC.loader.exec_module(bench_replay)


class BenchReplayTests(unittest.TestCase):
    def test_load_fixture(self) -> None:
        fixture = bench_replay.load_fixture(
            REPO_ROOT / "hack" / "bench" / "fixtures" / "replay" / "pods_startup_ok.json"
        )
        self.assertEqual(1, fixture["fixture_schema_version"])
        self.assertEqual("pods_startup", fixture["scenario"])

    def test_load_fixture_rejects_missing_anchor(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "broken.json"
            path.write_text(
                json.dumps(
                    {
                        "fixture_schema_version": 1,
                        "scenario": "pods_startup",
                        "mode": "cold",
                        "run_index": 1,
                        "status": "ok",
                        "error": None,
                        "events": [],
                        "samples": [],
                        "anchors": {},
                    }
                ),
                encoding="utf-8",
            )
            with self.assertRaisesRegex(ValueError, "anchors missing required fields"):
                bench_replay.load_fixture(path)

    def test_build_replay_result_uses_fixture_anchors(self) -> None:
        fixture = bench_replay.load_fixture(
            REPO_ROOT / "hack" / "bench" / "fixtures" / "replay" / "nodes_first_render_ok.json"
        )
        result = bench_replay.build_replay_result(fixture)
        self.assertEqual("replay", result["source_kind"])
        self.assertEqual(400, result["metrics"]["time_to_first_paint_ms"])
        self.assertEqual(450, result["metrics"]["time_to_first_useful_row_ms"])

    def test_run_replay_writes_artifacts(self) -> None:
        fixtures_dir = REPO_ROOT / "hack" / "bench" / "fixtures" / "replay"
        with tempfile.TemporaryDirectory() as tmp:
            root_dir, results = bench_replay.run_replay(fixtures_dir, "replay-validation", Path(tmp))
            self.assertEqual(4, len(results))

            run_json = root_dir / "runs" / "pods_startup" / "cold-01" / "run.json"
            data = json.loads(run_json.read_text(encoding="utf-8"))
            self.assertEqual("replay", data["source_kind"])

            csv_text = (root_dir / "summary" / "runs.csv").read_text(encoding="utf-8")
            self.assertIn("source_kind", csv_text.splitlines()[0])
            self.assertIn("replay,pods_startup,cold,1,ok", csv_text)

            report_text = (root_dir / "summary" / "report.md").read_text(encoding="utf-8")
            self.assertIn("Artifact source_kind(s): replay", report_text)
            self.assertIn("ok_runs: 0/1", report_text)
            self.assertIn("pod_events (cold)", report_text)


if __name__ == "__main__":
    unittest.main()
