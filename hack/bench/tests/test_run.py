import importlib.util
import json
import sys
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
RUN_PATH = REPO_ROOT / "hack" / "bench" / "run.py"
SPEC = importlib.util.spec_from_file_location("bench_run", RUN_PATH)
bench_run = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
sys.modules[SPEC.name] = bench_run
SPEC.loader.exec_module(bench_run)


class BenchRunTests(unittest.TestCase):
    def test_encode_key(self) -> None:
        self.assertEqual("\r", bench_run.encode_key("ENTER"))
        self.assertEqual("\x12", bench_run.encode_key("CTRL_R"))

    def test_load_manifest(self) -> None:
        manifest = bench_run.load_manifest(REPO_ROOT / "hack" / "bench" / "scenarios.json")
        self.assertIn("defaults", manifest)
        self.assertGreater(len(manifest["scenarios"]), 0)

    def test_derive_metrics_scopes_to_view_anchor(self) -> None:
        events = [
            {"type": "lifecycle_mark", "marker": "first_render_committed", "since_process_start_ms": 100, "view_seq": 1, "view_name": "Pod"},
            {"type": "kube_request_complete", "resource": "pods", "kube_verb": "list", "response_bytes": 200, "object_count": 2, "since_process_start_ms": 130},
            {"type": "lifecycle_mark", "marker": "first_useful_row", "since_process_start_ms": 150, "view_seq": 1, "view_name": "Pod"},
            {"type": "lifecycle_mark", "marker": "first_render_committed", "since_process_start_ms": 400, "view_seq": 2, "view_name": "Node"},
            {"type": "kube_request_complete", "resource": "nodes", "kube_verb": "list", "response_bytes": 300, "object_count": 3, "since_process_start_ms": 430},
            {"type": "lifecycle_mark", "marker": "first_useful_row", "since_process_start_ms": 450, "view_seq": 2, "view_name": "Node"},
            {"type": "kube_request_complete", "resource": "nodes", "kube_verb": "list", "response_bytes": 400, "object_count": 4, "since_process_start_ms": 500},
        ]
        samples = [{"rss_bytes": 1024, "cpu_percent": 10.5}, {"rss_bytes": 2048, "cpu_percent": 12.0}]
        anchor_event = {"marker": "first_useful_row", "view_seq": 2, "view_name": "Node"}
        metrics = bench_run.derive_metrics(
            events,
            samples,
            terminal_wait_event=anchor_event,
            view_anchor_event=anchor_event,
        )

        self.assertEqual(400, metrics["time_to_first_paint_ms"])
        self.assertEqual(450, metrics["time_to_first_useful_row_ms"])
        self.assertIsNone(metrics["time_to_stable_interactive_ms"])
        self.assertIsNone(metrics["filter_settle_ms"])
        self.assertIsNone(metrics["detail_ready_ms"])
        self.assertEqual(2048, metrics["peak_rss_bytes_30s"])
        self.assertEqual(12.0, metrics["peak_cpu_percent_30s"])
        self.assertEqual(3, metrics["api_request_count"])
        self.assertEqual(500, metrics["total_response_bytes_before_first_useful_row"])
        self.assertEqual(5, metrics["total_object_count_before_first_useful_row"])
        self.assertEqual(0, metrics["watch_count"])
        self.assertEqual(0, metrics["watches_before_first_useful_row"])

    def test_derive_metrics_scopes_to_latest_matching_view_seq(self) -> None:
        events = [
            {"type": "lifecycle_mark", "marker": "first_render_committed", "since_process_start_ms": 100, "view_seq": 1, "view_name": "Pod"},
            {"type": "lifecycle_mark", "marker": "first_useful_row", "since_process_start_ms": 150, "view_seq": 1, "view_name": "Pod"},
            {"type": "lifecycle_mark", "marker": "first_render_committed", "since_process_start_ms": 320, "view_seq": 2, "view_name": "Node"},
            {"type": "lifecycle_mark", "marker": "first_useful_row", "since_process_start_ms": 360, "view_seq": 2, "view_name": "Node"},
            {"type": "lifecycle_mark", "marker": "first_render_committed", "since_process_start_ms": 700, "view_seq": 3, "view_name": "Pod"},
            {"type": "lifecycle_mark", "marker": "first_useful_row", "since_process_start_ms": 750, "view_seq": 3, "view_name": "Pod"},
        ]

        terminal_event = {"marker": "first_useful_row", "view_seq": 3, "view_name": "Pod"}
        metrics = bench_run.derive_metrics(
            events,
            [],
            terminal_wait_event=terminal_event,
            view_anchor_event=terminal_event,
        )

        self.assertEqual(700, metrics["time_to_first_paint_ms"])
        self.assertEqual(750, metrics["time_to_first_useful_row_ms"])

    def test_derive_metrics_pairs_detail_with_terminal_kind(self) -> None:
        events = [
            {"type": "lifecycle_mark", "marker": "first_render_committed", "since_process_start_ms": 100, "view_seq": 1, "view_name": "Pod"},
            {"type": "lifecycle_mark", "marker": "first_useful_row", "since_process_start_ms": 150, "view_seq": 1, "view_name": "Pod"},
            {"type": "lifecycle_mark", "marker": "detail_open_start", "since_process_start_ms": 250, "detail_kind": "yaml"},
            {"type": "lifecycle_mark", "marker": "detail_content_ready", "since_process_start_ms": 290, "detail_kind": "yaml"},
            {"type": "lifecycle_mark", "marker": "detail_open_start", "since_process_start_ms": 300, "detail_kind": "logs"},
            {"type": "lifecycle_mark", "marker": "detail_content_ready", "since_process_start_ms": 360, "detail_kind": "logs"},
        ]

        metrics = bench_run.derive_metrics(
            events,
            [],
            terminal_wait_event={"marker": "detail_content_ready", "detail_kind": "logs"},
            view_anchor_event={"marker": "first_useful_row", "view_seq": 1, "view_name": "Pod"},
        )

        self.assertEqual(100, metrics["time_to_first_paint_ms"])
        self.assertEqual(150, metrics["time_to_first_useful_row_ms"])
        self.assertEqual(60, metrics["detail_ready_ms"])

    def test_derive_metrics_returns_none_without_view_anchor(self) -> None:
        events = [
            {"type": "lifecycle_mark", "marker": "first_render_committed", "since_process_start_ms": 100, "view_seq": 1, "view_name": "Pod"},
            {"type": "lifecycle_mark", "marker": "first_useful_row", "since_process_start_ms": 150, "view_seq": 1, "view_name": "Pod"},
        ]
        metrics = bench_run.derive_metrics(events, [], terminal_wait_event=None, view_anchor_event=None)

        self.assertIsNone(metrics["time_to_first_paint_ms"])
        self.assertIsNone(metrics["time_to_first_useful_row_ms"])

    def test_update_derived_vars(self) -> None:
        state = {}
        bench_run.update_derived_vars(
            state,
            {
                "view_name": "Pod",
                "selected_path": "big/pod-a",
            },
        )
        self.assertEqual("Pod", state["active_view"])
        self.assertEqual("big/pod-a", state["selected_path"])
        self.assertEqual("big", state["selected_namespace"])
        self.assertEqual("pod-a", state["selected_name"])
        self.assertEqual("big/pod-a", state["origin_selected_path"])
        self.assertEqual("big", state["origin_selected_namespace"])
        self.assertEqual("pod-a", state["origin_selected_name"])

        bench_run.update_derived_vars(
            state,
            {
                "view_name": "Event",
                "selected_path": "big/event-b",
            },
        )
        self.assertEqual("Event", state["active_view"])
        self.assertEqual("big/event-b", state["selected_path"])
        self.assertEqual("event-b", state["selected_name"])
        self.assertEqual("pod-a", state["origin_selected_name"])

    def test_trace_reader_polls_appended_events(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "trace.jsonl"
            reader = bench_run.TraceReader(path)

            path.write_text(json.dumps({"type": "lifecycle_mark", "marker": "a"}) + "\n", encoding="utf-8")
            first = reader.poll()
            self.assertEqual(1, len(first))

            with path.open("a", encoding="utf-8") as handle:
                handle.write(json.dumps({"type": "lifecycle_mark", "marker": "b"}) + "\n")
            second = reader.poll()
            self.assertEqual(1, len(second))
            self.assertEqual("b", second[0]["marker"])

    def test_write_report(self) -> None:
        aggregates = {
            "pods_startup:cold": {
                "scenario": "pods_startup",
                "mode": "cold",
                "statuses": ["ok"],
                "status_counts": {"ok": 1},
                "ok_run_count": 1,
                "total_run_count": 1,
                "time_to_first_useful_row_ms": {"min": 1.0, "median": 1.0, "mean": 1.0, "max": 1.0},
            }
        }
        with tempfile.TemporaryDirectory() as tmp:
            path = Path(tmp) / "report.md"
            bench_run.write_report(path, aggregates)
            text = path.read_text(encoding="utf-8")
            self.assertIn("pods_startup", text)
            self.assertIn("Aggregates include only runs with status `ok`.", text)
            self.assertIn("status_counts: ok=1", text)

    def test_aggregate_runs_uses_only_ok_results(self) -> None:
        results = [
            {
                "scenario": "pod_events",
                "mode": "cold",
                "status": "ok",
                "metrics": {
                    "time_to_first_useful_row_ms": 100.0,
                    "api_request_count": 10,
                },
            },
            {
                "scenario": "pod_events",
                "mode": "cold",
                "status": "no_data",
                "metrics": {
                    "time_to_first_useful_row_ms": 1.0,
                    "api_request_count": 1,
                },
            },
            {
                "scenario": "pod_events",
                "mode": "cold",
                "status": "failed",
                "metrics": {
                    "time_to_first_useful_row_ms": 2.0,
                    "api_request_count": 2,
                },
            },
        ]

        aggregate = bench_run.aggregate_runs(results)["pod_events:cold"]
        self.assertEqual({"failed": 1, "no_data": 1, "ok": 1}, aggregate["status_counts"])
        self.assertEqual(1, aggregate["ok_run_count"])
        self.assertEqual(3, aggregate["total_run_count"])
        self.assertEqual(100.0, aggregate["time_to_first_useful_row_ms"]["mean"])
        self.assertEqual(10.0, aggregate["api_request_count"]["mean"])


if __name__ == "__main__":
    unittest.main()
